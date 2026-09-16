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
