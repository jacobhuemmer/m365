package cli

import (
	"strings"
	"testing"
)

func TestCalendarFilesHelp(t *testing.T) {
	d, out, _ := testDeps()
	Run([]string{"m365", "--help"}, d)
	s := out.String()
	if !strings.Contains(s, "calendar") || !strings.Contains(s, "files") {
		t.Fatal(s)
	}
	out.Reset()
	Run([]string{"m365", "calendar", "--help"}, d)
	s = out.String()
	if !strings.Contains(s, "default 10") || !strings.Contains(s, "7 days") || !strings.Contains(s, "dry-run") {
		t.Fatal(s)
	}
	out.Reset()
	Run([]string{"m365", "files", "--help"}, d)
	s = out.String()
	if !strings.Contains(s, "100 MiB") || !strings.Contains(s, "default 20") {
		t.Fatal(s)
	}
	out.Reset()
	Run([]string{"m365", "mail", "list", "--help"}, d)
	if !strings.Contains(out.String(), "default 10") {
		t.Fatal(out.String())
	}
	out.Reset()
	Run([]string{"m365", "teams", "list", "--help"}, d)
	if !strings.Contains(out.String(), "default 20") {
		t.Fatal(out.String())
	}
}
