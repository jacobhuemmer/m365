package graph

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/domain"
	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

func TestReplyAttachmentsWire(t *testing.T) {
	for _, all := range []bool{false, true} {
		endpoint := "reply"
		if all {
			endpoint = "replyAll"
		}
		for _, mode := range []string{"files", "no-files", "later-unreadable", "graph-error"} {
			t.Run(endpoint+"/"+mode, func(t *testing.T) {
				dir := t.TempDir()
				files := []domain.OutboundFile{{Path: filepath.Join(dir, "first.txt"), Name: "first.txt"}, {Path: filepath.Join(dir, "second.bin"), Name: "second.bin"}}
				for i, data := range [][]byte{[]byte("synthetic attachment"), {0, 255, 10, 42}} {
					if err := os.WriteFile(files[i].Path, data, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "no-files" {
					files = nil
				}
				if mode == "later-unreadable" {
					files[1].Path = filepath.Join(dir, "missing.bin")
				}
				requests := 0
				var payload map[string]any
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++
					if r.Method != http.MethodPost || r.URL.Path != "/me/messages/msg-1/"+endpoint {
						t.Errorf("request %s %s", r.Method, r.URL.Path)
					}
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					if mode == "graph-error" {
						w.WriteHeader(http.StatusBadRequest)
					} else {
						w.WriteHeader(http.StatusAccepted)
					}
				}))
				defer srv.Close()
				c := &HTTPClient{Base: srv.URL, Token: "synthetic-token"}
				id, err := c.Reply(context.Background(), mail.ReplyInput{ID: "msg-1", All: all, Files: files, Rendered: msgbody.Rendered{Content: "<p>Authored reply.</p>"}})
				if mode == "later-unreadable" {
					if domain.ClassOf(err) != domain.ClassUsage || id != "" || requests != 0 {
						t.Fatalf("unreadable: id=%q err=%v requests=%d", id, err, requests)
					}
					return
				}
				if mode == "graph-error" {
					if domain.ClassOf(err) != domain.ClassService || id != "" {
						t.Fatalf("rejection: id=%q err=%v", id, err)
					}
				} else if err != nil || id != "sent" {
					t.Fatalf("id=%q err=%v", id, err)
				}
				if requests != 1 {
					t.Fatalf("requests=%d want one send and no fallback", requests)
				}
				if payload["comment"] != "<p>Authored reply.</p>" {
					t.Fatalf("comment: %v", payload)
				}
				if mode == "no-files" {
					if !reflect.DeepEqual(payload, map[string]any{"comment": "<p>Authored reply.</p>"}) {
						t.Fatalf("no-file payload=%v", payload)
					}
					return
				}
				message, _ := payload["message"].(map[string]any)
				atts, _ := message["attachments"].([]any)
				if len(payload) != 2 || len(message) != 1 || len(atts) != 2 {
					t.Fatalf("want comment and message.attachments only, got %v", payload)
				}
				for i, want := range []string{"synthetic attachment", string([]byte{0, 255, 10, 42})} {
					att := atts[i].(map[string]any)
					encoded, _ := att["contentBytes"].(string)
					b, e := base64.StdEncoding.DecodeString(encoded)
					if e != nil || string(b) != want || att["name"] != files[i].Name || att["@odata.type"] != "#microsoft.graph.fileAttachment" {
						t.Fatalf("attachment %d: %v bytes=%v err=%v", i, att, b, e)
					}
				}
			})
		}
	}
}
