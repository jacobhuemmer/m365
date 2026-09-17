package cli

import (
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/domain"
)

func loginAll(t *testing.T, d *Deps) {
	t.Helper()
	d.Login = graph.FakeLoginAll(true, true, true, true)
	if c := Run([]string{"m365", "auth", "login"}, *d); c != 0 {
		t.Fatal(c)
	}
}

func TestCalendarListGet(t *testing.T) {
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()
	if c := Run([]string{"m365", "calendar", "calendars"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var page map[string]any
	if err := json.Unmarshal(out.Bytes(), &page); err != nil {
		t.Fatal(err, out.String())
	}
	if page["limit"].(float64) != 20 {
		t.Fatalf("calendars limit %v", page["limit"])
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "list", "--start", "2026-09-16T00:00:00Z", "--end", "2026-09-23T00:00:00Z"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	if err := json.Unmarshal(out.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page["limit"].(float64) != 10 {
		t.Fatalf("events limit %v", page["limit"])
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "get", "ev-1"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "get", "missing"}, d); c != domain.ExitNotFound {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "list", "--top", "51"}, d); c != domain.ExitUsage {
		t.Fatal(c)
	}
}
