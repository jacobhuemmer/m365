package msgbody

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Converter output from clean input must never trip a format rule (FR-012).
func TestCleanInputHasNoProblems(t *testing.T) {
	files, err := filepath.Glob("testdata/clean/*")
	if err != nil || len(files) == 0 {
		t.Fatalf("no fixtures: %v", err)
	}
	for _, f := range files {
		src, err := os.ReadFile(f) // #nosec G304 -- test fixture path
		if err != nil {
			t.Fatal(err)
		}
		mode := Plain
		if strings.HasSuffix(f, ".md") {
			mode = Markdown
		}
		for _, target := range []Target{Mail, Teams} {
			r := Render(mode, target, string(src))
			if r.Content == "" {
				t.Fatalf("%s: rendered nothing", f)
			}
			if p := Lint(r.Content); len(p) != 0 {
				t.Errorf("%s target %d: %v\n%s", f, target, p, r.Content)
			}
		}
	}
}
