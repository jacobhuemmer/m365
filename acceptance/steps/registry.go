package steps

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/masonhuemmer/m365/acceptance/runtime"
	"github.com/masonhuemmer/m365/internal/adapters/cli"
	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/adapters/keychain"
	"github.com/masonhuemmer/m365/internal/config"
	"github.com/masonhuemmer/m365/internal/domain"
)

func init() {
	RegisterAll()
}

func RegisterAll() {
	runtime.Register("the CLI is available", func(w *runtime.World, _ string) error {
		w.T.Helper()
		return nil
	})
	runtime.Register("I run", func(w *runtime.World, text string) error {
		cmd := strings.TrimPrefix(text, "I run ")
		cmd = strings.Trim(cmd, `"`)
		parts := strings.Fields(cmd)
		mem := graph.Seed()
		out, errw := &bytes.Buffer{}, &bytes.Buffer{}
		d := cli.Deps{
			Config:   config.Config{ClientID: "x", TenantID: "y"},
			Store:    &keychain.Fake{},
			Mail:     graph.MailAPI{Memory: mem},
			Teams:    graph.TeamsAPI{Memory: mem},
			Calendar: graph.CalendarAPI{Memory: mem},
			Files:    graph.FilesAPI{Memory: mem},
			Login:    graph.FakeLoginAll(true, true, true, true),
			Stdout:   out, Stderr: errw,
		}
		if !strings.Contains(text, "help") && !strings.Contains(text, "--help") {
			_ = cli.Run([]string{"m365", "auth", "login"}, d)
			out.Reset()
			errw.Reset()
		}
		code := cli.Run(append([]string{"m365"}, parts...), d)
		w.Code, w.Out, w.Err = code, out.String(), errw.String()
		return nil
	})
	runtime.Register("the command succeeds", func(w *runtime.World, _ string) error {
		if w.Code != domain.ExitOK {
			w.T.Fatalf("exit %d stderr %s stdout %s", w.Code, w.Err, w.Out)
		}
		return nil
	})
	runtime.Register("stdout is JSON", func(w *runtime.World, _ string) error {
		var v any
		if err := json.Unmarshal([]byte(strings.TrimSpace(w.Out)), &v); err != nil {
			w.T.Fatalf("not JSON: %s", w.Out)
		}
		return nil
	})
	runtime.Register("exit code", func(w *runtime.World, text string) error {
		want := strings.TrimSpace(strings.TrimPrefix(text, "exit code"))
		if want != "" && strconv.Itoa(w.Code) != want {
			w.T.Fatalf("exit %d want %s", w.Code, want)
		}
		return nil
	})
}
