package cli

import (
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestDrawPreviewContractExample(t *testing.T) {
	want := "┌──────────────────────────────────────┐\n" +
		"│ Chat: chat-1                         │\n" +
		"├──────────────────────────────────────┤\n" +
		"│ Hi team                              │\n" +
		"│                                      │\n" +
		"│ • one                                │\n" +
		"│ • two                                │\n" +
		"│                                      │\n" +
		"│ See docs (https://example.com)       │\n" +
		"└──────────────────────────────────────┘\n"
	body := "Hi team\n\n• one\n• two\n\nSee docs (https://example.com)"
	if got := drawPreview([]string{"Chat: chat-1"}, body, nil, 40); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestDrawPreviewMinimumWidth(t *testing.T) {
	got := drawPreview([]string{"Chat: chat-1"}, "a long body line", nil, 10)
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if n := utf8.RuneCountInString(line); n != 20 {
			t.Fatalf("line %q is %d wide, want 20\n%s", line, n, got)
		}
	}
}

func TestTermWidthNotATerminal(t *testing.T) {
	var b strings.Builder
	if got := termWidth(&b); got != 80 {
		t.Fatalf("buffer width %d want 80", got)
	}
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if got := termWidth(f); got != 80 {
		t.Fatalf("regular file width %d want 80", got)
	}
}

func TestPreviewGolden(t *testing.T) {
	cases := []struct {
		golden string
		args   []string
	}{
		{"mail", []string{"mail", "send", "--to", "a@example.com", "--to", "b@example.com", "--cc", "c@example.com",
			"--subject", "Hi", "--body-file", "testdata/preview/body.txt", "--attach", "testdata/preview/note.txt", "--preview"}},
		{"mail-problems", []string{"mail", "send", "--to", "a@example.com", "--subject", "Hi", "--html", "--body", "<p>hi", "--preview"}},
		{"reply", []string{"mail", "reply", "msg-1", "--body-file", "testdata/preview/body.txt", "--preview"}},
		{"reply-all", []string{"mail", "reply", "msg-1", "--all", "--body", "hi", "--preview"}},
		{"teams", []string{"teams", "send", "chat-1", "--format", "md", "--text-file", "testdata/preview/body.md", "--preview"}},
		{"teams-to", []string{"teams", "send", "--to", "Alice", "--text", "hi", "--preview"}},
		{"teams-problems", []string{"teams", "send", "chat-1", "--html", "--text", "<p>hi", "--preview"}},
	}
	for _, tc := range cases {
		t.Run(tc.golden, func(t *testing.T) {
			want, err := os.ReadFile("testdata/preview/" + tc.golden + ".golden")
			if err != nil {
				t.Fatal(err)
			}
			d, out, errw := testDeps()
			loginAll(t, &d)
			out.Reset()
			if code := Run(append([]string{"m365"}, tc.args...), d); code != 0 {
				t.Fatalf("exit %d: %s", code, errw.String())
			}
			if out.String() != string(want) {
				t.Fatalf("got\n%s\nwant\n%s", out.String(), want)
			}
			if n := len(d.Mail.(graph.MailAPI).Memory.Sent); n != 0 {
				t.Fatalf("preview sent %d", n)
			}
		})
	}
}
