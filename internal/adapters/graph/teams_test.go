package graph

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestHTTPTeamsListGetMessagesViaFakeServer(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both"}}
	p, err := c.ListChats(context.Background(), 20, "")
	if err != nil || p.Count == 0 {
		t.Fatalf("list: %+v %v", p, err)
	}
	ch, err := c.GetChat(context.Background(), "chat-1")
	if err != nil || ch.ID != "chat-1" {
		t.Fatalf("get: %+v %v", ch, err)
	}
	msgs, err := c.Messages(context.Background(), teams.MessageQuery{ChatID: "chat-1", Top: 20})
	if err != nil || msgs.Count == 0 {
		t.Fatalf("messages: %+v %v", msgs, err)
	}
}

func TestMemoryChats(t *testing.T) {
	m := TeamsAPI{Memory: Seed()}
	p, err := m.ListChats(context.Background(), 20, "")
	if err != nil || p.Count == 0 {
		t.Fatal(err)
	}
	_, err = m.Messages(context.Background(), teams.MessageQuery{ChatID: "chat-1", Top: 20})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPWatch(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both"}}
	ev, err := c.Watch(context.Background(), teams.WatchQuery{Top: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(ev) == 0 {
		t.Fatal("expected watch events from 1:1 seed chat")
	}
	_, err = c.Watch(context.Background(), teams.WatchQuery{Chats: []string{"missing-chat"}, Top: 5})
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatalf("unknown chat: %v", err)
	}
}

func TestHTTPWatchMentionReason(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both"}}
	ev, err := c.Watch(context.Background(), teams.WatchQuery{Top: 20})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range ev {
		if e.Reason == "mention" && e.ChatID == "chat-2" && e.MessageID == "cmsg-2" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected mention event from group chat, got %+v", ev)
	}
}
