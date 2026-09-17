package graph

import (
	"context"
	"testing"

	"github.com/masonhuemmer/m365/internal/app/calendar"
)

func TestHTTPCalendarWriteViaFake(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPCalendar{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-calendar"}}
	id, err := c.CreateEvent(context.Background(), calendar.WriteInput{Subject: "n", Start: "2026-09-16T10:00:00Z", End: "2026-09-16T11:00:00Z"})
	if err != nil || id == "" {
		t.Fatal(err, id)
	}
}
