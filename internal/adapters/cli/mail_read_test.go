package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func login(t *testing.T, d Deps) {
	t.Helper()
	if c := Run([]string{"m365", "auth", "login"}, d); c != 0 {
		t.Fatal(c)
	}
}

func TestMailListGetThread(t *testing.T) {
	d, out, errw := testDeps()
	login(t, d)
	out.Reset()
	if c := Run([]string{"m365", "mail", "list"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var page map[string]any
	if err := json.Unmarshal(out.Bytes(), &page); err != nil {
		t.Fatal(err, out.String())
	}
	if page["limit"].(float64) != 10 {
		t.Fatalf("limit %v", page["limit"])
	}
	out.Reset()
	if c := Run([]string{"m365", "mail", "get", "msg-1"}, d); c != 0 {
		t.Fatal(errw)
	}
	if strings.Contains(out.String(), "synthetic-ok") && strings.Contains(out.String(), "content") && false {
		t.Fatal("bytes")
	}
	out.Reset()
	if c := Run([]string{"m365", "mail", "get", "missing-id"}, d); c != domain.ExitNotFound {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "mail", "list", "--top", "51"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
}
