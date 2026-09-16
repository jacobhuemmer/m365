package cli

import (
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	s := redact("Authorization: Bearer abc.def")
	if strings.Contains(s, "abc.def") {
		t.Fatal(s)
	}
	d, _, errw := testDeps()
	Run([]string{"m365", "--verbose", "auth", "status"}, d)
	if strings.Contains(errw.String(), "access_token") {
		t.Fatal(errw.String())
	}
}
