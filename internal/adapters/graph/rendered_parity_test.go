package graph

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

type bodyMode struct {
	name     string
	body     string
	html, md bool
}

var parityModes = []bodyMode{
	{name: "plain", body: "Hi team,\n\nThe change is in UAT.\nRollback is ready."},
	{name: "markdown", body: "# Status\n\n**Done**, see `make verify`.\n\n- one\n- two\n\n[ticket](https://example.com/t?a=1&b=2)", md: true},
	{name: "html", body: "<p>Hello</p><ul><li>one</li></ul>", html: true},
}

func paritySess() domain.Session {
	return domain.Session{SignedIn: true, SessionUsable: true, MailConsented: true, TeamsConsented: true}
}

func dryContent(t *testing.T, out any) string {
	t.Helper()
	r, ok := out.(map[string]any)["rendered"].(msgbody.Rendered)
	if !ok || r.ContentType != "html" {
		t.Fatalf("rendered %#v", out.(map[string]any)["rendered"])
	}
	return r.Content
}

func TestMailSendDryRunMatchesPayload(t *testing.T) {
	for _, m := range parityModes {
		t.Run(m.name, func(t *testing.T) {
			in := mail.SendInput{To: []string{"a@example.com"}, Subject: "t", Body: m.body, HTML: m.html, MD: m.md}
			srv, cap := captureJSONServer(t)
			c := &HTTPClient{Base: srv.URL, Token: "t"}
			dry := in
			dry.DryRun = true
			out, err := mail.Send(context.Background(), c, paritySess(), dry)
			if err != nil {
				t.Fatal(err)
			}
			want := dryContent(t, out)
			if cap.Method != "" {
				t.Fatal("dry-run reached Graph")
			}
			if _, err := mail.Send(context.Background(), c, paritySess(), in); err != nil {
				t.Fatal(err)
			}
			msg, _ := cap.Body["message"].(map[string]any)
			body, _ := msg["body"].(map[string]any)
			if body["contentType"] != "HTML" || body["content"] != want {
				t.Fatalf("sent %#v\nwant HTML %q", body, want)
			}
		})
	}
}

func TestMailReplyDryRunMatchesComment(t *testing.T) {
	for _, m := range parityModes {
		t.Run(m.name, func(t *testing.T) {
			in := mail.ReplyInput{ID: "msg-1", Body: m.body, HTML: m.html, MD: m.md}
			srv, cap := captureJSONServer(t)
			c := &HTTPClient{Base: srv.URL, Token: "t"}
			dry := in
			dry.DryRun = true
			out, err := mail.Reply(context.Background(), c, paritySess(), dry)
			if err != nil {
				t.Fatal(err)
			}
			want := dryContent(t, out)
			if _, err := mail.Reply(context.Background(), c, paritySess(), in); err != nil {
				t.Fatal(err)
			}
			if len(cap.Body) != 1 || cap.Body["comment"] != want {
				t.Fatalf("sent %#v\nwant only comment %q", cap.Body, want)
			}
		})
	}
}

func TestTeamsSendDryRunMatchesPayload(t *testing.T) {
	for _, m := range parityModes {
		t.Run(m.name, func(t *testing.T) {
			in := teams.SendInput{ChatID: "chat-1", Text: m.body, HTML: m.html, MD: m.md}
			srv, cap := captureJSONServer(t)
			c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "t"}}
			dry := in
			dry.DryRun = true
			out, err := teams.Send(context.Background(), c, paritySess(), dry)
			if err != nil {
				t.Fatal(err)
			}
			want := dryContent(t, out)
			if _, err := teams.Send(context.Background(), c, paritySess(), in); err != nil {
				t.Fatal(err)
			}
			body, _ := cap.Body["body"].(map[string]any)
			if body["contentType"] != "html" || body["content"] != want {
				t.Fatalf("sent %#v\nwant html %q", body, want)
			}
		})
	}
}
