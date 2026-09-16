package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFeature(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.feature")
	if err := os.WriteFile(p, []byte("Feature: f\n  Scenario: s\n    Given a\n    When b\n    Then c\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := Parse(p)
	if err != nil || f.Name != "f" || len(f.Scenarios) != 1 || len(f.Scenarios[0].Steps) != 3 {
		t.Fatalf("%+v %v", f, err)
	}
}
