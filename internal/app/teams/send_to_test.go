package teams

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestSendToUniqueAndSeveral(t *testing.T) {
	st := &chatMem{chats: sampleChats()}
	sess := domain.Session{SignedIn: true, SessionUsable: true, TeamsConsented: true}
	out, err := Send(context.Background(), st, sess, SendInput{To: "Ajay", Text: "ping", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	m := out.(map[string]any)
	if m["dry_run"] != true || m["chat_id"] != "chat-ajay" || m["to"] != "Ajay" {
		t.Fatal(m)
	}
	st.chats = append(st.chats, domain.Chat{
		ID: "chat-ajay-b", Type: "oneOnOne",
		Members: []domain.Person{{Name: "Ajay Singh", Address: "ajay.singh@example.com"}},
	})
	_, err = Send(context.Background(), st, sess, SendInput{To: "Ajay", Text: "ping"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestSendToNoneAndBoth(t *testing.T) {
	st := &chatMem{chats: sampleChats()}
	sess := domain.Session{SignedIn: true, SessionUsable: true, TeamsConsented: true}
	_, err := Send(context.Background(), st, sess, SendInput{To: "Nobody", Text: "x"})
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
	_, err = Send(context.Background(), st, sess, SendInput{To: "Ajay", ChatID: "chat-1", Text: "x"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestSendNoteToSelfUsesNotes(t *testing.T) {
	st := &chatMem{chats: sampleChats()}
	sess := domain.Session{SignedIn: true, SessionUsable: true, TeamsConsented: true, Account: "Mason.Huemmer@Sesami.io"}
	out, err := Send(context.Background(), st, sess, SendInput{Text: "ping", NoteToSelf: true, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	m := out.(map[string]any)
	if m["chat_id"] != SelfChatID || m["dry_run"] != true {
		t.Fatal(m)
	}
	_, err = Send(context.Background(), st, sess, SendInput{Text: "ping", NoteToSelf: true, To: "Ajay"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	_, err = Send(context.Background(), st, sess, SendInput{Text: "ping", NoteToSelf: true, ChatID: "chat-1"})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestExactRecipientDoesNotMatchMemberWithoutAddressByID(t *testing.T) {
	st := &chatMem{chats: []domain.Chat{{ID: "chat-1", Type: "oneOnOne", Members: []domain.Person{
		{ID: "01234567-89ab-cdef-0123-456789abcdef"},
		{ID: "fedcba98-7654-3210-fedc-ba9876543210", Address: "other@example.com"},
	}}}}
	sess := domain.Session{SignedIn: true, SessionUsable: true, TeamsConsented: true, Account: "self@example.com"}
	_, err := Send(context.Background(), st, sess, SendInput{To: "01234567-89ab-cdef-0123-456789abcdef", Text: "ping", ExactRecipient: true})
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatalf("addressless member ID matched: %v", err)
	}
}

func TestExactRecipientIDRequiresIdentifiableSelfMember(t *testing.T) {
	selfID := "01234567-89ab-cdef-0123-456789abcdef"
	st := &chatMem{chats: []domain.Chat{{ID: "chat-1", Type: "oneOnOne", Members: []domain.Person{
		{ID: selfID, Address: "alias@example.com"},
		{ID: "fedcba98-7654-3210-fedc-ba9876543210", Address: "other@example.com"},
	}}}}
	sess := domain.Session{SignedIn: true, SessionUsable: true, TeamsConsented: true, Account: "self@example.com"}
	_, err := Send(context.Background(), st, sess, SendInput{To: selfID, Text: "ping", ExactRecipient: true})
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatalf("self member ID matched despite unknown account mapping: %v", err)
	}
}
