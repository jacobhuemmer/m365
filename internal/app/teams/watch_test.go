package teams

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestWatchOneShot(t *testing.T) {
	st := &stub{}
	ev, err := Watch(context.Background(), st, sess(), WatchQuery{})
	if err != nil || len(ev) != 0 {
		t.Fatal(err)
	}
	_, err = Watch(context.Background(), st, domain.SignedOut(), WatchQuery{})
	if domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
}
