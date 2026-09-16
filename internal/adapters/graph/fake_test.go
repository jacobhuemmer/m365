package graph

import (
	"io"
	"net/http"
	"testing"
)

func TestFakeConsentAndStatus(t *testing.T) {
	mem := Seed()
	srv := NewFakeServer(mem)
	defer srv.Close()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/chats", nil)
	req.Header.Set("Authorization", "Bearer fake-mail")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("status %d", res.StatusCode)
	}
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/me/messages", nil)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("want 401 got %d %s", res.StatusCode, b)
	}
}
