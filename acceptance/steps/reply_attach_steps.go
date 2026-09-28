package steps

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/masonhuemmer/m365/acceptance/runtime"
	"github.com/masonhuemmer/m365/internal/adapters/cli"
	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/adapters/keychain"
	"github.com/masonhuemmer/m365/internal/config"
	"github.com/masonhuemmer/m365/internal/domain"
)

func init() {
	runtime.Register("I exercise reply attachments", func(w *runtime.World, text string) error {
		p := filepath.Join(w.T.TempDir(), "note.txt")
		if err := os.WriteFile(p, []byte("synthetic attachment"), 0o600); err != nil {
			return err
		}
		requests := 0
		var payload map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			requests++
			endpoint := "reply"
			if strings.Contains(text, "all") {
				endpoint = "replyAll"
			}
			if r.Method != "POST" || r.URL.Path != "/me/messages/msg-1/"+endpoint {
				w.T.Errorf("request %s %s", r.Method, r.URL.Path)
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.T.Error(err)
			}
			if strings.Contains(text, "rejected") {
				rw.WriteHeader(400)
			} else {
				rw.WriteHeader(202)
			}
		}))
		defer srv.Close()
		out, errw := &bytes.Buffer{}, &bytes.Buffer{}
		d := cli.Deps{Config: config.Config{ClientID: "x", TenantID: "y"}, Store: &keychain.Fake{}, Login: graph.FakeLogin(true, true), Mail: &graph.HTTPClient{Base: srv.URL, Token: "synthetic-token"}, Stdout: out, Stderr: errw}
		if code := cli.Run([]string{"m365", "auth", "login"}, d); code != 0 {
			w.T.Fatalf("login=%d %s", code, errw.String())
		}
		out.Reset()
		errw.Reset()
		args := []string{"m365", "mail", "reply", "msg-1", "--body", "Authored reply."}
		if strings.Contains(text, "all") {
			args = append(args, "--all")
		}
		if !strings.Contains(text, "without files") {
			args = append(args, "--attach", p)
		}
		if strings.Contains(text, "missing") {
			args = append(args, "--attach", p+".missing")
		}
		if strings.Contains(text, "preview") {
			args = append(args, "--dry-run")
		}
		w.Code = cli.Run(args, d)
		w.Out, w.Err, w.Sent = out.String(), errw.String(), requests
		if requests != 0 {
			if payload["comment"] != "<p>Authored reply.</p>" {
				w.T.Fatalf("comment=%v", payload)
			}
			if strings.Contains(text, "without files") {
				if len(payload) != 1 {
					w.T.Fatalf("no-files=%v", payload)
				}
			} else {
				msg, _ := payload["message"].(map[string]any)
				atts, _ := msg["attachments"].([]any)
				if len(payload) != 2 || len(msg) != 1 || len(atts) != 1 {
					w.T.Fatalf("attachments only alongside comment: %v", payload)
				}
				att := atts[0].(map[string]any)
				if att["@odata.type"] != "#microsoft.graph.fileAttachment" || att["name"] != "note.txt" || att["contentBytes"] != "c3ludGhldGljIGF0dGFjaG1lbnQ=" {
					w.T.Fatalf("attachment=%v", att)
				}
			}
		}
		return nil
	})
	runtime.Register("the attachment reply is sent once", func(w *runtime.World, _ string) error {
		var result map[string]any
		if w.Code != 0 || w.Sent != 1 || w.Err != "" || json.Unmarshal([]byte(w.Out), &result) != nil || result["sent"] != true {
			w.T.Fatalf("code=%d requests=%d out=%s err=%s", w.Code, w.Sent, w.Out, w.Err)
		}
		return nil
	})
	runtime.Register("the attachment reply fails without fallback", func(w *runtime.World, text string) error {
		want, count, class := domain.ExitUsage, 0, "usage"
		if strings.Contains(text, "service") {
			want, count, class = domain.ExitService, 1, "service"
		}
		var failure map[string]any
		if w.Code != want || w.Sent != count || w.Out != "" || json.Unmarshal([]byte(w.Err), &failure) != nil || failure["class"] != class {
			w.T.Fatalf("code=%d requests=%d out=%s err=%s", w.Code, w.Sent, w.Out, w.Err)
		}
		return nil
	})
	runtime.Register("the attachment reply preview has metadata only", func(w *runtime.World, _ string) error {
		var preview map[string]any
		if w.Code != 0 || w.Sent != 0 || json.Unmarshal([]byte(w.Out), &preview) != nil || preview["dry_run"] != true || !strings.Contains(w.Out, "note.txt") || !strings.Contains(w.Out, `"size":`) || strings.Contains(w.Out, "synthetic attachment") || strings.Contains(w.Out, "contentBytes") || strings.Contains(w.Out, "c3ludGhldGlj") {
			w.T.Fatalf("code=%d requests=%d preview=%s", w.Code, w.Sent, w.Out)
		}
		return nil
	})
}
