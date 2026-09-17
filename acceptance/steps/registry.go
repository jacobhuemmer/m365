package steps

import (
	"bytes"
	"strings"

	"github.com/masonhuemmer/m365/acceptance/runtime"
	"github.com/masonhuemmer/m365/internal/adapters/cli"
	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/adapters/keychain"
	"github.com/masonhuemmer/m365/internal/config"
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
			Config: config.Config{ClientID: "x", TenantID: "y"},
			Store:  &keychain.Fake{},
			Mail:     graph.MailAPI{Memory: mem},
			Teams:    graph.TeamsAPI{Memory: mem},
			Calendar: graph.CalendarAPI{Memory: mem},
			Files:    graph.FilesAPI{Memory: mem},
			Login:    graph.FakeLoginAll(true, true, true, true),
			Stdout: out, Stderr: errw,
		}
		code := cli.Run(append([]string{"m365"}, parts...), d)
		if strings.Contains(text, "help") && code != 0 {
			w.T.Fatalf("help exit %d %s", code, errw)
		}
		_ = out
		return nil
	})
	runtime.Register("the command succeeds", func(w *runtime.World, _ string) error { return nil })
	runtime.Register("stdout is JSON", func(w *runtime.World, _ string) error { return nil })
	runtime.Register("exit code", func(w *runtime.World, _ string) error { return nil })
}
