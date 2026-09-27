package cli

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPRejectsPreview(t *testing.T) {
	for _, v := range []bool{true, false} {
		d, _, _ := testDeps()
		loginAll(t, &d)
		cs := connectMCP(t, d)
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_run", Arguments: runIn{
				Namespace: "mail", Verb: "send",
				Flags: map[string]any{"to": "user@example.com", "subject": "t", "body": "b", "preview": v},
			},
		})
		if err != nil || !res.IsError {
			t.Fatalf("preview=%v: want IsError: %v %s", v, err, toolText(t, res))
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(toolText(t, res)), &got); err != nil {
			t.Fatal(err, toolText(t, res))
		}
		want := map[string]any{
			"class": "usage", "message": "preview is a terminal flag",
			"hint": "use dry-run; its JSON has rendered and format_problems",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("preview=%v: got %#v", v, got)
		}
	}
}
