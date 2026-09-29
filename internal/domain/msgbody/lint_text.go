package msgbody

import (
	"regexp"
	"sort"
	"strings"
)

var (
	reLeftBold    = regexp.MustCompile(`\*\*[^*]+\*\*`)
	reLeftLink    = regexp.MustCompile(`\[[^\]]+\]\([^)]+\)`)
	reLeftCode    = regexp.MustCompile("`[^`]+`")
	reLeftHeading = regexp.MustCompile(`^#{1,3} `)
	literalEscape = []string{`\n`, `\t`, `\"`}
)

const excerptLen = 40

// LintSubject checks a mail subject for leftover markdown, literal
// escape sequences, and newlines.
func LintSubject(subject string) Problems {
	located := locatedTextProblems(subject, true, "subject: ")
	if at := strings.IndexAny(subject, "\r\n"); at >= 0 {
		located = append(located, locatedProblem{Problem{Rule: RuleNewlineInSubject, Detail: "subject"}, at})
	}
	return byOffset(located)
}

// textProblems applies rules 4 and 5 to visible text. lineStart says
// whether s begins a line, for the markdown-heading check.
func textProblems(s string, lineStart bool, prefix string) Problems {
	return byOffset(locatedTextProblems(s, lineStart, prefix))
}

// byOffset returns the problems in the order they occur in the text.
func byOffset(located []locatedProblem) Problems {
	sort.SliceStable(located, func(i, j int) bool { return located[i].off < located[j].off })
	out := make(Problems, 0, len(located))
	for _, lp := range located {
		out = append(out, lp.Problem)
	}
	return out
}

// locatedProblem is a problem with the byte offset in the text where it
// starts, so the linter can place it at the token it came from.
type locatedProblem struct {
	Problem
	off int
}

func locatedTextProblems(s string, lineStart bool, prefix string) []locatedProblem {
	var out []locatedProblem
	add := func(off int, rule, detail string) {
		out = append(out, locatedProblem{Problem{Rule: rule, Detail: prefix + detail}, off})
	}
	lineOff := 0
	for i, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if (i > 0 || lineStart) && reLeftHeading.MatchString(trimmed) {
			add(lineOff+len(line)-len(trimmed), RuleLeftoverMarkdown, cut(strings.TrimSpace(line)))
		}
		for _, re := range []*regexp.Regexp{reLeftBold, reLeftLink, reLeftCode} {
			for _, m := range re.FindAllStringIndex(line, -1) {
				add(lineOff+m[0], RuleLeftoverMarkdown, cut(line[m[0]:m[1]]))
			}
		}
		lineOff += len(line) + 1
	}
	for _, seq := range literalEscape {
		if at := strings.Index(s, seq); at >= 0 {
			add(at, RuleLiteralEscape, seq+` in "`+cut(strings.TrimSpace(s))+`"`)
		}
	}
	return out
}

// cut shortens an excerpt to excerptLen characters.
func cut(s string) string {
	r := []rune(s)
	if len(r) > excerptLen {
		return string(r[:excerptLen])
	}
	return s
}
