package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// fn is one function's complexity, coverage and CRAP score.
type fn struct {
	File, Name string
	Complexity int
	Coverage   float64
	Score      float64
}

// key identifies a function in the baseline: its file and name.
func (f fn) key() string { return f.File + " " + f.Name }

// violation is one reason the gate fails.
type violation struct {
	Kind            string // "new", "rose" or "stale"
	File, Name      string
	Score, Baseline float64
}

// Lines are matched by structure, not split on spaces, so a file path
// may contain spaces: `go tool cover -func` prints "path:line:<tab>name
// <tab>pct%", gocyclo prints "complexity pkg name path:line:col".
var (
	coverLine = regexp.MustCompile(`^(.+):(\d+):\s+\S+\s+([\d.]+)%$`)
	cycloLine = regexp.MustCompile(`^(\d+) (\S+) (\S+) (.+):(\d+):(\d+)$`)
)

// crapScore is complexity² × (1 − coverage)³ + complexity, rounded up to
// 0.1: rounding never hides a crossing of the threshold or of a baseline
// entry. The epsilon keeps exact values (110) from rounding up.
func crapScore(complexity int, coverage float64) float64 {
	c := float64(complexity)
	return math.Ceil((c*c*math.Pow(1-coverage, 3)+c)*10-1e-9) / 10
}

// parseCover reads `go tool cover -func` output into coverage by
// "file:line" of each function's start, with the module prefix removed.
func parseCover(r io.Reader, module string) (map[string]float64, error) {
	out := map[string]float64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "total:") {
			continue
		}
		m := coverLine.FindStringSubmatch(text)
		if m == nil {
			return nil, fmt.Errorf("cover: malformed line %q", sc.Text())
		}
		pct, err := strconv.ParseFloat(m[3], 64)
		if err != nil {
			return nil, fmt.Errorf("cover: %w", err)
		}
		out[strings.TrimPrefix(m[1], module+"/")+":"+m[2]] = pct / 100
	}
	return out, sc.Err()
}

// parseCyclo reads gocyclo output ("complexity pkg name file:line:col")
// and joins it with coverage; a function with no coverage data counts as
// uncovered.
func parseCyclo(r io.Reader, cover map[string]float64) ([]fn, error) {
	var out []fn
	var lines []int
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}
		m := cycloLine.FindStringSubmatch(text)
		if m == nil {
			return nil, fmt.Errorf("gocyclo: malformed line %q", sc.Text())
		}
		c, _ := strconv.Atoi(m[1])    // the pattern matched digits
		line, _ := strconv.Atoi(m[5]) // the pattern matched digits
		file := m[4]
		cov := cover[file+":"+m[5]]
		out = append(out, fn{File: file, Name: m[3], Complexity: c, Coverage: cov, Score: crapScore(c, cov)})
		lines = append(lines, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	numberSameNames(out, lines)
	return out, nil
}

// numberSameNames gives functions that share a file and name (several
// init funcs) distinct keys by position in the file: the first keeps the
// name, the rest get #2, #3 and so on. Position, not score, so a rising
// function can't swap numbers with one that dropped.
func numberSameNames(fns []fn, lines []int) {
	groups := map[string][]int{}
	for i, f := range fns {
		groups[f.key()] = append(groups[f.key()], i)
	}
	for _, idx := range groups {
		if len(idx) < 2 {
			continue
		}
		sort.SliceStable(idx, func(a, b int) bool { return lines[idx[a]] < lines[idx[b]] })
		for n, i := range idx[1:] {
			fns[i].Name = fmt.Sprintf("%s#%d", fns[i].Name, n+2)
		}
	}
}

// parseBaseline reads "score file name" lines; # starts a comment.
func parseBaseline(r io.Reader) (map[string]float64, error) {
	out := map[string]float64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("baseline: malformed line %q", line)
		}
		score, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return nil, fmt.Errorf("baseline: %w", err)
		}
		if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 {
			return nil, fmt.Errorf("baseline: score must be a finite number >= 0 in %q", line)
		}
		out[fields[1]+" "+fields[2]] = score
	}
	return out, sc.Err()
}

// check compares scores above threshold with the baseline. A function
// over the threshold must be in the baseline and must not score higher
// than its entry; an entry whose function is now at or under the
// threshold, or gone, is stale and must be removed.
func check(fns []fn, baseline map[string]float64, threshold float64) []violation {
	var out []violation
	seen := map[string]bool{}
	for _, f := range fns {
		seen[f.key()] = true
		base, listed := baseline[f.key()]
		switch {
		case f.Score <= threshold && listed:
			out = append(out, violation{Kind: "stale", File: f.File, Name: f.Name, Score: f.Score, Baseline: base})
		case f.Score <= threshold:
		case !listed:
			out = append(out, violation{Kind: "new", File: f.File, Name: f.Name, Score: f.Score})
		case f.Score > base:
			out = append(out, violation{Kind: "rose", File: f.File, Name: f.Name, Score: f.Score, Baseline: base})
		}
	}
	for k, base := range baseline {
		if !seen[k] {
			file, name, _ := strings.Cut(k, " ")
			out = append(out, violation{Kind: "stale", File: file, Name: name, Baseline: base})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Name < out[j].Name
	})
	return out
}
