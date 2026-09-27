package cli

import (
	"io"

	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

// drawPreview draws the preview box: header lines, a rule, the body text,
// then any format problems below the box.
func drawPreview(headers []string, body string, problems msgbody.Problems, width int) string {
	return ""
}

// termWidth is the terminal's column count, or 80 when w is not a terminal.
func termWidth(w io.Writer) int {
	return 0
}
