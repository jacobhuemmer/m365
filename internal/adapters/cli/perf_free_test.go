package cli

import (
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestFreeAndWhenUnderDeadline(t *testing.T) {
	t.Setenv("M365_NOW", "2026-09-16T12:00:00-05:00")
	d, _, errw := testDeps()
	loginAll(t, &d)
	start := time.Now()
	if c := Run([]string{"m365", "calendar", "free"}, d); c != domain.ExitOK {
		t.Fatal(errw.String())
	}
	if time.Since(start) >= 15*time.Second {
		t.Fatal("free slow")
	}
	start = time.Now()
	if c := Run([]string{"m365", "calendar", "create", "--dry-run", "--subject", "Sync", "--when", "tomorrow at 1:30 pm"}, d); c != domain.ExitOK {
		t.Fatal(errw.String())
	}
	if time.Since(start) >= 5*time.Second {
		t.Fatal("dry-run slow")
	}
}
