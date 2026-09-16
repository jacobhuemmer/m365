package mail

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestThreadAndEmpty(t *testing.T) {
	st := &stub{th: domain.MailThread{Count: 0, Items: []domain.MailMessage{}}}
	th, err := Thread(context.Background(), st, sessMail(), "msg-1", false)
	if err != nil || th.Count != 0 {
		t.Fatalf("%+v %v", th, err)
	}
}
