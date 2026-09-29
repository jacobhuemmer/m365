package msgbody

import "testing"

func TestRender(t *testing.T) {
	mdAll := "# T\n\n## S\n\n### U\n\n**b** and `x`\n\n- one\n* two\n\n1. first\n\nSee [docs](https://example.com)\n\n```\n<code>\n```"
	mdAllHTML := "<h1>T</h1>\n<h2>S</h2>\n<h3>U</h3>\n<p><strong>b</strong> and <code>x</code></p>\n" +
		"<ul><li>one</li><li>two</li></ul>\n<ol><li>first</li></ol>\n" +
		"<p>See <a href=\"https://example.com\">docs</a></p>\n<pre><code>&lt;code&gt;</code></pre>"
	cases := []struct {
		name   string
		mode   Mode
		target Target
		src    string
		want   string
	}{
		{"plain mail paragraphs", Plain, Mail, "a\n\nb\nc", "<p>a</p>\n<p>b<br>c</p>"},
		{"plain teams one paragraph", Plain, Teams, "a\n\nb\nc", "<p>a<br><br>b<br>c</p>"},
		{"plain teams line break", Plain, Teams, "a\nb", "<p>a<br>b</p>"},
		{"plain whitespace-only lines are blank", Plain, Mail, "a\n \n\t\nb", "<p>a</p>\n<p>b</p>"},
		{"plain whitespace-only lines teams", Plain, Teams, "a\n \n\t\nb", "<p>a<br><br>b</p>"},
		{"plain escapes", Plain, Mail, "x < y & z", "<p>x &lt; y &amp; z</p>"},
		{"plain blank is empty", Plain, Mail, "   ", ""},
		{"plain blank is empty teams", Plain, Teams, " \n\n ", ""},
		{"md all constructs mail", Markdown, Mail, mdAll, mdAllHTML},
		{"md escapes prose", Markdown, Mail, "a <b> & c", "<p>a &lt;b&gt; &amp; c</p>"},
		{"md escapes inside bold", Markdown, Mail, "**a<b**", "<p><strong>a&lt;b</strong></p>"},
		{"md link href escaped once", Markdown, Mail, "[q](https://x.com/?a=1&b=2)", "<p><a href=\"https://x.com/?a=1&amp;b=2\">q</a></p>"},
		{"md inline code keeps markdown literal", Markdown, Mail, "`**x** [l](u)`", "<p><code>**x** [l](u)</code></p>"},
		{"md inline code escapes", Markdown, Mail, "`a<b`", "<p><code>a&lt;b</code></p>"},
		{"md teams merges adjacent paragraphs", Markdown, Teams, "p1\n\np2\n\n- i\n\np3\n\np4", "<p>p1<br><br>p2</p>\n<ul><li>i</li></ul>\n<p>p3<br><br>p4</p>"},
		{"md teams keeps line breaks", Markdown, Teams, "a\nb\n\nc", "<p>a<br>b<br><br>c</p>"},
		{"md teams heading stays a block", Markdown, Teams, "# H\n\np", "<h1>H</h1>\n<p>p</p>"},
		{"md blank is empty", Markdown, Mail, "  \n", ""},
		{"html mail unchanged", HTML, Mail, "<p>x</p>\n<p>y</p>", "<p>x</p>\n<p>y</p>"},
		{"html teams unchanged", HTML, Teams, "<p>x</p>\n<p>y</p>", "<p>x</p>\n<p>y</p>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Render(tc.mode, tc.target, tc.src)
			want := Rendered{ContentType: "html", Content: tc.want}
			if got != want {
				t.Fatalf("got %#v\nwant %#v", got, want)
			}
		})
	}
}
