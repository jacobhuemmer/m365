// Command crap-gate fails when a function's CRAP score is over the
// threshold and it is not in the baseline, when a baselined score rises,
// or when a baseline entry is stale (see scripts/crap-baseline.txt).
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses flags, loads the inputs and reports; it returns the exit
// code: 0 ok, 1 violations, 2 bad flags or unreadable input.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("crap-gate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	module := fs.String("module", "", "module path to strip from coverage positions")
	coverPath := fs.String("cover", "", "`go tool cover -func` output")
	cycloPath := fs.String("cyclo", "", "gocyclo output")
	baselinePath := fs.String("baseline", "", "baseline file")
	threshold := fs.Float64("threshold", 15, "highest CRAP score allowed without a baseline entry")
	printOffenders := fs.Bool("print-offenders", false, "print current offenders in baseline format and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	fns, err := load(*module, *coverPath, *cycloPath)
	if err != nil {
		fmt.Fprintln(stderr, "crap-gate:", err)
		return 2
	}
	if *printOffenders {
		sort.SliceStable(fns, func(i, j int) bool { return fns[i].Score > fns[j].Score })
		for _, f := range fns {
			if f.Score > *threshold {
				fmt.Fprintf(stdout, "%.1f %s %s\n", f.Score, f.File, f.Name)
			}
		}
		return 0
	}
	baseline, err := loadBaseline(*baselinePath)
	if err != nil {
		fmt.Fprintln(stderr, "crap-gate:", err)
		return 2
	}
	vs := check(fns, baseline, *threshold)
	for _, v := range vs {
		fmt.Fprintln(stderr, "crap:", describe(v, *threshold, *baselinePath))
	}
	if len(vs) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "crap: ok (%d functions, %d in the baseline)\n", len(fns), len(baseline))
	return 0
}

func load(module, coverPath, cycloPath string) ([]fn, error) {
	cf, err := os.Open(coverPath) // #nosec G304 -- path from scripts/crap.sh
	if err != nil {
		return nil, err
	}
	defer cf.Close()
	cover, err := parseCover(cf, module)
	if err != nil {
		return nil, err
	}
	yf, err := os.Open(cycloPath) // #nosec G304 -- path from scripts/crap.sh
	if err != nil {
		return nil, err
	}
	defer yf.Close()
	return parseCyclo(yf, cover)
}

func loadBaseline(path string) (map[string]float64, error) {
	f, err := os.Open(path) // #nosec G304 -- path from scripts/crap.sh
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseBaseline(f)
}

func describe(v violation, threshold float64, baselinePath string) string {
	switch {
	case v.Kind == "new":
		return fmt.Sprintf("%s %s scores %.1f (> %.0f); split it or cover it with tests", v.File, v.Name, v.Score, threshold)
	case v.Kind == "rose":
		return fmt.Sprintf("%s %s rose to %.1f (baseline %.1f); bring it back down", v.File, v.Name, v.Score, v.Baseline)
	case v.Score == 0:
		return fmt.Sprintf("%s %s is in the baseline but no longer exists; remove it from %s", v.File, v.Name, baselinePath)
	default:
		return fmt.Sprintf("%s %s now scores %.1f (baseline %.1f); remove it from %s", v.File, v.Name, v.Score, v.Baseline, baselinePath)
	}
}
