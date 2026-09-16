package cli

import (
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

func TestTeamsDryRun(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	out.Reset()
	c := Run([]string{"m365", "teams", "send", "chat-1", "--dry-run", "--text", "hi"}, d)
	if c != 0 {
		t.Fatal(errw.String())
	}
	var m map[string]any
	_ = json.Unmarshal(out.Bytes(), &m)
	if m["dry_run"] != true {
		t.Fatal(out.String())
	}
	if len(d.Teams.(graph.TeamsAPI).Memory.Sent) != 0 {
		t.Fatal("sent")
	}
}
