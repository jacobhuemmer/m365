package keychain

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stubStore struct {
	blob    Blob
	ok      bool
	putErr  error
	deleted bool
}

func (s *stubStore) Get() (Blob, bool, error) { return s.blob, s.ok, nil }
func (s *stubStore) Put(b Blob) error {
	if s.putErr != nil {
		return s.putErr
	}
	s.blob = b
	s.ok = true
	return nil
}
func (s *stubStore) Delete() error { s.deleted = true; s.ok = false; s.blob = Blob{}; return nil }

func TestFallbackPutUsesFileWhenKeychainTooBig(t *testing.T) {
	primary := &stubStore{putErr: errors.New("data passed to Set was too big")}
	path := filepath.Join(t.TempDir(), "session.json")
	f := &Fallback{Primary: primary, Secondary: &FileStore{Path: path}}
	big := Blob{Account: "user@example.com", Usable: true, AccessToken: strings.Repeat("a", 8000)}
	if err := f.Put(big); err != nil {
		t.Fatal(err)
	}
	got, ok, err := f.Get()
	if err != nil || !ok || len(got.AccessToken) != 8000 {
		t.Fatalf("got ok=%v err=%v len=%d", ok, err, len(got.AccessToken))
	}
	st, err := os.Stat(path)
	if err != nil || st.Mode().Perm() != 0o600 {
		t.Fatalf("session file: %v", err)
	}
}

func TestFallbackPutUsesFileWhenSecretsServiceFails(t *testing.T) {
	primary := &stubStore{putErr: errors.New("failed to unlock correct collection '/org/freedesktop/secrets/aliases/default'")}
	path := filepath.Join(t.TempDir(), "session.json")
	f := &Fallback{Primary: primary, Secondary: &FileStore{Path: path}}
	blob := Blob{Account: "user@example.com", Usable: true, AccessToken: "access-token"}
	if err := f.Put(blob); err != nil {
		t.Fatal(err)
	}
	got, ok, err := f.Get()
	if err != nil || !ok || got.AccessToken != "access-token" || got.Account != "user@example.com" {
		t.Fatalf("got ok=%v err=%v blob=%+v", ok, err, got)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", st.Mode().Perm())
	}
}

func TestFallbackPutPrefersPrimary(t *testing.T) {
	primary := &stubStore{}
	path := filepath.Join(t.TempDir(), "session.json")
	f := &Fallback{Primary: primary, Secondary: &FileStore{Path: path}}
	blob := Blob{Account: "user@example.com", Usable: true, AccessToken: "tok"}
	if err := f.Put(blob); err != nil {
		t.Fatal(err)
	}
	if !primary.ok || primary.blob.AccessToken != "tok" {
		t.Fatal("primary did not store blob")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("secondary written on primary success: %v", err)
	}
	got, ok, err := f.Get()
	if err != nil || !ok || got.AccessToken != "tok" {
		t.Fatalf("got ok=%v err=%v blob=%+v", ok, err, got)
	}
}

func TestFallbackDeleteClearsBoth(t *testing.T) {
	primary := &stubStore{blob: Blob{Account: "k"}, ok: true}
	path := filepath.Join(t.TempDir(), "session.json")
	sec := &FileStore{Path: path}
	if err := sec.Put(Blob{Account: "f", Usable: true}); err != nil {
		t.Fatal(err)
	}
	f := &Fallback{Primary: primary, Secondary: sec}
	if err := f.Delete(); err != nil {
		t.Fatal(err)
	}
	if !primary.deleted {
		t.Fatal("primary not deleted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("secondary still present: %v", err)
	}
}
