package graph

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/app/teams"
)

func TestHTTPSendUsesFakeServer(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPClient{Base: srv.URL, Token: "fake-both"}
	id, err := c.Send(context.Background(), mail.SendInput{
		To: []string{"a@example.com"}, Subject: "t", Body: "b",
	})
	if err != nil || id != "sent" {
		t.Fatalf("%s %v", id, err)
	}
	tc := &HTTPTeams{HTTPClient: c}
	id, err = tc.Send(context.Background(), teams.SendInput{ChatID: "chat-1", Text: "hi"})
	if err != nil || id != "sent" {
		t.Fatalf("teams %s %v", id, err)
	}
}

func TestTokenFn(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPClient{Base: srv.URL, TokenFn: func() string { return "fake-both" }}
	p, err := c.List(context.Background(), mail.ListQuery{Top: 10})
	if err != nil || p.Count == 0 {
		t.Fatal(err)
	}
}
