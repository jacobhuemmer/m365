package keychain

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/masonhuemmer/m365/internal/domain"
)

type FileStore struct {
	Path string
}

func LiveSessionPath() string {
	if d := os.Getenv("XDG_STATE_HOME"); d != "" {
		return filepath.Join(d, "m365", "session.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "m365", "session.json")
}

func (f *FileStore) Get() (Blob, bool, error) {
	b, err := os.ReadFile(f.Path)
	if err != nil {
		return Blob{}, false, nil
	}
	var blob Blob
	if err := json.Unmarshal(b, &blob); err != nil {
		return Blob{}, false, nil
	}
	return blob, true, nil
}

func (f *FileStore) Put(blob Blob) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o700); err != nil {
		return domain.Auth("could not save session")
	}
	b, err := json.Marshal(blob)
	if err != nil {
		return domain.Auth("could not save session")
	}
	if err := os.WriteFile(f.Path, b, 0o600); err != nil {
		return domain.Auth("could not save session")
	}
	return nil
}

func (f *FileStore) Delete() error {
	_ = os.Remove(f.Path)
	return nil
}
