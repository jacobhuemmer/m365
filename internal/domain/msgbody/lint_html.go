package msgbody

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// allowedTags is the allow-list set by the live client probe (design
// decision 8): what Outlook and Teams render as intended.
var allowedTags = map[string]bool{
	"p": true, "br": true, "h1": true, "h2": true, "h3": true, "ul": true, "ol": true, "li": true,
	"pre": true, "code": true, "a": true, "strong": true, "b": true, "em": true, "i": true,
	"u": true, "s": true, "blockquote": true, "hr": true,
}

// voidTags never take an end tag.
var voidTags = map[string]bool{
	"br": true, "hr": true, "img": true, "input": true, "meta": true, "link": true, "wbr": true,
	"col": true, "area": true, "base": true, "embed": true, "source": true, "track": true, "param": true,
}

// lineTags start a new line of visible text, for the markdown-heading rule.
var lineTags = map[string]bool{
	"p": true, "br": true, "h1": true, "h2": true, "h3": true, "ul": true, "ol": true, "li": true,
	"pre": true, "blockquote": true, "hr": true,
}

var linkSchemes = map[string]bool{"http": true, "https": true, "mailto": true}

type openTag struct {
	name    string
	content bool
}

// linter walks the tokens as written. html.Parse would repair broken
// markup, so the tokenizer and an explicit stack are used instead.
type linter struct {
	out       Problems
	stack     []openTag
	codeDepth int
	lineStart bool
	brRun     int
	// Visible text is collected across inline tags and checked per line,
	// so markdown split by a tag (**bo<strong>ld</strong>**) is still seen.
	pending      strings.Builder
	pendingStart bool
	// tok counts tokens; each problem records the token where it starts
	// (pendingTok for buffered text) so Lint can return document order.
	tok   int
	brTok int
	segs  []textSeg
	seqs  []position
}

// position orders problems: the token where a problem starts, then the
// byte offset within buffered text.
type position struct{ tok, off int }

// textSeg maps a byte offset in the pending text to its source token.
type textSeg struct{ off, tok int }

// Lint checks a rendered body against the format rules
// (contracts/format-rules.md) and lists problems in document order.
func Lint(content string) Problems {
	l := &linter{lineStart: true}
	z := html.NewTokenizer(strings.NewReader(content))
	for {
		tt := z.Next()
		l.tok++
		switch tt {
		case html.ErrorToken:
			l.flushText()
			l.endBreaks()
			for _, o := range l.stack {
				l.add(RuleBrokenHTML, "unclosed <"+o.name+">")
			}
			return l.ordered()
		case html.StartTagToken, html.SelfClosingTagToken:
			l.start(z.Token(), tt == html.SelfClosingTagToken)
		case html.EndTagToken:
			l.end(z.Token().Data)
		case html.TextToken:
			l.text(string(z.Text()))
		}
	}
}

func (l *linter) add(rule, detail string) {
	l.addAt(position{tok: l.tok}, rule, detail)
}

func (l *linter) addAt(pos position, rule, detail string) {
	l.out = append(l.out, Problem{Rule: rule, Detail: detail})
	l.seqs = append(l.seqs, pos)
}

// ordered returns the problems sorted by the token where each starts;
// problems from the same token keep the order they were found in.
func (l *linter) ordered() Problems {
	idx := make([]int, len(l.out))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		pa, pb := l.seqs[idx[a]], l.seqs[idx[b]]
		return pa.tok < pb.tok || (pa.tok == pb.tok && pa.off < pb.off)
	})
	out := make(Problems, 0, len(l.out))
	for _, i := range idx {
		out = append(out, l.out[i])
	}
	return out
}

func (l *linter) start(tok html.Token, selfClosing bool) {
	name := tok.Data
	if !allowedTags[name] {
		l.add(RuleTagNotAllowed, "<"+name+">")
	}
	for _, a := range tok.Attr {
		if name == "a" && a.Key == "href" {
			l.checkHref(a.Val)
			continue
		}
		l.add(RuleAttributeNotAllowed, a.Key+" on <"+name+">")
	}
	if lineTags[name] || name == "code" {
		l.flushText()
	}
	if name == "br" {
		if l.brRun == 0 {
			l.brTok = l.tok
		}
		l.brRun++
		l.lineStart = true
		return
	}
	l.endBreaks()
	l.markContent()
	if lineTags[name] {
		l.lineStart = true
	}
	if voidTags[name] || selfClosing {
		return
	}
	if name == "code" || name == "pre" {
		l.codeDepth++
	}
	l.stack = append(l.stack, openTag{name: name})
}

func (l *linter) end(name string) {
	if lineTags[name] || name == "code" {
		l.flushText()
	}
	l.endBreaks()
	i := len(l.stack) - 1
	for i >= 0 && l.stack[i].name != name {
		i--
	}
	if i < 0 {
		if !allowedTags[name] {
			l.add(RuleTagNotAllowed, "</"+name+">")
		}
		l.add(RuleBrokenHTML, "stray </"+name+">")
		return
	}
	if top := l.stack[len(l.stack)-1]; i != len(l.stack)-1 {
		l.add(RuleBrokenHTML, "</"+name+"> closes <"+top.name+">")
	}
	if name == "p" && !l.stack[i].content {
		l.add(RuleExtraBlankLines, "empty <p>")
	}
	for _, o := range l.stack[i:] {
		if o.name == "code" || o.name == "pre" {
			l.codeDepth--
		}
	}
	l.stack = l.stack[:i]
	if lineTags[name] {
		l.lineStart = true
	}
}

func (l *linter) text(s string) {
	if strings.TrimSpace(s) == "" {
		if l.pending.Len() > 0 {
			l.segs = append(l.segs, textSeg{off: l.pending.Len(), tok: l.tok})
			l.pending.WriteString(s)
		}
		// A newline-only token still starts a new line for the heading rule.
		if strings.Contains(s, "\n") {
			l.lineStart = true
		}
		return
	}
	l.endBreaks()
	l.markContent()
	if l.codeDepth == 0 {
		if l.pending.Len() == 0 {
			l.pendingStart = l.lineStart
		}
		l.segs = append(l.segs, textSeg{off: l.pending.Len(), tok: l.tok})
		l.pending.WriteString(s)
	}
	l.lineStart = strings.HasSuffix(s, "\n")
}

// flushText checks the visible text collected since the last line break,
// block tag or code boundary.
func (l *linter) flushText() {
	if l.pending.Len() == 0 {
		return
	}
	for _, p := range locatedTextProblems(l.pending.String(), l.pendingStart, "") {
		tok := l.segs[0].tok
		for _, sg := range l.segs {
			if sg.off <= p.off {
				tok = sg.tok
			}
		}
		l.addAt(position{tok: tok, off: p.off}, p.Rule, p.Detail)
	}
	l.segs = nil
	l.pending.Reset()
}

// markContent records that every open element holds something visible.
func (l *linter) markContent() {
	for i := range l.stack {
		l.stack[i].content = true
	}
}

// endBreaks closes a run of <br>; three or more read as stray blank lines.
func (l *linter) endBreaks() {
	if l.brRun >= 3 {
		l.addAt(position{tok: l.brTok}, RuleExtraBlankLines, fmt.Sprintf("%d <br> in a row", l.brRun))
	}
	l.brRun = 0
}

func (l *linter) checkHref(v string) {
	href := strings.TrimSpace(v)
	if href == "" {
		l.add(RuleLinkScheme, "empty href")
		return
	}
	scheme, _, ok := strings.Cut(href, ":")
	if !ok || !linkSchemes[strings.ToLower(scheme)] {
		l.add(RuleLinkScheme, cut(href))
	}
}
