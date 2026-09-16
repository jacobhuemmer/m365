package watchstate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestCheckpoint0600NoBodies(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "watch.json")
	f := &File{Path: p}
	cp := domain.NewCheckpoint()
	cp.Marks["chat-1"] = domain.WatchMark{MessageID: "cmsg-1", Created: "t"}
	if err := f.Save(cp); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", st.Mode().Perm())
	}
	b, _ := os.ReadFile(p)
	if string(b) == "" || containsToken(string(b)) {
		t.Fatalf("bad file %s", b)
	}
	got, err := f.Load()
	if err != nil || got.Marks["chat-1"].MessageID != "cmsg-1" {
		t.Fatalf("%+v %v", got, err)
	}
}

func containsToken(s string) bool {
	return false
}
