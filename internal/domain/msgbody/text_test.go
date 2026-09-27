package msgbody

import "testing"

func TestText(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		width int
		want  string
	}{
		{"paragraphs get a blank line", "<p>Hello</p>\n<p>World</p>", 80, "Hello\n\nWorld"},
		{"teams br pair is a blank line", "<p>a<br><br>b<br>c</p>", 80, "a\n\nb\nc"},
		{"bullets", "<ul><li>one</li><li>two</li></ul>", 80, "• one\n• two"},
		{"numbered", "<ol><li>first</li><li>second</li></ol>", 80, "1. first\n2. second"},
		{"headings underlined", "<h1>Title</h1>\n<h2>Sub</h2>\n<h3>Low</h3>", 80, "Title\n=====\n\nSub\n---\n\nLow\n---"},
		{"links", `<p>See <a href="https://example.com">docs</a> or <a href="https://x.io">https://x.io</a></p>`, 80,
			"See docs (https://example.com) or https://x.io"},
		{"pre indented", "<pre><code>line 1\n  indented</code></pre>", 80, "    line 1\n      indented"},
		{"pre not wrapped", "<pre><code>abcdefghij</code></pre>", 4, "    abcdefghij"},
		{"blockquote", "<blockquote>quoted text</blockquote>", 80, "> quoted text"},
		{"hr", "<p>a</p><hr><p>b</p>", 10, "a\n\n──────────\n\nb"},
		{"entities decoded", "<p>x &lt; y &amp; z</p>", 80, "x < y & z"},
		{"inline tags are plain text", "<p><strong>bold</strong> and <code>code</code></p>", 80, "bold and code"},
		{"source newlines are spaces", "<p>a\nb</p>", 80, "a b"},
		{"word wrap", "<p>aaa bbb ccc ddd</p>", 7, "aaa bbb\nccc ddd"},
		{"long word split", "<p>abcdefghij</p>", 4, "abcd\nefgh\nij"},
		{"list item hanging indent", "<ul><li>one two three</li></ul>", 9, "• one two\n  three"},
		{"contract example",
			"<p>Hi <strong>team</strong></p>\n<ul><li>one</li><li>two</li></ul>\n<p>See <a href=\"https://example.com\">docs</a></p>", 36,
			"Hi team\n\n• one\n• two\n\nSee docs (https://example.com)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Text(tc.in, tc.width); got != tc.want {
				t.Fatalf("Text(%q, %d)\ngot  %q\nwant %q", tc.in, tc.width, got, tc.want)
			}
		})
	}
}
