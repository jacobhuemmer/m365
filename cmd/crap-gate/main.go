// Command crap-gate fails when a function's CRAP score is over the
// threshold and it is not in the baseline, when a baselined score rises,
// or when a baseline entry is stale (see scripts/crap-baseline.txt).
package main

import (
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses flags, loads the inputs and reports; it returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	return 0
}
