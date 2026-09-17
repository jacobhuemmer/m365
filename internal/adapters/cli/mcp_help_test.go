package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPHelpNoSession(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_help", Arguments: helpIn{}})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	root := toolText(t, res)
	if !strings.Contains(root, "mcp") {
		t.Fatal(root)
	}
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_help", Arguments: helpIn{Namespace: "calendar"},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	cal := toolText(t, res)
	for _, want := range []string{"list", "create", "free"} {
		if !strings.Contains(cal, want) {
			t.Fatalf("missing %s in %s", want, cal)
		}
	}
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_help", Arguments: helpIn{Namespace: "mcp"},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	if !strings.Contains(toolText(t, res), "write_opt_in") {
		t.Fatal(toolText(t, res))
	}
}
