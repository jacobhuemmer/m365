package cli

import "regexp"

var secretRE = regexp.MustCompile(`(?i)(access_token|refresh_token|authorization_code|client_secret|bearer\s+[A-Za-z0-9._\-]+)`)

func redact(s string) string {
	return secretRE.ReplaceAllString(s, "[redacted]")
}

func redactAny(v any) any { return v }
