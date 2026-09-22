package graph

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestHTTPMailDeltaRejectsMissingRevision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"value":            []any{map[string]any{"id": "msg-1"}},
			"@odata.deltaLink": "https://graph.microsoft.com/v1.0/cursor",
		})
	}))
	defer server.Close()

	adapter := &HTTPMailDelta{HTTPClient: &HTTPClient{Base: server.URL, Token: "fake", Client: server.Client()}}
	_, err := adapter.Delta(context.Background(), mail.DeltaQuery{Folder: "inbox"})
	if domain.ExitOf(err) != domain.ExitService {
		t.Fatalf("missing revision error = %v", err)
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
