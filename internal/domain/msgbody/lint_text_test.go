package msgbody

import (
	"reflect"
	"testing"
)

func TestLintLeftoverMarkdown(t *testing.T) {
	runLint(t, []lintCase{
		{"bold", `<p>a **bold** b</p>`, Problems{{RuleLeftoverMarkdown, "**bold**"}}},
		{"heading at line start", `<p># Title</p>`, Problems{{RuleLeftoverMarkdown, "# Title"}}},
		{"level-3 heading after a break", `<p>a<br>### Sub</p>`, Problems{{RuleLeftoverMarkdown, "### Sub"}}},
		{"hash mid-line passes", `<p>C# and <strong>b</strong> # 3</p>`, nil},
		{"hash without space passes", `<p>#1 priority</p>`, nil},
		{"link", `<p>see [docs](https://example.com)</p>`, Problems{{RuleLeftoverMarkdown, "[docs](https://example.com)"}}},
		{"code span", "<p>run `make`</p>", Problems{{RuleLeftoverMarkdown, "`make`"}}},
		{"dash list line passes", `<p>- one<br>- two</p>`, nil},
		{"inside code passes", "<p><code>**x** [l](u) `y`</code></p>", nil},
		{"inside pre passes", "<pre><code># T\n**x**</code></pre>", nil},
		{"entity-escaped text is checked", `<p>&#42;&#42;x&#42;&#42;</p>`, Problems{{RuleLeftoverMarkdown, "**x**"}}},
		{"long excerpt cut to 40", `<p># ` + "0123456789012345678901234567890123456789XYZ" + `</p>`,
			Problems{{RuleLeftoverMarkdown, "# 01234567890123456789012345678901234567"}}},
	})
}

func TestLintLiteralEscape(t *testing.T) {
	runLint(t, []lintCase{
		{"backslash n", `<p>line one\nline two</p>`, Problems{{RuleLiteralEscape, `\n in "line one\nline two"`}}},
		{"backslash t", `<p>a\tb</p>`, Problems{{RuleLiteralEscape, `\t in "a\tb"`}}},
		{"backslash quote", `<p>say \"hi\"</p>`, Problems{{RuleLiteralEscape, `\" in "say \"hi\""`}}},
		{"inside code passes", `<p><code>a\nb</code></p>`, nil},
		{"real newline passes", "<p>a\nb</p>", nil},
	})
}

func TestLintSubject(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Problems
	}{
		{"clean", "Status: UAT is done", nil},
		{"bold", "**urgent**", Problems{{RuleLeftoverMarkdown, "subject: **urgent**"}}},
		{"escape", `a\nb`, Problems{{RuleLiteralEscape, `subject: \n in "a\nb"`}}},
		{"newline", "a\nb", Problems{{RuleNewlineInSubject, "subject"}}},
		{"carriage return", "a\rb", Problems{{RuleNewlineInSubject, "subject"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LintSubject(tc.in)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v want %#v", got, tc.want)
			}
		})
	}
}
