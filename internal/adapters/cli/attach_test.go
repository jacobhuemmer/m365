package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachDryRunNoBytes(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	dir := t.TempDir()
	p := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(p, []byte("synthetic-ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	c := Run([]string{"m365", "mail", "send", "--dry-run", "--to", "a@b.c", "--subject", "t", "--body", "b", "--attach", p}, d)
	if c != 0 {
		t.Fatal(errw.String())
	}
	if strings.Contains(out.String(), "synthetic-ok") {
		t.Fatal("bytes on stdout")
	}
	var m map[string]any
	_ = json.Unmarshal(out.Bytes(), &m)
	if m["dry_run"] != true {
		t.Fatal(out.String())
	}
}
