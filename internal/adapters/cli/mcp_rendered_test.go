package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FR-022: MCP secret redaction still covers the rendered body. The
// existing redaction also consumes the closing JSON quote, so the output
// is checked for the secret and the redacted rendered content, not parsed.
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
	if strings.Contains(text, "abc123") {
		t.Fatalf("secret leaked: %s", text)
	}
	if !strings.Contains(text, `"rendered":{"content_type":"html","content":"<p>[redacted]`) {
		t.Fatalf("rendered content not redacted in place: %s", text)
	}
}
