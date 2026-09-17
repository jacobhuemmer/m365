package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestFilesDownloadAndDryUpload(t *testing.T) {
	d, out, errw := testDeps()
	loginAll(t, &d)
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.txt")
	out.Reset()
	if c := Run([]string{"m365", "files", "download", "file-1", "--out", dest}, d); c != 0 {
		t.Fatal(errw.String())
	}
	b, err := os.ReadFile(dest)
	if err != nil || string(b) != "synthetic-ok" {
		t.Fatalf("%s %v", b, err)
	}
	out.Reset()
	if c := Run([]string{"m365", "files", "download", "file-1", "--out", dest}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	src := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(src, []byte("synthetic-ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errw.Reset()
	if c := Run([]string{"m365", "files", "upload", "--dry-run", "--file", src}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var v map[string]any
	if err := json.Unmarshal(out.Bytes(), &v); err != nil || v["dry_run"] != true {
		t.Fatal(out.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "files", "download", "file-1"}, d); c != domain.ExitUsage {
		t.Fatal(c)
	}
}
