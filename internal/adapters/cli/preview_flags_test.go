package cli

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestPreviewWithJSONIsUsage(t *testing.T) {
	for _, args := range [][]string{
		{"mail", "send", "--to", "a@example.com", "--subject", "Hi", "--body", "hi", "--preview", "--json"},
		{"--json", "mail", "reply", "msg-1", "--body", "hi", "--preview"},
		{"teams", "send", "chat-1", "--text", "hi", "--json", "--preview"},
	} {
		d, out, errw := testDeps()
		loginAll(t, &d)
		out.Reset()
		if code := Run(append([]string{"m365"}, args...), d); code != 3 {
			t.Fatalf("%v: exit %d want 3", args, code)
		}
		var got map[string]any
		if err := json.Unmarshal(errw.Bytes(), &got); err != nil {
			t.Fatal(err, errw.String())
		}
		want := map[string]any{"class": "usage", "message": "use only one of --preview or --json"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%v: stderr %#v", args, got)
		}
		if n := len(d.Mail.(graph.MailAPI).Memory.Sent); n != 0 {
			t.Fatalf("sent %d", n)
		}
	}
}

func TestPreviewIgnoresHumanAndDryRun(t *testing.T) {
	base := []string{"m365", "teams", "send", "chat-1", "--text", "hi", "--preview"}
	var outputs []string
	for _, extra := range [][]string{nil, {"--human"}, {"--dry-run"}} {
		d, out, errw := testDeps()
		loginAll(t, &d)
		out.Reset()
		if code := Run(append(append([]string{}, base...), extra...), d); code != 0 {
			t.Fatalf("%v: exit %d %s", extra, code, errw.String())
		}
		outputs = append(outputs, out.String())
	}
	if outputs[0] != outputs[1] || outputs[0] != outputs[2] || !strings.HasPrefix(outputs[0], "┌") {
		t.Fatalf("outputs differ or are not a box:\n%s\n%s\n%s", outputs[0], outputs[1], outputs[2])
	}
}

func TestSendHelpListsPreview(t *testing.T) {
	for _, args := range [][]string{{"mail", "send"}, {"mail", "reply"}, {"teams", "send"}} {
		d, out, _ := testDeps()
		if code := Run(append(append([]string{"m365"}, args...), "--help"), d); code != 0 {
			t.Fatalf("%v help exit %d", args, code)
		}
		if !strings.Contains(out.String(), "--preview") {
			t.Fatalf("%v help lacks --preview:\n%s", args, out.String())
		}
	}
}
