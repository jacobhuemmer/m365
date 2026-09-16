package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestReadAttachAndWrite(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(p, []byte("synthetic-ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := ReadAttach(p)
	if err != nil || f.Name != "note.txt" || f.Size != 12 {
		t.Fatalf("%+v %v", f, err)
	}
	if _, err := ReadAttach("https://example.com/a"); domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out.txt")
	if err := WriteFile(out, []byte("x"), false); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(out, []byte("y"), false); domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal("refuse overwrite")
	}
	b, _ := os.ReadFile(out)
	if string(b) != "x" {
		t.Fatalf("changed: %s", b)
	}
	if err := WriteFile(out, []byte("y"), true); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile("", []byte("z"), false); domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal("empty path")
	}
}
