package cli

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestMCPBlockedSendIsUsageError(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	mem := d.Mail.(graph.MailAPI).Memory
	cs := connectMCP(t, d)
	flags := map[string]any{"to": "user@example.com", "subject": "t", "html": true, "body": "<p>hi"}

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "mail", Verb: "send", Flags: flags, WriteOptIn: true},
	})
	if err != nil || !res.IsError {
		t.Fatalf("want IsError: %v %s", err, toolText(t, res))
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(toolText(t, res)), &got); err != nil {
		t.Fatal(err, toolText(t, res))
	}
	want := map[string]any{"class": "usage", "message": "1 format problem: broken-html: unclosed <p>", "hint": "run with --preview to see them"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	if len(mem.Sent) != 0 {
		t.Fatalf("sent %d", len(mem.Sent))
	}

	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{Namespace: "mail", Verb: "send", Flags: flags},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	var dry map[string]any
	if err := json.Unmarshal([]byte(toolText(t, res)), &dry); err != nil {
		t.Fatal(err, toolText(t, res))
	}
	wantProblems := []any{map[string]any{"rule": "broken-html", "detail": "unclosed <p>"}}
	if !reflect.DeepEqual(dry["format_problems"], wantProblems) {
		t.Fatalf("format_problems %#v", dry["format_problems"])
	}
}
