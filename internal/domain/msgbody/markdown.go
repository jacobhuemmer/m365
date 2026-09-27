package msgbody

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

var (
	reLink = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reBold = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reCode = regexp.MustCompile("`([^`]+)`")
)

// MDSubsetToHTML converts a documented markdown subset to HTML.
// Headings (# / ## / ###), bold (**text**), lists (- / * / 1.), links,
// inline and fenced code, and blank-line paragraphs. Text outside code is
// escaped, so raw HTML shows literally. Tables and images are left as
// text. Not a full markdown implementation.
func MDSubsetToHTML(src string) string {
	return joinBlocks(mdBlocks(src), Mail)
}

func mdBlocks(src string) []block {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.TrimRight(src, "\n")
	if strings.TrimSpace(src) == "" {
		return nil
	}
	lines := strings.Split(src, "\n")
	var out []block
	for i := 0; i < len(lines); {
		line := lines[i]
		if isFence(line) {
			body, next := consumeFence(lines, i)
			out = append(out, block{html: "<pre><code>" + html.EscapeString(body) + "</code></pre>"})
			i = next
			continue
		}
		if level, text, ok := heading(line); ok {
			tag := "h" + strconv.Itoa(level)
			out = append(out, block{html: "<" + tag + ">" + inlineMD(text) + "</" + tag + ">"})
			i++
			continue
		}
		if text, ok := ulItem(line); ok {
			var items []string
			items, i = consumeList(lines, i, ulItem)
			out = append(out, listBlock("ul", append([]string{text}, items...)))
			continue
		}
		if text, ok := olItem(line); ok {
			var items []string
			items, i = consumeList(lines, i, olItem)
			out = append(out, listBlock("ol", append([]string{text}, items...)))
			continue
		}
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		var para []string
		para, i = consumeParagraph(lines, i)
		for j, l := range para {
			para[j] = inlineMD(l)
		}
		out = append(out, block{html: strings.Join(para, "<br>"), para: true})
	}
	return out
}

func listBlock(tag string, items []string) block {
	var b strings.Builder
	b.WriteString("<" + tag + ">")
	for _, it := range items {
		b.WriteString("<li>" + inlineMD(it) + "</li>")
	}
	b.WriteString("</" + tag + ">")
	return block{html: b.String()}
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

// inlineMD renders one line: code spans verbatim (escaped), then links
// and bold on escaped text.
func inlineMD(s string) string {
	var b strings.Builder
	last := 0
	for _, m := range reCode.FindAllStringSubmatchIndex(s, -1) {
		b.WriteString(applyLinks(s[last:m[0]]))
		b.WriteString("<code>" + html.EscapeString(s[m[2]:m[3]]) + "</code>")
		last = m[1]
	}
	b.WriteString(applyLinks(s[last:]))
	return b.String()
}

func applyLinks(s string) string {
	var b strings.Builder
	last := 0
	for _, m := range reLink.FindAllStringSubmatchIndex(s, -1) {
		if m[0] > 0 && s[m[0]-1] == '!' {
			continue
		}
		b.WriteString(boldText(s[last:m[0]]))
		b.WriteString(`<a href="` + html.EscapeString(s[m[4]:m[5]]) + `">` + boldText(s[m[2]:m[3]]) + `</a>`)
		last = m[1]
	}
	b.WriteString(boldText(s[last:]))
	return b.String()
}

func boldText(s string) string {
	return reBold.ReplaceAllString(html.EscapeString(s), "<strong>$1</strong>")
}
