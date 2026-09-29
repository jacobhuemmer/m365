package cli

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

var sendCommands = []struct {
	name string
	args []string
}{
	{"mail send", []string{"mail", "send", "--to", "user@example.com", "--subject", "t", "--body", "b"}},
	{"mail reply", []string{"mail", "reply", "msg-1", "--body", "b"}},
	{"teams send", []string{"teams", "send", "chat-1", "--text", "b"}},
}

func TestDryRunShowsRendered(t *testing.T) {
	for _, c := range sendCommands {
		t.Run(c.name, func(t *testing.T) {
			d, out, errw := testDeps()
			loginAll(t, &d)
			out.Reset()
			if code := Run(append(append([]string{"m365"}, c.args...), "--dry-run"), d); code != 0 {
				t.Fatalf("exit %d: %s", code, errw.String())
			}
			var m map[string]any
			if err := json.Unmarshal(out.Bytes(), &m); err != nil {
				t.Fatal(err, out.String())
			}
			want := map[string]any{"content_type": "html", "content": "<p>b</p>"}
			if !reflect.DeepEqual(m["rendered"], want) {
				t.Fatalf("rendered %#v want %#v", m["rendered"], want)
			}
			if !reflect.DeepEqual(m["format_problems"], []any{}) {
				t.Fatalf("format_problems %#v want []", m["format_problems"])
			}
			if n := len(d.Mail.(graph.MailAPI).Memory.Sent); n != 0 {
				t.Fatalf("dry-run sent %d", n)
			}
		})
	}
}

func TestSendDeliversRendered(t *testing.T) {
	for _, c := range sendCommands {
		t.Run(c.name, func(t *testing.T) {
			d, _, errw := testDeps()
			loginAll(t, &d)
			if code := Run(append([]string{"m365"}, c.args...), d); code != 0 {
				t.Fatalf("exit %d: %s", code, errw.String())
			}
			sent := d.Mail.(graph.MailAPI).Memory.Sent
			if len(sent) != 1 {
				t.Fatalf("sent %d", len(sent))
			}
			var got msgbody.Rendered
			switch in := sent[0].(type) {
			case mail.SendInput:
				got = in.Rendered
			case mail.ReplyInput:
				got = in.Rendered
			case teams.SendInput:
				got = in.Rendered
			default:
				t.Fatalf("unexpected %T", in)
			}
			if want := (msgbody.Rendered{ContentType: "html", Content: "<p>b</p>"}); got != want {
				t.Fatalf("delivered %#v want %#v", got, want)
			}
		})
	}
}
