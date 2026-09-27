package mail

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
	"github.com/masonhuemmer/m365/internal/domain/msgbody"
)

func wantBlocked(t *testing.T, err error, msg string) {
	t.Helper()
	var de *domain.Error
	want := domain.Error{Class: domain.ClassUsage, Message: msg, Hint: "run with --preview to see them"}
	if !errors.As(err, &de) || *de != want {
		t.Fatalf("got %#v want %#v", err, want)
	}
}

func wantProblems(t *testing.T, out any, want msgbody.Problems) {
	t.Helper()
	got, _ := out.(map[string]any)["format_problems"].(msgbody.Problems)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("format_problems %#v want %#v", got, want)
	}
}

var unclosedP = msgbody.Problems{{Rule: msgbody.RuleBrokenHTML, Detail: "unclosed <p>"}}

func TestSendFormatProblems(t *testing.T) {
	st := &stub{}
	in := SendInput{To: []string{"user@example.com"}, Subject: "t", Body: "<p>hi", HTML: true}
	dry := in
	dry.DryRun = true
	out, err := Send(context.Background(), st, sessMail(), dry)
	if err != nil {
		t.Fatal(err)
	}
	wantProblems(t, out, unclosedP)
	_, err = Send(context.Background(), st, sessMail(), in)
	wantBlocked(t, err, "1 format problem: broken-html: unclosed <p>")
	_, err = Send(context.Background(), st, sessMail(), SendInput{To: []string{"user@example.com"}, Subject: "**urgent**", Body: "b"})
	wantBlocked(t, err, "1 format problem: leftover-markdown: subject: **urgent**")
	if st.sent != 0 {
		t.Fatalf("sent %d", st.sent)
	}
}

func TestSendSubjectProblemsComeFirst(t *testing.T) {
	out, err := Send(context.Background(), &stub{}, sessMail(), SendInput{
		To: []string{"user@example.com"}, Subject: "**urgent**", Body: "<p>hi", HTML: true, DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantProblems(t, out, msgbody.Problems{
		{Rule: msgbody.RuleLeftoverMarkdown, Detail: "subject: **urgent**"},
		{Rule: msgbody.RuleBrokenHTML, Detail: "unclosed <p>"},
	})
}

func TestReplyFormatProblems(t *testing.T) {
	st := &replyRec{stub: &stub{}}
	out, err := Reply(context.Background(), st, sessMail(), ReplyInput{ID: "msg-1", Body: "<p>hi", HTML: true, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	wantProblems(t, out, unclosedP)
	_, err = Reply(context.Background(), st, sessMail(), ReplyInput{ID: "msg-1", Body: "<p>hi", HTML: true})
	wantBlocked(t, err, "1 format problem: broken-html: unclosed <p>")
	if st.sent != 0 {
		t.Fatalf("sent %d", st.sent)
	}
}
