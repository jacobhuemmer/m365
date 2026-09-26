package cli

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPExactRecipientsAllowsKnownMailRecipients(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	mem := d.Mail.(graph.MailAPI).Memory
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := NewMCPServerWithPolicy(d, MCPPolicy{ExactRecipients: true, Allow: map[string]bool{"mail.send": true, "mail.reply": true}}).Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	for _, in := range []runIn{
		{Namespace: "mail", Verb: "send", WriteOptIn: true, Flags: map[string]any{"to": "approved@example.com", "cc": []string{"copy@example.com"}, "subject": "Synthetic", "body": "Test"}},
		{Namespace: "mail", Verb: "reply", Args: []string{"msg-1"}, WriteOptIn: true, Flags: map[string]any{"body": "Synthetic reply"}},
	} {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "m365_run", Arguments: in})
		if err != nil || res.IsError {
			t.Fatalf("valid %s was rejected: %v, %v", in.Verb, err, res)
		}
	}
	if got := len(mem.Sent); got != 2 {
		t.Fatalf("valid sends = %d, want 2", got)
	}
}
