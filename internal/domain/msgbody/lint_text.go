package msgbody

import (
	"regexp"
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
	out := textProblems(subject, true, "subject: ")
	if strings.ContainsAny(subject, "\r\n") {
		out = append(out, Problem{Rule: RuleNewlineInSubject, Detail: "subject"})
	}
	return out
}

// textProblems applies rules 4 and 5 to visible text. lineStart says
// whether s begins a line, for the markdown-heading check.
func textProblems(s string, lineStart bool, prefix string) Problems {
	var out Problems
	add := func(rule, detail string) {
		out = append(out, Problem{Rule: rule, Detail: prefix + detail})
	}
	for i, line := range strings.Split(s, "\n") {
		if (i > 0 || lineStart) && reLeftHeading.MatchString(strings.TrimLeft(line, " \t")) {
			add(RuleLeftoverMarkdown, cut(strings.TrimSpace(line)))
		}
		for _, re := range []*regexp.Regexp{reLeftBold, reLeftLink, reLeftCode} {
			for _, m := range re.FindAllString(line, -1) {
				add(RuleLeftoverMarkdown, cut(m))
			}
		}
	}
	for _, seq := range literalEscape {
		if strings.Contains(s, seq) {
			add(RuleLiteralEscape, seq+` in "`+cut(strings.TrimSpace(s))+`"`)
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
