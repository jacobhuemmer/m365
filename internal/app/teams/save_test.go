package teams

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestSaveRefuse(t *testing.T) {
	st := &stub{}
	_, err := Save(context.Background(), st, sess(), "chat-1", "cmsg-1", "att-1", "", false, nil)
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}
