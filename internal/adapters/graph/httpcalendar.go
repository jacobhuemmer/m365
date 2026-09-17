package graph

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/masonhuemmer/m365/internal/app/calendar"
	"github.com/masonhuemmer/m365/internal/domain"
)

type HTTPCalendar struct{ *HTTPClient }

func (c *HTTPCalendar) ListCalendars(ctx context.Context, top int, _ string) (domain.CalendarPage, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/calendars?$top="+strconv.Itoa(top))
	if err != nil {
		return domain.CalendarPage{}, err
	}
	defer res.Body.Close()
	var raw struct {
		Value []struct {
			ID, Name          string
			IsDefaultCalendar bool   `json:"isDefaultCalendar"`
			TimeZone          string `json:"timeZone"`
		} `json:"value"`
		Next string `json:"@odata.nextLink"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return domain.CalendarPage{}, domain.Service("invalid calendars")
	}
	items := make([]domain.Calendar, 0, len(raw.Value))
	for _, g := range raw.Value {
		items = append(items, domain.Calendar{ID: g.ID, Name: g.Name, IsDefault: g.IsDefaultCalendar, Timezone: mapTZ(g.TimeZone)})
	}
	p := domain.CalendarPage{Limit: top, Count: len(items), Items: items}
	if tok := encodeNext(raw.Next); tok != "" {
		p.NextPage = &tok
	}
	return p, nil
}

func (c *HTTPCalendar) ListEvents(ctx context.Context, q calendar.ListEventsQuery) (domain.EventPage, error) {
	path := "/me/calendar/calendarView?startDateTime=" + url.QueryEscape(q.Start) + "&endDateTime=" + url.QueryEscape(q.End) + "&$top=" + strconv.Itoa(q.Top)
	if q.Calendar != "" && q.Calendar != "default" {
		path = "/me/calendars/" + url.PathEscape(q.Calendar) + "/calendarView?startDateTime=" + url.QueryEscape(q.Start) + "&endDateTime=" + url.QueryEscape(q.End) + "&$top=" + strconv.Itoa(q.Top)
	}
	res, err := c.do(ctx, http.MethodGet, path)
	if err != nil {
		return domain.EventPage{}, err
	}
	defer res.Body.Close()
	var raw struct {
		Value []graphCalEvent `json:"value"`
		Next  string          `json:"@odata.nextLink"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return domain.EventPage{}, domain.Service("invalid events")
	}
	items := make([]domain.CalendarEvent, 0, len(raw.Value))
	for _, g := range raw.Value {
		ev := g.toEvent()
		ev.Body = ""
		items = append(items, ev)
	}
	p := domain.EventPage{Limit: q.Top, Count: len(items), Items: items}
	if tok := encodeNext(raw.Next); tok != "" {
		p.NextPage = &tok
	}
	return p, nil
}

func (c *HTTPCalendar) GetEvent(ctx context.Context, id string) (domain.CalendarEvent, error) {
	res, err := c.do(ctx, http.MethodGet, "/me/events/"+url.PathEscape(id))
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	defer res.Body.Close()
	var g graphCalEvent
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return domain.CalendarEvent{}, domain.Service("invalid event")
	}
	return g.toEvent(), nil
}

func (c *HTTPCalendar) CreateEvent(ctx context.Context, in calendar.WriteInput) (string, error) {
	body, _ := json.Marshal(eventBody(in))
	path := "/me/events"
	if in.CalendarID != "" {
		path = "/me/calendars/" + url.PathEscape(in.CalendarID) + "/events"
	}
	res, err := c.doBody(ctx, http.MethodPost, path, body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var g struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return "", domain.Service("invalid create")
	}
	return g.ID, nil
}

func (c *HTTPCalendar) UpdateEvent(ctx context.Context, in calendar.WriteInput) error {
	body, _ := json.Marshal(eventBody(in))
	res, err := c.doBody(ctx, http.MethodPatch, "/me/events/"+url.PathEscape(in.ID), body)
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	return nil
}

func (c *HTTPCalendar) DeleteEvent(ctx context.Context, id string) error {
	res, err := c.do(ctx, http.MethodDelete, "/me/events/"+url.PathEscape(id))
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	return nil
}

type graphCalEvent struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Start   struct {
		DateTime string `json:"dateTime"`
	} `json:"start"`
	End struct {
		DateTime string `json:"dateTime"`
	} `json:"end"`
	Location struct {
		DisplayName string `json:"displayName"`
	} `json:"location"`
	Organizer struct {
		EmailAddress struct{ Name, Address string } `json:"emailAddress"`
	} `json:"organizer"`
	Body struct {
		Content string `json:"content"`
	} `json:"body"`
}

func (g graphCalEvent) toEvent() domain.CalendarEvent {
	return domain.CalendarEvent{
		ID: g.ID, Subject: g.Subject, Start: g.Start.DateTime, End: g.End.DateTime,
		Location: g.Location.DisplayName, Body: g.Body.Content,
		Organizer: domain.Person{Name: g.Organizer.EmailAddress.Name, Address: g.Organizer.EmailAddress.Address},
	}
}

func eventBody(in calendar.WriteInput) map[string]any {
	m := map[string]any{}
	if in.Subject != "" {
		m["subject"] = in.Subject
	}
	if in.Start != "" {
		m["start"] = map[string]string{"dateTime": in.Start, "timeZone": "UTC"}
	}
	if in.End != "" {
		m["end"] = map[string]string{"dateTime": in.End, "timeZone": "UTC"}
	}
	if in.Location != "" {
		m["location"] = map[string]string{"displayName": in.Location}
	}
	if in.Body != "" {
		m["body"] = map[string]string{"content": in.Body, "contentType": "text"}
	}
	if len(in.Attendees) > 0 {
		var at []map[string]any
		for _, a := range in.Attendees {
			at = append(at, map[string]any{"emailAddress": map[string]string{"address": a}})
		}
		m["attendees"] = at
	}
	return m
}

func (c *HTTPClient) doBody(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	res, err := c.request(ctx, method, strings.TrimRight(c.Base, "/")+path, body, "application/json", "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		b, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		return nil, MapGraphError(res.StatusCode, b)
	}
	return res, nil
}

func mapTZ(s string) string {
	if s == "" {
		return "America/Chicago"
	}
	if loc, err := time.LoadLocation(s); err == nil {
		return loc.String()
	}
	switch s {
	case "Central Standard Time", "Central America Standard Time":
		return "America/Chicago"
	case "Eastern Standard Time":
		return "America/New_York"
	case "Pacific Standard Time":
		return "America/Los_Angeles"
	default:
		return s
	}
}
