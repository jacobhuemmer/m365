package graph

import (
	"context"
	"testing"
)

func TestListCalendarsTimezone(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	c := &HTTPCalendar{HTTPClient: &HTTPClient{Base: srv.URL, Token: "fake-calendar"}}
	p, err := c.ListCalendars(context.Background(), 20, "")
	if err != nil || len(p.Items) == 0 || p.Items[0].Timezone != "America/Chicago" {
		t.Fatalf("%+v %v", p, err)
	}
}

func TestMapTZ(t *testing.T) {
	if mapTZ("Central Standard Time") != "America/Chicago" {
		t.Fatal(mapTZ("Central Standard Time"))
	}
	if mapTZ("America/Chicago") != "America/Chicago" {
		t.Fatal(mapTZ("America/Chicago"))
	}
}
