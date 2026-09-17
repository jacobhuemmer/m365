package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPromptsListFourRecipes(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	pl, err := cs.ListPrompts(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, p := range pl.Prompts {
		got[p.Name] = true
	}
	for _, name := range recipeNames {
		if !got[name] {
			t.Fatalf("missing prompt %s in %#v", name, got)
		}
	}
	if len(pl.Prompts) != 4 {
		t.Fatalf("count %d", len(pl.Prompts))
	}
	tl, err := cs.ListTools(context.Background(), nil)
	if err != nil || len(tl.Tools) != 3 {
		t.Fatal(err, len(tl.Tools))
	}
}

func TestPromptBodyMatchesHelpTopic(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	for _, name := range recipeNames {
		pr, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{Name: name})
		if err != nil || pr == nil || len(pr.Messages) == 0 {
			t.Fatal(name, err)
		}
		tc, ok := pr.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatalf("%T", pr.Messages[0].Content)
		}
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name: "m365_help", Arguments: helpIn{Topic: name},
		})
		if err != nil || res.IsError {
			t.Fatal(name, err, toolText(t, res))
		}
		if strings.TrimSpace(tc.Text) != strings.TrimSpace(toolText(t, res)) {
			t.Fatalf("%s mismatch", name)
		}
		low := strings.ToLower(tc.Text)
		if strings.Contains(low, "token") || strings.Contains(low, "bearer") || strings.Contains(tc.Text, "synthetic-ok") {
			t.Fatal(tc.Text)
		}
	}
}
