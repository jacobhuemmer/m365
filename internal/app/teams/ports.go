package teams

import (
	"context"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/domain"
)

type ListQuery struct {
	Top, PageToken string
	TopN           int
	PageTokenVal   string
}

type MessageQuery struct {
	ChatID        string
	Top           int
	PageToken     string
	IncludeSystem bool
}

type SendInput struct {
	ChatID         string
	To             string
	Text           string
	HTML           bool
	MD             bool
	DryRun         bool
	NoteToSelf     bool
	ExactRecipient bool
	Files          []domain.OutboundFile
}

type WatchQuery struct {
	Since string
	Chats []string
	Top   int
}

type Store interface {
	ListChats(ctx context.Context, top int, page string) (domain.ChatPage, error)
	GetChat(ctx context.Context, id string) (domain.Chat, error)
	Messages(ctx context.Context, q MessageQuery) (domain.ChatMessagePage, error)
	Send(ctx context.Context, in SendInput) (string, error)
	Watch(ctx context.Context, q WatchQuery) ([]domain.WatchEvent, error)
	Attachments(ctx context.Context, chatID, messageID string) ([]domain.Attachment, error)
	Download(ctx context.Context, chatID, messageID, attach string) ([]byte, domain.Attachment, error)
}

func List(ctx context.Context, st Store, sess domain.Session, top int, page string) (domain.ChatPage, error) {
	return ListMapped(ctx, st, nil, sess, top, page)
}

func ListMapped(ctx context.Context, st Store, m ChatMap, sess domain.Session, top int, page string) (domain.ChatPage, error) {
	if err := auth.Require(sess, false, true); err != nil {
		return domain.ChatPage{}, err
	}
	n, err := domain.NormalizeTop(top, domain.DefaultTeamsTop)
	if err != nil {
		return domain.ChatPage{}, err
	}
	p, err := st.ListChats(ctx, n, page)
	if err != nil {
		return domain.ChatPage{}, err
	}
	rememberChats(m, sess.Account, p.Items)
	return p, nil
}

func Get(ctx context.Context, st Store, sess domain.Session, id string) (domain.Chat, error) {
	if err := auth.Require(sess, false, true); err != nil {
		return domain.Chat{}, err
	}
	if strings.TrimSpace(id) == "" {
		return domain.Chat{}, domain.Usage("chat id is required")
	}
	return st.GetChat(ctx, id)
}

func Messages(ctx context.Context, st Store, sess domain.Session, q MessageQuery) (domain.ChatMessagePage, error) {
	if err := auth.Require(sess, false, true); err != nil {
		return domain.ChatMessagePage{}, err
	}
	if q.ChatID == "" {
		return domain.ChatMessagePage{}, domain.Usage("chat id is required")
	}
	n, err := domain.NormalizeTop(q.Top, domain.DefaultTeamsTop)
	if err != nil {
		return domain.ChatMessagePage{}, err
	}
	q.Top = n
	return st.Messages(ctx, q)
}

func Send(ctx context.Context, st Store, sess domain.Session, in SendInput) (any, error) {
	return SendMapped(ctx, st, nil, sess, in)
}

func SendMapped(ctx context.Context, st Store, m ChatMap, sess domain.Session, in SendInput) (any, error) {
	if err := auth.Require(sess, false, true); err != nil {
		return nil, err
	}
	if in.NoteToSelf {
		if in.To != "" || in.ChatID != "" {
			return nil, domain.Usage("use --note-to-self or --to or a chat id, not both")
		}
		in.ChatID = SelfChatID
	}
	if in.To != "" && in.ChatID != "" {
		return nil, domain.Usage("use --to or a chat id, not both")
	}
	if in.To != "" {
		if in.ExactRecipient {
			chat, err := findExactRecipient(ctx, st, sess, in.To)
			if err != nil {
				return nil, err
			}
			in.ChatID = chat.ID
		} else {
			q, err := ParseQuery(in.To, false)
			if err != nil {
				return nil, err
			}
			r, err := FindMapped(ctx, st, m, sess, q, 0)
			if err != nil {
				return nil, err
			}
			if r.Incomplete {
				return nil, domain.Usage("chat search incomplete")
			}
			if r.Count == 0 {
				return nil, domain.NotFound("no matching chat")
			}
			if r.Count > 1 {
				ids := make([]string, 0, r.Count)
				for _, c := range r.Items {
					ids = append(ids, c.ID)
				}
				return nil, domain.Usagef("several matches (%s); pass a chat id", strings.Join(ids, ", "))
			}
			in.ChatID = r.Items[0].ID
		}
	}
	if in.ChatID == "" || in.Text == "" {
		return nil, domain.Usage("chat id and text are required")
	}
	if err := domain.ValidateOutbound(in.Files); err != nil {
		return nil, err
	}
	if in.DryRun {
		atts := []map[string]any{}
		for _, f := range in.Files {
			atts = append(atts, map[string]any{"name": f.Name, "size": f.Size})
		}
		out := map[string]any{"dry_run": true, "chat_id": in.ChatID, "text": in.Text, "attachments": atts}
		if in.To != "" {
			out["to"] = in.To
		}
		return out, nil
	}
	id, err := st.Send(ctx, in)
	if err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "sent": true}, nil
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
			if !strings.EqualFold(chat.Type, "oneOnOne") {
				continue
			}
			for _, member := range chat.Members {
				if strings.EqualFold(member.Address, sess.Account) {
					continue
				}
				if strings.EqualFold(member.Address, recipient) || strings.EqualFold(member.ID, recipient) {
					if found.ID != "" && found.ID != chat.ID {
						return domain.Chat{}, domain.Usage("several exact recipient chats; pass a chat id")
					}
					found = chat
				}
			}
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

func Watch(ctx context.Context, st Store, sess domain.Session, q WatchQuery) ([]domain.WatchEvent, error) {
	if err := auth.Require(sess, false, true); err != nil {
		return nil, err
	}
	n, err := domain.NormalizeTop(q.Top, domain.DefaultTeamsTop)
	if err != nil {
		return nil, err
	}
	q.Top = n
	return st.Watch(ctx, q)
}

func ListAttachments(ctx context.Context, st Store, sess domain.Session, chatID, msgID string) ([]domain.Attachment, error) {
	if err := auth.Require(sess, false, true); err != nil {
		return nil, err
	}
	if chatID == "" || msgID == "" {
		return nil, domain.Usage("chat id and message id are required")
	}
	return st.Attachments(ctx, chatID, msgID)
}

func Save(ctx context.Context, st Store, sess domain.Session, chatID, msgID, attach, dest string, overwrite bool, write func(string, []byte, bool) error) (string, error) {
	if err := auth.Require(sess, false, true); err != nil {
		return "", err
	}
	if chatID == "" || msgID == "" || attach == "" {
		return "", domain.Usage("chat id, message id, and attachment are required")
	}
	if dest == "" {
		return "", domain.Usage("destination path is required")
	}
	data, _, err := st.Download(ctx, chatID, msgID, attach)
	if err != nil {
		return "", err
	}
	if err := write(dest, data, overwrite); err != nil {
		return "", err
	}
	return dest, nil
}
