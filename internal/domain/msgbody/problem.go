package msgbody

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/masonhuemmer/m365/internal/domain"
)

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

// MarshalJSON writes an empty list, never null, and leaves < and > as-is
// so the caller's encoder decides on HTML escaping.
func (p Problems) MarshalJSON() ([]byte, error) {
	if p == nil {
		return []byte("[]"), nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode([]Problem(p)); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// Err is the usage error for a send blocked by these problems, or nil.
func (p Problems) Err() error {
	if len(p) == 0 {
		return nil
	}
	parts := make([]string, len(p))
	for i, pr := range p {
		parts[i] = pr.Rule + ": " + pr.Detail
	}
	noun := "format problems"
	if len(p) == 1 {
		noun = "format problem"
	}
	return &domain.Error{
		Class:   domain.ClassUsage,
		Message: fmt.Sprintf("%d %s: %s", len(p), noun, strings.Join(parts, "; ")),
		Hint:    "run with --preview to see them",
	}
}
