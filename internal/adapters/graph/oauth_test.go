package graph

import (
	"context"
	"net/http"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
	"golang.org/x/oauth2"
)

func TestPKCEHelpers(t *testing.T) {
	cfg := PKCEConfig("id", "tenant", "http://127.0.0.1/")
	v := oauth2.GenerateVerifier()
	u := cfg.AuthCodeURL("state", oauth2.S256ChallengeOption(v), oauth2.AccessTypeOffline)
	if u == "" || v == "" {
		t.Fatal("pkce")
	}
	fn := FakeLogin(true, false)
	b, err := fn(context.Background())
	if err != nil || !b.Mail || b.Teams {
		t.Fatalf("%+v %v", b, err)
	}
	if domain.ExitOf(Denied()) != domain.ExitAuth {
		t.Fatal("denied")
	}
}

func TestStartLoopbackReceivesCode(t *testing.T) {
	redirect, wait, err := StartLoopback()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = http.Get(redirect + "?code=abc")
	}()
	code, err := wait(context.Background())
	if err != nil || code != "abc" {
		t.Fatalf("%q %v", code, err)
	}
}
