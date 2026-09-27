package cli

import "testing"

// Leaks found in the Codex review of PR #14.
func TestRedactReviewFindings(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"json embedded in a string value", `{"body":"{\"access_token\":\"xyz\"}"}`, `{"body":"{\"[redacted]}"}`},
		{"json line followed by plain text", "{\"access_token\":\"xyz\"}\nstatus", "{\"access_token\":\"[redacted]\"}\nstatus"},
		{"plain value in angle brackets", `access_token=<abc>`, `[redacted]`},
		{"json inside a plain log line", `status {"access_token":"xyz"}`, `status {"[redacted]}`},
		{"plain bearer across a newline", "Authorization: Bearer\nabc.def", "Authorization: [redacted]"},
		{"quoted plain value keeps surrounding text", `say "access_token=abc" now`, `say "[redacted]" now`},
		{"quoted value with spaces", `access_token="abc def" end`, `[redacted] end`},
		{"single-quoted value", `client_secret='s3cr3t'`, `[redacted]`},
		{"json escapes in untouched values are kept", "{\"e\":\"\\u003cp\\u003e\",\"t\":\"access_token=x\"}", "{\"e\":\"\\u003cp\\u003e\",\"t\":\"[redacted]\"}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := redact(tc.in); got != tc.want {
				t.Fatalf("redact(%q)\ngot  %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}
