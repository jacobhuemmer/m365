package mailwatchstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/masonhuemmer/m365/internal/domain"
)

type File struct {
	Path string

	mu     sync.Mutex
	rename func(string, string) error
}

type diskState struct {
	Entries []diskEntry `json:"entries"`
}

type diskEntry struct {
	Account string                `json:"account"`
	Folder  string                `json:"folder"`
	State   domain.MailWatchState `json:"state"`
}

func DefaultPath() string {
	if directory := os.Getenv("XDG_STATE_HOME"); directory != "" {
		return filepath.Join(directory, "m365", "mail-watch.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "m365", "mail-watch.json")
}

func (f *File) path() string {
	if f != nil && f.Path != "" {
		return f.Path
	}
	return DefaultPath()
}

func (f *File) Load(account, folder string) (domain.MailWatchState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	contents, err := f.read()
	if err != nil {
		return domain.MailWatchState{}, err
	}
	account, folder = normalizeKey(account, folder)
	for _, entry := range contents.Entries {
		if entry.Account == account && entry.Folder == folder {
			return normalizedState(entry.State), nil
		}
	}
	return domain.NewMailWatchState(), nil
}

func (f *File) Save(account, folder string, state domain.MailWatchState) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	contents, err := f.read()
	if err != nil {
		return err
	}
	account, folder = normalizeKey(account, folder)
	state = normalizedState(state)
	found := false
	for i := range contents.Entries {
		entry := &contents.Entries[i]
		if entry.Account == account && entry.Folder == folder {
			entry.State = state
			found = true
			break
		}
	}
	if !found {
		contents.Entries = append(contents.Entries, diskEntry{Account: account, Folder: folder, State: state})
	}
	sort.Slice(contents.Entries, func(i, j int) bool {
		if contents.Entries[i].Account != contents.Entries[j].Account {
			return contents.Entries[i].Account < contents.Entries[j].Account
		}
		return contents.Entries[i].Folder < contents.Entries[j].Folder
	})
	payload, err := json.Marshal(contents)
	if err != nil {
		return err
	}
	return f.replace(payload)
}

func (f *File) read() (diskState, error) {
	payload, err := os.ReadFile(f.path())
	if os.IsNotExist(err) {
		return diskState{Entries: []diskEntry{}}, nil
	}
	if err != nil {
		return diskState{}, err
	}
	var contents diskState
	if err := json.Unmarshal(payload, &contents); err != nil {
		return diskState{}, err
	}
	if contents.Entries == nil {
		contents.Entries = []diskEntry{}
	}
	return contents, nil
}

func (f *File) replace(payload []byte) error {
	path := f.path()
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(directory, 0o700); err != nil {
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
	rename := os.Rename
	if f.rename != nil {
		rename = f.rename
	}
	return rename(temporaryPath, path)
}

func normalizeKey(account, folder string) (string, string) {
	return strings.ToLower(strings.TrimSpace(account)), strings.TrimSpace(folder)
}

func normalizedState(state domain.MailWatchState) domain.MailWatchState {
	copyState := domain.MailWatchState{
		Cursor:        state.Cursor,
		Revisions:     make(map[string]string, len(state.Revisions)),
		RevisionOrder: make([]string, 0, len(state.Revisions)),
	}
	for id, revision := range state.Revisions {
		copyState.Revisions[id] = revision
	}
	seen := make(map[string]bool, len(copyState.Revisions))
	for _, id := range state.RevisionOrder {
		if _, ok := copyState.Revisions[id]; ok && !seen[id] {
			seen[id] = true
			copyState.RevisionOrder = append(copyState.RevisionOrder, id)
		}
	}
	missing := make([]string, 0, len(copyState.Revisions)-len(copyState.RevisionOrder))
	for id := range copyState.Revisions {
		if !seen[id] {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	copyState.RevisionOrder = append(copyState.RevisionOrder, missing...)
	for len(copyState.RevisionOrder) > domain.MaxMailWatchRevisions {
		oldest := copyState.RevisionOrder[0]
		copyState.RevisionOrder = copyState.RevisionOrder[1:]
		delete(copyState.Revisions, oldest)
	}
	return copyState
}
