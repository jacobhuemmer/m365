package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestTeamsListAndChatAlias(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	out.Reset()
	if c := Run([]string{"m365", "teams", "list"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var page map[string]any
	_ = json.Unmarshal(out.Bytes(), &page)
	if page["limit"].(float64) != 20 {
		t.Fatalf("%v", page["limit"])
	}
	out.Reset()
	if c := Run([]string{"m365", "chat", "list"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "teams", "get", "missing"}, d); c != domain.ExitNotFound {
		t.Fatal(c, errw.String())
	}
}

func TestTeamsMessagesHelpAndSender(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	out.Reset()
	if c := Run([]string{"m365", "teams", "messages", "--help"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	if !strings.Contains(out.String(), "from:") {
		t.Fatalf("help: %s", out.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "teams", "messages", "chat-1"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var page domain.ChatMessagePage
	if err := json.Unmarshal(out.Bytes(), &page); err != nil || page.Count == 0 || page.Items[0].From != "Alice" {
		t.Fatalf("%v %s", err, out.String())
	}
}
