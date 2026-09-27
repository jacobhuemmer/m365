package cli

import (
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

// Bidi controls can reorder how a line reads; show them as escapes.
func TestDrawPreviewShowsBidiControlsAsText(t *testing.T) {
	got := drawPreview([]string{"Chat: chat-1"}, "abc\u202etxt \u2066x\u2069", nil, 40)
	for _, r := range []string{"\u202e", "\u2066", "\u2069"} {
		if strings.Contains(got, r) {
			t.Fatalf("raw bidi control %U reached the terminal:\n%s", []rune(r)[0], got)
		}
	}
	if !strings.Contains(got, `abc\u202etxt \u2066x\u2069`) {
		t.Fatalf("bidi controls not shown as escapes:\n%s", got)
	}
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if w := msgbody.StringWidth(line); w != 40 {
			t.Fatalf("line %q is %d columns, want 40", line, w)
		}
	}
}
