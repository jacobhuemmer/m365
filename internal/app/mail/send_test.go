package mail

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestDryRunDoesNotSend(t *testing.T) {
	st := &stub{}
	out, err := Send(context.Background(), st, sessMail(), SendInput{
		To: []string{"user@example.com"}, Subject: "t", Body: "b", DryRun: true,
		Files: []domain.OutboundFile{{Name: "note.txt", Size: 12}},
	})
	if err != nil {
		t.Fatal(err)
	}
	m := out.(map[string]any)
	if m["dry_run"] != true || st.sent != 0 {
		t.Fatalf("sent=%d out=%v", st.sent, out)
	}
	_, err = Send(context.Background(), st, sessMail(), SendInput{Subject: "t", Body: "b"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	_, err = Send(context.Background(), st, sessMail(), SendInput{
		To: []string{"a@b.c"}, Subject: "t", Body: "b",
	})
	if err != nil || st.sent != 1 {
		t.Fatalf("real send %d %v", st.sent, err)
	}
}

func TestNoteToSelfUsesAccount(t *testing.T) {
	st := &stub{}
	sess := sessMail()
	sess.Account = "Mason.Huemmer@Sesami.io"
	out, err := Send(context.Background(), st, sess, SendInput{Body: "remember", NoteToSelf: true, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	m := out.(map[string]any)
	if m["dry_run"] != true {
		t.Fatal(m)
	}
	_, err = Send(context.Background(), st, sess, SendInput{Body: "remember", NoteToSelf: true})
	if err != nil || st.sent != 1 {
		t.Fatalf("sent=%d err=%v", st.sent, err)
	}
	if len(st.last.To) != 1 || st.last.To[0] != sess.Account || st.last.Subject != "Note to self" {
		t.Fatalf("%+v", st.last)
	}
	_, err = Send(context.Background(), st, sess, SendInput{
		To: []string{"other@example.com"}, Body: "x", NoteToSelf: true,
	})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}
