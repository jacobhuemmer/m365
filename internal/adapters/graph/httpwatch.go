package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
)

func (c *HTTPTeams) Watch(ctx context.Context, q teams.WatchQuery) ([]domain.WatchEvent, error) {
	top := q.Top
	if top <= 0 {
		top = domain.DefaultTeamsTop
	}
	me, err := c.signedInID(ctx)
	if err != nil {
		return nil, err
	}
	page, err := c.ListChats(ctx, 50, "")
	if err != nil {
		return nil, err
	}
	chats := map[string]domain.Chat{}
	for _, ch := range page.Items {
		chats[ch.ID] = ch
	}
	forced := map[string]bool{}
	for _, id := range q.Chats {
		if _, ok := chats[id]; !ok {
			ch, err := c.GetChat(ctx, id)
			if err != nil {
				return nil, err
			}
			chats[id] = ch
		}
		forced[id] = true
	}
	var ev []domain.WatchEvent
	for id, ch := range chats {
		reasonAll := ""
		t := strings.ToLower(ch.Type)
		if t == "oneonone" || t == "one_on_one" {
			reasonAll = "one_to_one"
		} else if forced[id] {
			reasonAll = "watched_chat"
		}
		msgs, err := c.watchMessages(ctx, id, top)
		if err != nil {
			return nil, err
		}
		for _, m := range msgs {
			if q.Since != "" && m.Created < q.Since {
				continue
			}
			reason := reasonAll
			if reason == "" && mentionsUser(m, me) {
				reason = "mention"
			}
			if reason == "" {
				continue
			}
			ev = append(ev, domain.WatchEvent{
				ChatID: id, MessageID: m.ID, From: m.FromName, Text: m.Text,
				Created: m.Created, Reason: reason, Topic: ch.Topic, ChatType: ch.Type,
			})
		}
	}
	sort.Slice(ev, func(i, j int) bool { return ev[i].Created > ev[j].Created })
	if len(ev) > top {
		ev = ev[:top]
	}
	return ev, nil
}

func (c *HTTPTeams) signedInID(ctx context.Context) (string, error) {
	res, err := c.do(ctx, http.MethodGet, "/me")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var me struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&me); err != nil || me.ID == "" {
		return "", domain.Service("invalid profile")
	}
	return me.ID, nil
}

type watchMsg struct {
	ID         string
	Created    string
	Text       string
	FromName   string
	MentionIDs []string
}

func (c *HTTPTeams) watchMessages(ctx context.Context, chatID string, top int) ([]watchMsg, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/chats/"+url.PathEscape(chatID)+"/messages?$top="+strconv.Itoa(top))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var raw struct {
		Value []struct {
			ID      string `json:"id"`
			Created string `json:"createdDateTime"`
			Body    struct {
				Content string `json:"content"`
			} `json:"body"`
			From struct {
				User struct {
					DisplayName string `json:"displayName"`
				} `json:"user"`
			} `json:"from"`
			Mentions []struct {
				Mentioned struct {
					User struct {
						ID string `json:"id"`
					} `json:"user"`
				} `json:"mentioned"`
			} `json:"mentions"`
		} `json:"value"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, domain.Service("invalid messages")
	}
	out := make([]watchMsg, 0, len(raw.Value))
	for _, g := range raw.Value {
		ids := make([]string, 0, len(g.Mentions))
		for _, n := range g.Mentions {
			if n.Mentioned.User.ID != "" {
				ids = append(ids, n.Mentioned.User.ID)
			}
		}
		out = append(out, watchMsg{
			ID: g.ID, Created: g.Created, Text: g.Body.Content,
			FromName: g.From.User.DisplayName, MentionIDs: ids,
		})
	}
	return out, nil
}

func mentionsUser(m watchMsg, me string) bool {
	for _, id := range m.MentionIDs {
		if id == me {
			return true
		}
	}
	return false
}
