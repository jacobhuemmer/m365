package cli

import (
	"encoding/json"
	"testing"
)

func TestCalendarDryRunCreate(t *testing.T) {
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()
	if c := Run([]string{"m365", "calendar", "create", "--dry-run", "--subject", "t", "--start", "2026-09-16T10:00:00Z", "--end", "2026-09-16T11:00:00Z"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var v map[string]any
	if err := json.Unmarshal(out.Bytes(), &v); err != nil || v["dry_run"] != true {
		t.Fatal(out.String(), err)
	}
}
