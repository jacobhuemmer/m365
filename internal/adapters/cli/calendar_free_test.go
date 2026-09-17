package cli

import (
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestCalendarFree(t *testing.T) {
	t.Setenv("M365_NOW", "2026-09-16T12:00:00-05:00")
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()
	if c := Run([]string{"m365", "calendar", "free"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var p map[string]any
	if err := json.Unmarshal(out.Bytes(), &p); err != nil {
		t.Fatal(err, out.String())
	}
	if p["limit"].(float64) != 5 {
		t.Fatalf("%v", p["limit"])
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "free", "--dry-run"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "free", "--top", "21"}, d); c != domain.ExitUsage {
		t.Fatal(c)
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "free", "--when", "2026-09-19"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	_ = json.Unmarshal(out.Bytes(), &p)
	if p["count"].(float64) != 0 {
		t.Fatalf("booked %v", p["count"])
	}
}
