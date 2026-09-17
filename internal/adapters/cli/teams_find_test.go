package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestTeamsFindAjay(t *testing.T) {
	d, out, errw := testDeps()
	if c := Run([]string{"m365", "teams", "find", "Ajay"}, d); c != domain.ExitAuth {
		t.Fatal(c)
	}
	login(t, d)
	out.Reset()
	if c := Run([]string{"m365", "teams", "find", "Ajay"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var r map[string]any
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatal(err, out.String())
	}
	if r["intent"] != "person" || r["count"].(float64) != 1 {
		t.Fatal(out.String())
	}
	items := r["items"].([]any)
	if items[0].(map[string]any)["id"] != "chat-ajay" {
		t.Fatal(out.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "teams", "find", "--top", "21", "Ajay"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "teams", "find"}, d); c != domain.ExitUsage {
		t.Fatal(c)
	}
	out.Reset()
	if c := Run([]string{"m365", "chat", "find", "Ajay"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "teams", "find", "--group", "NOC"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	s := out.String()
	if !strings.Contains(s, "chat-noc-dev") && !strings.Contains(s, "chat-2") {
		t.Fatal(s)
	}
}

func TestTeamsFindHelp(t *testing.T) {
	d, out, _ := testDeps()
	Run([]string{"m365", "teams", "find", "--help"}, d)
	s := out.String()
	for _, w := range []string{"Ajay", "--group NOC", "default 10", "max 20"} {
		if !strings.Contains(s, w) {
			t.Fatal(s)
		}
	}
}
