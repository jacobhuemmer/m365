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
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := redact(tc.in); got != tc.want {
				t.Fatalf("redact(%q)\ngot  %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}
