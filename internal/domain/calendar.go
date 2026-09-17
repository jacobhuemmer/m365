package domain

import "time"

type Calendar struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
}

type CalendarEvent struct {
	ID         string   `json:"id"`
	CalendarID string   `json:"calendar_id,omitempty"`
	Subject    string   `json:"subject"`
	Start      string   `json:"start"`
	End        string   `json:"end"`
	Location   string   `json:"location,omitempty"`
	Organizer  Person   `json:"organizer,omitempty"`
	Attendees  []Person `json:"attendees,omitempty"`
	Body       string   `json:"body,omitempty"`
}

type CalendarPage struct {
	Limit    int        `json:"limit"`
	Count    int        `json:"count"`
	NextPage *string    `json:"next_page"`
	Items    []Calendar `json:"items"`
}

type EventPage struct {
	Limit       int             `json:"limit"`
	Count       int             `json:"count"`
	NextPage    *string         `json:"next_page"`
	WindowStart string          `json:"window_start,omitempty"`
	WindowEnd   string          `json:"window_end,omitempty"`
	Items       []CalendarEvent `json:"items"`
}

func DefaultEventWindow(now time.Time) (start, end time.Time) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return now, now.Add(7 * 24 * time.Hour)
}

func ParseEventTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, Usage("event time is required")
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, Usage("event time must include a timezone or UTC instant")
}

func ValidateEventRange(start, end string) error {
	st, err := ParseEventTime(start)
	if err != nil {
		return err
	}
	en, err := ParseEventTime(end)
	if err != nil {
		return err
	}
	if !st.Before(en) {
		return Usage("start must be before end")
	}
	return nil
}
