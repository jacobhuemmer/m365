package msgbody

import (
	"html"
	"strings"
)

// PlainTextToHTML escapes plain text and keeps its layout: blank lines
// become paragraphs and single newlines become <br>. Outlook reply comments
// and Teams messages otherwise collapse newlines into one line.
func PlainTextToHTML(src string) string {
	return joinBlocks(plainBlocks(src), Mail)
}

// plainBlocks splits plain text into paragraphs. A line holding only
// spaces or tabs counts as blank, so it never becomes a run of <br>.
func plainBlocks(src string) []block {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			lines[i] = ""
		}
	}
	var out []block
	for _, para := range strings.Split(strings.Join(lines, "\n"), "\n\n") {
		para = strings.Trim(para, "\n")
		if para == "" {
			continue
		}
		pl := strings.Split(para, "\n")
		for i, l := range pl {
			pl[i] = html.EscapeString(l)
		}
		out = append(out, block{html: strings.Join(pl, "<br>"), para: true})
	}
	return out
}
