package msgbody

import (
	"strings"
	"testing"
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

func BenchmarkRenderLintText(b *testing.B) {
	src := largeBody()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderLintText(src)
	}
}
