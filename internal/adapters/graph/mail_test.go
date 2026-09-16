package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
