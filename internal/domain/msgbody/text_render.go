package msgbody

import (
	"strings"
	"unicode"

	"golang.org/x/text/width"
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
			for j, piece := range wrapWords(l.text, t.width-StringWidth(l.first)) {
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
		for StringWidth(w) > width {
			if cur != "" {
				lines = append(lines, cur)
				cur = ""
			}
			var head string
			head, w = SplitWidth(w, width)
			lines = append(lines, head)
		}
		switch {
		case cur == "":
			cur = w
		case StringWidth(cur)+1+StringWidth(w) <= width:
			cur += " " + w
		default:
			lines = append(lines, cur)
			cur = w
		}
	}
	return append(lines, cur)
}

// StringWidth is the number of terminal columns s occupies: wide and
// fullwidth East Asian characters take two, combining marks none.
func StringWidth(s string) int {
	n := 0
	for _, r := range s {
		n += runeWidth(r)
	}
	return n
}

func runeWidth(r rune) int {
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
		return 0
	}
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	}
	return 1
}

// SplitWidth cuts s after at most w columns (at least one rune).
func SplitWidth(s string, w int) (string, string) {
	n := 0
	for i, r := range s {
		rw := runeWidth(r)
		if n+rw > w && i > 0 {
			return s[:i], s[i:]
		}
		n += rw
	}
	return s, ""
}
