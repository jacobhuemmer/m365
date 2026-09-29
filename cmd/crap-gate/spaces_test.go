package main

import (
	"reflect"
	"strings"
	"testing"
)

// Codex review of PR #22 (round 2): a Go file name may contain a space.
func TestParsersAcceptPathsWithSpaces(t *testing.T) {
	cover, err := parseCover(strings.NewReader(module+"/internal/a b.go:3:\t\tF\t\t50.0%\n"), module)
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]float64{"internal/a b.go:3": 0.5}; !reflect.DeepEqual(cover, want) {
		t.Fatalf("cover %v want %v", cover, want)
	}
	fns, err := parseCyclo(strings.NewReader("4 p F internal/a b.go:3:1\n"), cover)
	if err != nil {
		t.Fatal(err)
	}
	want := []fn{{File: "internal/a b.go", Name: "F", Complexity: 4, Coverage: 0.5, Score: crapScore(4, 0.5)}}
	if !reflect.DeepEqual(fns, want) {
		t.Fatalf("fns %+v want %+v", fns, want)
	}
}

// The baseline can record a path with spaces: the score is the first
// field and the function name the last; the file is everything between.
func TestParseBaselineAcceptsPathsWithSpaces(t *testing.T) {
	got, err := parseBaseline(strings.NewReader("20.0 internal/a b.go F\n"))
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]float64{"internal/a b.go F": 20}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
