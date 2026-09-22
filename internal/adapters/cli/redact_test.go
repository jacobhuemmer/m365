package cli

import (
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	s := redact("Authorization: Bearer abc.def TYPESAFE_API_KEY=jev-secret access_token=graph-secret")
	for _, secret := range []string{"abc.def", "jev-secret", "graph-secret"} {
		if strings.Contains(s, secret) {
			t.Fatal(s)
		}
	}
	d, _, errw := testDeps()
	Run([]string{"m365", "--verbose", "auth", "status"}, d)
	if strings.Contains(errw.String(), "access_token") {
		t.Fatal(errw.String())
	}
}
