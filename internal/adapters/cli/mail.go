package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/masonhuemmer/m365/internal/adapters/fs"
	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/domain"
)

func runMail(args []string, d Deps, format string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeHelp(d.Stdout, mailListHelp+mailWatchHelp+mailSendHelp)
	}
	verb, args := args[0], args[1:]
	sess, err := session(d)
	if err != nil {
		return fail(d, err)
	}
	switch verb {
	case "list":
		if hasHelp(args) {
			return writeHelp(d.Stdout, mailListHelp)
		}
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
	case "get":
		if len(args) < 1 {
			return fail(d, domain.Usage("message id is required"))
		}
		msg, err := mail.Get(ctx(), d.Mail, sess, args[0])
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, msg)
	case "thread":
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
	case "watch":
		if hasHelp(args) {
			return writeHelp(d.Stdout, mailWatchHelp)
		}
		return mailWatch(args, d, sess, format)
	case "send":
		if hasHelp(args) {
			return writeHelp(d.Stdout, mailSendHelp)
		}
		return mailSend(args, d, sess, format)
	case "reply":
		return mailReply(args, d, sess, format)
	case "attachments":
		if len(args) < 1 {
			return fail(d, domain.Usage("message id is required"))
		}
		atts, err := mail.ListAttachments(ctx(), d.Mail, sess, args[0])
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, map[string]any{"limit": len(atts), "count": len(atts), "items": atts})
	case "save-attachment":
		return mailSave(args, d, sess, format)
	default:
		return fail(d, domain.Usagef("unknown mail verb %q", verb))
	}
}

func mailWatch(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("mail watch", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	folder := fsset.String("folder", "inbox", "")
	includeExisting := fsset.Bool("include-existing", false, "")
	classify := fsset.Bool("classify", false, "")
	var targetAddresses, targetNames []string
	fsset.Func("target-address", "", func(value string) error {
		targetAddresses = append(targetAddresses, value)
		return nil
	})
	fsset.Func("target-name", "", func(value string) error {
		targetNames = append(targetNames, value)
		return nil
	})
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	if fsset.NArg() != 0 {
		return fail(d, domain.Usage("mail watch does not accept positional arguments"))
	}
	if *classify && d.ConfigError != nil {
		return fail(d, d.ConfigError)
	}
	sink := &mailWatchSink{writer: d.Stdout, human: format == "human"}
	classificationConfig := d.Config.Experimental.MailResponseClassification
	err := mail.Watch(ctx(), d.MailChanges, d.MailThreads, d.MailClassifier, d.MailWatchState, sink, sess, mail.WatchConfig{
		ClassificationEnabled: classificationConfig.Enabled,
		ActionableThreshold:   classificationConfig.ActionableThreshold,
	}, mail.WatchQuery{
		Folder:          *folder,
		IncludeExisting: *includeExisting,
		Classify:        *classify,
		TargetAddresses: targetAddresses,
		TargetNames:     targetNames,
	})
	if err != nil {
		return fail(d, err)
	}
	return domain.ExitOK
}

type mailWatchSink struct {
	writer io.Writer
	human  bool
}

func (s *mailWatchSink) Emit(_ context.Context, event domain.MailWatchEvent) error {
	if s.human {
		_, err := fmt.Fprintf(s.writer, "%s\t%s\t%s\n", event.Received, event.From.Address, event.Subject)
		return err
	}
	encoder := json.NewEncoder(s.writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(event)
}

func mailSend(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("mail send", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	subject := fsset.String("subject", "", "")
	body := fsset.String("body", "", "")
	bodyFile := fsset.String("body-file", "", "")
	html := fsset.Bool("html", false, "")
	dry := fsset.Bool("dry-run", false, "")
	var to, cc, attach []string
	fsset.Func("to", "", func(s string) error { to = append(to, s); return nil })
	fsset.Func("cc", "", func(s string) error { cc = append(cc, s); return nil })
	fsset.Func("attach", "", func(s string) error { attach = append(attach, s); return nil })
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	b, err := readBody(*body, *bodyFile, d.Stdin)
	if err != nil {
		return fail(d, err)
	}
	files, err := loadAttach(attach)
	if err != nil {
		return fail(d, err)
	}
	out, err := mail.Send(ctx(), d.Mail, sess, mail.SendInput{
		To: to, CC: cc, Subject: *subject, Body: b, HTML: *html, DryRun: *dry, Files: files,
	})
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, out)
}

func mailReply(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("mail reply", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	body := fsset.String("body", "", "")
	bodyFile := fsset.String("body-file", "", "")
	all := fsset.Bool("all", false, "")
	html := fsset.Bool("html", false, "")
	dry := fsset.Bool("dry-run", false, "")
	var attach []string
	fsset.Func("attach", "", func(s string) error { attach = append(attach, s); return nil })
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	id := fsset.Arg(0)
	b, err := readBody(*body, *bodyFile, d.Stdin)
	if err != nil {
		return fail(d, err)
	}
	files, err := loadAttach(attach)
	if err != nil {
		return fail(d, err)
	}
	out, err := mail.Reply(ctx(), d.Mail, sess, mail.ReplyInput{
		ID: id, Body: b, All: *all, HTML: *html, DryRun: *dry, Files: files,
	})
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, out)
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

func loadAttach(paths []string) ([]domain.OutboundFile, error) {
	var files []domain.OutboundFile
	for _, p := range paths {
		f, err := fs.ReadAttach(p)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
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
