package cli

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestBlockedSendExitsUsage(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
	}{
		{"mail send", []string{"mail", "send", "--to", "user@example.com", "--subject", "t", "--html", "--body", "<p>hi"}},
		{"mail reply", []string{"mail", "reply", "msg-1", "--html", "--body", "<p>hi"}},
		{"teams send", []string{"teams", "send", "chat-1", "--html", "--text", "<p>hi"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			d, out, errw := testDeps()
			loginAll(t, &d)
			out.Reset()
			if code := Run(append([]string{"m365"}, c.args...), d); code != 3 {
				t.Fatalf("exit %d want 3; stderr %s", code, errw.String())
			}
			if out.Len() != 0 {
				t.Fatalf("stdout %q", out.String())
			}
			var got map[string]any
			if err := json.Unmarshal(errw.Bytes(), &got); err != nil {
				t.Fatalf("stderr %q: %v", errw.String(), err)
			}
			want := map[string]any{
				"class":   "usage",
				"message": "1 format problem: broken-html: unclosed <p>",
				"hint":    "run with --preview to see them",
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("stderr %#v want %#v", got, want)
			}
			if n := len(d.Mail.(graph.MailAPI).Memory.Sent); n != 0 {
				t.Fatalf("sent %d", n)
			}
		})
	}
}
