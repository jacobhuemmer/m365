package graph

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
)

type shareServer struct {
	calls      []string
	upload     []byte
	invite     map[string]any
	message    map[string]any
	inviteCode int
}

func newShareServer(t *testing.T, s *shareServer) *httptest.Server {
	t.Helper()
	if s.inviteCode == 0 {
		s.inviteCode = http.StatusOK
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.calls = append(s.calls, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, ":/content"):
			s.upload = body
			_, _ = w.Write([]byte(`{"id":"ITEM1","name":"a b.pdf","webUrl":"https://contoso-my.sharepoint.com/personal/me/Documents/a%20b.pdf","eTag":"\"{74D20C7F-34AA-4A7F-B74E-2B30004247C5},1\""}`))
		case strings.HasSuffix(r.URL.Path, "/invite"):
			_ = json.Unmarshal(body, &s.invite)
			w.WriteHeader(s.inviteCode)
			_, _ = w.Write([]byte(`{"value":[]}`))
		case strings.HasSuffix(r.URL.Path, "/messages"):
			_ = json.Unmarshal(body, &s.message)
			_, _ = w.Write([]byte(`{"id":"MSG1"}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
	}))
}

func shareInput(t *testing.T) teams.SendInput {
	t.Helper()
	path := filepath.Join(t.TempDir(), "a b.pdf")
	if err := os.WriteFile(path, []byte("%PDF-fake"), 0o600); err != nil {
		t.Fatal(err)
	}
	return teams.SendInput{
		ChatID: "19:chat@thread.v2", ShareWith: []string{"ajay@example.com", "bo@example.com"},
		Files: []domain.OutboundFile{{Path: path, Name: "a b.pdf", Size: 9}},
	}
}

func TestHTTPTeamsSendUploadsSharesAndAttaches(t *testing.T) {
	s := &shareServer{}
	srv := newShareServer(t, s)
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both", Client: srv.Client()}}
	in := shareInput(t)
	in.Rendered.Content = "<p>Report</p>"

	if _, err := c.Send(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"PUT /me/drive/root:/Microsoft Teams Chat Files/a b.pdf:/content?@microsoft.graph.conflictBehavior=rename",
		"POST /me/drive/items/ITEM1/invite?",
		"POST /me/chats/19:chat@thread.v2/messages?",
	}
	if strings.Join(s.calls, "\n") != strings.Join(want, "\n") {
		t.Fatalf("calls:\n%s\nwant:\n%s", strings.Join(s.calls, "\n"), strings.Join(want, "\n"))
	}
	if string(s.upload) != "%PDF-fake" {
		t.Fatalf("upload body %q", s.upload)
	}
	if s.invite["sendInvitation"] != false || s.invite["requireSignIn"] != true {
		t.Fatalf("invite must not email anyone and must require sign-in: %v", s.invite)
	}
	if roles, _ := s.invite["roles"].([]any); len(roles) != 1 || roles[0] != "read" {
		t.Fatalf("roles %v", s.invite["roles"])
	}
	rec, _ := s.invite["recipients"].([]any)
	if len(rec) != 2 || rec[0].(map[string]any)["email"] != "ajay@example.com" || rec[1].(map[string]any)["email"] != "bo@example.com" {
		t.Fatalf("recipients %v", rec)
	}
	const id = "74d20c7f34aa4a7fb74e2b30004247c5"
	body, _ := s.message["body"].(map[string]any)
	if body["content"] != `<p>Report</p><attachment id="`+id+`"></attachment>` {
		t.Fatalf("body %v", body)
	}
	atts, _ := s.message["attachments"].([]any)
	a, _ := atts[0].(map[string]any)
	if len(atts) != 1 || a["id"] != id || a["contentType"] != "reference" || a["name"] != "a b.pdf" ||
		a["contentUrl"] != "https://contoso-my.sharepoint.com/personal/me/Documents/a%20b.pdf" {
		t.Fatalf("attachments %v", atts)
	}
}

func TestHTTPTeamsSendSharingFailureNeverPosts(t *testing.T) {
	s := &shareServer{inviteCode: http.StatusForbidden}
	srv := newShareServer(t, s)
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both", Client: srv.Client()}}

	id, err := c.Send(context.Background(), shareInput(t))
	if err == nil || id != "" {
		t.Fatalf("sharing failed but Send returned id=%q err=%v", id, err)
	}
	if s.message != nil {
		t.Fatal("message was posted although the file was not shared")
	}
	if !strings.Contains(err.Error(), "uploaded") {
		t.Fatalf("error should say the file was uploaded: %v", err)
	}
}

func TestHTTPTeamsSendWithoutRecipientsSkipsInvite(t *testing.T) {
	s := &shareServer{}
	srv := newShareServer(t, s)
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both", Client: srv.Client()}}
	in := shareInput(t)
	in.ShareWith = nil
	if _, err := c.Send(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	for _, call := range s.calls {
		if strings.Contains(call, "/invite") {
			t.Fatalf("invite called with nobody to share with: %v", s.calls)
		}
	}
}
