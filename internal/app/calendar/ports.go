package calendar

import (
	"context"
	"strings"
	"time"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/domain"
)

type ListCalendarsQuery struct {
	Top, PageToken string
	TopN           int
}

type ListEventsQuery struct {
	Calendar  string
	Start     string
	End       string
	Top       int
	PageToken string
	Now       time.Time
}

type WriteInput struct {
	CalendarID string
	ID         string
	Subject    string
	Start      string
	End        string
	Location   string
	Body       string
	Attendees  []string
	DryRun     bool
}

type Store interface {
	ListCalendars(ctx context.Context, top int, page string) (domain.CalendarPage, error)
	ListEvents(ctx context.Context, q ListEventsQuery) (domain.EventPage, error)
	GetEvent(ctx context.Context, id string) (domain.CalendarEvent, error)
	CreateEvent(ctx context.Context, in WriteInput) (string, error)
	UpdateEvent(ctx context.Context, in WriteInput) error
	DeleteEvent(ctx context.Context, id string) error
}

func ListCalendars(ctx context.Context, st Store, sess domain.Session, top int, page string) (domain.CalendarPage, error) {
	if err := auth.RequireCalendar(sess); err != nil {
		return domain.CalendarPage{}, err
	}
	n, err := domain.NormalizeTop(top, domain.DefaultCalendarsTop)
	if err != nil {
		return domain.CalendarPage{}, err
	}
	return st.ListCalendars(ctx, n, page)
}

func ListEvents(ctx context.Context, st Store, sess domain.Session, q ListEventsQuery) (domain.EventPage, error) {
	if err := auth.RequireCalendar(sess); err != nil {
		return domain.EventPage{}, err
	}
	n, err := domain.NormalizeTop(q.Top, domain.DefaultCalendarTop)
	if err != nil {
		return domain.EventPage{}, err
	}
	q.Top = n
	now := q.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	ws, we := domain.DefaultEventWindow(now)
	if q.Start != "" {
		ws, err = domain.ParseEventTime(q.Start)
		if err != nil {
			return domain.EventPage{}, err
		}
	}
	if q.End != "" {
		we, err = domain.ParseEventTime(q.End)
		if err != nil {
			return domain.EventPage{}, err
		}
	}
	if !ws.Before(we) {
		return domain.EventPage{}, domain.Usage("start must be before end")
	}
	q.Start, q.End = ws.Format(time.RFC3339), we.Format(time.RFC3339)
	p, err := st.ListEvents(ctx, q)
	if err != nil {
		return domain.EventPage{}, err
	}
	p.WindowStart, p.WindowEnd = q.Start, q.End
	p.Limit = n
	return p, nil
}

func GetEvent(ctx context.Context, st Store, sess domain.Session, id string) (domain.CalendarEvent, error) {
	if err := auth.RequireCalendar(sess); err != nil {
		return domain.CalendarEvent{}, err
	}
	if strings.TrimSpace(id) == "" {
		return domain.CalendarEvent{}, domain.Usage("event id is required")
	}
	return st.GetEvent(ctx, id)
}

func Create(ctx context.Context, st Store, sess domain.Session, in WriteInput) (string, error) {
	if err := auth.RequireCalendar(sess); err != nil {
		return "", err
	}
	if strings.TrimSpace(in.Subject) == "" || in.Start == "" || in.End == "" {
		return "", domain.Usage("subject, start, and end are required")
	}
	if err := domain.ValidateEventRange(in.Start, in.End); err != nil {
		return "", err
	}
	if in.DryRun {
		return "", nil
	}
	return st.CreateEvent(ctx, in)
}

func Update(ctx context.Context, st Store, sess domain.Session, in WriteInput) error {
	if err := auth.RequireCalendar(sess); err != nil {
		return err
	}
	if strings.TrimSpace(in.ID) == "" {
		return domain.Usage("event id is required")
	}
	if in.Subject == "" && in.Start == "" && in.End == "" && in.Location == "" && in.Body == "" && len(in.Attendees) == 0 {
		return domain.Usage("at least one field is required")
	}
	if in.Start != "" || in.End != "" {
		stt, end := in.Start, in.End
		if stt == "" || end == "" {
			ev, err := st.GetEvent(ctx, in.ID)
			if err != nil {
				return err
			}
			if stt == "" {
				stt = ev.Start
			}
			if end == "" {
				end = ev.End
			}
		}
		if err := domain.ValidateEventRange(stt, end); err != nil {
			return err
		}
	}
	if in.DryRun {
		return nil
	}
	return st.UpdateEvent(ctx, in)
}

func Delete(ctx context.Context, st Store, sess domain.Session, id string, dry bool) error {
	if err := auth.RequireCalendar(sess); err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return domain.Usage("event id is required")
	}
	if dry {
		return nil
	}
	return st.DeleteEvent(ctx, id)
}
