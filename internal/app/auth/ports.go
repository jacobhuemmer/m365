package auth

import (
	"context"

	"github.com/masonhuemmer/m365/internal/domain"
)

type Blob struct {
	Account      string `json:"account"`
	Mail         bool   `json:"mail"`
	Teams        bool   `json:"teams"`
	Usable       bool   `json:"usable"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type Store interface {
	Get() (Blob, bool, error)
	Put(Blob) error
	Delete() error
}

type LoginFn func(ctx context.Context) (Blob, error)

// BrowserOpener opens an authorize URL (system browser in production, fake in tests).
type BrowserOpener func(authorizeURL string) error

// LoopbackStart binds 127.0.0.1 and returns the redirect URL plus a wait for the code.
type LoopbackStart func() (redirectURL string, wait func(context.Context) (code string, err error), err error)

func Status(s Store) (domain.Session, error) {
	b, ok, err := s.Get()
	if err != nil {
		return domain.Session{}, err
	}
	if !ok {
		return domain.SignedOut(), nil
	}
	return domain.Session{
		SignedIn:       true,
		SessionUsable:  b.Usable,
		Account:        b.Account,
		MailConsented:  b.Mail,
		TeamsConsented: b.Teams,
	}, nil
}

func Logout(s Store) error {
	return s.Delete()
}

func Login(ctx context.Context, s Store, fn LoginFn) (domain.Session, error) {
	if fn == nil {
		return domain.Session{}, domain.Auth("login not available")
	}
	b, err := fn(ctx)
	if err != nil {
		return domain.Session{}, err
	}
	b.Usable = true
	if err := s.Put(b); err != nil {
		return domain.Session{}, err
	}
	return Status(s)
}

func Require(sess domain.Session, mail, teams bool) error {
	if !sess.SignedIn || !sess.SessionUsable {
		return domain.Auth("not signed in")
	}
	if mail && !sess.MailConsented {
		return domain.Auth("missing mail consent")
	}
	if teams && !sess.TeamsConsented {
		return domain.Auth("missing Teams consent")
	}
	return nil
}
