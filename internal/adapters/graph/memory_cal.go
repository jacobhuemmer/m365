package graph

import (
	"context"
	"strconv"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/calendar"
	"github.com/masonhuemmer/m365/internal/domain"
)

func (m *Memory) ListCalendars(_ context.Context, top int, _ string) (domain.CalendarPage, error) {
	if err := m.fail(); err != nil {
		return domain.CalendarPage{}, err
	}
	items := m.Calendars
	if top > 0 && len(items) > top {
		items = items[:top]
	}
	return domain.CalendarPage{Limit: top, Count: len(items), Items: items}, nil
}

func (m *Memory) ListEvents(_ context.Context, q calendar.ListEventsQuery) (domain.EventPage, error) {
	if err := m.fail(); err != nil {
		return domain.EventPage{}, err
	}
	cal := q.Calendar
	if cal == "" {
		for _, c := range m.Calendars {
			if c.IsDefault {
				cal = c.ID
				break
			}
		}
	}
	var items []domain.CalendarEvent
	for _, e := range m.CalEvents {
		if e.CalendarID != cal && e.CalendarID != "" && cal != e.CalendarID {
			found := false
			for _, c := range m.Calendars {
				if (c.ID == cal || strings.EqualFold(c.Name, cal)) && e.CalendarID == c.ID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if q.Start != "" && e.Start < q.Start {
			continue
		}
		if q.End != "" && e.Start >= q.End {
			continue
		}
		cp := e
		cp.Body = ""
		items = append(items, cp)
	}
	if q.Top > 0 && len(items) > q.Top {
		items = items[:q.Top]
	}
	known := false
	for _, c := range m.Calendars {
		if c.ID == cal || strings.EqualFold(c.Name, cal) {
			known = true
			break
		}
	}
	if cal != "" && !known {
		return domain.EventPage{}, domain.NotFound("calendar not found")
	}
	return domain.EventPage{Limit: q.Top, Count: len(items), Items: items}, nil
}

func (m *Memory) GetEvent(_ context.Context, id string) (domain.CalendarEvent, error) {
	if err := m.fail(); err != nil {
		return domain.CalendarEvent{}, err
	}
	for _, e := range m.CalEvents {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.CalendarEvent{}, domain.NotFound("event not found")
}

func (m *Memory) CreateEvent(_ context.Context, in calendar.WriteInput) (string, error) {
	if err := m.fail(); err != nil {
		return "", err
	}
	id := "ev-" + strconv.Itoa(len(m.CalEvents)+1)
	m.mu.Lock()
	m.CalEvents = append(m.CalEvents, domain.CalendarEvent{
		ID: id, CalendarID: in.CalendarID, Subject: in.Subject, Start: in.Start, End: in.End, Location: in.Location, Body: in.Body,
	})
	m.mu.Unlock()
	return id, nil
}

func (m *Memory) UpdateEvent(_ context.Context, in calendar.WriteInput) error {
	if err := m.fail(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, e := range m.CalEvents {
		if e.ID == in.ID {
			if in.Subject != "" {
				e.Subject = in.Subject
			}
			if in.Start != "" {
				e.Start = in.Start
			}
			if in.End != "" {
				e.End = in.End
			}
			if in.Location != "" {
				e.Location = in.Location
			}
			if in.Body != "" {
				e.Body = in.Body
			}
			m.CalEvents[i] = e
			return nil
		}
	}
	return domain.NotFound("event not found")
}

func (m *Memory) DeleteEvent(_ context.Context, id string) error {
	if err := m.fail(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, e := range m.CalEvents {
		if e.ID == id {
			m.CalEvents = append(m.CalEvents[:i], m.CalEvents[i+1:]...)
			return nil
		}
	}
	return domain.NotFound("event not found")
}

type CalendarAPI struct{ *Memory }
