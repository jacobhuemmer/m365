package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
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

// crapScore is complexity² × (1 − coverage)³ + complexity, to 0.1.
func crapScore(complexity int, coverage float64) float64 {
	c := float64(complexity)
	return math.Round((c*c*math.Pow(1-coverage, 3)+c)*10) / 10
}

// parseCover reads `go tool cover -func` output into coverage by
// "file:line" of each function's start, with the module prefix removed.
func parseCover(r io.Reader, module string) (map[string]float64, error) {
	out := map[string]float64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 || fields[0] == "total:" {
			continue
		}
		if len(fields) < 3 || !strings.HasSuffix(fields[len(fields)-1], "%") {
			return nil, fmt.Errorf("cover: malformed line %q", sc.Text())
		}
		file, line, ok := splitPos(strings.TrimPrefix(fields[0], module+"/"))
		if !ok {
			return nil, fmt.Errorf("cover: malformed position in %q", sc.Text())
		}
		pct, err := strconv.ParseFloat(strings.TrimSuffix(fields[len(fields)-1], "%"), 64)
		if err != nil {
			return nil, fmt.Errorf("cover: %w", err)
		}
		out[file+":"+line] = pct / 100
	}
	return out, sc.Err()
}

// parseCyclo reads gocyclo output ("complexity pkg name file:line:col")
// and joins it with coverage; a function with no coverage data counts as
// uncovered.
func parseCyclo(r io.Reader, cover map[string]float64) ([]fn, error) {
	var out []fn
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 4 {
			return nil, fmt.Errorf("gocyclo: malformed line %q", sc.Text())
		}
		c, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, fmt.Errorf("gocyclo: %w", err)
		}
		file, line, ok := splitPos(fields[3])
		if !ok {
			return nil, fmt.Errorf("gocyclo: malformed position in %q", sc.Text())
		}
		cov := cover[file+":"+line]
		out = append(out, fn{File: file, Name: fields[2], Complexity: c, Coverage: cov, Score: crapScore(c, cov)})
	}
	return out, sc.Err()
}

// splitPos splits "file:line" or "file:line:col" into file and line.
func splitPos(pos string) (file, line string, ok bool) {
	parts := strings.Split(pos, ":")
	if len(parts) < 2 || parts[0] == "" {
		return "", "", false
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return "", "", false
	}
	return parts[0], parts[1], true
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
