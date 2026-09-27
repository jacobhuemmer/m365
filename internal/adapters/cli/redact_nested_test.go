package cli

import (
	"encoding/json"
	"testing"
)

// JSON nested in a string value is redacted as JSON, so it still parses.
func TestRedactNestedJSONStaysValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"nested array", `{"items":"[{\"refresh_token\":\"r1\"},\"access_token=a1\"]"}`,
			`{"items":"[{\"refresh_token\":\"[redacted]\"},\"[redacted]\"]"}`},
		{"doubly nested", `{"a":"{\"b\":\"{\\\"client_secret\\\":\\\"s\\\"}\"}"}`,
			`{"a":"{\"b\":\"{\\\"client_secret\\\":\\\"[redacted]\\\"}\"}"}`},
		{"nested without secret is untouched", `{"body":"{\"a\":1}"}`, `{"body":"{\"a\":1}"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := redact(tc.in)
			if got != tc.want {
				t.Fatalf("redact(%s)\ngot  %s\nwant %s", tc.in, got, tc.want)
			}
			var outer map[string]string
			if err := json.Unmarshal([]byte(got), &outer); err != nil {
				t.Fatal(err)
			}
			for k, v := range outer {
				if !json.Valid([]byte(v)) {
					t.Fatalf("nested %s is not JSON: %s", k, v)
				}
			}
		})
	}
}
