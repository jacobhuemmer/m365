package cli

import (
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestListUnder15s(t *testing.T) {
	d, _, errw := testDeps()
	login(t, d)
	start := time.Now()
	if c := Run([]string{"m365", "mail", "list"}, d); c != domain.ExitOK {
		t.Fatal(errw.String())
	}
	if c := Run([]string{"m365", "teams", "list"}, d); c != domain.ExitOK {
		t.Fatal(errw.String())
	}
	if time.Since(start) >= 15*time.Second {
		t.Fatalf("lists took %s", time.Since(start))
	}
}
