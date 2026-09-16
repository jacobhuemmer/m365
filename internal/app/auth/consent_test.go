package auth

import (
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestConsentIsAuthNotService(t *testing.T) {
	mailOnly := domain.Session{SignedIn: true, SessionUsable: true, MailConsented: true}
	if err := Require(mailOnly, false, true); domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
	teamsOnly := domain.Session{SignedIn: true, SessionUsable: true, TeamsConsented: true}
	if err := Require(teamsOnly, true, false); domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
}
