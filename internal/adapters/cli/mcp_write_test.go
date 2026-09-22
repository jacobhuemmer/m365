package cli

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestMCPWriteGateDryRun(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	mem := d.Mail.(graph.MailAPI).Memory
	before := len(mem.Sent)
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{
			Namespace: "mail", Verb: "send",
			Flags: map[string]any{"to": "user@example.com", "subject": "t", "body": "b"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(toolText(t, res)), &v); err != nil || v["dry_run"] != true {
		t.Fatal(toolText(t, res))
	}
	if len(mem.Sent) != before {
		t.Fatal("sent")
	}
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{
			Namespace: "mail", Verb: "reply", Args: []string{"msg-1"},
			Flags: map[string]any{"body": "reviewed draft"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	if err := json.Unmarshal([]byte(toolText(t, res)), &v); err != nil || v["dry_run"] != true {
		t.Fatal(toolText(t, res))
	}
	if len(mem.Sent) != before {
		t.Fatal("reply sent without write opt-in")
	}
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{
			Namespace: "calendar", Verb: "create",
			Flags: map[string]any{"subject": "t", "start": "2026-09-16T10:00:00Z", "end": "2026-09-16T11:00:00Z"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	if err := json.Unmarshal([]byte(toolText(t, res)), &v); err != nil || v["dry_run"] != true {
		t.Fatal(toolText(t, res))
	}
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{
			Namespace: "mail", Verb: "send", WriteOptIn: true,
			Flags: map[string]any{"to": "user@example.com", "subject": "t", "body": "b"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	if len(mem.Sent) == before {
		t.Fatal("expected send")
	}
}
