package watchstate

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/masonhuemmer/m365/internal/domain"
)

type File struct {
	Path string
}

func DefaultPath() string {
	if d := os.Getenv("XDG_STATE_HOME"); d != "" {
		return filepath.Join(d, "m365", "watch.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "m365", "watch.json")
}

func (f *File) path() string {
	if f != nil && f.Path != "" {
		return f.Path
	}
	return DefaultPath()
}

func (f *File) Load() (domain.WatchCheckpoint, error) {
	b, err := os.ReadFile(f.path())
	if err != nil {
		return domain.NewCheckpoint(), nil
	}
	var cp domain.WatchCheckpoint
	if err := json.Unmarshal(b, &cp); err != nil {
		return domain.NewCheckpoint(), nil
	}
	if cp.Marks == nil {
		cp.Marks = map[string]domain.WatchMark{}
	}
	return cp, nil
}

func (f *File) Save(cp domain.WatchCheckpoint) error {
	p := f.path()
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	b, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o600)
}
