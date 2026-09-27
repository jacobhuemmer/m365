package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Redacting JSON output must leave valid JSON: MCP callers parse it.
func TestRedactJSONStaysValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"secret in a value", `{"body":"access_token=abc123","dry_run":true}`, `{"body":"[redacted]","dry_run":true}`},
		{"secret inside html", `{"c":"<p>access_token=abc123</p>"}`, `{"c":"<p>[redacted]</p>"}`},
		{"secret under a key", `{"access_token":"xyz","n":1}`, `{"access_token":"[redacted]","n":1}`},
		{"secret under a key with spaces", `{"Refresh_Token" : "xyz"}`, `{"Refresh_Token" : "[redacted]"}`},
		{"bearer in a value", `{"h":"Bearer abc.def"}`, `{"h":"[redacted]"}`},
		{"escaped quotes around", `{"b":"say \"access_token=abc\" now"}`, `{"b":"say \"[redacted]\" now"}`},
		{"json lines", "{\"a\":\"access_token=x1\"}\n{\"a\":\"ok\"}\n", "{\"a\":\"[redacted]\"}\n{\"a\":\"ok\"}\n"},
		{"indented", "{\n  \"a\": \"access_token=x\"\n}", "{\n  \"a\": \"[redacted]\"\n}"},
		{"no secret is byte-identical", `{"c":"<p>x</p>","e":"<p>","q":"a\"b"}`, `{"c":"<p>x</p>","e":"<p>","q":"a\"b"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := redact(tc.in)
			if got != tc.want {
				t.Fatalf("redact(%s)\ngot  %s\nwant %s", tc.in, got, tc.want)
			}
			for _, line := range strings.Split(strings.TrimSpace(got), "\n") {
				if !json.Valid([]byte(line)) && !json.Valid([]byte(got)) {
					t.Fatalf("not valid JSON: %s", got)
				}
			}
		})
	}
}

func TestMCPRedactedOutputIsJSON(t *testing.T) {
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
	var v map[string]any
	if err := json.Unmarshal([]byte(text), &v); err != nil {
		t.Fatalf("MCP output is not JSON after redaction: %v\n%s", err, text)
	}
	if v["body"] != "[redacted]" || strings.Contains(text, "abc123") {
		t.Fatalf("body %v in %s", v["body"], text)
	}
}
