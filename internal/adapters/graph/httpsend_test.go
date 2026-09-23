package graph

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/app/teams"
)

func TestHTTPSendUsesFakeServer(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPClient{Base: srv.URL, Token: "fake-both"}
	id, err := c.Send(context.Background(), mail.SendInput{
		To: []string{"a@example.com"}, Subject: "t", Body: "b",
	})
	if err != nil || id != "sent" {
		t.Fatalf("%s %v", id, err)
	}
	tc := &HTTPTeams{HTTPClient: c}
	id, err = tc.Send(context.Background(), teams.SendInput{ChatID: "chat-1", Text: "hi"})
	if err != nil || id != "sent" {
		t.Fatalf("teams %s %v", id, err)
	}
}

func TestTokenFn(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPClient{Base: srv.URL, TokenFn: func() string { return "fake-both" }}
	p, err := c.List(context.Background(), mail.ListQuery{Top: 10})
	if err != nil || p.Count == 0 {
		t.Fatal(err)
	}
}

type capturedPOST struct {
	Method string
	Path   string
	Body   map[string]any
}

func captureJSONServer(t *testing.T) (*httptest.Server, *capturedPOST) {
	t.Helper()
	cap := &capturedPOST{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.Method = r.Method
		cap.Path = r.URL.Path
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if len(b) > 0 {
			if err := json.Unmarshal(b, &cap.Body); err != nil {
				t.Errorf("json: %v body=%s", err, b)
			}
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)
	return srv, cap
}

func TestHTTPTeamsSendTextHTMLMDPayloads(t *testing.T) {
	md := "Hello.\n\n- one\n- two\n\nSee [docs](https://example.com).\n\n```\ncode\n```\n"
	cases := []struct {
		name        string
		in          teams.SendInput
		wantType    string
		wantContent string
		wantHTML    bool
	}{
		{
			name:        "text keeps line breaks as html",
			in:          teams.SendInput{ChatID: "chat-1", Text: "SDO-559: approve?\n\nRollback is removing it.\nThanks"},
			wantType:    "html",
			wantContent: "<p>SDO-559: approve?</p>\n<p>Rollback is removing it.<br>Thanks</p>",
		},
		{
			name:        "html",
			in:          teams.SendInput{ChatID: "chat-1", Text: "<p>hi</p>", HTML: true},
			wantType:    "html",
			wantContent: "<p>hi</p>",
		},
		{
			name:     "md",
			in:       teams.SendInput{ChatID: "chat-1", Text: md, MD: true},
			wantType: "html",
			wantHTML: true,
		},
		{
			name:        "html wins over md",
			in:          teams.SendInput{ChatID: "chat-1", Text: "<p>already</p>", HTML: true, MD: true},
			wantType:    "html",
			wantContent: "<p>already</p>",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, cap := captureJSONServer(t)
			c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "t"}}
			if _, err := c.Send(context.Background(), tc.in); err != nil {
				t.Fatal(err)
			}
			if cap.Method != http.MethodPost {
				t.Fatalf("method %s", cap.Method)
			}
			if !strings.HasSuffix(cap.Path, "/me/chats/chat-1/messages") {
				t.Fatalf("path %s", cap.Path)
			}
			body, _ := cap.Body["body"].(map[string]any)
			if body["contentType"] != tc.wantType {
				t.Fatalf("contentType %v want %s", body["contentType"], tc.wantType)
			}
			content, _ := body["content"].(string)
			if tc.wantHTML {
				for _, want := range []string{"<p>Hello.</p>", "<ul>", "<a href=\"https://example.com\">docs</a>", "<pre><code>code</code></pre>"} {
					if !strings.Contains(content, want) {
						t.Fatalf("md content missing %q: %q", want, content)
					}
				}
				return
			}
			if content != tc.wantContent {
				t.Fatalf("content %q want %q", content, tc.wantContent)
			}
		})
	}
}

func TestHTTPMailReplyKeepsQuotedThread(t *testing.T) {
	cases := []struct {
		name string
		in   mail.ReplyInput
		want string
	}{
		{
			name: "plain text becomes html comment",
			in:   mail.ReplyInput{ID: "msg-1", Body: "Hi all,\n\nNot DNS.\n- Iulia: retry.\n\nThanks,\nMason"},
			want: "<p>Hi all,</p>\n<p>Not DNS.<br>- Iulia: retry.</p>\n<p>Thanks,<br>Mason</p>",
		},
		{
			name: "html goes in comment unchanged",
			in:   mail.ReplyInput{ID: "msg-1", Body: "<p>hello</p>", HTML: true},
			want: "<p>hello</p>",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, cap := captureJSONServer(t)
			c := &HTTPClient{Base: srv.URL, Token: "t"}
			if _, err := c.Reply(context.Background(), tc.in); err != nil {
				t.Fatal(err)
			}
			if cap.Method != http.MethodPost {
				t.Fatalf("method %s", cap.Method)
			}
			if !strings.HasSuffix(cap.Path, "/me/messages/msg-1/reply") {
				t.Fatalf("path %s", cap.Path)
			}
			if cap.Body["comment"] != tc.want {
				t.Fatalf("comment %q want %q", cap.Body["comment"], tc.want)
			}
			// message.body replaces the reply body and drops the quoted thread.
			if _, ok := cap.Body["message"]; ok {
				t.Fatalf("reply must not send message.body: %+v", cap.Body)
			}
		})
	}
}
