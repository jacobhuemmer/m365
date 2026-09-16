package graph

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/teams"
)

func TestTeamSend(t *testing.T) {
	m := TeamsAPI{Memory: Seed()}
	id, err := m.Send(context.Background(), teams.SendInput{ChatID: "chat-1", Text: "hi"})
	if err != nil || id == "" {
		t.Fatal(err)
	}
}
