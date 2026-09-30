package teams

import (
	"context"
	"reflect"
	"strings"
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

var pdf = []domain.OutboundFile{{Path: "/tmp/a.pdf", Name: "a.pdf", Size: 10}}

func selfSess() domain.Session {
	se := sess()
	se.Account = "self@example.com"
	return se
}

func chatOf(members ...domain.Person) *stub {
	return &stub{chat: domain.Chat{ID: "c", Type: "group", Members: members}}
}

func TestSendWithFilesSharesWithOtherMembersOnly(t *testing.T) {
	st := chatOf(
		domain.Person{Name: "Me", Address: "Self@Example.com"},
		domain.Person{Name: "Ajay", Address: "ajay@example.com"},
		domain.Person{Name: "Bo", Address: "bo@example.com"},
	)
	if _, err := Send(context.Background(), st, selfSess(), SendInput{ChatID: "c", Text: "t", Files: pdf}); err != nil {
		t.Fatal(err)
	}
	if st.sent != 1 || !reflect.DeepEqual(st.last.ShareWith, []string{"ajay@example.com", "bo@example.com"}) {
		t.Fatalf("sent=%d share_with=%v", st.sent, st.last.ShareWith)
	}
}

func TestSendWithFilesFailsBeforeSendWhenMemberHasNoEmail(t *testing.T) {
	st := chatOf(domain.Person{Address: "self@example.com"}, domain.Person{Name: "Guest Gail"})
	_, err := Send(context.Background(), st, selfSess(), SendInput{ChatID: "c", Text: "t", Files: pdf})
	if domain.ExitOf(err) != domain.ExitUsage || st.sent != 0 || !strings.Contains(err.Error(), "Guest Gail") {
		t.Fatalf("err=%v sent=%d", err, st.sent)
	}
}

func TestSendWithFilesFailsWhenChatMembersUnknown(t *testing.T) {
	st := chatOf()
	_, err := Send(context.Background(), st, selfSess(), SendInput{ChatID: "c", Text: "t", Files: pdf})
	if domain.ExitOf(err) != domain.ExitUsage || st.sent != 0 {
		t.Fatalf("err=%v sent=%d", err, st.sent)
	}
}

func TestSendWithFilesToNotesSharesWithNobody(t *testing.T) {
	st := &stub{}
	if _, err := Send(context.Background(), &notesStub{st}, selfSess(), SendInput{ChatID: SelfChatID, Text: "t", Files: pdf}); err != nil {
		t.Fatal(err)
	}
	if st.sent != 1 || len(st.last.ShareWith) != 0 {
		t.Fatalf("sent=%d share_with=%v", st.sent, st.last.ShareWith)
	}
}

// notesStub has no chat to look up, like Graph and 48:notes; sending must not need one.
type notesStub struct{ *stub }

func (n *notesStub) GetChat(context.Context, string) (domain.Chat, error) {
	return domain.Chat{}, domain.NotFound("no such chat")
}

func TestDryRunWithFilesShowsWhoGetsAccess(t *testing.T) {
	st := chatOf(domain.Person{Address: "self@example.com"}, domain.Person{Address: "ajay@example.com"})
	out, err := Send(context.Background(), st, selfSess(), SendInput{ChatID: "c", Text: "t", Files: pdf, DryRun: true})
	if err != nil || st.sent != 0 {
		t.Fatalf("err=%v sent=%d", err, st.sent)
	}
	if got := out.(map[string]any)["share_with"]; !reflect.DeepEqual(got, []string{"ajay@example.com"}) {
		t.Fatalf("share_with=%v", got)
	}
}
