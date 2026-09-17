package domain

import (
	"testing"
	"time"
)

func TestDefaultEventWindow(t *testing.T) {
	now := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	s, e := DefaultEventWindow(now)
	if !s.Equal(now) || e.Sub(s) != 7*24*time.Hour {
		t.Fatalf("%v %v", s, e)
	}
}

func TestValidateEventRange(t *testing.T) {
	if err := ValidateEventRange("2026-09-16T10:00:00Z", "2026-09-16T11:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if ExitOf(ValidateEventRange("2026-09-16T11:00:00Z", "2026-09-16T10:00:00Z")) != ExitUsage {
		t.Fatal("start after end")
	}
	if ExitOf(ValidateEventRange("nope", "2026-09-16T10:00:00Z")) != ExitUsage {
		t.Fatal("naive")
	}
}
