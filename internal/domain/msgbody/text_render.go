package msgbody

import (
	"strings"
	"unicode/utf8"
)

func (t *textWriter) render() string {
	var out []string
	for i, b := range t.blocks {
		if i > 0 {
			out = append(out, "")
		}
		for _, l := range b {
			if l.text == "" {
				out = append(out, strings.TrimRight(l.first, " "))
				continue
			}
			if !l.wrap {
				out = append(out, l.first+l.text)
				continue
			}
			for j, piece := range wrapWords(l.text, t.width-utf8.RuneCountInString(l.first)) {
				prefix := l.first
				if j > 0 {
					prefix = l.rest
				}
				out = append(out, prefix+piece)
			}
		}
	}
	return strings.Join(out, "\n")
}

// wrapWords wraps at spaces and splits words longer than width.
func wrapWords(s string, width int) []string {
	if width < 1 {
		width = 1
	}
	var lines []string
	cur := ""
	for _, w := range strings.Split(s, " ") {
		for utf8.RuneCountInString(w) > width {
			if cur != "" {
				lines = append(lines, cur)
				cur = ""
			}
			r := []rune(w)
			lines = append(lines, string(r[:width]))
			w = string(r[width:])
		}
		switch {
		case cur == "":
			cur = w
		case utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(w) <= width:
			cur += " " + w
		default:
			lines = append(lines, cur)
			cur = w
		}
	}
	return append(lines, cur)
}

// StringWidth is the number of terminal columns s occupies.
func StringWidth(s string) int {
	return 0
}
