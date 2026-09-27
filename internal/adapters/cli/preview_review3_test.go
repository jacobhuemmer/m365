package cli

import (
	"strings"
	"testing"
)

// Third Codex review of PR #20: a body of invisible format characters
// must not look blank in the preview.
func TestDrawPreviewShowsFormatCharactersAsText(t *testing.T) {
	body := "ZWJ" + string(rune(0x200d)) + string(rune(0x200d)) + " ZWSP" + string(rune(0x200b)) + " BOM" + string(rune(0xfeff))
	got := drawPreview([]string{"Chat: chat-1"}, body, nil, 40)
	for _, r := range []rune{0x200d, 0x200b, 0xfeff} {
		if strings.ContainsRune(got, r) {
			t.Fatalf("raw %U reached the preview:\n%s", r, got)
		}
	}
	want := `ZWJ\u200d\u200d ZWSP\u200b BOM\ufeff`
	if !strings.Contains(got, want) {
		t.Fatalf("format characters not shown as escapes:\n%s", got)
	}
}
