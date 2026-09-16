package teams

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestMessagesEmpty(t *testing.T) {
	st := &stub{msgs: domain.ChatMessagePage{Items: nil, Count: 0}}
	p, err := Messages(context.Background(), st, sess(), MessageQuery{ChatID: "chat-1"})
	if err != nil || p.Count != 0 {
		t.Fatal(err)
	}
	st.err = domain.NotFound("chat")
	_, err = Messages(context.Background(), st, sess(), MessageQuery{ChatID: "missing"})
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
}
