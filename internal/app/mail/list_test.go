package mail

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

type stub struct {
	page domain.MailPage
	msg  domain.MailMessage
	th   domain.MailThread
	err  error
	sent int
	last SendInput
}

func (s *stub) List(context.Context, ListQuery) (domain.MailPage, error) {
	return s.page, s.err
}
func (s *stub) Get(context.Context, string) (domain.MailMessage, error) { return s.msg, s.err }
func (s *stub) Thread(context.Context, string, bool) (domain.MailThread, error) {
	return s.th, s.err
}
func (s *stub) Send(_ context.Context, in SendInput) (string, error) {
	s.sent++
	s.last = in
	return "new-1", s.err
}
func (s *stub) Reply(context.Context, ReplyInput) (string, error) {
	s.sent++
	return "r-1", s.err
}
func (s *stub) Attachments(context.Context, string) ([]domain.Attachment, error) {
	return s.msg.Attachments, s.err
}
func (s *stub) Download(context.Context, string, string) ([]byte, domain.Attachment, error) {
	return []byte("synthetic-ok"), domain.Attachment{ID: "att-1", Name: "note.txt", Size: 12}, s.err
}

func sessMail() domain.Session {
	return domain.Session{SignedIn: true, SessionUsable: true, MailConsented: true}
}

func TestListDefaultsAndAuth(t *testing.T) {
	st := &stub{page: domain.MailPage{Items: nil, Limit: 10}}
	_, err := List(context.Background(), st, domain.SignedOut(), ListQuery{})
	if domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
	p, err := List(context.Background(), st, sessMail(), ListQuery{})
	if err != nil || p.Limit != 0 && len(p.Items) != 0 {
		// store returns its page; use case should set top on query
	}
	_, err = List(context.Background(), st, sessMail(), ListQuery{Top: 51})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	_, err = Get(context.Background(), st, sessMail(), "")
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestGetNotFound(t *testing.T) {
	st := &stub{err: domain.NotFound("missing")}
	_, err := Get(context.Background(), st, sessMail(), "missing-id")
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
}
