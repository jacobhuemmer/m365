package calendar

import (
	"testing"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

func freeze() (time.Time, *time.Location) {
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		panic(err)
	}
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, loc)
	return now, loc
}

func TestResolveWhenValid(t *testing.T) {
	now, loc := freeze()
	iv, err := ResolveWhen("tomorrow at 1:30 pm", now, loc, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if iv.Start.Hour() != 13 || iv.Start.Minute() != 30 || iv.Start.Day() != 17 {
		t.Fatalf("start %v", iv.Start)
	}
	if iv.End.Sub(iv.Start) != 30*time.Minute {
		t.Fatalf("dur %v", iv.End.Sub(iv.Start))
	}
	iv, err = ResolveWhen("tomorrow at 1:30 pm", now, loc, time.Hour, "")
	if err != nil || iv.End.Hour() != 14 || iv.End.Minute() != 30 {
		t.Fatal(err, iv.End)
	}
	iv, err = ResolveWhen("tomorrow at 1:30 pm", now, loc, 0, "2:00 pm")
	if err != nil || iv.End.Hour() != 14 {
		t.Fatal(err, iv.End)
	}
	iv, err = ResolveWhen("2026-09-18 13:30", now, loc, 0, "")
	if err != nil || iv.Start.Day() != 18 || iv.Start.Hour() != 13 {
		t.Fatal(err, iv.Start)
	}
}

func TestResolveWhenInvalid(t *testing.T) {
	now, loc := freeze()
	for _, p := range []string{"1:30", "Friday", "next week", "after lunch"} {
		_, err := ResolveWhen(p, now, loc, 0, "")
		if domain.ExitOf(err) != domain.ExitUsage {
			t.Fatalf("%q: %v", p, err)
		}
	}
	_, err := ResolveWhen("today at 9am", now, loc, 0, "")
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal("past today 9am", err)
	}
}

func TestParseDuration(t *testing.T) {
	d, err := ParseDuration("1 hour")
	if err != nil || d != time.Hour {
		t.Fatal(d, err)
	}
	_, err = ParseDuration("0 minutes")
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestResolveDay(t *testing.T) {
	now, loc := freeze()
	d, err := ResolveDay("tomorrow", now, loc)
	if err != nil || d.Day() != 17 {
		t.Fatal(err, d)
	}
	d, err = ResolveDay("", now, loc)
	if err != nil || d.Day() != 17 {
		t.Fatal("default tomorrow", err, d)
	}
}
