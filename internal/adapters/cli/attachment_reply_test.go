package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestReplyAttachmentCLIAndMCP(t *testing.T) {
	for _, seam := range []string{"cli", "mcp"} {
		for _, all := range []bool{false, true} {
			for _, mode := range []string{"send", "dry-run", "missing", "service"} {
				t.Run(seam+"/"+map[bool]string{false: "reply", true: "replyAll"}[all]+"/"+mode, func(t *testing.T) {
					p := filepath.Join(t.TempDir(), "note.txt")
					if err := os.WriteFile(p, []byte("synthetic attachment"), 0o600); err != nil {
						t.Fatal(err)
					}
					if mode == "missing" {
						p += ".missing"
					}
					requests := 0
					var payload map[string]any
					srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests++
						endpoint := "reply"
						if all {
							endpoint = "replyAll"
						}
						if r.Method != "POST" || r.URL.Path != "/me/messages/msg-1/"+endpoint {
							t.Errorf("%s %s", r.Method, r.URL.Path)
						}
						if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
							t.Error(err)
						}
						if mode == "service" {
							w.WriteHeader(400)
						} else {
							w.WriteHeader(202)
						}
					}))
					defer srv.Close()
					d, out, errw := testDeps()
					login(t, d)
					out.Reset()
					errw.Reset()
					d.Mail = &graph.HTTPClient{Base: srv.URL, Token: "synthetic-token"}
					code, result, diagnostic := 0, "", ""
					if seam == "cli" {
						args := []string{"m365", "mail", "reply", "msg-1", "--body", "Authored reply.", "--attach", p}
						if all {
							args = append(args, "--all")
						}
						if mode == "dry-run" {
							args = append(args, "--dry-run")
						}
						code = Run(args, d)
						result = out.String()
						diagnostic = errw.String()
					} else {
						cs := connectMCP(t, d)
						res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_run", Arguments: runIn{Namespace: "mail", Verb: "reply", Args: []string{"msg-1"}, WriteOptIn: mode != "dry-run", Flags: map[string]any{"body": "Authored reply.", "attach": []string{p}, "all": all}}})
						if err != nil {
							t.Fatal(err)
						}
						if res.IsError {
							diagnostic = toolText(t, res)
							code = domain.ExitService
							if mode == "missing" {
								code = domain.ExitUsage
							}
						} else {
							result = toolText(t, res)
						}
					}
					if mode == "missing" || mode == "service" {
						want, count := domain.ExitUsage, 0
						class := "usage"
						if mode == "service" {
							want, count, class = domain.ExitService, 1, "service"
						}
						var failure map[string]any
						if code != want || result != "" || requests != count || json.Unmarshal([]byte(diagnostic), &failure) != nil || failure["class"] != class {
							t.Fatalf("code=%d result=%s diagnostic=%s requests=%d", code, result, diagnostic, requests)
						}
						return
					}
					if code != 0 || diagnostic != "" {
						t.Fatalf("code=%d diagnostic=%s result=%s", code, diagnostic, result)
					}
					var success map[string]any
					if json.Unmarshal([]byte(result), &success) != nil {
						t.Fatal(result)
					}
					if mode == "dry-run" {
						if requests != 0 || success["dry_run"] != true || !strings.Contains(result, "note.txt") || !strings.Contains(result, `"size":`) || strings.Contains(result, "synthetic attachment") || strings.Contains(result, "c3ludGhldGlj") || strings.Contains(result, "contentBytes") {
							t.Fatalf("preview=%s requests=%d", result, requests)
						}
						return
					}
					msg, _ := payload["message"].(map[string]any)
					atts, _ := msg["attachments"].([]any)
					if requests != 1 || payload["comment"] != "<p>Authored reply.</p>" || len(payload) != 2 || len(msg) != 1 || len(atts) != 1 {
						t.Fatalf("payload=%v requests=%d", payload, requests)
					}
					att := atts[0].(map[string]any)
					if att["name"] != "note.txt" || att["contentBytes"] != "c3ludGhldGljIGF0dGFjaG1lbnQ=" || att["@odata.type"] != "#microsoft.graph.fileAttachment" || success["id"] != "sent" {
						t.Fatalf("payload=%v success=%v", payload, success)
					}
				})
			}
		}
	}
}
