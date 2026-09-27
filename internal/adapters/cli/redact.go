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
	jsonColonRE  = regexp.MustCompile(`^\s*:\s*`)
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

// redactJSONStrings rewrites only what holds a secret: the whole value
// after a secret-named key (any JSON type becomes "[redacted]"), and any
// other string literal whose text contains a secret. Every other byte is
// kept, so unchanged output is byte-identical.
func redactJSONStrings(s string) string {
	var b strings.Builder
	last, pos := 0, 0
	for {
		m := jsonStringRE.FindStringIndex(s[pos:])
		if m == nil {
			break
		}
		start, end := pos+m[0], pos+m[1]
		pos = end
		var v string
		if err := json.Unmarshal([]byte(s[start:end]), &v); err != nil {
			continue
		}
		if secretKeyRE.MatchString(v) {
			if c := jsonColonRE.FindStringIndex(s[end:]); c != nil {
				valStart := end + c[1]
				if n := jsonValueLen(s[valStart:]); n > 0 {
					b.WriteString(s[last:valStart])
					b.WriteString(`"[redacted]"`)
					last, pos = valStart+n, valStart+n
					continue
				}
			}
		}
		if out := redactValue(v); out != v {
			b.WriteString(s[last:start])
			b.WriteString(jsonQuote(out))
			last = end
		}
	}
	b.WriteString(s[last:])
	return b.String()
}

// jsonValueLen is the byte length of the JSON value at the start of s,
// or 0 if there is none.
func jsonValueLen(s string) int {
	dec := json.NewDecoder(strings.NewReader(s))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return 0
	}
	return int(dec.InputOffset())
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
