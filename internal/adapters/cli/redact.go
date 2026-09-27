package cli

import (
	"bytes"
	"encoding/json"
	"regexp"
	"slices"
	"strings"
)

const secretNames = `access_token|refresh_token|authorization_code|client_secret|typesafe_api_key|typesafe_ai_token`

var (
	// A value starting with '<' is taken whole (access_token=<abc>);
	// otherwise it stops at '<' so HTML around a secret survives. The
	// optional quote after the name catches quoted keys ("access_token":"x").
	secretRE     = regexp.MustCompile(`(?i)(?:bearer\s+[A-Za-z0-9._\-]+|(?:` + secretNames + `)["']?\s*[:=]\s*["']?(?:<[^"'\s,}\]]+|[^"'\s,}\]<]+)["']?)`)
	secretKeyRE  = regexp.MustCompile(`(?i)^(?:` + secretNames + `)$`)
	jsonStringRE = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	jsonColonRE  = regexp.MustCompile(`^\s*:\s*$`)
)

// redact removes secrets from text written to stderr or MCP. JSON is
// redacted inside its string values so it stays valid JSON. When JSON
// lines are mixed with plain lines, each line is handled on its own;
// all-plain text is redacted as a whole, so a match may span lines.
func redact(s string) string {
	if isJSON(s) {
		return redactJSONStrings(s)
	}
	lines := strings.SplitAfter(s, "\n")
	if !slices.ContainsFunc(lines, isJSON) {
		return secretRE.ReplaceAllString(s, "[redacted]")
	}
	for i, l := range lines {
		if isJSON(l) {
			lines[i] = redactJSONStrings(l)
		} else {
			lines[i] = secretRE.ReplaceAllString(l, "[redacted]")
		}
	}
	return strings.Join(lines, "")
}

func redactAny(v any) any { return v }

func isJSON(s string) bool {
	t := strings.TrimSpace(s)
	return t != "" && json.Valid([]byte(t))
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
		out := secretRE.ReplaceAllString(v, "[redacted]")
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

func jsonQuote(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSuffix(buf.String(), "\n")
}
