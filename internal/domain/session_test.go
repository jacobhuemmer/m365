package domain

import "testing"

func TestSignedOutHasNoConsent(t *testing.T) {
	s := SignedOut()
	if s.SignedIn || s.SessionUsable || s.MailConsented || s.TeamsConsented {
		t.Fatalf("signed out: %+v", s)
	}
}
