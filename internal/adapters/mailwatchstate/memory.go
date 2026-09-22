package mailwatchstate

import (
	"sync"

	"github.com/masonhuemmer/m365/internal/domain"
)

type Memory struct {
	mu     sync.Mutex
	states map[stateKey]domain.MailWatchState
}

type stateKey struct {
	account string
	folder  string
}

func (m *Memory) Load(account, folder string) (domain.MailWatchState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	account, folder = normalizeKey(account, folder)
	if m.states == nil {
		return domain.NewMailWatchState(), nil
	}
	state, ok := m.states[stateKey{account: account, folder: folder}]
	if !ok {
		return domain.NewMailWatchState(), nil
	}
	return normalizedState(state), nil
}

func (m *Memory) Save(account, folder string, state domain.MailWatchState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.states == nil {
		m.states = map[stateKey]domain.MailWatchState{}
	}
	account, folder = normalizeKey(account, folder)
	m.states[stateKey{account: account, folder: folder}] = normalizedState(state)
	return nil
}
