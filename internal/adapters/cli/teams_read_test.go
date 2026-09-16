package cli

import (
	"encoding/json"
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
