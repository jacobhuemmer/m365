package domain

import "time"

const (
	DefaultMeeting = 30 * time.Minute
	MaxMeeting     = 8 * time.Hour
	DefaultFreeTop = 5
	MaxFreeTop     = 20
)

type WhenPhrase struct {
	Original string
}

type ResolvedInterval struct {
	When     string `json:"when,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	Start    time.Time
	End      time.Time
}

type FreeSlot struct {
	CalendarID  string `json:"calendar_id,omitempty"`
	Start       string `json:"start"`
	End         string `json:"end"`
	DurationMin int    `json:"duration_minutes,omitempty"`
}

func NormalizeFreeTop(raw int) (int, error) {
	if raw == 0 {
		return DefaultFreeTop, nil
	}
	if raw < 1 || raw > MaxFreeTop {
		return 0, Usagef("top must be 1..%d, got %d", MaxFreeTop, raw)
	}
	return raw, nil
}

func ValidateResolved(iv ResolvedInterval, now time.Time) error {
	if !iv.Start.Before(iv.End) {
		return Usage("start must be before end")
	}
	if !iv.Start.After(now) {
		return Usage("start is in the past")
	}
	if iv.End.Sub(iv.Start) > MaxMeeting {
		return Usage("duration must be at most 8 hours")
	}
	return nil
}
