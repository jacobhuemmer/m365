package cli

import (
	"strconv"
	"strings"

	"github.com/masonhuemmer/m365/internal/domain"
)

// markdownFlag validates --format against --html and reports whether the
// body is markdown. Only md is a format; --html is its own mode.
func markdownFlag(html bool, format string) (bool, error) {
	switch {
	case format == "":
		return false, nil
	case format != "md":
		return false, domain.Usagef("unsupported --format %q; only md", format)
	case html:
		return false, domain.Usage("use --html or --format md, not both")
	}
	return true, nil
}

// previewRequested reports a true --preview in any form the flag package
// accepts: --preview, -preview, --preview=true, -preview=1.
func previewRequested(args []string) bool {
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			continue
		}
		name, val, hasVal := strings.Cut(strings.TrimLeft(a, "-"), "=")
		if name != "preview" {
			continue
		}
		if !hasVal {
			return true
		}
		if on, err := strconv.ParseBool(val); err == nil && on {
			return true
		}
	}
	return false
}
