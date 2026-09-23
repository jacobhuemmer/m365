package cli

import (
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestVersion(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"m365", "--version"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	got := strings.TrimSpace(out.String())
	if got != "dev" {
		t.Fatalf("got %q", got)
	}

	out.Reset()
	errw.Reset()
	if c := Run([]string{"m365", "--human", "mcp", "serve"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
}
