package teams

import (
	"strings"

	"github.com/masonhuemmer/m365/internal/domain"
)

// ChatMap persists 1:1 person keys to chat IDs for the signed-in account.
type ChatMap interface {
	Lookup(account, key string) (string, bool, error)
	Remember(account, key, chatID string) error
}

func chatMapKeys(c domain.Chat, self string) []string {
	if !strings.EqualFold(c.Type, "oneOnOne") {
		return nil
	}
	if c.ID == SelfChatID {
		return nil
	}
	var keys []string
	seen := map[string]bool{}
	add := func(raw string) {
		k := chatMapKey(raw)
		if k == "" || seen[k] {
			return
		}
		seen[k] = true
		keys = append(keys, k)
	}
	for _, p := range c.Members {
		if self != "" && strings.EqualFold(p.Address, self) {
			continue
		}
		add(p.Address)
		add(p.Name)
	}
	return keys
}

func chatMapKey(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}
	if i := strings.IndexByte(s, '@'); i > 0 {
		s = s[:i]
	}
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func rememberChats(m ChatMap, account string, chats []domain.Chat) {
	if m == nil || account == "" {
		return
	}
	for _, c := range chats {
		for _, key := range chatMapKeys(c, account) {
			_ = m.Remember(account, key, c.ID)
		}
	}
}
