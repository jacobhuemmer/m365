package cli

import (
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestMailDryRun(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	out.Reset()
	c := Run([]string{"m365", "mail", "send", "--dry-run", "--to", "user@example.com", "--subject", "t", "--body", "b"}, d)
	if c != 0 {
		t.Fatal(errw.String())
	}
	var m map[string]any
	_ = json.Unmarshal(out.Bytes(), &m)
	if m["dry_run"] != true {
		t.Fatal(out.String())
	}
	mem := d.Mail.(graph.MailAPI).Memory
	if len(mem.Sent) != 0 {
		t.Fatal("sent")
	}
}

func TestMailNoteToSelfDryRun(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	out.Reset()
	c := Run([]string{"m365", "mail", "send", "--note-to-self", "--body", "remember", "--dry-run"}, d)
	if c != 0 {
		t.Fatal(errw.String())
	}
	var m map[string]any
	if err := json.Unmarshal(out.Bytes(), &m); err != nil || m["dry_run"] != true {
		t.Fatal(out.String(), err)
	}
	out.Reset()
	c = Run([]string{"m365", "mail", "send", "--note-to-self", "--to", "a@b.c", "--body", "x"}, d)
	if c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
}
