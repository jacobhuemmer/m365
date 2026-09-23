package cli

import (
	"flag"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/teams"
	"github.com/masonhuemmer/m365/internal/domain"
)

func teamsFind(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("teams find", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	group := fsset.Bool("group", false, "")
	top := fsset.Int("top", 0, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	qraw := strings.TrimSpace(strings.Join(fsset.Args(), " "))
	q, err := teams.ParseQuery(qraw, *group)
	if err != nil {
		return fail(d, err)
	}
	r, err := teams.FindMapped(ctx(), d.Teams, d.ChatMap, sess, q, *top)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, r)
}
