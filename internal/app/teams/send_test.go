package teams

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestDryRunDoesNotSend(t *testing.T) {
	st := &stub{}
	out, err := Send(context.Background(), st, sess(), SendInput{ChatID: "chat-1", Text: "hi", DryRun: true})
	if err != nil || st.sent != 0 {
		t.Fatal(err)
	}
	if out.(map[string]any)["dry_run"] != true {
		t.Fatal(out)
	}
	_, err = Send(context.Background(), st, sess(), SendInput{ChatID: "chat-1"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	_, err = Send(context.Background(), st, sess(), SendInput{ChatID: "chat-1", Text: "hi"})
	if err != nil || st.sent != 1 {
		t.Fatal(err)
	}
}
