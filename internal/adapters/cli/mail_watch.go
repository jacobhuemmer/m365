package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/masonhuemmer/m365/internal/app/mail"
	"github.com/masonhuemmer/m365/internal/domain"
)

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
	if *classify && d.MailClassifierError != nil {
		return fail(d, d.MailClassifierError)
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
