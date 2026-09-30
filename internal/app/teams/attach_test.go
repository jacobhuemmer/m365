package teams

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestAttachOversize(t *testing.T) {
	st := &stub{}
	_, err := Send(context.Background(), st, sess(), SendInput{
		ChatID: "c", Text: "t",
		Files: []domain.OutboundFile{{Name: "x", Size: domain.MaxAttachBytes + 1}},
	})
	if domain.ExitOf(err) != domain.ExitUsage || st.sent != 0 {
		t.Fatal(err)
	}
}

// Teams file sharing is not implemented: a live send must fail rather than post
// text only and report success. Dry-run still previews the files.
func TestLiveSendWithFilesIsRejected(t *testing.T) {
	files := []domain.OutboundFile{{Path: "/tmp/a.pdf", Name: "a.pdf", Size: 10}}
	st := &stub{}
	_, err := Send(context.Background(), st, sess(), SendInput{ChatID: "c", Text: "t", Files: files})
	if domain.ExitOf(err) != domain.ExitUsage || st.sent != 0 {
		t.Fatalf("live send with files: err=%v sent=%d", err, st.sent)
	}
	out, err := Send(context.Background(), st, sess(), SendInput{ChatID: "c", Text: "t", Files: files, DryRun: true})
	if err != nil || out.(map[string]any)["dry_run"] != true {
		t.Fatalf("dry-run with files: out=%v err=%v", out, err)
	}
}
