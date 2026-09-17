package calendar

import (
	"context"
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

type stub struct {
	cals domain.CalendarPage
	evs  domain.EventPage
	ev   domain.CalendarEvent
	err  error
	mut  int
}

func (s *stub) ListCalendars(context.Context, int, string) (domain.CalendarPage, error) {
	return s.cals, s.err
}
func (s *stub) ListEvents(context.Context, ListEventsQuery) (domain.EventPage, error) {
	return s.evs, s.err
}
func (s *stub) GetEvent(context.Context, string) (domain.CalendarEvent, error) { return s.ev, s.err }
func (s *stub) CreateEvent(context.Context, WriteInput) (string, error) {
	s.mut++
	return "ev-new", s.err
}
func (s *stub) UpdateEvent(context.Context, WriteInput) error { s.mut++; return s.err }
func (s *stub) DeleteEvent(context.Context, string) error     { s.mut++; return s.err }

func sessCal() domain.Session {
	return domain.Session{SignedIn: true, SessionUsable: true, CalendarConsented: true}
}

func TestListDefaultsAndAuth(t *testing.T) {
	st := &stub{cals: domain.CalendarPage{Count: 1, Items: []domain.Calendar{{ID: "cal-1"}}}}
	p, err := ListCalendars(context.Background(), st, sessCal(), 0, "")
	if err != nil || p.Count != 1 {
		t.Fatal(err, p)
	}
	_, err = ListCalendars(context.Background(), st, domain.SignedOut(), 0, "")
	if domain.ExitOf(err) != domain.ExitAuth {
		t.Fatal(err)
	}
	_, err = ListCalendars(context.Background(), st, sessCal(), 51, "")
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	ep, err := ListEvents(context.Background(), st, sessCal(), ListEventsQuery{Now: now})
	if err != nil || ep.WindowStart == "" {
		t.Fatal(err, ep)
	}
}
