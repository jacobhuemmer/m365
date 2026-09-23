package steps

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/acceptance/runtime"
)

func init() {
	runtime.Register("the client calls MCP help with no topic", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_help", Arguments: map[string]any{}})
		cur.res = res
		return err
	})
	runtime.Register("MCP help names recipe topics", func(w *runtime.World, _ string) error {
		s := mcpText()
		for _, want := range []string{"mail-search", "teams-find", "calendar", "files", "mail-write", "teams-write"} {
			if !strings.Contains(s, want) {
				w.T.Fatalf("missing %s in %s", want, s)
			}
		}
		return nil
	})
	runtime.Register("the client calls MCP help topic mail-search", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_help", Arguments: map[string]any{"topic": "mail-search"},
		})
		cur.res = res
		return err
	})
	runtime.Register("MCP help includes from:ajay", func(w *runtime.World, _ string) error {
		if !strings.Contains(mcpText(), "from:ajay") {
			w.T.Fatalf("%s", mcpText())
		}
		return nil
	})
	runtime.Register("the client lists MCP prompts", func(w *runtime.World, _ string) error {
		list, err := cur.cs.ListPrompts(context.Background(), nil)
		if err != nil {
			return err
		}
		cur.tools = nil
		for _, p := range list.Prompts {
			cur.tools = append(cur.tools, p.Name)
		}
		return nil
	})
	runtime.Register("the MCP recipes are six", func(w *runtime.World, _ string) error {
		if len(cur.tools) != 6 {
			w.T.Fatalf("prompts %v", cur.tools)
		}
		return nil
	})
}
