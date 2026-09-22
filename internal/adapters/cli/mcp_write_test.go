package cli

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/config"
	"github.com/masonhuemmer/m365/internal/domain"
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

func TestMCPActionableClassificationDoesNotAuthorizeReply(t *testing.T) {
	d, _, _ := testDeps()
	d.Config.Experimental.MailResponseClassification = config.MailResponseClassification{
		Enabled: true, Provider: "jev", Model: "jev-latest", ActionableThreshold: 0.8,
	}
	loginAll(t, &d)
	mem := d.Mail.(graph.MailAPI).Memory
	client := connectMCP(t, d)

	classified, err := client.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{
			Namespace: "mail", Verb: "watch",
			Flags: map[string]any{
				"classify":         true,
				"include-existing": true,
				"target-address":   []string{"user@example.com"},
			},
		},
	})
	if err != nil || classified.IsError {
		t.Fatal(err, toolText(t, classified))
	}
	var event domain.MailWatchEvent
	if err := json.Unmarshal([]byte(toolText(t, classified)), &event); err != nil || event.Actionable == nil || !*event.Actionable {
		t.Fatalf("classified event = %s, %v", toolText(t, classified), err)
	}

	before := len(mem.Sent)
	reply, err := client.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{
			Namespace: "mail", Verb: "reply", Args: []string{"msg-1"},
			Flags: map[string]any{"body": "reviewed draft"},
		},
	})
	if err != nil || reply.IsError {
		t.Fatal(err, toolText(t, reply))
	}
	var preview map[string]any
	if err := json.Unmarshal([]byte(toolText(t, reply)), &preview); err != nil || preview["dry_run"] != true {
		t.Fatalf("reply = %s, %v", toolText(t, reply), err)
	}
	if len(mem.Sent) != before {
		t.Fatal("actionable classification bypassed MCP reply write gate")
	}
}
