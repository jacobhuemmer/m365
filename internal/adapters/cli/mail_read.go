package cli

import (
	"flag"
	"strings"

	"github.com/masonhuemmer/m365/internal/adapters/fs"
	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/domain"
)

func mailList(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("mail list", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	folder := fsset.String("folder", "inbox", "")
	unread := fsset.Bool("unread", false, "")
	search := fsset.String("search", "", "")
	top := fsset.Int("top", 0, "")
	page := fsset.String("page-token", "", "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	p, err := mail.List(ctx(), d.Mail, sess, mail.ListQuery{
		Folder: *folder, Unread: *unread, Search: *search, Top: *top, PageToken: *page,
	})
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, p)
}

func mailGet(args []string, d Deps, sess domain.Session, format string) int {
	if len(args) < 1 {
		return fail(d, domain.Usage("message id is required"))
	}
	msg, err := mail.Get(ctx(), d.Mail, sess, args[0])
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, msg)
}

func mailThread(args []string, d Deps, sess domain.Session, format string) int {
	bodies := false
	id := ""
	for _, a := range args {
		if a == "--bodies" {
			bodies = true
			continue
		}
		if !strings.HasPrefix(a, "-") {
			id = a
		}
	}
	th, err := mail.Thread(ctx(), d.Mail, sess, id, bodies)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, th)
}

func mailAttachments(args []string, d Deps, sess domain.Session, format string) int {
	if len(args) < 1 {
		return fail(d, domain.Usage("message id is required"))
	}
	atts, err := mail.ListAttachments(ctx(), d.Mail, sess, args[0])
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, map[string]any{"limit": len(atts), "count": len(atts), "items": atts})
}

func mailSave(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("mail save-attachment", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	outp := fsset.String("out", "", "")
	ov := fsset.Bool("overwrite", false, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	id, att := fsset.Arg(0), fsset.Arg(1)
	write := d.Write
	if write == nil {
		write = fs.WriteFile
	}
	path, err := mail.Save(ctx(), d.Mail, sess, id, att, *outp, *ov, write)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, map[string]any{"path": path})
}
