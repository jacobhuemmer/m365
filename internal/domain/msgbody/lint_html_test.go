package msgbody

import (
	"reflect"
	"testing"
)

type lintCase struct {
	name string
	in   string
	want Problems
}

func runLint(t *testing.T, cases []lintCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Lint(tc.in)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Lint(%q)\ngot  %#v\nwant %#v", tc.in, got, tc.want)
			}
		})
	}
}

func TestLintTagNotAllowed(t *testing.T) {
	runLint(t, []lintCase{
		{"every allowed tag passes",
			`<h1>a</h1><h2>b</h2><h3>c</h3><p>d<br>e<br/>f</p><ul><li>g</li></ul><ol><li>h</li></ol>` +
				`<pre><code>i</code></pre><p><a href="https://example.com">j</a> <strong>k</strong> <b>l</b> ` +
				`<em>m</em> <i>n</i> <u>o</u> <s>p</s></p><blockquote>q</blockquote><hr>`, nil},
		{"table", `<table><tr><td>x</td></tr></table>`, Problems{
			{RuleTagNotAllowed, "<table>"}, {RuleTagNotAllowed, "<tr>"}, {RuleTagNotAllowed, "<td>"}}},
		{"span", `<p><span>x</span></p>`, Problems{{RuleTagNotAllowed, "<span>"}}},
		{"div", `<div>x</div>`, Problems{{RuleTagNotAllowed, "<div>"}}},
		{"upper-case allowed tag passes", `<P>x</P>`, nil},
		{"upper-case disallowed tag", `<DIV>x</DIV>`, Problems{{RuleTagNotAllowed, "<div>"}}},
	})
}

func TestLintAttributeNotAllowed(t *testing.T) {
	runLint(t, []lintCase{
		{"href on a passes", `<p><a href="https://example.com">x</a></p>`, nil},
		{"style on p", `<p style="color:red">x</p>`, Problems{{RuleAttributeNotAllowed, "style on <p>"}}},
		{"class on strong", `<p><strong class="x">y</strong></p>`, Problems{{RuleAttributeNotAllowed, "class on <strong>"}}},
		{"target on a", `<p><a href="https://example.com" target="_blank">x</a></p>`, Problems{{RuleAttributeNotAllowed, "target on <a>"}}},
		{"href on p", `<p href="https://example.com">x</p>`, Problems{{RuleAttributeNotAllowed, "href on <p>"}}},
		{"upper-case attribute", `<p STYLE="x">y</p>`, Problems{{RuleAttributeNotAllowed, "style on <p>"}}},
	})
}

func TestLintBrokenHTML(t *testing.T) {
	runLint(t, []lintCase{
		{"balanced passes", `<p><strong>x</strong> <em>y</em></p>`, nil},
		{"void tags need no end tag", `<p>a<br>b<br/>c</p><hr>`, nil},
		{"unclosed", `<p>hi`, Problems{{RuleBrokenHTML, "unclosed <p>"}}},
		{"unclosed nested reported in order", `<ul><li>x`, Problems{
			{RuleBrokenHTML, "unclosed <ul>"}, {RuleBrokenHTML, "unclosed <li>"}}},
		{"mismatched", `<p><em><strong>x</em></strong></p>`, Problems{
			{RuleBrokenHTML, "</em> closes <strong>"}, {RuleBrokenHTML, "stray </strong>"}}},
		{"stray", `<p>x</p></p>`, Problems{{RuleBrokenHTML, "stray </p>"}}},
	})
}

func TestLintExtraBlankLines(t *testing.T) {
	runLint(t, []lintCase{
		{"two breaks pass", `<p>a<br><br>b</p>`, nil},
		{"empty p", `<p>a</p><p> </p>`, Problems{{RuleExtraBlankLines, "empty <p>"}}},
		{"p holding only a break", `<p><br></p>`, Problems{{RuleExtraBlankLines, "empty <p>"}}},
		{"three breaks", `<p>a<br><br><br>b</p>`, Problems{{RuleExtraBlankLines, "3 <br> in a row"}}},
		{"breaks separated by spaces", `<p>a<br> <br> <br>b</p>`, Problems{{RuleExtraBlankLines, "3 <br> in a row"}}},
		{"four breaks reported once", `<p>a<br><br><br><br>b</p>`, Problems{{RuleExtraBlankLines, "4 <br> in a row"}}},
	})
}

func TestLintLinkScheme(t *testing.T) {
	runLint(t, []lintCase{
		{"http https mailto pass",
			`<p><a href="http://example.com">a</a> <a href="HTTPS://example.com">b</a> <a href="mailto:x@example.com">c</a></p>`, nil},
		{"javascript", `<p><a href="javascript:alert(1)">x</a></p>`, Problems{{RuleLinkScheme, "javascript:alert(1)"}}},
		{"data", `<p><a href="data:text/html,x">x</a></p>`, Problems{{RuleLinkScheme, "data:text/html,x"}}},
		{"relative", `<p><a href="/path">x</a></p>`, Problems{{RuleLinkScheme, "/path"}}},
		{"empty", `<p><a href="">x</a></p>`, Problems{{RuleLinkScheme, "empty href"}}},
	})
}
