package cli

import (
	"testing"
	"time"
)

func TestTeamsFindUnder15s(t *testing.T) {
	d, _, errw := testDeps()
	login(t, d)
	start := time.Now()
	if c := Run([]string{"m365", "teams", "find", "Ajay"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	if time.Since(start) > 15*time.Second {
		t.Fatal("find slow")
	}
}
