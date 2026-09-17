package cli

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestMCPRunUnknownVerbAndTop(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "mail", Verb: "nope"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertClass(t, res, "usage")
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "mail", Verb: "list", Flags: map[string]any{"top": float64(51)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertClass(t, res, "usage")
}

func TestMCPRunMissingConsentIsAuth(t *testing.T) {
	d, _, _ := testDeps()
	d.Login = graph.FakeLoginAll(true, false, true, true)
	if c := Run([]string{"m365", "auth", "login"}, d); c != 0 {
		t.Fatal(c)
	}
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "teams", Verb: "list"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertClass(t, res, "auth")
}

func TestMCPRunExpiredSessionIsAuth(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	b, ok, err := d.Store.Get()
	if err != nil || !ok {
		t.Fatal(err, ok)
	}
	b.Usable = false
	if err := d.Store.Put(b); err != nil {
		t.Fatal(err)
	}
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "mail", Verb: "list"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertClass(t, res, "auth")
}

func assertClass(t *testing.T, res *mcp.CallToolResult, class string) {
	t.Helper()
	if !res.IsError {
		t.Fatal(toolText(t, res))
	}
	var e map[string]any
	if err := json.Unmarshal([]byte(toolText(t, res)), &e); err != nil {
		t.Fatal(toolText(t, res))
	}
	if e["class"] != class {
		t.Fatalf("%v", e)
	}
}
