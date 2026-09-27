package main

import (
	"reflect"
	"strings"
	"testing"
)

// Findings from the Codex review of PR #22.

// Scores round up to 0.1, so rounding never hides a crossing of the
// threshold or of a baseline entry.
func TestCrapScoreRoundsUp(t *testing.T) {
	for _, tc := range []struct {
		c    int
		cov  float64
		want float64
	}{
		{5, 0.263, 15.1}, // 15.0078
		{5, 0.087, 24.1}, // 24.026
	} {
		if got := crapScore(tc.c, tc.cov); got != tc.want {
			t.Errorf("crapScore(%d, %v) = %v, want %v", tc.c, tc.cov, got, tc.want)
		}
	}
	fns := []fn{{File: "a.go", Name: "Near", Score: crapScore(5, 0.263)}, {File: "b.go", Name: "Up", Score: crapScore(5, 0.087)}}
	got := check(fns, map[string]float64{"b.go Up": 24.0}, 15)
	want := []violation{
		{Kind: "new", File: "a.go", Name: "Near", Score: 15.1},
		{Kind: "rose", File: "b.go", Name: "Up", Score: 24.1, Baseline: 24.0},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

// Functions with the same name in one file (several init funcs) get
// their own baseline keys, numbered by score, highest first.
func TestSameNameFunctionsAreNumbered(t *testing.T) {
	in := "4 p init f.go:3:1\n10 p init f.go:20:1\n2 p other f.go:40:1\n"
	fns, err := parseCyclo(strings.NewReader(in), nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range fns {
		names = append(names, f.Name)
	}
	if want := []string{"init#2", "init", "other"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("names %v want %v", names, want)
	}
	got := check(fns, map[string]float64{"f.go init": 110}, 15)
	want := []violation{{Kind: "new", File: "f.go", Name: "init#2", Score: 20}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("a second init must not share the first one's entry: got %+v", got)
	}
}

func TestParseBaselineRejectsNonFiniteScores(t *testing.T) {
	for _, line := range []string{"NaN f.go F\n", "+Inf f.go F\n", "-1 f.go F\n"} {
		if _, err := parseBaseline(strings.NewReader(line)); err == nil {
			t.Errorf("want an error for %q", line)
		}
	}
}
