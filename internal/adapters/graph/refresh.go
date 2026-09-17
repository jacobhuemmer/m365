package graph

import (
	"context"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/domain"
)

type Refresher struct {
	Store  auth.Store
	Config *oauth2.Config
	Now    func() time.Time

	mu sync.Mutex
}

func (r *Refresher) Token(force bool) (string, error) {
	if r == nil || r.Store == nil || r.Config == nil {
		return "", domain.Auth("not signed in")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok, err := r.Store.Get()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", domain.Auth("not signed in")
	}
	now := time.Now()
	if r.Now != nil {
		now = r.Now()
	}
	fresh := b.AccessToken != "" && (b.ExpiresAt.IsZero() || b.ExpiresAt.After(now.Add(time.Minute)))
	if !force && fresh {
		return b.AccessToken, nil
	}
	if b.RefreshToken == "" {
		if b.AccessToken != "" && !force {
			return b.AccessToken, nil
		}
		return "", domain.Auth("session expired")
	}
	cfg := *r.Config
	cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	src := cfg.TokenSource(context.Background(), &oauth2.Token{
		AccessToken:  b.AccessToken,
		RefreshToken: b.RefreshToken,
		Expiry:       now.Add(-time.Minute),
	})
	tok, err := src.Token()
	if err != nil || tok == nil || tok.AccessToken == "" {
		return "", domain.Auth("session expired")
	}
	b.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		b.RefreshToken = tok.RefreshToken
	}
	b.ExpiresAt = tok.Expiry
	b.Usable = true
	if err := r.Store.Put(b); err != nil {
		return "", err
	}
	return b.AccessToken, nil
}
