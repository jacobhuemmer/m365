package steps

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/acceptance/runtime"
	"github.com/masonhuemmer/m365/internal/adapters/cli"
)

func init() {
	runtime.Register("a signed-in read-only MCP server", func(w *runtime.World, _ string) error {
		return startMCP(w, true, true, cli.MCPPolicy{ReadOnly: true})
	})
	runtime.Register("a signed-in MCP server allowing only", func(w *runtime.World, text string) error {
		verb := strings.TrimPrefix(text, "a signed-in MCP server allowing only ")
		return startMCP(w, true, true, cli.MCPPolicy{Allow: map[string]bool{verb: true}})
	})
	runtime.Register("a signed-in exact-recipients MCP server", func(w *runtime.World, _ string) error {
		return startMCP(w, true, true, cli.MCPPolicy{ExactRecipients: true})
	})
	runtime.Register("the client opts in to an MCP mail send", func(w *runtime.World, _ string) error {
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_run", Arguments: map[string]any{
			"namespace": "mail", "verb": "send", "write_opt_in": true,
			"flags": map[string]any{"to": "user@example.com", "subject": "t", "body": "b"},
		}})
		cur.res = res
		return err
	})
	runtime.Register("the client opts in to an MCP teams send to", func(w *runtime.World, text string) error {
		to := strings.TrimPrefix(text, "the client opts in to an MCP teams send to ")
		res, err := cur.cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_run", Arguments: map[string]any{
			"namespace": "teams", "verb": "send", "write_opt_in": true,
			"flags": map[string]any{"to": to, "text": "ping"},
		}})
		cur.res = res
		return err
	})
}
