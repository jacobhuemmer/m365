package cli

import (
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestCalendarFilesListUnder15s(t *testing.T) {
	d, _, errw := testDeps()
	loginAll(t, &d)
	start := time.Now()
	if c := Run([]string{"m365", "calendar", "list", "--start", "2026-09-16T00:00:00Z", "--end", "2026-09-23T00:00:00Z"}, d); c != domain.ExitOK {
		t.Fatal(errw.String())
	}
	if c := Run([]string{"m365", "files", "list"}, d); c != domain.ExitOK {
		t.Fatal(errw.String())
	}
	if time.Since(start) >= 15*time.Second {
		t.Fatalf("lists took %s", time.Since(start))
	}
}
