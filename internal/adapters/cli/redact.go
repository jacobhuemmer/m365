package cli

import "regexp"

var secretRE = regexp.MustCompile(`(?i)(?:bearer\s+[A-Za-z0-9._\-]+|(?:access_token|refresh_token|authorization_code|client_secret|typesafe_api_key)\s*[:=]\s*["']?[^"'\s,}\]]+["']?)`)

func redact(s string) string {
	return secretRE.ReplaceAllString(s, "[redacted]")
}

func redactAny(v any) any { return v }
