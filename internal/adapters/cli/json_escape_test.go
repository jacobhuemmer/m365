package cli

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Every JSON the CLI writes leaves <, > and & as-is, like normal stdout,
// instead of \u003c, \u003e and \u0026.

func TestStderrErrorKeepsHTMLCharacters(t *testing.T) {
	d, _, errw := testDeps()
	loginAll(t, &d)
	errw.Reset()
	if code := Run([]string{"m365", "mail", "<x>&"}, d); code != 3 {
		t.Fatalf("exit %d", code)
	}
	want := `{"class":"usage","message":"unknown mail verb \"<x>&\""}` + "\n"
	if errw.String() != want {
		t.Fatalf("stderr\ngot  %q\nwant %q", errw.String(), want)
	}
}

func TestMCPErrorKeepsHTMLCharacters(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "m365_help", Arguments: map[string]any{"topic": "<x>&"},
	})
	if err != nil || !res.IsError {
		t.Fatalf("want IsError: %v", err)
	}
	want := `{"class":"usage","message":"unknown topic \"<x>&\""}`
	if got := toolText(t, res); got != want {
		t.Fatalf("MCP error\ngot  %q\nwant %q", got, want)
	}
}

func TestHumanOutputKeepsHTMLCharacters(t *testing.T) {
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()
	if code := Run([]string{"m365", "--human", "teams", "send", "chat-1", "--text", "<b>&", "--dry-run"}, d); code != 0 {
		t.Fatalf("exit %d: %s", code, errw.String())
	}
	want := "{\n  \"attachments\": [],\n  \"chat_id\": \"chat-1\",\n  \"dry_run\": true,\n  \"text\": \"<b>&\"\n}\n"
	if out.String() != want {
		t.Fatalf("--human\ngot  %q\nwant %q", out.String(), want)
	}
}
