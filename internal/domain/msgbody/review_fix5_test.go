package msgbody

import (
	"reflect"
	"testing"
)

// Fifth Codex review of PR #20: only visible text makes a paragraph
// non-empty; an empty inline element does not.
func TestLintParagraphWithOnlyEmptyInlineIsEmpty(t *testing.T) {
	empty := Problems{{Rule: RuleExtraBlankLines, Detail: "empty <p>"}}
	cases := map[string]Problems{
		"<p><strong></strong></p>":                  empty,
		`<p><a href="https://example.com"></a></p>`: empty,
		"<p><em> </em></p>":                         empty,
		"<p><strong>x</strong></p>":                 nil,
		"<p><code>x</code></p>":                     nil,
	}
	for in, want := range cases {
		got := Lint(in)
		if len(got) == 0 && len(want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Lint(%q)\ngot  %#v\nwant %#v", in, got, want)
		}
	}
}
