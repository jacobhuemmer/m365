package graph

import (
	"context"
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/app/calendar"
	"github.com/masonhuemmer/m365/internal/domain"
)

func TestHTTPCalendarViaFakeServer(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPCalendar{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-calendar"}}
	p, err := c.ListCalendars(context.Background(), 20, "")
	if err != nil || p.Count == 0 {
		t.Fatalf("calendars: %+v %v", p, err)
	}
	now := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	evs, err := c.ListEvents(context.Background(), calendar.ListEventsQuery{
		Start: now.Format(time.RFC3339), End: now.Add(7 * 24 * time.Hour).Format(time.RFC3339), Top: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range evs.Items {
		if e.ID == "ev-out" {
			t.Fatal("series dump")
		}
	}
	ev, err := c.GetEvent(context.Background(), "ev-1")
	if err != nil || ev.Subject == "" {
		t.Fatalf("%+v %v", ev, err)
	}
	_, err = c.GetEvent(context.Background(), "missing")
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
}
