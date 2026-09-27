package cli

import "testing"

// Any value under a secret-named key is redacted, whatever its JSON type.
func TestRedactSecretKeyAnyType(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"number", `{"client_secret":12345}`, `{"client_secret":"[redacted]"}`},
		{"nested number", `{"body":"{\"client_secret\":12345}"}`, `{"body":"{\"client_secret\":\"[redacted]\"}"}`},
		{"object", `{"access_token":{"v":"x"},"n":1}`, `{"access_token":"[redacted]","n":1}`},
		{"array with spaces", `{"refresh_token" : [1, "x"] }`, `{"refresh_token" : "[redacted]" }`},
		{"bool", `{"typesafe_api_key":true}`, `{"typesafe_api_key":"[redacted]"}`},
		{"secret name as a value is not a key", `{"name":"access_token","v":"ok"}`, `{"name":"access_token","v":"ok"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := redact(tc.in); got != tc.want {
				t.Fatalf("redact(%s)\ngot  %s\nwant %s", tc.in, got, tc.want)
			}
		})
	}
}
