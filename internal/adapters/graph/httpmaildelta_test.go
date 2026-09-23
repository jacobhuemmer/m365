package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestHTTPMailDeltaMapsPagesAndEscapesFolder(t *testing.T) {
	var server *httptest.Server
	requests := 0
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		if requests == 1 {
			if !strings.Contains(r.URL.EscapedPath(), "/mailFolders/folder%2Fid/messages/delta") {
				t.Fatalf("delta path = %s", r.URL.EscapedPath())
			}
			selectValue := r.URL.Query().Get("$select")
			for _, field := range []string{"id", "changeKey", "conversationId", "receivedDateTime"} {
				if !strings.Contains(selectValue, field) {
					t.Fatalf("$select missing %s: %s", field, selectValue)
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"value": []any{map[string]any{
					"id": "msg-1", "changeKey": "rev-1", "conversationId": "conv-1",
					"subject": "Approval", "receivedDateTime": "2026-01-01T00:00:00Z",
					"from": map[string]any{"emailAddress": map[string]string{"name": "Alice", "address": "alice@example.com"}},
				}},
				"@odata.nextLink": server.URL + "/delta-next",
			})
			return
		}
		if r.URL.Path != "/delta-next" {
			t.Fatalf("continuation path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"value": []any{
				map[string]any{"id": "msg-deleted", "@removed": map[string]string{"reason": "deleted"}},
				map[string]any{"id": "msg-2", "changeKey": "rev-2", "conversationId": "conv-2", "receivedDateTime": "2026-01-01T01:00:00Z"},
			},
			"@odata.deltaLink": server.URL + "/delta-cursor",
		})
	}))
	defer server.Close()

	adapter := &HTTPMailDelta{HTTPClient: &HTTPClient{Base: server.URL, Token: "fake", Client: server.Client()}}
	first, err := adapter.Delta(context.Background(), mail.DeltaQuery{Folder: "folder/id"})
	if err != nil {
		t.Fatal(err)
	}
	if first.NextToken != server.URL+"/delta-next" || first.DeltaToken != "" || len(first.Changes) != 1 {
		t.Fatalf("first page = %+v", first)
	}
	change := first.Changes[0]
	if change.Revision != "rev-1" || change.Message.ID != "msg-1" || change.Message.From.Address != "alice@example.com" || change.Message.Body != "" {
		t.Fatalf("mapped change = %+v", change)
	}

	last, err := adapter.Delta(context.Background(), mail.DeltaQuery{Folder: "folder/id", Token: first.NextToken})
	if err != nil {
		t.Fatal(err)
	}
	if last.DeltaToken != server.URL+"/delta-cursor" || last.NextToken != "" || len(last.Changes) != 2 {
		t.Fatalf("last page = %+v", last)
	}
	if !last.Changes[0].Removed || last.Changes[0].Message.ID != "msg-deleted" {
		t.Fatalf("removed change = %+v", last.Changes[0])
	}
}

func TestHTTPMailDeltaMapsGoneToResetSignal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte(`{"error":{"code":"syncStateNotFound"}}`))
	}))
	defer server.Close()

	adapter := &HTTPMailDelta{HTTPClient: &HTTPClient{Base: server.URL, Token: "fake", Client: server.Client()}}
	_, err := adapter.Delta(context.Background(), mail.DeltaQuery{Folder: "inbox", Token: server.URL + "/expired"})
	if !errors.Is(err, domain.ErrMailDeltaReset) {
		t.Fatalf("gone error = %v", err)
	}
}

func TestHTTPMailDeltaSkipsPartialUpdateWithoutRevision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"value": []any{
				map[string]any{"@odata.type": "#microsoft.graph.message", "id": "msg-read", "isRead": true},
				map[string]any{"id": "msg-2", "changeKey": "rev-2", "conversationId": "conv-2", "receivedDateTime": "2026-01-01T01:00:00Z"},
			},
			"@odata.deltaLink": "https://graph.microsoft.com/v1.0/cursor",
		})
	}))
	defer server.Close()

	adapter := &HTTPMailDelta{HTTPClient: &HTTPClient{Base: server.URL, Token: "fake", Client: server.Client()}}
	page, err := adapter.Delta(context.Background(), mail.DeltaQuery{Folder: "inbox"})
	if err != nil {
		t.Fatal(err)
	}
	if page.DeltaToken != "https://graph.microsoft.com/v1.0/cursor" || len(page.Changes) != 1 || page.Changes[0].Message.ID != "msg-2" {
		t.Fatalf("partial update page = %+v", page)
	}
}

func TestMemoryMailDeltaReturnsStableRoundAndThenStaysQuiet(t *testing.T) {
	memory := Seed()
	adapter := MailDeltaAPI{Memory: memory}

	first, err := adapter.Delta(context.Background(), mail.DeltaQuery{Folder: "inbox"})
	if err != nil || len(first.Changes) != len(memory.Mails) || first.DeltaToken == "" {
		t.Fatalf("first memory delta = %+v, %v", first, err)
	}
	for _, change := range first.Changes {
		if change.Revision == "" || change.Message.Body != "" || change.Message.Attachments != nil {
			t.Fatalf("unsafe memory delta change = %+v", change)
		}
	}
	second, err := adapter.Delta(context.Background(), mail.DeltaQuery{Folder: "inbox", Token: first.DeltaToken})
	if err != nil || len(second.Changes) != 0 || second.DeltaToken != first.DeltaToken {
		t.Fatalf("second memory delta = %+v, %v", second, err)
	}
}

func TestHTTPMailDeltaLatestThreadKeepsLatestWindowOldestFirst(t *testing.T) {
	messages := make([]any, 0, 12)
	for i := 11; i >= 0; i-- {
		messages = append(messages, map[string]any{
			"id":               fmt.Sprintf("msg-%02d", i),
			"conversationId":   "conv-1",
			"receivedDateTime": fmt.Sprintf("2026-01-01T%02d:00:00Z", i),
			"body":             map[string]string{"content": fmt.Sprintf("body-%02d", i)},
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/me/messages/") {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "changed", "conversationId": "conv-1"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"value": messages})
	}))
	defer server.Close()

	adapter := &HTTPMailDelta{HTTPClient: &HTTPClient{Base: server.URL, Token: "fake", Client: server.Client()}}
	thread, err := adapter.LatestThread(context.Background(), "changed", 10)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(thread.Items))
	for _, message := range thread.Items {
		got = append(got, message.ID)
		if message.Body == "" {
			t.Fatalf("latest thread body missing: %+v", message)
		}
	}
	want := []string{"msg-02", "msg-03", "msg-04", "msg-05", "msg-06", "msg-07", "msg-08", "msg-09", "msg-10", "msg-11"}
	if !reflect.DeepEqual(got, want) || thread.Count != 12 {
		t.Fatalf("latest thread = count %d ids %v", thread.Count, got)
	}
}

func TestMemoryMailDeltaLatestThreadMatchesHTTPWindow(t *testing.T) {
	memory := &Memory{}
	for i := 11; i >= 0; i-- {
		memory.Mails = append(memory.Mails, domain.MailMessage{
			ID: fmt.Sprintf("msg-%02d", i), Conversation: "conv-1",
			Received: fmt.Sprintf("2026-01-01T%02d:00:00Z", i), Body: fmt.Sprintf("body-%02d", i),
		})
	}
	adapter := MailDeltaAPI{Memory: memory}
	thread, err := adapter.LatestThread(context.Background(), "msg-11", 10)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(thread.Items))
	for _, message := range thread.Items {
		got = append(got, message.ID)
	}
	want := []string{"msg-02", "msg-03", "msg-04", "msg-05", "msg-06", "msg-07", "msg-08", "msg-09", "msg-10", "msg-11"}
	if !reflect.DeepEqual(got, want) || thread.Count != 12 {
		t.Fatalf("latest memory thread = count %d ids %v", thread.Count, got)
	}
}
