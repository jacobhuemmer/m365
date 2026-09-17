package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
)

type HTTPTeams struct{ *HTTPClient }

func (c *HTTPTeams) ListChats(ctx context.Context, top int, page string) (domain.ChatPage, error) {
	path := "/me/chats?$expand=members&$top=" + strconv.Itoa(top)
	if page != "" {
		path += "&$skiptoken=" + url.QueryEscape(page)
	}
	res, err := c.do(ctx, http.MethodGet, path)
	if err != nil {
		return domain.ChatPage{}, err
	}
	defer res.Body.Close()
	var raw graphChatList
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return domain.ChatPage{}, domain.Service("invalid chat list")
	}
	items := make([]domain.Chat, 0, len(raw.Value))
	for _, g := range raw.Value {
		items = append(items, mapGraphChat(g))
	}
	p := domain.ChatPage{Limit: top, Count: len(items), Items: items}
	if raw.Next != "" {
		n := raw.Next
		p.NextPage = &n
	}
	return p, nil
}

func (c *HTTPTeams) GetChat(ctx context.Context, id string) (domain.Chat, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/chats/"+url.PathEscape(id))
	if err != nil {
		return domain.Chat{}, err
	}
	defer res.Body.Close()
	var g graphChat
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return domain.Chat{}, domain.Service("invalid chat")
	}
	return mapGraphChat(g), nil
}

func mapGraphChat(g graphChat) domain.Chat {
	c := domain.Chat{ID: g.ID, Topic: g.Topic, Type: g.ChatType}
	for _, m := range g.Members {
		c.Members = append(c.Members, domain.Person{Name: m.DisplayName, Address: m.Email})
	}
	return c
}

func (c *HTTPTeams) Messages(ctx context.Context, q teams.MessageQuery) (domain.ChatMessagePage, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/chats/"+url.PathEscape(q.ChatID)+"/messages?$top="+strconv.Itoa(q.Top))
	if err != nil {
		return domain.ChatMessagePage{}, err
	}
	defer res.Body.Close()
	var raw graphMsgList
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return domain.ChatMessagePage{}, domain.Service("invalid messages")
	}
	items := make([]domain.ChatMessage, 0, len(raw.Value))
	for _, g := range raw.Value {
		items = append(items, domain.ChatMessage{ID: g.ID, ChatID: q.ChatID, Text: g.Body.Content, Created: g.Created})
	}
	return domain.ChatMessagePage{Limit: q.Top, Count: len(items), Items: items}, nil
}

func (c *HTTPTeams) Attachments(context.Context, string, string) ([]domain.Attachment, error) {
	return nil, domain.Service("not wired")
}
func (c *HTTPTeams) Download(context.Context, string, string, string) ([]byte, domain.Attachment, error) {
	return nil, domain.Attachment{}, domain.Service("not wired")
}

type graphChatList struct {
	Value []graphChat `json:"value"`
	Next  string      `json:"@odata.nextLink"`
}
type graphChat struct {
	ID       string        `json:"id"`
	Topic    string        `json:"topic"`
	ChatType string        `json:"chatType"`
	Members  []graphMember `json:"members"`
}
type graphMember struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}
type graphMsgList struct {
	Value []graphChatMsg `json:"value"`
}
type graphChatMsg struct {
	ID      string `json:"id"`
	Created string `json:"createdDateTime"`
	Body    struct {
		Content string `json:"content"`
	} `json:"body"`
}
