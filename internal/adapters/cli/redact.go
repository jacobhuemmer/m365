package cli

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

const secretNames = `access_token|refresh_token|authorization_code|client_secret|typesafe_api_key|typesafe_ai_token`

var (
	// The optional quote after the name catches quoted keys
	// ("access_token":"x"). A quoted value is taken whole with its quotes,
	// or to the end of the line when the closing quote is missing; an
	// unquoted one never takes a closing quote that belongs to the
	// surrounding text. A value starting with '<' is taken whole
	// (access_token=<abc>); otherwise it stops at '<' so HTML survives.
	secretRE = regexp.MustCompile(`(?i)(?:bearer\s+[A-Za-z0-9._\-]+|(?:` + secretNames + `)["']?\s*[:=]\s*` +
		`(?:"[^"\n]*"?|'[^'\n]*'?|(?:<[^"'\s,}\]]+|[^"'\s,}\]<]+)))`)
	secretKeyRE  = regexp.MustCompile(`(?i)^(?:` + secretNames + `)$`)
	jsonStringRE = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	jsonColonRE  = regexp.MustCompile(`^\s*:\s*$`)
)

// redact removes secrets from text written to stderr or MCP. JSON is
// redacted inside its string values so it stays valid JSON. When JSON
// object or array lines are mixed with plain lines, each JSON line is
// handled on its own and each run of adjacent plain lines is redacted as
// one text, so a secret split across plain lines is still found.
func redact(s string) string {
	if isJSON(s) {
		return redactJSONStrings(s)
	}
	var b, plain strings.Builder
	flush := func() {
		b.WriteString(secretRE.ReplaceAllString(plain.String(), "[redacted]"))
		plain.Reset()
	}
	for _, line := range strings.SplitAfter(s, "\n") {
		if !isJSONLine(line) {
			plain.WriteString(line)
			continue
		}
		flush()
		b.WriteString(redactJSONStrings(line))
	}
	flush()
	return b.String()
}

func redactAny(v any) any { return v }

func isJSON(s string) bool {
	t := strings.TrimSpace(s)
	return t != "" && json.Valid([]byte(t))
}

// isJSONLine reports a JSON object or array on one line. Scalars such as
// "true" or "42" stay plain text, so they never split a plain run.
func isJSONLine(s string) bool {
	t := strings.TrimSpace(s)
	return (strings.HasPrefix(t, "{") || strings.HasPrefix(t, "[")) && json.Valid([]byte(t))
}

// redactJSONStrings rewrites only the string literals that hold a secret,
// or that are the value of a secret-named key; all other bytes are kept.
func redactJSONStrings(s string) string {
	var b strings.Builder
	last, prevEnd := 0, -1
	prev := ""
	for _, m := range jsonStringRE.FindAllStringIndex(s, -1) {
		var v string
		if err := json.Unmarshal([]byte(s[m[0]:m[1]]), &v); err != nil {
			continue
		}
		out := redactValue(v)
		if prevEnd >= 0 && secretKeyRE.MatchString(prev) && jsonColonRE.MatchString(s[prevEnd:m[0]]) {
			out = "[redacted]"
		}
		prev, prevEnd = v, m[1]
		if out == v {
			continue
		}
		b.WriteString(s[last:m[0]])
		b.WriteString(jsonQuote(out))
		last = m[1]
	}
	b.WriteString(s[last:])
	return b.String()
}

// redactValue redacts one decoded string value. A value that is itself
// a JSON object or array is redacted as JSON, so it still parses.
func redactValue(v string) string {
	if isJSONLine(v) {
		return redactJSONStrings(v)
	}
	return secretRE.ReplaceAllString(v, "[redacted]")
}

func jsonQuote(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSuffix(buf.String(), "\n")
}
