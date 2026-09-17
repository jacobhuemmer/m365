package graph

import (
	"context"
	"testing"
)

func TestListChatsExpandMembersAndPage(t *testing.T) {
	mem := Seed()
	srv := NewFakeServer(mem)
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both", Client: srv.Client()}}
	p, err := c.ListChats(context.Background(), 20, "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Count != 20 || p.NextPage == nil {
		t.Fatalf("%+v", p)
	}
	foundAjay := false
	for _, ch := range p.Items {
		if ch.ID == "chat-ajay" {
			foundAjay = true
		}
	}
	if foundAjay {
		t.Fatal("ajay on first page")
	}
	p2, err := c.ListChats(context.Background(), 20, *p.NextPage)
	if err != nil {
		t.Fatal(err)
	}
	ok := false
	for _, ch := range p2.Items {
		if ch.ID == "chat-ajay" && len(ch.Members) > 0 {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("%+v", p2)
	}
}
