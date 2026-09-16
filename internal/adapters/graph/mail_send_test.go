package graph

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/mail"
)

func TestSendRecords(t *testing.T) {
	m := Seed()
	id, err := m.Send(context.Background(), mail.SendInput{To: []string{"a@b.c"}, Subject: "t", Body: "b"})
	if err != nil || id == "" {
		t.Fatal(err)
	}
}
