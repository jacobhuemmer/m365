package calendar

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/masonhuemmer/m365/internal/domain"
)

var (
	clock12 = regexp.MustCompile(`^(\d{1,2})(?::(\d{2}))?\s*(am|pm)$`)
	clock24 = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	durRe   = regexp.MustCompile(`^(\d+)\s*(minutes?|hours?)$`)
)

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
	"wednesday": time.Wednesday, "thursday": time.Thursday, "friday": time.Friday, "saturday": time.Saturday,
}

func ParseDuration(s string) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return domain.DefaultMeeting, nil
	}
	m := durRe.FindStringSubmatch(s)
	if m == nil {
		return 0, domain.Usage("duration must be N minutes or N hours")
	}
	n, _ := strconv.Atoi(m[1])
	if n < 1 {
		return 0, domain.Usage("duration must be a positive whole number of minutes")
	}
	d := time.Duration(n) * time.Minute
	if strings.HasPrefix(m[2], "hour") {
		d = time.Duration(n) * time.Hour
	}
	if d > domain.MaxMeeting {
		return 0, domain.Usage("duration must be at most 8 hours")
	}
	return d, nil
}

func ResolveDay(phrase string, now time.Time, loc *time.Location) (time.Time, error) {
	p := norm(phrase)
	if p == "" {
		p = "tomorrow"
	}
	day, _, err := splitDayClock(p, now, loc, false)
	return day, err
}

func ResolveWhen(phrase string, now time.Time, loc *time.Location, dur time.Duration, until string) (domain.ResolvedInterval, error) {
	if loc == nil {
		loc = time.UTC
	}
	now = now.In(loc)
	p := norm(phrase)
	if p == "" {
		return domain.ResolvedInterval{}, domain.Usage("when-phrase is required")
	}
	if t, err := time.Parse(time.RFC3339, phrase); err == nil {
		if dur == 0 {
			dur = domain.DefaultMeeting
		}
		iv := domain.ResolvedInterval{When: phrase, Timezone: loc.String(), Start: t.In(loc), End: t.In(loc).Add(dur)}
		return iv, domain.ValidateResolved(iv, now)
	}
	day, clock, err := splitDayClock(p, now, loc, true)
	if err != nil {
		return domain.ResolvedInterval{}, err
	}
	start := time.Date(day.Year(), day.Month(), day.Day(), clock.hour, clock.min, 0, 0, loc)
	end := start.Add(dur)
	if dur == 0 {
		end = start.Add(domain.DefaultMeeting)
	}
	if u := norm(until); u != "" {
		ut, err := parseUntil(u, start, loc)
		if err != nil {
			return domain.ResolvedInterval{}, err
		}
		end = ut
	}
	iv := domain.ResolvedInterval{When: phrase, Timezone: loc.String(), Start: start, End: end}
	return iv, domain.ValidateResolved(iv, now)
}

type clockHM struct{ hour, min int }

func splitDayClock(p string, now time.Time, loc *time.Location, needClock bool) (time.Time, clockHM, error) {
	if t, err := time.ParseInLocation("2006-01-02 15:04", p, loc); err == nil {
		return t, clockHM{t.Hour(), t.Minute()}, nil
	}
	if t, err := time.ParseInLocation("2 January 2006 15:04", p, loc); err == nil {
		return t, clockHM{t.Hour(), t.Minute()}, nil
	}
	p = strings.ReplaceAll(p, " at ", " ")
	parts := strings.Fields(p)
	if len(parts) == 0 {
		return time.Time{}, clockHM{}, domain.Usage("when-phrase could not be used")
	}
	var clk clockHM
	var hasClock bool
	dayPart := p
	if c, ok := parseClock(parts[len(parts)-1]); ok {
		clk, hasClock = c, true
		dayPart = strings.TrimSpace(strings.TrimSuffix(p, parts[len(parts)-1]))
	} else if len(parts) >= 2 {
		if c, ok := parseClock(parts[len(parts)-2] + parts[len(parts)-1]); ok {
			clk, hasClock = c, true
			dayPart = strings.TrimSpace(strings.Join(parts[:len(parts)-2], " "))
		}
	}
	if needClock && !hasClock {
		return time.Time{}, clockHM{}, domain.Usage("when-phrase could not be used")
	}
	day, err := parseDay(dayPart, now, loc)
	if err != nil {
		return time.Time{}, clockHM{}, err
	}
	if hasClock {
		cand := time.Date(day.Year(), day.Month(), day.Day(), clk.hour, clk.min, 0, 0, loc)
		if wd, ok := weekdays[norm(dayPart)]; ok && !cand.After(now) {
			day = day.AddDate(0, 0, 7)
			_ = wd
		}
	}
	return day, clk, nil
}

func parseDay(s string, now time.Time, loc *time.Location) (time.Time, error) {
	s = norm(s)
	if s == "" || s == "today" {
		return now.In(loc), nil
	}
	if s == "tomorrow" {
		return now.In(loc).AddDate(0, 0, 1), nil
	}
	if wd, ok := weekdays[s]; ok {
		d := now.In(loc)
		for i := 0; i < 8; i++ {
			if d.Weekday() == wd {
				return d, nil
			}
			d = d.AddDate(0, 0, 1)
		}
	}
	if t, err := time.ParseInLocation("2006-01-02", s, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2 January 2006", s, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2 january 2006", s, loc); err == nil {
		return t, nil
	}
	return time.Time{}, domain.Usage("when-phrase could not be used")
}

func parseClock(s string) (clockHM, bool) {
	s = strings.ReplaceAll(norm(s), " ", "")
	if m := clock12.FindStringSubmatch(s); m != nil {
		h, _ := strconv.Atoi(m[1])
		min := 0
		if m[2] != "" {
			min, _ = strconv.Atoi(m[2])
		}
		if h < 1 || h > 12 || min > 59 {
			return clockHM{}, false
		}
		if m[3] == "pm" && h != 12 {
			h += 12
		}
		if m[3] == "am" && h == 12 {
			h = 0
		}
		return clockHM{h, min}, true
	}
	if m := clock24.FindStringSubmatch(s); m != nil {
		h, _ := strconv.Atoi(m[1])
		min, _ := strconv.Atoi(m[2])
		if h > 23 || min > 59 {
			return clockHM{}, false
		}
		return clockHM{h, min}, true
	}
	return clockHM{}, false
}

func parseUntil(u string, start time.Time, loc *time.Location) (time.Time, error) {
	if c, ok := parseClock(u); ok {
		end := time.Date(start.Year(), start.Month(), start.Day(), c.hour, c.min, 0, 0, loc)
		if !end.After(start) {
			return time.Time{}, domain.Usage("start must be before end")
		}
		return end, nil
	}
	iv, err := ResolveWhen(u, start.Add(-time.Second), loc, domain.DefaultMeeting, "")
	if err != nil {
		return time.Time{}, err
	}
	if iv.Start.Year() != start.Year() || iv.Start.YearDay() != start.YearDay() {
		return time.Time{}, domain.Usage("until must be on the same local date")
	}
	return iv.Start, nil
}

func norm(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), " ")
}

func Clock() time.Time {
	if s := strings.TrimSpace(os.Getenv("M365_NOW")); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
	}
	return time.Now()
}

func loadLoc(name string) *time.Location {
	if name == "" {
		name = "America/Chicago"
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
