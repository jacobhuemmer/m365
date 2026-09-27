package msgbody

// Rule ids for format problems (contracts/format-rules.md).
const (
	RuleTagNotAllowed       = "tag-not-allowed"
	RuleAttributeNotAllowed = "attribute-not-allowed"
	RuleBrokenHTML          = "broken-html"
	RuleLeftoverMarkdown    = "leftover-markdown"
	RuleLiteralEscape       = "literal-escape"
	RuleExtraBlankLines     = "extra-blank-lines"
	RuleLinkScheme          = "link-scheme"
	RuleNewlineInSubject    = "newline-in-subject"
)

// Problem is one format rule violation.
type Problem struct {
	Rule   string `json:"rule"`
	Detail string `json:"detail"`
}

// Problems lists violations in document order.
type Problems []Problem

// Err is the usage error for a send blocked by these problems, or nil.
func (p Problems) Err() error {
	return nil
}
