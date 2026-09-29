package cli

import (
	"strconv"

	"github.com/masonhuemmer/m365/internal/domain"
)

// mailVerb is one mail verb: its handler, and its help text when it
// answers --help (verbs without help keep their existing handling).
type mailVerb struct {
	run  func(args []string, d Deps, sess domain.Session, format string) int
	help string
}

var mailVerbs = map[string]mailVerb{
	"list":            {mailList, mailListHelp},
	"get":             {mailGet, ""},
	"thread":          {mailThread, ""},
	"watch":           {mailWatch, mailWatchHelp},
	"send":            {mailSend, mailSendHelp},
	"reply":           {mailReply, mailReplyHelp},
	"attachments":     {mailAttachments, ""},
	"save-attachment": {mailSave, ""},
}

func runMail(args []string, d Deps, format string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeHelp(d.Stdout, mailListHelp+mailWatchHelp+mailSendHelp)
	}
	verb, args := args[0], args[1:]
	sess, err := session(d)
	if err != nil {
		return fail(d, err)
	}
	v, ok := mailVerbs[verb]
	if !ok {
		return fail(d, domain.Usagef("unknown mail verb %q", verb))
	}
	if v.help != "" && hasHelp(args) {
		return writeHelp(d.Stdout, v.help)
	}
	return v.run(args, d, sess, format)
}

func hasHelp(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }
