//go:build !race

package msgbody

import (
	"testing"
	"time"
)

// Not built under -race: the race detector slows this code roughly
// tenfold, so a wall-clock limit there measures the runner, not the code.
// Generous deadline for CI; the benchmark reports the real cost
// (target under 100 ms, plan.md Performance Goals).
func TestLargeBodyDeadline(t *testing.T) {
	src := largeBody()
	start := time.Now()
	renderLintText(src)
	if d := time.Since(start); d > time.Second {
		t.Fatalf("render+lint+text of 256 KiB took %v, want under 1s", d)
	}
}
