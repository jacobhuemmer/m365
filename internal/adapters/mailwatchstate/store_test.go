package mailwatchstate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestFileStateIsolatesAccountAndFolderWithProtectedModes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "m365", "mail-watch.json")
	store := &File{Path: path}
	first := domain.MailWatchState{
		Cursor:        "cursor-a",
		Revisions:     map[string]string{"msg-a": "rev-a"},
		RevisionOrder: []string{"msg-a"},
	}
	second := domain.MailWatchState{
		Cursor:        "cursor-b",
		Revisions:     map[string]string{"msg-b": "rev-b"},
		RevisionOrder: []string{"msg-b"},
	}
	if err := store.Save("User@Example.com", "inbox", first); err != nil {
		t.Fatal(err)
	}
	if err := store.Save("user@example.com", "archive/2026", second); err != nil {
		t.Fatal(err)
	}

	gotFirst, err := store.Load(" user@example.COM ", "inbox")
	if err != nil || gotFirst.Cursor != "cursor-a" || gotFirst.Revisions["msg-a"] != "rev-a" {
		t.Fatalf("first state = %+v, %v", gotFirst, err)
	}
	gotSecond, err := store.Load("user@example.com", "archive/2026")
	if err != nil || gotSecond.Cursor != "cursor-b" || gotSecond.Revisions["msg-b"] != "rev-b" {
		t.Fatalf("second state = %+v, %v", gotSecond, err)
	}
	empty, err := store.Load("other@example.com", "inbox")
	if err != nil || empty.Cursor != "" || empty.Revisions == nil || empty.RevisionOrder == nil {
		t.Fatalf("missing state = %+v, %v", empty, err)
	}

	fileInfo, err := os.Stat(path)
	if err != nil || fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("file mode = %v, %v", fileInfo.Mode().Perm(), err)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil || dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("directory mode = %v, %v", dirInfo.Mode().Perm(), err)
	}
}

func TestFileStateReturnsMalformedAndUnreadableErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mail-watch.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := &File{Path: path}
	if _, err := store.Load("user@example.com", "inbox"); err == nil {
		t.Fatal("expected malformed state error")
	}

	directoryPath := filepath.Join(t.TempDir(), "state-as-directory")
	if err := os.Mkdir(directoryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := (&File{Path: directoryPath}).Load("user@example.com", "inbox"); err == nil {
		t.Fatal("expected unreadable state error")
	}
}

func TestFileStateFailedReplacementPreservesPreviousCheckpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mail-watch.json")
	store := &File{Path: path}
	if err := store.Save("user@example.com", "inbox", domain.MailWatchState{Cursor: "old"}); err != nil {
		t.Fatal(err)
	}
	store.rename = func(_, _ string) error { return errors.New("rename failed") }
	if err := store.Save("user@example.com", "inbox", domain.MailWatchState{Cursor: "new"}); err == nil {
		t.Fatal("expected replacement failure")
	}
	store.rename = nil
	got, err := store.Load("user@example.com", "inbox")
	if err != nil || got.Cursor != "old" {
		t.Fatalf("checkpoint after failed replacement = %+v, %v", got, err)
	}
	temporaries, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".mail-watch.json-*"))
	if err != nil || len(temporaries) != 0 {
		t.Fatalf("temporary files = %v, %v", temporaries, err)
	}
}

func TestFileStatePrunesOldestRevisions(t *testing.T) {
	state := domain.NewMailWatchState()
	for i := 0; i < domain.MaxMailWatchRevisions+2; i++ {
		id := fmt.Sprintf("msg-%04d", i)
		state.Revisions[id] = "revision"
		state.RevisionOrder = append(state.RevisionOrder, id)
	}
	store := &File{Path: filepath.Join(t.TempDir(), "mail-watch.json")}
	if err := store.Save("user@example.com", "inbox", state); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load("user@example.com", "inbox")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Revisions) != domain.MaxMailWatchRevisions || got.Revisions["msg-0000"] != "" || got.Revisions["msg-0001"] != "" {
		t.Fatalf("pruned state has %d revisions", len(got.Revisions))
	}
}

func TestMemoryStateCopiesValuesAcrossLoads(t *testing.T) {
	store := &Memory{}
	state := domain.MailWatchState{Cursor: "cursor", Revisions: map[string]string{"msg": "rev"}, RevisionOrder: []string{"msg"}}
	if err := store.Save("user@example.com", "inbox", state); err != nil {
		t.Fatal(err)
	}
	state.Revisions["msg"] = "mutated"
	got, err := store.Load("user@example.com", "inbox")
	if err != nil || got.Revisions["msg"] != "rev" {
		t.Fatalf("memory state = %+v, %v", got, err)
	}
	got.Revisions["msg"] = "also-mutated"
	again, _ := store.Load("user@example.com", "inbox")
	if again.Revisions["msg"] != "rev" {
		t.Fatalf("memory load was not copied: %+v", again)
	}
}
