package msgbody

import (
	"reflect"
	"testing"
)

// Fourth Codex review of PR #20: subject problems follow document order.
func TestLintSubjectDocumentOrder(t *testing.T) {
	cases := []struct {
		in   string
		want Problems
	}{
		{`\n **bad**`, Problems{
			{Rule: RuleLiteralEscape, Detail: `subject: \n in "\n **bad**"`},
			{Rule: RuleLeftoverMarkdown, Detail: "subject: **bad**"},
		}},
		{"**a**\n**b**", Problems{
			{Rule: RuleLeftoverMarkdown, Detail: "subject: **a**"},
			{Rule: RuleNewlineInSubject, Detail: "subject"},
			{Rule: RuleLeftoverMarkdown, Detail: "subject: **b**"},
		}},
	}
	for _, tc := range cases {
		if got := LintSubject(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("LintSubject(%q)\ngot  %#v\nwant %#v", tc.in, got, tc.want)
		}
	}
}
