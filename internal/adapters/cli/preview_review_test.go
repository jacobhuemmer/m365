package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

// Findings from the Codex review of PR #20.

func TestDrawPreviewShowsControlCharactersAsText(t *testing.T) {
	got := drawPreview([]string{"Subject: a\x1bb"}, "x\x1b[31my\tz", msgbody.Problems{{Rule: "r", Detail: "d\x07"}}, 40)
	if strings.ContainsAny(got, "\x1b\x07\t") {
		t.Fatalf("raw control characters reached the terminal: %q", got)
	}
	for _, want := range []string{`Subject: a\x1bb`, `x\x1b[31my z`, `- r: d\x07`} {
		if !strings.Contains(got, want) {
			t.Fatalf("output lacks %q:\n%s", want, got)
		}
	}
}

func TestDrawPreviewWideCharactersKeepBoxWidth(t *testing.T) {
	got := drawPreview([]string{"Chat: chat-1"}, "界界界\n"+strings.Repeat("界", 12), nil, 20)
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if w := msgbody.StringWidth(line); w != 20 {
			t.Fatalf("line %q is %d columns, want 20\n%s", line, w, got)
		}
	}
}

func TestPreviewWithJSONValueFormsIsUsage(t *testing.T) {
	for _, flagForm := range []string{"--preview=true", "-preview", "-preview=1"} {
		d, out, errw := testDeps()
		loginAll(t, &d)
		out.Reset()
		if code := Run([]string{"m365", "teams", "send", "chat-1", "--text", "hi", "--json", flagForm}, d); code != 3 {
			t.Fatalf("%s: exit %d want 3; stdout %q", flagForm, code, out.String())
		}
		var got map[string]any
		if err := json.Unmarshal(errw.Bytes(), &got); err != nil || got["message"] != "use only one of --preview or --json" {
			t.Fatalf("%s: stderr %q", flagForm, errw.String())
		}
	}
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()
	if code := Run([]string{"m365", "teams", "send", "chat-1", "--text", "hi", "--json", "--preview=false", "--dry-run"}, d); code != 0 || !json.Valid(out.Bytes()) {
		t.Fatalf("--preview=false with --json: exit %d, stdout %q, stderr %q", code, out.String(), errw.String())
	}
}
