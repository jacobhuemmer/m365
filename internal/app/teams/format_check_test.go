package teams

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

func TestSendFormatProblems(t *testing.T) {
	st := &sendRec{chatMem: &chatMem{chats: sampleChats()}}
	in := SendInput{ChatID: "chat-1", Text: "<p>hi", HTML: true}
	dry := in
	dry.DryRun = true
	out, err := SendMapped(context.Background(), st, nil, teamsSess(), dry)
	if err != nil {
		t.Fatal(err)
	}
	want := msgbody.Problems{{Rule: msgbody.RuleBrokenHTML, Detail: "unclosed <p>"}}
	if got, _ := out.(map[string]any)["format_problems"].(msgbody.Problems); !reflect.DeepEqual(got, want) {
		t.Fatalf("format_problems %#v want %#v", got, want)
	}
	_, err = SendMapped(context.Background(), st, nil, teamsSess(), in)
	var de *domain.Error
	wantErr := domain.Error{Class: domain.ClassUsage, Message: "1 format problem: broken-html: unclosed <p>", Hint: "run with --preview to see them"}
	if !errors.As(err, &de) || *de != wantErr {
		t.Fatalf("got %#v want %#v", err, wantErr)
	}
	if st.n != 0 {
		t.Fatalf("sent %d", st.n)
	}
}
