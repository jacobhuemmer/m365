package calendar

import (
	"context"
	"strings"
	"time"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/domain"
)

type FreeQuery struct {
	When     string
	Duration string
	Hours    string
	Calendar string
	Top      int
	Now      time.Time
	Location string
}

type FreePage struct {
	Limit    int               `json:"limit"`
	Count    int               `json:"count"`
	Timezone string            `json:"timezone,omitempty"`
	Items    []domain.FreeSlot `json:"items"`
}

func Free(ctx context.Context, st Store, sess domain.Session, q FreeQuery) (FreePage, error) {
	if err := auth.RequireCalendar(sess); err != nil {
		return FreePage{}, err
	}
	top, err := domain.NormalizeFreeTop(q.Top)
	if err != nil {
		return FreePage{}, err
	}
	dur, err := ParseDuration(q.Duration)
	if err != nil {
		return FreePage{}, err
	}
	loc := loadLoc(q.Location)
	now := q.Now
	if now.IsZero() {
		now = time.Now().In(loc)
	} else {
		now = now.In(loc)
	}
	day, err := ResolveDay(q.When, now, loc)
	if err != nil {
		return FreePage{}, err
	}
	whStart, whEnd, err := parseHours(q.Hours, day, loc)
	if err != nil {
		return FreePage{}, err
	}
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24 * time.Hour)
	evs, err := st.ListEvents(ctx, ListEventsQuery{
		Calendar: q.Calendar, Start: dayStart.Format(time.RFC3339), End: dayEnd.Format(time.RFC3339), Top: 50, Now: now,
	})
	if err != nil {
		return FreePage{}, err
	}
	busy := busyFrom(evs.Items, loc)
	var items []domain.FreeSlot
	for t := whStart; !t.Add(dur).After(whEnd); t = t.Add(30 * time.Minute) {
		end := t.Add(dur)
		if sameLocalDay(t, now) && !t.After(now) {
			continue
		}
		if overlaps(t, end, busy) {
			continue
		}
		items = append(items, domain.FreeSlot{
			CalendarID: q.Calendar, Start: t.Format(time.RFC3339), End: end.Format(time.RFC3339),
			DurationMin: int(dur / time.Minute),
		})
		if len(items) >= top {
			break
		}
	}
	return FreePage{Limit: top, Count: len(items), Timezone: loc.String(), Items: items}, nil
}

func parseHours(s string, day time.Time, loc *time.Location) (time.Time, time.Time, error) {
	startH, startM, endH, endM := 9, 0, 17, 0
	if s != "" {
		i := strings.Index(s, "-")
		if i <= 0 || i == len(s)-1 {
			return time.Time{}, time.Time{}, domain.Usage("hours must be start-end")
		}
		c1, ok1 := parseClock(strings.TrimSpace(s[:i]))
		c2, ok2 := parseClock(strings.TrimSpace(s[i+1:]))
		if !ok1 || !ok2 {
			return time.Time{}, time.Time{}, domain.Usage("hours must be start-end")
		}
		startH, startM, endH, endM = c1.hour, c1.min, c2.hour, c2.min
	}
	ws := time.Date(day.Year(), day.Month(), day.Day(), startH, startM, 0, 0, loc)
	we := time.Date(day.Year(), day.Month(), day.Day(), endH, endM, 0, 0, loc)
	if !ws.Before(we) {
		return time.Time{}, time.Time{}, domain.Usage("hours start must be before end")
	}
	return ws, we, nil
}

func busyFrom(evs []domain.CalendarEvent, loc *time.Location) [][2]time.Time {
	var out [][2]time.Time
	for _, e := range evs {
		st, err1 := time.Parse(time.RFC3339, e.Start)
		en, err2 := time.Parse(time.RFC3339, e.End)
		if err1 != nil || err2 != nil {
			continue
		}
		out = append(out, [2]time.Time{st.In(loc), en.In(loc)})
	}
	return out
}

func overlaps(s, e time.Time, busy [][2]time.Time) bool {
	for _, b := range busy {
		if s.Before(b[1]) && e.After(b[0]) {
			return true
		}
	}
	return false
}

func sameLocalDay(a, b time.Time) bool {
	a = a.In(b.Location())
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}
