package msgbody

import (
	"strings"
	"testing"
	"time"
)

// largeBody is a 256 KiB synthetic markdown body.
func largeBody() string {
	unit := "## Section\n\nSome **bold** text with a [link](https://example.com) and `code` < & >.\n\n- one\n- two\n\n```\nfn()\n```\n\n"
	return strings.Repeat(unit, 256*1024/len(unit)+1)[:256*1024]
}

func renderLintText(src string) {
	r := Render(Markdown, Teams, src)
	_ = Lint(r.Content)
	_ = Text(r.Content, 76)
}

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

func BenchmarkRenderLintText(b *testing.B) {
	src := largeBody()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderLintText(src)
	}
}
