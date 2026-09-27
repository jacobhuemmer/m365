package msgbody

import (
	"reflect"
	"testing"
)

// Findings from the Codex review of PR #20.

func TestTextLinkLabelAcrossBreak(t *testing.T) {
	in := `<p>before <a href="https://example.com">link<br>after</a></p>`
	if got, want := Text(in, 80), "before link\nafter (https://example.com)"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTextWrapsByDisplayWidth(t *testing.T) {
	if got, want := Text("<p>界界界界界</p>", 4), "界界\n界界\n界"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got, want := Text("<p>ab 界界 cd</p>", 6), "ab\n界界\ncd"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLintMarkdownSplitByInlineTags(t *testing.T) {
	got := Lint(`<p>**bo<strong>ld</strong>**</p>`)
	want := Problems{{Rule: RuleLeftoverMarkdown, Detail: "**bold**"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestLintStrayDisallowedEndTag(t *testing.T) {
	got := Lint(`<p>x</p></table>`)
	want := Problems{{Rule: RuleTagNotAllowed, Detail: "</table>"}, {Rule: RuleBrokenHTML, Detail: "stray </table>"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestStringWidth(t *testing.T) {
	for s, want := range map[string]int{"abc": 3, "界界": 4, "": 0, "é": 1} {
		if got := StringWidth(s); got != want {
			t.Errorf("StringWidth(%q) = %d, want %d", s, got, want)
		}
	}
}
