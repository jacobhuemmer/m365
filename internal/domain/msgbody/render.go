package msgbody

import "strings"

// Mode is how the caller wrote the body.
type Mode int

const (
	Plain Mode = iota
	Markdown
	HTML
)

// Target is where the body is delivered. It decides paragraph layout.
type Target int

const (
	Mail Target = iota
	Teams
)

// Rendered is the exact body a send delivers.
type Rendered struct {
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

// block is one top-level HTML block. For a paragraph, html is the inner
// content without the <p> wrapper, so Teams can merge paragraphs.
type block struct {
	html string
	para bool
}

// ModeFor picks the mode from the --html and --format md flags.
func ModeFor(html, md bool) Mode {
	switch {
	case html:
		return HTML
	case md:
		return Markdown
	}
	return Plain
}

// Render turns the caller's body into the HTML a send delivers.
func Render(mode Mode, target Target, src string) Rendered {
	content := src
	switch mode {
	case Plain:
		content = joinBlocks(plainBlocks(src), target)
	case Markdown:
		content = joinBlocks(mdBlocks(src), target)
	}
	return Rendered{ContentType: "html", Content: content}
}

// joinBlocks writes blocks one per line. Mail keeps one <p> per paragraph.
// Teams shows no gap between <p> blocks, so adjacent paragraphs are merged
// into one <p> separated by <br><br> (a blank line on every client).
func joinBlocks(blocks []block, target Target) string {
	var out []string
	var run []string
	flush := func() {
		if len(run) > 0 {
			out = append(out, "<p>"+strings.Join(run, "<br><br>")+"</p>")
			run = nil
		}
	}
	for _, b := range blocks {
		switch {
		case b.para && target == Teams:
			run = append(run, b.html)
		case b.para:
			out = append(out, "<p>"+b.html+"</p>")
		default:
			flush()
			out = append(out, b.html)
		}
	}
	flush()
	return strings.Join(out, "\n")
}
