package main

import "io"

// fn is one function's complexity, coverage and CRAP score.
type fn struct {
	File, Name string
	Complexity int
	Coverage   float64
	Score      float64
}

// violation is one reason the gate fails.
type violation struct {
	Kind            string // "new", "rose" or "stale"
	File, Name      string
	Score, Baseline float64
}

// crapScore is complexity² × (1 − coverage)³ + complexity, to 0.1.
func crapScore(complexity int, coverage float64) float64 { return 0 }

// parseCover reads `go tool cover -func` output into coverage by
// "file:line" of each function's start, with the module prefix removed.
func parseCover(r io.Reader, module string) (map[string]float64, error) { return nil, nil }

// parseCyclo reads gocyclo output and joins it with coverage; a function
// with no coverage data counts as uncovered.
func parseCyclo(r io.Reader, cover map[string]float64) ([]fn, error) { return nil, nil }

// parseBaseline reads "score file name" lines; # starts a comment.
func parseBaseline(r io.Reader) (map[string]float64, error) { return nil, nil }

// check compares scores above threshold with the baseline.
func check(fns []fn, baseline map[string]float64, threshold float64) []violation { return nil }
