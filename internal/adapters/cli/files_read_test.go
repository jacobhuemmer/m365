package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestFilesRootListGet(t *testing.T) {
	d, out, errw := testDeps()
	loginAll(t, &d)
	out.Reset()
	if c := Run([]string{"m365", "files", "root"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "files", "list"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	var page map[string]any
	if err := json.Unmarshal(out.Bytes(), &page); err != nil {
		t.Fatal(err, out.String())
	}
	if page["limit"].(float64) != 20 {
		t.Fatalf("limit %v", page["limit"])
	}
	if strings.Contains(out.String(), "synthetic-ok") {
		t.Fatal("bytes")
	}
	out.Reset()
	if c := Run([]string{"m365", "files", "get", "file-1"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	out.Reset()
	if c := Run([]string{"m365", "files", "get", "missing"}, d); c != domain.ExitNotFound {
		t.Fatal(c, errw.String())
	}
}
