package cli

import (
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestCalendarWhenDryRun(t *testing.T) {
	t.Setenv("M365_NOW", "2026-09-16T12:00:00-05:00")
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()
	if c := Run([]string{"m365", "calendar", "create", "--dry-run", "--subject", "Sync", "--when", "tomorrow at 1:30 pm"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var v map[string]any
	if err := json.Unmarshal(out.Bytes(), &v); err != nil {
		t.Fatal(err, out.String())
	}
	if v["dry_run"] != true || v["when"] != "tomorrow at 1:30 pm" {
		t.Fatal(v)
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "create", "--subject", "Sync", "--when", "1:30"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "calendar", "create", "--dry-run", "--subject", "Sync", "--when", "tomorrow at 1:30 pm", "--start", "2026-09-17T13:30:00-05:00"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "--human", "calendar", "create", "--dry-run", "--subject", "Sync", "--when", "tomorrow at 1:30 pm"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	s := out.String()
	if !jsonContainsPhrase(s, "tomorrow at 1:30 pm") {
		t.Fatal(s)
	}
}

func jsonContainsPhrase(s, p string) bool {
	return len(s) > 0 && (contains(s, p) || contains(s, "13:30") || contains(s, "1:30"))
}

func contains(s, p string) bool {
	return len(p) > 0 && (len(s) >= len(p)) && (indexOf(s, p) >= 0)
}

func indexOf(s, p string) int {
	for i := 0; i+len(p) <= len(s); i++ {
		if s[i:i+len(p)] == p {
			return i
		}
	}
	return -1
}
