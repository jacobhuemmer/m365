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

func TestMCPHelpTopics(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "m365_help", Arguments: helpIn{}})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	overview := toolText(t, res)
	for _, topic := range recipeNames {
		if !strings.Contains(overview, topic) {
			t.Fatalf("overview missing %s in %s", topic, overview)
		}
	}
	cases := map[string][]string{
		"mail-search": {"mail list --folder all --search 'from:ajay'", "mail get", "mail thread"},
		"teams-find":  {"teams list", "MUST NOT send", "Several matches"},
		"calendar":    {"calendar list", "calendar free", "calendar create --when 'tomorrow at 1:30 pm'"},
		"files":       {"files list", "files download", "--out", "upload", "--dry-run"},
	}
	for topic, wants := range cases {
		res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_help", Arguments: helpIn{Topic: topic},
		})
		if err != nil || res.IsError {
			t.Fatal(topic, err, toolText(t, res))
		}
		body := toolText(t, res)
		for _, w := range wants {
			if !strings.Contains(body, w) {
				t.Fatalf("%s missing %q in %s", topic, w, body)
			}
		}
		if topic == "teams-find" && strings.Contains(body, "teams find") {
			t.Fatal(body)
		}
	}
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_help", Arguments: helpIn{Topic: "nope"},
	})
	if err != nil || !res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_help", Arguments: helpIn{Topic: "calendar", Namespace: "calendar"},
	})
	if err != nil || !res.IsError {
		t.Fatal(err, toolText(t, res))
	}
}
