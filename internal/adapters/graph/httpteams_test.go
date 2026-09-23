package graph

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/teams"
)

const graphChatMessages = `{"value":[
{"id":"m-user","createdDateTime":"2026-01-01T00:00:00Z","messageType":"message",
 "body":{"content":"hi"},"from":{"user":{"id":"u-1","displayName":"Alice","userIdentityType":"aadUser"}}},
{"id":"m-app","createdDateTime":"2026-01-01T00:01:00Z","messageType":"message",
 "body":{"content":"alert"},"from":{"application":{"id":"a-1","displayName":"Datadog"}}},
{"id":"m-sys","createdDateTime":"2026-01-01T00:02:00Z","messageType":"unknownFutureValue",
 "body":{"content":"<systemEventMessage/>"},"from":null}
]}`

func TestHTTPTeamsMessagesSender(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(graphChatMessages))
	}))
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both", Client: srv.Client()}}

	p, err := c.Messages(context.Background(), teams.MessageQuery{ChatID: "chat-1", Top: 20})
	if err != nil {
		t.Fatal(err)
	}
	if p.Count != 2 {
		t.Fatalf("system message not filtered: %+v", p.Items)
	}
	if p.Items[0].From != "Alice" || p.Items[1].From != "Datadog" {
		t.Fatalf("senders: %+v", p.Items)
	}

	p, err = c.Messages(context.Background(), teams.MessageQuery{ChatID: "chat-1", Top: 20, IncludeSystem: true})
	if err != nil {
		t.Fatal(err)
	}
	if p.Count != 3 || !p.Items[2].System || p.Items[2].From != "" {
		t.Fatalf("include system: %+v", p.Items)
	}
}

func TestHTTPTeamsMessagesPage(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("$skiptoken") == "two" {
			_, _ = w.Write([]byte(`{"value":[{"id":"m-2","messageType":"message","from":{"user":{"displayName":"Bob"}}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"value":[{"id":"m-1","messageType":"message","from":{"user":{"displayName":"Alice"}}}],
"@odata.nextLink":"` + srv.URL + `/chats/chat-1/messages?$top=1&$skiptoken=two"}`))
	}))
	defer srv.Close()
	c := &HTTPTeams{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-both", Client: srv.Client()}}

	p, err := c.Messages(context.Background(), teams.MessageQuery{ChatID: "chat-1", Top: 1})
	if err != nil || p.NextPage == nil || p.Items[0].From != "Alice" {
		t.Fatalf("page 1: %+v %v", p, err)
	}
	p, err = c.Messages(context.Background(), teams.MessageQuery{ChatID: "chat-1", Top: 1, PageToken: *p.NextPage})
	if err != nil || p.NextPage != nil || p.Items[0].From != "Bob" {
		t.Fatalf("page 2: %+v %v", p, err)
	}
}
