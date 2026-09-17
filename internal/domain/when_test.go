package domain

import (
	"testing"
	"time"
)

func TestNormalizeFreeTop(t *testing.T) {
	n, err := NormalizeFreeTop(0)
	if err != nil || n != 5 {
		t.Fatal(n, err)
	}
	_, err = NormalizeFreeTop(21)
	if ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
}

func TestValidateResolved(t *testing.T) {
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	iv := ResolvedInterval{Start: now.Add(time.Hour), End: now.Add(90 * time.Minute)}
	if err := ValidateResolved(iv, now); err != nil {
		t.Fatal(err)
	}
	iv.End = iv.Start
	if ExitOf(ValidateResolved(iv, now)) != ExitUsage {
		t.Fatal("equal")
	}
	iv = ResolvedInterval{Start: now.Add(-time.Hour), End: now}
	if ExitOf(ValidateResolved(iv, now)) != ExitUsage {
		t.Fatal("past")
	}
}
