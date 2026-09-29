package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/m365/internal/adapters/graph"
)

// teams watch JSON lines keep <, > and & literal, like the other output.
func TestTeamsWatchKeepsHTMLCharacters(t *testing.T) {
	d, out, errw := testDeps()
	loginAll(t, &d)
	mem := d.Teams.(graph.TeamsAPI).Memory
	mem.Events[0].Text = "<b>&"
	var want bytes.Buffer
	enc := json.NewEncoder(&want)
	enc.SetEscapeHTML(false)
	for _, e := range mem.Events {
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
	out.Reset()
	if code := Run([]string{"m365", "teams", "watch"}, d); code != 0 {
		t.Fatalf("exit %d: %s", code, errw.String())
	}
	if out.String() != want.String() {
		t.Fatalf("teams watch\ngot  %q\nwant %q", out.String(), want.String())
	}
}
