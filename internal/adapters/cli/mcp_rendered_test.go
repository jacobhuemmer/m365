package cli

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FR-022: MCP secret redaction still covers the rendered body, and the
// output stays valid JSON (#14).
func TestMCPRedactsRenderedContent(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_run", Arguments: runIn{
			Namespace: "mail", Verb: "send",
			Flags: map[string]any{"to": "user@example.com", "subject": "t", "body": "access_token=abc123"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	text := toolText(t, res)
	var out struct {
		Body     string            `json:"body"`
		Rendered map[string]string `json:"rendered"`
	}
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("MCP output is not JSON: %v\n%s", err, text)
	}
	want := map[string]string{"content_type": "html", "content": "<p>[redacted]</p>"}
	if out.Body != "[redacted]" || !reflect.DeepEqual(out.Rendered, want) {
		t.Fatalf("body %q rendered %v, want [redacted] and %v", out.Body, out.Rendered, want)
	}
}
