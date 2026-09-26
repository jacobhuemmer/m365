package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPPolicyValidation(t *testing.T) {
	for _, raw := range []string{"mail.list", "mail.send,", "send", "teams.nope"} {
		if _, err := newMCPPolicy(false, raw, false); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	p, err := newMCPPolicy(false, "teams.send, mail.send", true)
	if err != nil || !p.Allow["teams.send"] || !p.Allow["mail.send"] || !p.ExactRecipients {
		t.Fatalf("policy %+v, %v", p, err)
	}
	d, _, _ := testDeps()
	if code := Run([]string{"m365", "mcp", "serve", "--allow", "mail.list"}, d); code == 0 {
		t.Fatal("serve accepted invalid allowlist")
	}
}

func TestMCPPolicyGates(t *testing.T) {
	readonly := MCPPolicy{ReadOnly: true}
	if err := readonly.check(&runIn{Namespace: "mail", Verb: "list", WriteOptIn: true}); err == nil {
		t.Fatal("read-only accepted opt-in on a read")
	}
	if err := readonly.check(&runIn{Namespace: "mail", Verb: "send"}); err != nil {
		t.Fatal(err)
	}
	allow := MCPPolicy{Allow: map[string]bool{"teams.send": true}}
	if err := allow.check(&runIn{Namespace: "mail", Verb: "send", WriteOptIn: true}); err == nil {
		t.Fatal("unlisted write allowed")
	}
	if err := allow.check(&runIn{Namespace: "mail", Verb: "send"}); err != nil {
		t.Fatal("dry-run blocked", err)
	}
	if err := allow.check(&runIn{Namespace: "chat", Verb: "send", WriteOptIn: true}); err != nil {
		t.Fatal("allowed alias blocked", err)
	}
}

func TestMCPExactRecipients(t *testing.T) {
	p := MCPPolicy{ExactRecipients: true}
	for _, to := range []string{"Ajay", "ajay@", "ajay@example.com.uk extra"} {
		in := runIn{Namespace: "teams", Verb: "send", Flags: map[string]any{"to": to}}
		if err := p.check(&in); err == nil {
			t.Fatalf("accepted fuzzy recipient %q", to)
		}
	}
	for _, to := range []string{"ajay@example.com", "01234567-89ab-cdef-0123-456789abcdef"} {
		in := runIn{Namespace: "teams", Verb: "send", Flags: map[string]any{"to": to}}
		if err := p.check(&in); err != nil || in.Flags["exact-recipient"] != true {
			t.Fatalf("rejected exact recipient %q: %v", to, err)
		}
	}
	in := runIn{Namespace: "teams", Verb: "send", Flags: map[string]any{"to": "19:chat@thread.v2"}}
	if err := p.check(&in); err != nil || len(in.Args) != 1 || in.Args[0] != "19:chat@thread.v2" {
		t.Fatalf("chat ID not converted: %+v, %v", in, err)
	}
}

func TestMCPExactRecipientDoesNotUseFuzzyMatch(t *testing.T) {
	d, _, _ := testDeps()
	loginAll(t, &d)
	mem := d.Teams.(graph.TeamsAPI).Memory
	for i := range mem.Chats {
		if mem.Chats[i].ID == "chat-ajay" {
			mem.Chats[i].Members[0].ID = "01234567-89ab-cdef-0123-456789abcdef"
		}
	}
	before := len(mem.Sent)
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := NewMCPServerWithPolicy(d, MCPPolicy{ExactRecipients: true}).Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	call := func(to string) *mcp.CallToolResult {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "m365_run", Arguments: runIn{Namespace: "teams", Verb: "send", WriteOptIn: true, Flags: map[string]any{"to": to, "text": "ping"}}})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	if res := call("ajay@example.co"); !res.IsError {
		t.Fatalf("fuzzy email sent: %s", toolText(t, res))
	}
	if len(mem.Sent) != before {
		t.Fatal("fuzzy email sent")
	}
	res := call("ajay@example.com")
	if res.IsError {
		t.Fatal(toolText(t, res))
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(toolText(t, res)), &v); err != nil || v["sent"] != true {
		t.Fatal(toolText(t, res))
	}
	if len(mem.Sent) != before+1 {
		t.Fatal("exact recipient not sent")
	}
	if res := call("01234567-89ab-cdef-0123-456789abcdef"); res.IsError {
		t.Fatal(toolText(t, res))
	}
	if len(mem.Sent) != before+2 {
		t.Fatal("user ID not sent")
	}
	if !strings.Contains(toolText(t, res), "sent") {
		t.Fatal(toolText(t, res))
	}
}
