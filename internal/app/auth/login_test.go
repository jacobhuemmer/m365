package auth

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

type mem struct{ b *Blob }

func (m *mem) Get() (Blob, bool, error) {
	if m.b == nil {
		return Blob{}, false, nil
	}
	return *m.b, true, nil
}
func (m *mem) Put(b Blob) error { m.b = &b; return nil }
func (m *mem) Delete() error    { m.b = nil; return nil }

func TestLoginStatusLogout(t *testing.T) {
	s := &mem{}
	st, err := Status(s)
	if err != nil || st.SignedIn {
		t.Fatalf("signed out: %+v %v", st, err)
	}
	st, err = Login(context.Background(), s, func(context.Context) (Blob, error) {
		return Blob{Account: "user@example.com", Mail: true, Teams: true, Calendar: true, Files: true}, nil
	})
	if err != nil || !st.SessionUsable || st.Account != "user@example.com" || !st.CalendarConsented || !st.FilesConsented {
		t.Fatalf("login: %+v %v", st, err)
	}
	if err := Logout(s); err != nil {
		t.Fatal(err)
	}
	st, _ = Status(s)
	if st.SignedIn {
		t.Fatal("still signed in")
	}
}

func TestRequire(t *testing.T) {
	if err := Require(domain.SignedOut(), true, false); domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
	sess := domain.Session{SignedIn: true, SessionUsable: true, MailConsented: true}
	if err := Require(sess, true, false); err != nil {
		t.Fatal(err)
	}
	if err := Require(sess, false, true); domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal("teams consent")
	}
	if err := Require(domain.Session{SignedIn: true, SessionUsable: false, MailConsented: true}, true, false); domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal("expired")
	}
}
