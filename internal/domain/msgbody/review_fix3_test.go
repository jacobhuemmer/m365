package msgbody

import (
	"reflect"
	"testing"
)

// Third Codex review of PR #20: problems follow where they occur.
func TestLintOrderWithinBufferedText(t *testing.T) {
	cases := []struct {
		in   string
		want Problems
	}{
		{"<p>ok <span>x</span> **bad**</p>", Problems{{Rule: RuleTagNotAllowed, Detail: "<span>"}, {Rule: RuleLeftoverMarkdown, Detail: "**bad**"}}},
		{"<p><br><br><br><span>x</span></p>", Problems{{Rule: RuleExtraBlankLines, Detail: "3 <br> in a row"}, {Rule: RuleTagNotAllowed, Detail: "<span>"}}},
		{"<p>**a** <span>x</span> **b**</p>", Problems{{Rule: RuleLeftoverMarkdown, Detail: "**a**"}, {Rule: RuleTagNotAllowed, Detail: "<span>"}, {Rule: RuleLeftoverMarkdown, Detail: "**b**"}}},
	}
	for _, tc := range cases {
		if got := Lint(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Lint(%q)\ngot  %#v\nwant %#v", tc.in, got, tc.want)
		}
	}
}
