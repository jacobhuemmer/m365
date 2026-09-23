package graph

import (
	"strings"
	"testing"
)

func TestMdSubsetToHTMLListLinkFenceParagraph(t *testing.T) {
	src := "Hello world.\n\n- one\n- two\n\nSee [docs](https://example.com).\n\n```\ncode line\n```\n"
	got := mdSubsetToHTML(src)
	for _, want := range []string{
		"<p>Hello world.</p>",
		"<ul><li>one</li><li>two</li></ul>",
		`<a href="https://example.com">docs</a>`,
		"<pre><code>code line</code></pre>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestMdSubsetToHTMLHeadingsBoldOrdered(t *testing.T) {
	src := "# Title\n\n## Sub\n\n**bold** text\n\n1. first\n2. second\n"
	got := mdSubsetToHTML(src)
	for _, want := range []string{
		"<h1>Title</h1>",
		"<h2>Sub</h2>",
		"<strong>bold</strong>",
		"<ol><li>first</li><li>second</li></ol>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestMdSubsetToHTMLSkipsImageSyntax(t *testing.T) {
	got := mdSubsetToHTML("See ![alt](https://example.com/x.png) please.")
	if strings.Contains(got, "<a href=") {
		t.Fatalf("image must not become a link: %q", got)
	}
	if !strings.Contains(got, "![alt](https://example.com/x.png)") {
		t.Fatalf("image markdown should pass through: %q", got)
	}
}

func TestMdSubsetToHTMLKeepsLineBreaksInParagraph(t *testing.T) {
	got := mdSubsetToHTML("Thanks,\nMason")
	if got != "<p>Thanks,<br>Mason</p>" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainTextToHTML(t *testing.T) {
	got := plainTextToHTML("Hi Jeff,\r\n\r\n\r\nUse <b> & **not md**.\n  indented\n\nThanks,\nMason\n")
	want := "<p>Hi Jeff,</p>\n<p>Use &lt;b&gt; &amp; **not md**.<br>  indented</p>\n<p>Thanks,<br>Mason</p>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if plainTextToHTML("\n\n") != "" {
		t.Fatal("blank input should be empty")
	}
}
