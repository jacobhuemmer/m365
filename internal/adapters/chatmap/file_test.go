package chatmap

import (
	"path/filepath"
	"testing"
)

func TestFileRememberLookup(t *testing.T) {
	f := &File{Path: filepath.Join(t.TempDir(), "chat-map.json")}
	if err := f.Remember("User@Example.com", "ajay kumar", "chat-ajay"); err != nil {
		t.Fatal(err)
	}
	id, ok, err := f.Lookup("user@example.com", "ajay kumar")
	if err != nil || !ok || id != "chat-ajay" {
		t.Fatalf("%s %v %v", id, ok, err)
	}
	_, ok, err = f.Lookup("user@example.com", "nobody")
	if err != nil || ok {
		t.Fatal(ok, err)
	}
}
