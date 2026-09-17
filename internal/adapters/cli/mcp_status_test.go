package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPStatusSignedOutAndIn(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_status"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatal(toolText(t, res))
	}
	text := toolText(t, res)
	if strings.Contains(strings.ToLower(text), "token") || strings.Contains(strings.ToLower(text), "bearer") {
		t.Fatal(text)
	}
	var st map[string]any
	if err := json.Unmarshal([]byte(text), &st); err != nil {
		t.Fatal(err, text)
	}
	if st["signed_in"] != false {
		t.Fatal(st)
	}
	loginAll(t, &d)
	cs = connectMCP(t, d)
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_status"})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	text = toolText(t, res)
	if err := json.Unmarshal([]byte(text), &st); err != nil {
		t.Fatal(text)
	}
	if st["signed_in"] != true || st["session_usable"] != true {
		t.Fatal(st)
	}
	ns := st["namespaces"].(map[string]any)
	for _, k := range []string{"mail", "teams", "calendar", "files"} {
		if ns[k] != true {
			t.Fatalf("%s %v", k, ns)
		}
	}
}
