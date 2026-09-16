package graph

import (
	"context"
	"strings"
	"sync"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
)

type Memory struct {
	mu       sync.Mutex
	Throttle bool
	Mails    []domain.MailMessage
	Chats    []domain.Chat
	Msgs     map[string][]domain.ChatMessage
	Sent     []any
	Events   []domain.WatchEvent
	Bytes    map[string][]byte
}

func Seed() *Memory {
	att := domain.Attachment{ID: "att-1", Name: "note.txt", Size: 12, ContentType: "text/plain"}
	m1 := domain.MailMessage{
		ID: "msg-1", Conversation: "conv-1", Subject: "Hello",
		From: domain.Person{Address: "a@example.com"}, Received: "2026-01-01T00:00:00Z",
		HasAttachments: true, Body: "body-1", Attachments: []domain.Attachment{att},
	}
	m2 := domain.MailMessage{
		ID: "msg-2", Conversation: "conv-1", Subject: "Re: Hello",
		From: domain.Person{Address: "user@example.com"}, Received: "2026-01-01T01:00:00Z",
		Body: "own-reply",
	}
	c1 := domain.Chat{ID: "chat-1", Type: "oneOnOne", Topic: "Alice"}
	cm := domain.ChatMessage{
		ID: "cmsg-1", ChatID: "chat-1", From: "Alice", Created: "2026-01-01T00:00:00Z",
		Text: "hi", Attachments: []domain.Attachment{att},
	}
	c2m := domain.ChatMessage{
		ID: "cmsg-2", ChatID: "chat-2", From: "Bob", Created: "2026-01-01T02:00:00Z",
		Text: "please look",
	}
	return &Memory{
		Mails: []domain.MailMessage{m1, m2},
		Chats: []domain.Chat{c1, {ID: "chat-2", Type: "group", Topic: "NOC"}},
		Msgs:  map[string][]domain.ChatMessage{"chat-1": {cm}, "chat-2": {c2m}},
		Bytes: map[string][]byte{"att-1": []byte("synthetic-ok")},
		Events: []domain.WatchEvent{{
			ChatID: "chat-1", MessageID: "cmsg-1", From: "Alice", Text: "hi",
			Created: "2026-01-01T00:00:00Z", Reason: "one_to_one",
		}},
	}
}

func (m *Memory) fail() error {
	if m.Throttle {
		return domain.Service("throttled")
	}
	return nil
}

func (m *Memory) List(_ context.Context, q mail.ListQuery) (domain.MailPage, error) {
	if err := m.fail(); err != nil {
		return domain.MailPage{}, err
	}
	var items []domain.MailMessage
	for _, msg := range m.Mails {
		if q.Unread && msg.IsRead {
			continue
		}
		if q.Search != "" && !strings.Contains(strings.ToLower(msg.Subject+" "+msg.Body), strings.ToLower(q.Search)) {
			continue
		}
		cp := msg
		cp.Body = ""
		items = append(items, cp)
	}
	return pageMail(items, q.Top), nil
}

func pageMail(items []domain.MailMessage, top int) domain.MailPage {
	if top <= 0 {
		top = domain.DefaultMailTop
	}
	p := domain.MailPage{Limit: top, Items: []domain.MailMessage{}}
	if len(items) > top {
		p.Items = items[:top]
		n := "next"
		p.NextPage = &n
	} else {
		p.Items = items
	}
	p.Count = len(p.Items)
	return p
}

func (m *Memory) Get(_ context.Context, id string) (domain.MailMessage, error) {
	if err := m.fail(); err != nil {
		return domain.MailMessage{}, err
	}
	for _, msg := range m.Mails {
		if msg.ID == id {
			return msg, nil
		}
	}
	return domain.MailMessage{}, domain.NotFound("message not found")
}

func (m *Memory) Thread(_ context.Context, id string, bodies bool) (domain.MailThread, error) {
	msg, err := m.Get(context.Background(), id)
	if err != nil {
		return domain.MailThread{}, err
	}
	var items []domain.MailMessage
	for _, x := range m.Mails {
		if x.Conversation == msg.Conversation {
			cp := x
			if !bodies {
				cp.Body = ""
			}
			items = append(items, cp)
		}
	}
	return domain.MailThread{ConversationID: msg.Conversation, Count: len(items), Items: items}, nil
}

func (m *Memory) Send(_ context.Context, in mail.SendInput) (string, error) {
	if err := m.fail(); err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	id := "sent-1"
	m.Sent = append(m.Sent, in)
	return id, nil
}

func (m *Memory) Reply(_ context.Context, in mail.ReplyInput) (string, error) {
	if _, err := m.Get(context.Background(), in.ID); err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Sent = append(m.Sent, in)
	return "reply-1", nil
}

func (m *Memory) Attachments(_ context.Context, messageID string) ([]domain.Attachment, error) {
	msg, err := m.Get(context.Background(), messageID)
	if err != nil {
		return nil, err
	}
	if msg.Attachments == nil {
		return []domain.Attachment{}, nil
	}
	return msg.Attachments, nil
}

func (m *Memory) Download(_ context.Context, messageID, attach string) ([]byte, domain.Attachment, error) {
	atts, err := m.Attachments(context.Background(), messageID)
	if err != nil {
		return nil, domain.Attachment{}, err
	}
	var hits []domain.Attachment
	for _, a := range atts {
		if a.ID == attach || a.Name == attach {
			hits = append(hits, a)
		}
	}
	if len(hits) == 0 {
		return nil, domain.Attachment{}, domain.NotFound("attachment not found")
	}
	if len(hits) > 1 {
		return nil, domain.Attachment{}, domain.Usage("ambiguous attachment name; use id")
	}
	return m.Bytes[hits[0].ID], hits[0], nil
}

func (m *Memory) ListChats(_ context.Context, top int, _ string) (domain.ChatPage, error) {
	if err := m.fail(); err != nil {
		return domain.ChatPage{}, err
	}
	items := m.Chats
	p := domain.ChatPage{Limit: top, Items: items}
	if len(items) > top {
		p.Items = items[:top]
		n := "next"
		p.NextPage = &n
	}
	p.Count = len(p.Items)
	return p, nil
}

func (m *Memory) GetChat(_ context.Context, id string) (domain.Chat, error) {
	for _, c := range m.Chats {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.Chat{}, domain.NotFound("chat not found")
}

func (m *Memory) Messages(_ context.Context, q teams.MessageQuery) (domain.ChatMessagePage, error) {
	if _, err := m.GetChat(context.Background(), q.ChatID); err != nil {
		return domain.ChatMessagePage{}, err
	}
	items := m.Msgs[q.ChatID]
	if items == nil {
		items = []domain.ChatMessage{}
	}
	var out []domain.ChatMessage
	for _, x := range items {
		if x.System && !q.IncludeSystem {
			continue
		}
		out = append(out, x)
	}
	p := domain.ChatMessagePage{Limit: q.Top, Items: out, Count: len(out)}
	return p, nil
}

func (m *Memory) SendChat(_ context.Context, in teams.SendInput) (string, error) {
	if _, err := m.GetChat(context.Background(), in.ChatID); err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Sent = append(m.Sent, in)
	return "cmsg-new", nil
}

func (m *Memory) Watch(_ context.Context, q teams.WatchQuery) ([]domain.WatchEvent, error) {
	if err := m.fail(); err != nil {
		return nil, err
	}
	for _, id := range q.Chats {
		if _, err := m.GetChat(context.Background(), id); err != nil {
			return nil, err
		}
	}
	ev := m.Events
	if q.Top > 0 && len(ev) > q.Top {
		ev = ev[:q.Top]
	}
	return ev, nil
}

func (m *Memory) TeamAttachments(_ context.Context, chatID, messageID string) ([]domain.Attachment, error) {
	if _, err := m.GetChat(context.Background(), chatID); err != nil {
		return nil, err
	}
	for _, msg := range m.Msgs[chatID] {
		if msg.ID == messageID {
			if msg.Attachments == nil {
				return []domain.Attachment{}, nil
			}
			return msg.Attachments, nil
		}
	}
	return nil, domain.NotFound("message not found")
}

func (m *Memory) AttachmentsChat(ctx context.Context, chatID, messageID string) ([]domain.Attachment, error) {
	return m.TeamAttachments(ctx, chatID, messageID)
}

func (m *Memory) DownloadChat(_ context.Context, chatID, messageID, attach string) ([]byte, domain.Attachment, error) {
	atts, err := m.TeamAttachments(context.Background(), chatID, messageID)
	if err != nil {
		return nil, domain.Attachment{}, err
	}
	for _, a := range atts {
		if a.ID == attach || a.Name == attach {
			return m.Bytes[a.ID], a, nil
		}
	}
	return nil, domain.Attachment{}, domain.NotFound("attachment not found")
}

func (m *Memory) DownloadTeam(ctx context.Context, chatID, messageID, attach string) ([]byte, domain.Attachment, error) {
	return m.DownloadChat(ctx, chatID, messageID, attach)
}

type MailAPI struct{ *Memory }
type TeamsAPI struct{ *Memory }

func (t TeamsAPI) Send(ctx context.Context, in teams.SendInput) (string, error) {
	return t.SendChat(ctx, in)
}

func (t TeamsAPI) Attachments(ctx context.Context, chatID, messageID string) ([]domain.Attachment, error) {
	return t.TeamAttachments(ctx, chatID, messageID)
}

func (t TeamsAPI) Download(ctx context.Context, chatID, messageID, attach string) ([]byte, domain.Attachment, error) {
	return t.DownloadChat(ctx, chatID, messageID, attach)
}
