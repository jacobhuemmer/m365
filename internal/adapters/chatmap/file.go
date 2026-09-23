package chatmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type File struct {
	Path string

	mu sync.Mutex
}

type diskState struct {
	Accounts map[string]map[string]string `json:"accounts"`
}

func DefaultPath() string {
	if directory := os.Getenv("XDG_STATE_HOME"); directory != "" {
		return filepath.Join(directory, "m365", "chat-map.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "m365", "chat-map.json")
}

func (f *File) path() string {
	if f != nil && f.Path != "" {
		return f.Path
	}
	return DefaultPath()
}

func (f *File) Lookup(account, key string) (string, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	contents, err := f.read()
	if err != nil {
		return "", false, err
	}
	account, key = normalize(account), chatMapLookupKey(key)
	ids := contents.Accounts[account]
	if ids == nil {
		return "", false, nil
	}
	id, ok := ids[key]
	return id, ok && id != "", nil
}

func (f *File) Remember(account, key, chatID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	contents, err := f.read()
	if err != nil {
		return err
	}
	account, key = normalize(account), chatMapLookupKey(key)
	chatID = strings.TrimSpace(chatID)
	if account == "" || key == "" || chatID == "" {
		return nil
	}
	if contents.Accounts == nil {
		contents.Accounts = map[string]map[string]string{}
	}
	if contents.Accounts[account] == nil {
		contents.Accounts[account] = map[string]string{}
	}
	contents.Accounts[account][key] = chatID
	payload, err := json.Marshal(contents)
	if err != nil {
		return err
	}
	return f.replace(payload)
}

func (f *File) read() (diskState, error) {
	payload, err := os.ReadFile(f.path())
	if os.IsNotExist(err) {
		return diskState{Accounts: map[string]map[string]string{}}, nil
	}
	if err != nil {
		return diskState{}, err
	}
	var contents diskState
	if err := json.Unmarshal(payload, &contents); err != nil {
		return diskState{}, err
	}
	if contents.Accounts == nil {
		contents.Accounts = map[string]map[string]string{}
	}
	return contents, nil
}

func (f *File) replace(payload []byte) error {
	path := f.path()
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(directory, 0o700); err != nil { // #nosec G302 -- state directories require owner execute permission
		return err
	}
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	closed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		_ = os.Remove(temporaryPath)
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	closed = true
	return os.Rename(temporaryPath, path)
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func chatMapLookupKey(raw string) string {
	s := normalize(raw)
	if i := strings.IndexByte(s, '@'); i > 0 {
		s = s[:i]
	}
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.Join(strings.Fields(s), " ")
}
