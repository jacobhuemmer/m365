package cli

import (
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestMCPHelpAndHuman(t *testing.T) {
	d, out, errw := testDeps()
	if c := Run([]string{"m365", "--help"}, d); c != 0 {
		t.Fatal(c)
	}
	if !strings.Contains(out.String(), "mcp") {
		t.Fatal(out.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "mcp", "--help"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	s := out.String()
	for _, want := range []string{"m365_status", "m365_help", "m365_run", "write_opt_in", "serve", "mail-search", "teams-find"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
	out.Reset()
	if c := Run([]string{"m365", "mcp", "serve", "--help"}, d); c != 0 {
		t.Fatal(c, errw.String())
	}
	if c := Run([]string{"m365", "mcp", "nope"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	if c := Run([]string{"m365", "--human", "mcp", "serve"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
}
