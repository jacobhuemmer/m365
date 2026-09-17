package domain

import "testing"

func TestSignedOutHasNoConsent(t *testing.T) {
	s := SignedOut()
	if s.SignedIn || s.SessionUsable || s.MailConsented || s.TeamsConsented || s.CalendarConsented || s.FilesConsented {
		t.Fatalf("signed out: %+v", s)
	}
}
