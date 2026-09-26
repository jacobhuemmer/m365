package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/acceptance/runtime"
	"github.com/masonhuemmer/m365/internal/adapters/cli"
	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/adapters/keychain"
	"github.com/masonhuemmer/m365/internal/config"
)

type mcpWorld struct {
	cs     *mcp.ClientSession
	d      cli.Deps
	res    *mcp.CallToolResult
	tools  []string
	cancel func()
}

var (
	mcpMu sync.Mutex
	cur   *mcpWorld
)

func mcpDeps(teams bool) cli.Deps {
	mem := graph.Seed()
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	return cli.Deps{
		Config:   config.Config{ClientID: "x", TenantID: "y"},
		Store:    &keychain.Fake{},
		Mail:     graph.MailAPI{Memory: mem},
		Teams:    graph.TeamsAPI{Memory: mem},
		Calendar: graph.CalendarAPI{Memory: mem},
		Files:    graph.FilesAPI{Memory: mem},
		Login:    graph.FakeLoginAll(true, teams, true, true),
		Stdout:   out,
		Stderr:   errw,
	}
}

func startMCP(w *runtime.World, login, teams bool, policies ...cli.MCPPolicy) error {
	mcpMu.Lock()
	defer mcpMu.Unlock()
	if cur != nil && cur.cancel != nil {
		cur.cancel()
	}
	d := mcpDeps(teams)
	if login {
		if c := cli.Run([]string{"m365", "auth", "login"}, d); c != 0 {
			w.T.Fatalf("login %d", c)
		}
	}
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	policy := cli.MCPPolicy{}
	if len(policies) > 0 {
		policy = policies[0]
	}
	ss, err := cli.NewMCPServerWithPolicy(d, policy).Connect(ctx, t1, nil)
	if err != nil {
		return err
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "acc", Version: "1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		_ = ss.Close()
		return err
	}
	cancel := func() { _ = cs.Close(); _ = ss.Close() }
	w.T.Cleanup(cancel)
	cur = &mcpWorld{cs: cs, d: d, cancel: cancel}
	return nil
}

func mcpText() string {
	if cur == nil || cur.res == nil || len(cur.res.Content) == 0 {
		return ""
	}
	tc, ok := cur.res.Content[0].(*mcp.TextContent)
	if !ok {
		return ""
	}
	return tc.Text
}

func mcpClass() string {
	var e map[string]any
	if err := json.Unmarshal([]byte(mcpText()), &e); err != nil {
		return ""
	}
	s, _ := e["class"].(string)
	return s
}

func init() {
	runtime.Register("an MCP server", func(w *runtime.World, text string) error {
		if strings.Contains(text, "without teams consent") {
			return startMCP(w, true, false)
		}
		if strings.Contains(text, "signed-in") {
			return startMCP(w, true, true)
		}
		return startMCP(w, false, true)
	})
	runtime.Register("a signed-in MCP server", func(w *runtime.World, _ string) error {
		return startMCP(w, true, true)
	})
	runtime.Register("the client lists MCP tools", func(w *runtime.World, _ string) error {
		list, err := cur.cs.ListTools(context.Background(), nil)
		if err != nil {
			return err
		}
		cur.tools = nil
		for _, tl := range list.Tools {
			cur.tools = append(cur.tools, tl.Name)
		}
		return nil
	})
	runtime.Register("the MCP catalog is compact", func(w *runtime.World, _ string) error {
		if len(cur.tools) != 3 {
			w.T.Fatalf("tools %v", cur.tools)
		}
		return nil
	})
	runtime.Register("the client calls MCP status", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_status"})
		cur.res = res
		return err
	})
	runtime.Register("MCP status is signed out", func(w *runtime.World, _ string) error {
		if cur.res.IsError || !strings.Contains(mcpText(), `"signed_in":false`) && !strings.Contains(mcpText(), `"signed_in": false`) {
			var st map[string]any
			_ = json.Unmarshal([]byte(mcpText()), &st)
			if st["signed_in"] != false {
				w.T.Fatalf("%s", mcpText())
			}
		}
		return nil
	})
	runtime.Register("MCP status is signed in", func(w *runtime.World, _ string) error {
		var st map[string]any
		if err := json.Unmarshal([]byte(mcpText()), &st); err != nil || st["signed_in"] != true {
			w.T.Fatalf("%s", mcpText())
		}
		return nil
	})
	runtime.Register("the client runs MCP mail list", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_run", Arguments: map[string]any{"namespace": "mail", "verb": "list"},
		})
		cur.res = res
		return err
	})
	runtime.Register("the client runs MCP mail nope", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_run", Arguments: map[string]any{"namespace": "mail", "verb": "nope"},
		})
		cur.res = res
		return err
	})
	runtime.Register("the client runs MCP teams list", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_run", Arguments: map[string]any{"namespace": "teams", "verb": "list"},
		})
		cur.res = res
		return err
	})
	runtime.Register("the MCP run succeeds", func(w *runtime.World, _ string) error {
		if cur.res.IsError {
			w.T.Fatalf("%s", mcpText())
		}
		return nil
	})
	runtime.Register("the MCP run is usage", func(w *runtime.World, _ string) error {
		if !cur.res.IsError || mcpClass() != "usage" {
			w.T.Fatalf("%s", mcpText())
		}
		return nil
	})
	runtime.Register("the MCP run is auth", func(w *runtime.World, _ string) error {
		if !cur.res.IsError || mcpClass() != "auth" {
			w.T.Fatalf("%s", mcpText())
		}
		return nil
	})
	runtime.Register("the client runs MCP mail send without opt-in", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_run", Arguments: map[string]any{
				"namespace": "mail", "verb": "send",
				"flags": map[string]any{"to": "user@example.com", "subject": "t", "body": "b"},
			},
		})
		cur.res = res
		return err
	})
	runtime.Register("the client runs MCP calendar create without opt-in", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_run", Arguments: map[string]any{
				"namespace": "calendar", "verb": "create",
				"flags": map[string]any{"subject": "t", "start": "2026-09-16T10:00:00Z", "end": "2026-09-16T11:00:00Z"},
			},
		})
		cur.res = res
		return err
	})
	runtime.Register("the MCP run is dry-run", func(w *runtime.World, _ string) error {
		var v map[string]any
		if cur.res.IsError || json.Unmarshal([]byte(mcpText()), &v) != nil || v["dry_run"] != true {
			w.T.Fatalf("%s", mcpText())
		}
		return nil
	})
	runtime.Register("the client calls MCP help for calendar", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_help", Arguments: map[string]any{"namespace": "calendar"},
		})
		cur.res = res
		return err
	})
	runtime.Register("MCP help names calendar verbs", func(w *runtime.World, _ string) error {
		s := mcpText()
		for _, want := range []string{"list", "create", "free"} {
			if !strings.Contains(s, want) {
				w.T.Fatalf("missing %s in %s", want, s)
			}
		}
		return nil
	})
}
