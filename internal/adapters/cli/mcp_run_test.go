package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestMCPRunMailListParity(t *testing.T) {
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()
	if c := Run([]string{"m365", "mail", "list"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var cliPage map[string]any
	if err := json.Unmarshal(out.Bytes(), &cliPage); err != nil {
		t.Fatal(err)
	}
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "mail", Verb: "list"},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	var mcpPage map[string]any
	if err := json.Unmarshal([]byte(toolText(t, res)), &mcpPage); err != nil {
		t.Fatal(toolText(t, res))
	}
	if mcpPage["count"] != cliPage["count"] {
		t.Fatalf("%v vs %v", mcpPage["count"], cliPage["count"])
	}
	cliIDs := idsOf(cliPage["items"])
	mcpIDs := idsOf(mcpPage["items"])
	if !equalIDs(cliIDs, mcpIDs) {
		t.Fatalf("%v vs %v", mcpIDs, cliIDs)
	}
}

func TestMCPRunChatAliasAndDownloadPathOnly(t *testing.T) {
	d, _, errw := testDeps()
	loginAll(t, &d)
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "chat", Verb: "list"},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	dest := filepath.Join(t.TempDir(), "out.txt")
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{
			Namespace: "files", Verb: "download", Args: []string{"file-1"},
			Flags: map[string]any{"out": dest},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res), errw.String())
	}
	text := toolText(t, res)
	if strings.Contains(text, "synthetic-ok") {
		t.Fatal(text)
	}
	b, err := os.ReadFile(dest)
	if err != nil || string(b) != "synthetic-ok" {
		t.Fatal(b, err)
	}
}

func idsOf(v any) []string {
	arr, _ := v.([]any)
	var ids []string
	for _, it := range arr {
		m, _ := it.(map[string]any)
		if id, _ := m["id"].(string); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func equalIDs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMCPRunLoginRefused(t *testing.T) {
	d, _, _ := testDeps()
	d.Login = graph.FakeLoginAll(true, true, true, true)
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "auth", Verb: "login"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal(toolText(t, res))
	}
}
