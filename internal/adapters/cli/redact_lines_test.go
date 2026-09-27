package cli

import "testing"

// Mixed JSON and plain output: plain lines next to each other are
// redacted together, so a secret split across them is still found.
func TestRedactMixedLinesSpanPlainRuns(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"scalar line then split bearer", "true\nAuthorization: Bearer\nabc.def", "true\nAuthorization: [redacted]"},
		{"json object then split bearer", "{\"a\":\"ok\"}\nBearer\nabc.def", "{\"a\":\"ok\"}\n[redacted]"},
		{"number line then split key", "42\naccess_token=\nxyz", "42\n[redacted]"},
		{"plain run between json lines", "{\"a\":1}\nBearer\nabc.def\n{\"b\":\"access_token=x\"}", "{\"a\":1}\n[redacted]\n{\"b\":\"[redacted]\"}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := redact(tc.in); got != tc.want {
				t.Fatalf("redact(%q)\ngot  %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}
