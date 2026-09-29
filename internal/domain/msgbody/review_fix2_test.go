package msgbody

import (
	"reflect"
	"testing"
)

// Findings from the second Codex review of PR #20.

func TestLintHeadingAfterNewlineToken(t *testing.T) {
	got := Lint("<p><code>x</code>\n<strong># Heading</strong></p>")
	want := Problems{{Rule: RuleLeftoverMarkdown, Detail: "# Heading"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestLintProblemsInDocumentOrder(t *testing.T) {
	got := Lint("<p>**bad**<span>x</span></p>")
	want := Problems{{Rule: RuleLeftoverMarkdown, Detail: "**bad**"}, {Rule: RuleTagNotAllowed, Detail: "<span>"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestStringWidthFormatCharactersAreZero(t *testing.T) {
	for s, want := range map[string]int{"a\u200bb": 2, "a\u2060b": 2, "\ufeffx": 1} {
		if got := StringWidth(s); got != want {
			t.Errorf("StringWidth(%q) = %d, want %d", s, got, want)
		}
	}
}
