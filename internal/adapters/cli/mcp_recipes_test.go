package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPromptsListSixRecipes(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	pl, err := cs.ListPrompts(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, p := range pl.Prompts {
		got[p.Name] = p.Description
	}
	for _, name := range recipeNames {
		if got[name] == "" {
			t.Fatalf("missing prompt %s in %#v", name, got)
		}
	}
	if len(pl.Prompts) != 6 {
		t.Fatalf("count %d", len(pl.Prompts))
	}
	if got["mail-search"] != "Lookup recipe mail-search" {
		t.Fatalf("mail-search desc %q", got["mail-search"])
	}
	if got["mail-write"] != "Write recipe mail-write" {
		t.Fatalf("mail-write desc %q", got["mail-write"])
	}
	if got["teams-write"] != "Write recipe teams-write" {
		t.Fatalf("teams-write desc %q", got["teams-write"])
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

func TestPromptBodyWriteTopicsUseHTML(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	for _, name := range []string{"mail-write", "teams-write"} {
		pr, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{Name: name})
		if err != nil || pr == nil || len(pr.Messages) == 0 {
			t.Fatal(name, err)
		}
		tc, ok := pr.Messages[0].Content.(*mcp.TextContent)
		if !ok {
			t.Fatalf("%T", pr.Messages[0].Content)
		}
		body := tc.Text
		if !strings.Contains(body, "<p>") && !strings.Contains(body, "<ul>") {
			t.Fatalf("%s missing HTML tags in %s", name, body)
		}
		if !strings.Contains(body, "--html") {
			t.Fatalf("%s missing --html in %s", name, body)
		}
		if strings.Contains(body, "--body '# ") || strings.Contains(body, "--text '# ") {
			t.Fatalf("%s markdown heading send body in %s", name, body)
		}
		if name == "mail-write" {
			for _, w := range []string{"mail send --", "--html", "--body", "--dry-run", "mail reply", "flags.body", "body-file=-"} {
				if w == "body-file=-" {
					if !strings.Contains(body, "body-file=-") {
						t.Fatalf("%s missing %q in %s", name, w, body)
					}
					continue
				}
				if !strings.Contains(body, w) {
					t.Fatalf("%s missing %q in %s", name, w, body)
				}
			}
			if strings.Contains(body, "write_opt_in true") || strings.Contains(body, "write_opt_in=true") {
				t.Fatalf("%s happy path must not set write_opt_in true: %s", name, body)
			}
		}
		if name == "teams-write" {
			for _, w := range []string{"teams send --to Ajay --format md", "--html", "flags.text", "text-file=-"} {
				if !strings.Contains(body, w) {
					t.Fatalf("%s missing %q in %s", name, w, body)
				}
			}
		}
	}
}
