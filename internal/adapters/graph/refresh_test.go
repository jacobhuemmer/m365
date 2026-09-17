package graph

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/app/mail"
	"golang.org/x/oauth2"
)

type memStore struct{ b *auth.Blob }

func (m *memStore) Get() (auth.Blob, bool, error) {
	if m.b == nil {
		return auth.Blob{}, false, nil
	}
	return *m.b, true, nil
}
func (m *memStore) Put(b auth.Blob) error { m.b = &b; return nil }
func (m *memStore) Delete() error         { m.b = nil; return nil }

func TestRefreshSendsRefreshTokenWhenExpired(t *testing.T) {
	var grant, refresh string
	tokSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		grant = r.Form.Get("grant_type")
		refresh = r.Form.Get("refresh_token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-at","token_type":"Bearer","expires_in":3600,"refresh_token":"new-rt"}`))
	}))
	defer tokSrv.Close()
	st := &memStore{b: &auth.Blob{
		AccessToken:  "old-at",
		RefreshToken: "old-rt",
		ExpiresAt:    time.Now().Add(-time.Minute),
		Usable:       true,
	}}
	r := &Refresher{Store: st, Config: &oauth2.Config{
		ClientID: "id",
		Endpoint: oauth2.Endpoint{TokenURL: tokSrv.URL, AuthStyle: oauth2.AuthStyleInParams},
	}}
	tok, err := r.Token(false)
	if err != nil || tok != "new-at" {
		t.Fatalf("%q %v", tok, err)
	}
	if grant != "refresh_token" || refresh != "old-rt" {
		t.Fatalf("grant=%q refresh=%q", grant, refresh)
	}
	got, _, _ := st.Get()
	if got.AccessToken != "new-at" || got.RefreshToken != "new-rt" || got.ExpiresAt.IsZero() {
		t.Fatalf("%+v", got)
	}
}

func TestRefreshSkipsWhenAccessTokenValid(t *testing.T) {
	var n atomic.Int32
	tokSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tokSrv.Close()
	st := &memStore{b: &auth.Blob{
		AccessToken:  "live-at",
		RefreshToken: "old-rt",
		ExpiresAt:    time.Now().Add(time.Hour),
		Usable:       true,
	}}
	r := &Refresher{Store: st, Config: &oauth2.Config{
		ClientID: "id",
		Endpoint: oauth2.Endpoint{TokenURL: tokSrv.URL, AuthStyle: oauth2.AuthStyleInParams},
	}}
	tok, err := r.Token(false)
	if err != nil || tok != "live-at" || n.Load() != 0 {
		t.Fatalf("%q n=%d %v", tok, n.Load(), err)
	}
}

func TestHTTPRetriesUnauthorizedWithRefreshToken(t *testing.T) {
	var sawNew bool
	graphSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer new-at" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		sawNew = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"value":[]}`)
	}))
	defer graphSrv.Close()
	var grant string
	tokSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		grant = r.Form.Get("grant_type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-at","token_type":"Bearer","expires_in":3600,"refresh_token":"old-rt"}`))
	}))
	defer tokSrv.Close()
	st := &memStore{b: &auth.Blob{
		AccessToken:  "old-at",
		RefreshToken: "old-rt",
		Usable:       true,
	}}
	ref := &Refresher{Store: st, Config: &oauth2.Config{
		ClientID: "id",
		Endpoint: oauth2.Endpoint{TokenURL: tokSrv.URL, AuthStyle: oauth2.AuthStyleInParams},
	}}
	c := &HTTPClient{Base: graphSrv.URL, Refresh: ref.Token}
	p, err := c.List(context.Background(), mail.ListQuery{Top: 10})
	if err != nil || !sawNew || grant != "refresh_token" || p.Count != 0 {
		t.Fatalf("err=%v sawNew=%v grant=%q page=%+v", err, sawNew, grant, p)
	}
}
