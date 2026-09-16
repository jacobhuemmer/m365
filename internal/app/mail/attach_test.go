package mail

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestAttachCapsAndNoUploadOnDryRun(t *testing.T) {
	st := &stub{}
	big := domain.OutboundFile{Name: "big.bin", Size: domain.MaxAttachBytes + 1}
	_, err := Send(context.Background(), st, sessMail(), SendInput{
		To: []string{"a@b.c"}, Subject: "t", Body: "b", Files: []domain.OutboundFile{big},
	})
	if domain.ExitOf(err) != domain.ExitUsage || st.sent != 0 {
		t.Fatal(err)
	}
	empty := domain.OutboundFile{Name: "e", Size: 0}
	_, err = Send(context.Background(), st, sessMail(), SendInput{
		To: []string{"a@b.c"}, Subject: "t", Body: "b", Files: []domain.OutboundFile{empty},
	})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}
