package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func connectMCP(t *testing.T, d Deps) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := NewMCPServer(d).Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func toolText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if res == nil || len(res.Content) == 0 {
		t.Fatal("no content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("%T", res.Content[0])
	}
	return tc.Text
}

func TestToolsListCompactCatalog(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	list, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tools) != 3 {
		t.Fatalf("count %d", len(list.Tools))
	}
	got := map[string]string{}
	for _, tl := range list.Tools {
		got[tl.Name] = tl.Description
	}
	for _, name := range []string{"m365_status", "m365_help", "m365_run"} {
		if got[name] == "" {
			t.Fatalf("missing %s in %#v", name, got)
		}
	}
	if !strings.Contains(got["m365_run"], "mail-search") {
		t.Fatal(got["m365_run"])
	}
}
