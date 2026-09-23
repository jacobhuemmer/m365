package graph

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

var (
	reLink = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reBold = regexp.MustCompile(`\*\*([^*]+)\*\*`)
)

// mdSubsetToHTML converts a documented markdown subset to HTML.
// Headings (# / ## / ###), bold (**text**), lists (- / * / 1.), links,
// fenced code, and blank-line paragraphs. Tables, images, and raw HTML
// are left as-is. Not a full markdown implementation.
func mdSubsetToHTML(src string) string {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.TrimRight(src, "\n")
	if src == "" {
		return ""
	}
	lines := strings.Split(src, "\n")
	var b strings.Builder
	for i := 0; i < len(lines); {
		line := lines[i]
		if isFence(line) {
			body, next := consumeFence(lines, i)
			writeBlock(&b, "<pre><code>"+html.EscapeString(body)+"</code></pre>")
			i = next
			continue
		}
		if level, text, ok := heading(line); ok {
			tag := "h" + strconv.Itoa(level)
			writeBlock(&b, "<"+tag+">"+inlineMD(text)+"</"+tag+">")
			i++
			continue
		}
		if text, ok := ulItem(line); ok {
			var items []string
			items, i = consumeList(lines, i, ulItem)
			writeList(&b, "ul", append([]string{text}, items...))
			continue
		}
		if text, ok := olItem(line); ok {
			var items []string
			items, i = consumeList(lines, i, olItem)
			writeList(&b, "ol", append([]string{text}, items...))
			continue
		}
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		var para []string
		para, i = consumeParagraph(lines, i)
		writeBlock(&b, "<p>"+inlineMD(strings.Join(para, "<br>"))+"</p>")
	}
	return b.String()
}

// plainTextToHTML escapes plain text and keeps its layout: blank lines
// become paragraphs and single newlines become <br>. Outlook reply comments
// and Teams messages otherwise collapse newlines into one line.
func plainTextToHTML(src string) string {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	var b strings.Builder
	for _, block := range strings.Split(src, "\n\n") {
		block = strings.Trim(block, "\n")
		if strings.TrimSpace(block) == "" {
			continue
		}
		lines := strings.Split(block, "\n")
		for i, l := range lines {
			lines[i] = html.EscapeString(l)
		}
		writeBlock(&b, "<p>"+strings.Join(lines, "<br>")+"</p>")
	}
	return b.String()
}

func writeBlock(b *strings.Builder, s string) {
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(s)
}

func writeList(b *strings.Builder, tag string, items []string) {
	var inner strings.Builder
	inner.WriteByte('<')
	inner.WriteString(tag)
	inner.WriteByte('>')
	for _, it := range items {
		inner.WriteString("<li>")
		inner.WriteString(inlineMD(it))
		inner.WriteString("</li>")
	}
	inner.WriteString("</")
	inner.WriteString(tag)
	inner.WriteByte('>')
	writeBlock(b, inner.String())
}

func isFence(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "```")
}

func consumeFence(lines []string, i int) (string, int) {
	i++
	var body []string
	for i < len(lines) {
		if isFence(lines[i]) {
			return strings.Join(body, "\n"), i + 1
		}
		body = append(body, lines[i])
		i++
	}
	return strings.Join(body, "\n"), i
}

func heading(line string) (int, string, bool) {
	s := strings.TrimLeft(line, " ")
	n := 0
	for n < len(s) && s[n] == '#' {
		n++
	}
	if n < 1 || n > 3 {
		return 0, "", false
	}
	if n == len(s) || s[n] != ' ' {
		return 0, "", false
	}
	return n, strings.TrimSpace(s[n+1:]), true
}

func ulItem(line string) (string, bool) {
	s := strings.TrimLeft(line, " ")
	if strings.HasPrefix(s, "- ") || strings.HasPrefix(s, "* ") {
		return s[2:], true
	}
	return "", false
}

func olItem(line string) (string, bool) {
	s := strings.TrimLeft(line, " ")
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 || i+1 >= len(s) || s[i] != '.' || s[i+1] != ' ' {
		return "", false
	}
	return s[i+2:], true
}

func consumeList(lines []string, i int, item func(string) (string, bool)) ([]string, int) {
	i++
	var rest []string
	for i < len(lines) {
		text, ok := item(lines[i])
		if !ok {
			break
		}
		rest = append(rest, text)
		i++
	}
	return rest, i
}

func consumeParagraph(lines []string, i int) ([]string, int) {
	var para []string
	for i < len(lines) {
		line := lines[i]
		if strings.TrimSpace(line) == "" || isFence(line) {
			break
		}
		if _, _, ok := heading(line); ok {
			break
		}
		if _, ok := ulItem(line); ok {
			break
		}
		if _, ok := olItem(line); ok {
			break
		}
		para = append(para, line)
		i++
	}
	return para, i
}

func inlineMD(s string) string {
	s = applyLinks(s)
	return reBold.ReplaceAllString(s, "<strong>$1</strong>")
}

func applyLinks(s string) string {
	matches := reLink.FindAllStringSubmatchIndex(s, -1)
	if matches == nil {
		return s
	}
	var b strings.Builder
	last := 0
	for _, m := range matches {
		if m[0] > 0 && s[m[0]-1] == '!' {
			continue
		}
		b.WriteString(s[last:m[0]])
		b.WriteString(`<a href="`)
		b.WriteString(html.EscapeString(s[m[4]:m[5]]))
		b.WriteString(`">`)
		b.WriteString(s[m[2]:m[3]])
		b.WriteString(`</a>`)
		last = m[1]
	}
	b.WriteString(s[last:])
	return b.String()
}
