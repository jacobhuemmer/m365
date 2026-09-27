package teams

import (
	"context"
	"errors"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

// sendRec records the send input; chatMem discards it.
type sendRec struct {
	*chatMem
	n    int
	last SendInput
}

func (r *sendRec) Send(_ context.Context, in SendInput) (string, error) {
	r.n++
	r.last = in
	return "id", nil
}

func teamsSess() domain.Session {
	return domain.Session{SignedIn: true, SessionUsable: true, TeamsConsented: true}
}

func TestSendRendersBeforeDryRun(t *testing.T) {
	want := msgbody.Rendered{ContentType: "html", Content: "<p>a<br><br>b</p>"}
	for _, in := range []SendInput{
		{ChatID: "chat-1", Text: "a\n\nb"},
		{To: "Ajay", Text: "a\n\nb"},
	} {
		st := &sendRec{chatMem: &chatMem{chats: sampleChats()}}
		dry := in
		dry.DryRun = true
		out, err := SendMapped(context.Background(), st, nil, teamsSess(), dry)
		if err != nil {
			t.Fatal(err)
		}
		m := out.(map[string]any)
		if got := m["rendered"]; got != want {
			t.Fatalf("%+v: rendered %#v want %#v", in, got, want)
		}
		if fp, ok := m["format_problems"].(msgbody.Problems); !ok || len(fp) != 0 {
			t.Fatalf("format_problems %#v", m["format_problems"])
		}
		if m["text"] != "a\n\nb" {
			t.Fatalf("text %v want raw", m["text"])
		}
		if _, err := SendMapped(context.Background(), st, nil, teamsSess(), in); err != nil {
			t.Fatal(err)
		}
		if st.n != 1 || st.last.Rendered != want {
			t.Fatalf("store n=%d got %#v want %#v", st.n, st.last.Rendered, want)
		}
	}
}

func TestWhitespaceTextIsMissing(t *testing.T) {
	st := &sendRec{chatMem: &chatMem{chats: sampleChats()}}
	_, err := SendMapped(context.Background(), st, nil, teamsSess(), SendInput{ChatID: "chat-1", Text: " \n ", DryRun: true})
	var de *domain.Error
	if !errors.As(err, &de) || de.Class != domain.ClassUsage || de.Message != "chat id and text are required" {
		t.Fatalf("got %v", err)
	}
	if st.n != 0 {
		t.Fatalf("sent %d", st.n)
	}
}
