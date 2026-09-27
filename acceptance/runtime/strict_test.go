package runtime

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Codex review of PR #19 (SDO-566).

// withHandlers registers handlers for one test and restores the registry.
func withHandlers(t *testing.T, prefixes ...string) {
	t.Helper()
	saved := handlers
	t.Cleanup(func() { handlers = saved })
	handlers = nil
	for _, p := range prefixes {
		Register(p, func(*World, string) error { return nil })
	}
}

func writeFeature(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "x.feature")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseReadsButAndStarSteps(t *testing.T) {
	p := writeFeature(t, "Feature: f\n  Scenario: s\n    Given a\n    But b\n    * c\n")
	f, err := Parse(p)
	if err != nil {
		t.Fatal(err)
	}
	want := []Step{{Kind: "Given", Text: "a"}, {Kind: "But", Text: "b"}, {Kind: "*", Text: "c"}}
	if !reflect.DeepEqual(f.Scenarios[0].Steps, want) {
		t.Fatalf("steps %+v want %+v", f.Scenarios[0].Steps, want)
	}
}

func TestParseRejectsUnsupportedSyntax(t *testing.T) {
	cases := map[string]string{
		"background":           "Feature: f\n  Background:\n    Given a\n  Scenario: s\n    Given b\n",
		"scenario outline":     "Feature: f\n  Scenario Outline: s\n    Given <x>\n  Examples:\n    | x |\n",
		"rule":                 "Feature: f\n  Rule: r\n  Scenario: s\n    Given a\n",
		"step before scenario": "Feature: f\n  Given a\n  Scenario: s\n    Given b\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(writeFeature(t, body)); err == nil {
				t.Fatal("want an error: the runner cannot check or run this syntax")
			}
		})
	}
}

func TestHandlerPrefixNeedsWordBoundary(t *testing.T) {
	withHandlers(t, "I run")
	for text, want := range map[string]bool{
		"I run":                true,
		`I run "mail list"`:    true,
		"I run-something else": false,
		"I running":            false,
	} {
		if _, ok := handlerFor(text); ok != want {
			t.Errorf("handlerFor(%q) = %v, want %v", text, ok, want)
		}
	}
}

func TestCheckFeatureReportsEveryUnmatchedStep(t *testing.T) {
	withHandlers(t, "a known step")
	p := writeFeature(t, "Feature: f\n  Scenario: one\n    Given a known step\n    When a typo\n  Scenario: two\n    Then another typo\n")
	_, err := checkFeature(p)
	if err == nil {
		t.Fatal("want an error listing unmatched steps")
	}
	for _, want := range []string{p, "one: When a typo", "two: Then another typo"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q lacks %q", err, want)
		}
	}
}

func TestCheckFeatureAcceptsFullyMatchedFeature(t *testing.T) {
	withHandlers(t, "a known step")
	p := writeFeature(t, "Feature: f\n  Scenario: one\n    Given a known step\n")
	f, err := checkFeature(p)
	if err != nil || f.Name != "f" || len(f.Scenarios) != 1 {
		t.Fatalf("got %+v, %v", f, err)
	}
}

func TestCheckFeatureReportsParseErrors(t *testing.T) {
	withHandlers(t, "a")
	if _, err := checkFeature(writeFeature(t, "Feature: f\n  Background:\n    Given a\n")); err == nil {
		t.Fatal("want the parse error")
	}
}
