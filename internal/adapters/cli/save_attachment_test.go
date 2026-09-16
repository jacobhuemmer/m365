package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestSaveRefuseAndPath(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.txt")
	out.Reset()
	if c := Run([]string{"m365", "mail", "save-attachment", "msg-1", "att-1"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "mail", "save-attachment", "msg-1", "att-1", "--out", dest}, d); c != 0 {
		t.Fatal(errw.String())
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatal(err)
	}
	if c := Run([]string{"m365", "mail", "save-attachment", "msg-1", "att-1", "--out", dest}, d); c != domain.ExitUsage {
		t.Fatal("overwrite", errw.String())
	}
}
