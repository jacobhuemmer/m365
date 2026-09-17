package cli

import (
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestTeamsSendToDryRun(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	out.Reset()
	if c := Run([]string{"m365", "teams", "send", "--to", "Ajay", "--text", "ping", "--dry-run"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var v map[string]any
	if err := json.Unmarshal(out.Bytes(), &v); err != nil || v["dry_run"] != true || v["chat_id"] != "chat-ajay" || v["to"] != "Ajay" {
		t.Fatal(out.String(), err)
	}
	out.Reset()
	if c := Run([]string{"m365", "teams", "send", "chat-1", "--to", "Ajay", "--text", "x"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "teams", "send", "--to", "Nobody", "--text", "x"}, d); c != domain.ExitNotFound {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "teams", "send", "chat-1", "--dry-run", "--text", "hi"}, d); c != 0 {
		t.Fatal(errw.String())
	}
}
