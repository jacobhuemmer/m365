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
	calOnly := domain.Session{SignedIn: true, SessionUsable: true, CalendarConsented: true}
	if err := RequireFiles(calOnly); domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
	if err := RequireCalendar(calOnly); err != nil {
		t.Fatal(err)
	}
	filesOnly := domain.Session{SignedIn: true, SessionUsable: true, FilesConsented: true}
	if err := RequireCalendar(filesOnly); domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
}
