package cli

import "github.com/masonhuemmer/m365/internal/domain"

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
