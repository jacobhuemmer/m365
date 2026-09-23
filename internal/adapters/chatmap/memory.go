package chatmap

import "sync"

type Memory struct {
	mu       sync.Mutex
	accounts map[string]map[string]string
}

func (m *Memory) Lookup(account, key string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	account, key = normalize(account), chatMapLookupKey(key)
	ids := m.accounts[account]
	if ids == nil {
		return "", false, nil
	}
	id, ok := ids[key]
	return id, ok && id != "", nil
}

func (m *Memory) Remember(account, key, chatID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	account, key = normalize(account), chatMapLookupKey(key)
	if account == "" || key == "" || chatID == "" {
		return nil
	}
	if m.accounts == nil {
		m.accounts = map[string]map[string]string{}
	}
	if m.accounts[account] == nil {
		m.accounts[account] = map[string]string{}
	}
	m.accounts[account][key] = chatID
	return nil
}
