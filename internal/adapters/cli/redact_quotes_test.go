package cli

import "testing"

// An opening quote with no closing quote runs to the end of the line, so
// nothing after a space leaks.
func TestRedactUnterminatedQuote(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"double quote with a space", `access_token="abc def`, `[redacted]`},
		{"single quote with a space", "client_secret='abc def\nnext line", "[redacted]\nnext line"},
		{"terminated value keeps the rest", `access_token="abc def" end`, `[redacted] end`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := redact(tc.in); got != tc.want {
				t.Fatalf("redact(%q)\ngot  %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}
