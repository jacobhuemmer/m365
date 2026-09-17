package calendar

import (
	"context"
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestFreeSlotsSkipBusy(t *testing.T) {
	now, loc := freeze()
	st := &stub{evs: domain.EventPage{Items: []domain.CalendarEvent{
		{Start: "2026-09-17T09:00:00-05:00", End: "2026-09-17T10:00:00-05:00"},
		{Start: "2026-09-17T13:00:00-05:00", End: "2026-09-17T14:00:00-05:00"},
	}}}
	p, err := Free(context.Background(), st, sessCal(), FreeQuery{When: "tomorrow", Now: now, Location: loc.String()})
	if err != nil {
		t.Fatal(err)
	}
	if p.Limit != 5 || p.Count == 0 || p.Count > 5 {
		t.Fatalf("%+v", p)
	}
	for _, it := range p.Items {
		stt, _ := time.Parse(time.RFC3339, it.Start)
		if stt.Hour() == 9 {
			t.Fatal("overlap 9am")
		}
	}
}

func TestFreeFullyBooked(t *testing.T) {
	now, loc := freeze()
	st := &stub{evs: domain.EventPage{Items: []domain.CalendarEvent{
		{Start: "2026-09-19T09:00:00-05:00", End: "2026-09-19T17:00:00-05:00"},
	}}}
	p, err := Free(context.Background(), st, sessCal(), FreeQuery{When: "2026-09-19", Now: now, Location: loc.String()})
	if err != nil || p.Count != 0 {
		t.Fatal(err, p)
	}
}

func TestFreeTopMax(t *testing.T) {
	now, loc := freeze()
	_, err := Free(context.Background(), &stub{}, sessCal(), FreeQuery{Top: 21, Now: now, Location: loc.String()})
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}
