package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestHTTPMailListGetThreadViaFakeServer(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPClient{Base: srv.URL, Token: "fake-both"}
	p, err := c.List(context.Background(), mail.ListQuery{Top: 10})
	if err != nil || p.Count == 0 {
		t.Fatalf("list: %+v %v", p, err)
	}
	msg, err := c.Get(context.Background(), "msg-1")
	if err != nil || msg.ID != "msg-1" || msg.Subject == "" {
		t.Fatalf("get: %+v %v", msg, err)
	}
	if len(msg.To) != 1 || msg.To[0].Address != "user@example.com" || len(msg.CC) != 1 || msg.CC[0].Address != "ops@example.com" {
		t.Fatalf("get recipients: %+v", msg)
	}
	if msg.Received != "2026-01-01T00:00:00Z" || !msg.IsRead {
		t.Fatalf("get state: %+v", msg)
	}
	th, err := c.Thread(context.Background(), "msg-1", true)
	if err != nil || th.Count == 0 {
		t.Fatalf("thread: %+v %v", th, err)
	}
	for _, it := range p.Items {
		if it.Body != "" {
			t.Fatal("list must not include body")
		}
	}
	if p.NextPage == nil || strings.Contains(*p.NextPage, "graph.microsoft.com") || strings.Contains(*p.NextPage, "http") {
		t.Fatalf("next_page must be opaque, got %v", p.NextPage)
	}
	_, err = c.Get(context.Background(), "not-a-real-id-000")
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatalf("unknown id: %v", err)
	}
}

func TestHTTPMailListAndThreadAttachmentMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expand := strings.Contains(r.URL.RawQuery, "expand=attachments") || strings.Contains(r.URL.RawQuery, "expand%3Dattachments")
		atts := []any{}
		if expand {
			atts = []any{map[string]any{"id": "att-1", "name": "note.txt", "size": int64(12), "contentType": "text/plain"}}
		}
		item := map[string]any{
			"id": "msg-1", "subject": "Hello", "conversationId": "conv-1",
			"hasAttachments": true, "attachments": atts,
			"from": map[string]any{"emailAddress": map[string]string{"address": "a@example.com"}},
		}
		if strings.HasPrefix(r.URL.Path, "/me/messages/") && !strings.Contains(r.URL.Path, "/me/messages?") {
			_ = json.NewEncoder(w).Encode(item)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"value": []any{item}})
	}))
	defer srv.Close()
	c := &HTTPClient{Base: srv.URL, Token: "fake-both"}
	p, err := c.List(context.Background(), mail.ListQuery{Top: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) == 0 || len(p.Items[0].Attachments) == 0 || p.Items[0].Attachments[0].Name != "note.txt" {
		t.Fatalf("list must include attachment metadata, got %+v", p.Items)
	}
	if p.Items[0].Body != "" {
		t.Fatal("list must not include body")
	}
	th, err := c.Thread(context.Background(), "msg-1", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(th.Items) == 0 || len(th.Items[0].Attachments) == 0 || th.Items[0].Attachments[0].Name != "note.txt" {
		t.Fatalf("thread must include attachment metadata, got %+v", th.Items)
	}
}

func TestMemoryMailList(t *testing.T) {
	m := Seed()
	p, err := m.List(context.Background(), mail.ListQuery{Top: 10})
	if err != nil || p.Count == 0 {
		t.Fatal(err)
	}
}

func TestGraphMailKeepsEmptyRecipientArrays(t *testing.T) {
	var raw graphMsg
	if err := json.Unmarshal([]byte(`{"id":"msg-1","toRecipients":[],"ccRecipients":[]}`), &raw); err != nil {
		t.Fatal(err)
	}
	msg := raw.toMail()
	if msg.To == nil || msg.CC == nil {
		t.Fatalf("empty Graph recipient arrays must remain empty arrays: %+v", msg)
	}
}

func TestHTTPMailThreadFollowsPagesAndSortsOldestFirst(t *testing.T) {
	var srv *httptest.Server
	page := 0
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/me/messages/") {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "root", "conversationId": "conv-1"})
			return
		}
		page++
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(map[string]any{"value": []any{
				map[string]any{"id": "msg-a", "conversationId": "conv-1", "receivedDateTime": "2026-01-01T00:00:00Z", "body": map[string]string{"content": "a"}},
			}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"value": []any{
				map[string]any{"id": "msg-b", "conversationId": "conv-1", "receivedDateTime": "2026-01-01T01:00:00Z", "body": map[string]string{"content": "b"}},
				map[string]any{"id": "msg-c", "conversationId": "conv-1", "receivedDateTime": "2026-01-01T00:00:00Z", "body": map[string]string{"content": "c"}},
			},
			"@odata.nextLink": srv.URL + "/me/messages?page=2",
		})
	}))
	defer srv.Close()

	c := &HTTPClient{Base: srv.URL, Token: "fake-both", Client: srv.Client()}
	thread, err := c.Thread(context.Background(), "root", true)
	if err != nil {
		t.Fatal(err)
	}
	if page != 2 {
		t.Fatalf("thread pages = %d, want 2", page)
	}
	got := make([]string, 0, len(thread.Items))
	for _, msg := range thread.Items {
		got = append(got, msg.ID)
		if msg.Body == "" {
			t.Fatalf("thread body missing: %+v", msg)
		}
	}
	want := []string{"msg-a", "msg-c", "msg-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("thread order = %v, want %v", got, want)
	}
}

func TestMemoryMailThreadSortsOldestFirst(t *testing.T) {
	m := &Memory{Mails: []domain.MailMessage{
		{ID: "msg-b", Conversation: "conv-1", Received: "2026-01-01T01:00:00Z", Body: "b"},
		{ID: "msg-c", Conversation: "conv-1", Received: "2026-01-01T00:00:00Z", Body: "c"},
		{ID: "msg-a", Conversation: "conv-1", Received: "2026-01-01T00:00:00Z", Body: "a"},
	}}
	thread, err := m.Thread(context.Background(), "msg-b", false)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(thread.Items))
	for _, msg := range thread.Items {
		got = append(got, msg.ID)
		if msg.Body != "" {
			t.Fatalf("thread body must be omitted: %+v", msg)
		}
	}
	want := []string{"msg-a", "msg-c", "msg-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("thread order = %v, want %v", got, want)
	}
}
