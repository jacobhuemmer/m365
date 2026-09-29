package teams

import (
	"context"
	"strings"

	"github.com/masonhuemmer/m365/internal/domain"
)

func resolveRecipient(ctx context.Context, st Store, m ChatMap, sess domain.Session, in SendInput) (string, error) {
	if in.ExactRecipient {
		chat, err := findExactRecipient(ctx, st, sess, in.To)
		return chat.ID, err
	}
	q, err := ParseQuery(in.To, false)
	if err != nil {
		return "", err
	}
	r, err := FindMapped(ctx, st, m, sess, q, 0)
	if err != nil {
		return "", err
	}
	if r.Incomplete {
		return "", domain.Usage("chat search incomplete")
	}
	if r.Count == 0 {
		return "", domain.NotFound("no matching chat")
	}
	if r.Count > 1 {
		ids := make([]string, 0, r.Count)
		for _, c := range r.Items {
			ids = append(ids, c.ID)
		}
		return "", domain.Usagef("several matches (%s); pass a chat id", strings.Join(ids, ", "))
	}
	return r.Items[0].ID, nil
}

// selfIsMember reports whether the signed-in account is among the chat members.
// Exact matching needs it: without it, a self alias could pick another person's chat.
func selfIsMember(chat domain.Chat, account string) bool {
	if account == "" {
		return false
	}
	for _, member := range chat.Members {
		if strings.EqualFold(member.Address, account) {
			return true
		}
	}
	return false
}

func isExactMatch(chat domain.Chat, account, recipient string) bool {
	for _, member := range chat.Members {
		if strings.EqualFold(member.Address, account) {
			continue
		}
		if strings.EqualFold(member.Address, recipient) || strings.EqualFold(member.ID, recipient) {
			return true
		}
	}
	return false
}

func findExactRecipient(ctx context.Context, st Store, sess domain.Session, recipient string) (domain.Chat, error) {
	var found domain.Chat
	page := ""
	for i := 0; i < ScanMaxPages; i++ {
		chats, err := st.ListChats(ctx, ScanPageSize, page)
		if err != nil {
			return domain.Chat{}, err
		}
		for _, chat := range chats.Items {
			if !strings.EqualFold(chat.Type, "oneOnOne") || !selfIsMember(chat, sess.Account) || !isExactMatch(chat, sess.Account, recipient) {
				continue
			}
			if found.ID != "" && found.ID != chat.ID {
				return domain.Chat{}, domain.Usage("several exact recipient chats; pass a chat id")
			}
			found = chat
		}
		if chats.NextPage == nil || *chats.NextPage == "" {
			if found.ID == "" {
				return domain.Chat{}, domain.NotFound("no exact recipient chat")
			}
			return found, nil
		}
		page = *chats.NextPage
	}
	return domain.Chat{}, domain.Usage("chat search incomplete; pass a chat id")
}
