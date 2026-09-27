package msgbody

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// Text approximates a rendered body as plain text wrapped to width, for
// the terminal preview (contracts/preview.md). Blocks are separated by a
// blank line; unknown or broken markup shows as its text.
func Text(content string, width int) string {
	if width < 1 {
		width = 1
	}
	t := &textWriter{width: width}
	z := html.NewTokenizer(strings.NewReader(content))
	for {
		switch z.Next() {
		case html.ErrorToken:
			t.flushBlock()
			return t.render()
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := z.Token()
			t.start(tok.Data, tok.Attr)
		case html.EndTagToken:
			t.end(z.Token().Data)
		case html.TextToken:
			t.text(string(z.Text()))
		}
	}
}

type textLine struct {
	text, first, rest string
	wrap              bool
}

type listCtx struct {
	ordered bool
	n       int
}

type textWriter struct {
	width        int
	blocks       [][]textLine
	cur          []textLine
	line         strings.Builder
	space        bool
	first, rest  string
	lists        []listCtx
	quote        int
	pre          bool
	href         string
	inLink       bool
	labelStart   int
	headingLevel int
}

func (t *textWriter) base() string {
	p := strings.Repeat("> ", t.quote)
	if t.pre {
		p += "    "
	}
	return p
}

func (t *textWriter) start(name string, attrs []html.Attribute) {
	switch name {
	case "p":
		t.flushBlock()
	case "br":
		t.endLine(true)
	case "h1", "h2", "h3":
		t.flushBlock()
		t.headingLevel = int(name[1] - '0')
	case "ul", "ol":
		if len(t.lists) == 0 {
			t.flushBlock()
		} else {
			t.endLine(false)
		}
		t.lists = append(t.lists, listCtx{ordered: name == "ol"})
	case "li":
		t.endLine(false)
		if len(t.lists) > 0 {
			ctx := &t.lists[len(t.lists)-1]
			ctx.n++
			marker := "• "
			if ctx.ordered {
				marker = fmt.Sprintf("%d. ", ctx.n)
			}
			indent := t.base() + strings.Repeat("  ", len(t.lists)-1)
			t.first = indent + marker
			t.rest = indent + strings.Repeat(" ", utf8.RuneCountInString(marker))
		}
	case "pre":
		t.flushBlock()
		t.pre = true
		t.first, t.rest = t.base(), t.base()
	case "blockquote":
		t.flushBlock()
		t.quote++
		t.first, t.rest = t.base(), t.base()
	case "hr":
		t.flushBlock()
		t.cur = []textLine{{text: strings.Repeat("─", t.width)}}
		t.flushBlock()
	case "a":
		t.href = ""
		for _, a := range attrs {
			if a.Key == "href" {
				t.href = a.Val
			}
		}
		t.writeSpace()
		t.inLink = true
		t.labelStart = t.line.Len()
	}
}

func (t *textWriter) end(name string) {
	switch name {
	case "p":
		t.flushBlock()
	case "h1", "h2", "h3":
		t.endLine(false)
		if n := len(t.cur); n > 0 {
			ch := "-"
			if t.headingLevel == 1 {
				ch = "="
			}
			w := min(StringWidth(t.cur[n-1].text), t.width)
			t.cur = append(t.cur, textLine{text: strings.Repeat(ch, w), first: t.first})
		}
		t.flushBlock()
	case "ul", "ol":
		t.endLine(false)
		if len(t.lists) > 0 {
			t.lists = t.lists[:len(t.lists)-1]
		}
		if len(t.lists) == 0 {
			t.flushBlock()
		}
	case "li":
		t.endLine(false)
		t.first, t.rest = t.base(), t.base()
	case "pre":
		t.flushBlock()
		t.pre = false
		t.first, t.rest = t.base(), t.base()
	case "blockquote":
		t.flushBlock()
		if t.quote > 0 {
			t.quote--
		}
		t.first, t.rest = t.base(), t.base()
	case "a":
		if !t.inLink {
			return
		}
		t.inLink = false
		if label := t.line.String()[t.labelStart:]; t.href != "" && label != t.href {
			t.line.WriteString(" (" + t.href + ")")
		}
	}
}

func (t *textWriter) text(s string) {
	if t.pre {
		for i, seg := range strings.Split(s, "\n") {
			if i > 0 {
				t.endLine(true)
			}
			t.line.WriteString(seg)
		}
		return
	}
	for _, r := range s {
		if unicode.IsSpace(r) {
			t.space = t.line.Len() > 0
			continue
		}
		t.writeSpace()
		t.line.WriteRune(r)
	}
}

func (t *textWriter) writeSpace() {
	if t.space {
		t.line.WriteByte(' ')
		t.space = false
	}
}

// endLine closes the current line. A forced end (a <br> or a pre newline)
// keeps an empty line; otherwise empty lines are dropped.
func (t *textWriter) endLine(force bool) {
	if t.line.Len() > 0 || force {
		t.cur = append(t.cur, textLine{text: t.line.String(), first: t.first, rest: t.rest, wrap: !t.pre})
		t.first = t.rest
	}
	t.line.Reset()
	t.space = false
	// A link label that spans a line break continues on the new line.
	if t.inLink {
		t.labelStart = 0
	}
}

func (t *textWriter) flushBlock() {
	t.endLine(false)
	if len(t.cur) > 0 {
		t.blocks = append(t.blocks, t.cur)
	}
	t.cur = nil
	t.first, t.rest = t.base(), t.base()
}
