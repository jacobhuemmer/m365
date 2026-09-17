package cli

import (
	"flag"

	"github.com/masonhuemmer/m365/internal/app/calendar"
	"github.com/masonhuemmer/m365/internal/domain"
)

func runCalendar(args []string, d Deps, format string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeHelp(d.Stdout, calendarHelp)
	}
	if d.Calendar == nil {
		return fail(d, domain.Usage("calendar not available"))
	}
	verb, args := args[0], args[1:]
	sess, err := session(d)
	if err != nil {
		return fail(d, err)
	}
	switch verb {
	case "calendars":
		if hasHelp(args) {
			return writeHelp(d.Stdout, calendarHelp)
		}
		fsset := flag.NewFlagSet("calendar calendars", flag.ContinueOnError)
		fsset.SetOutput(d.Stderr)
		top := fsset.Int("top", 0, "")
		page := fsset.String("page-token", "", "")
		if err := parseMixed(fsset, args); err != nil {
			return fail(d, domain.Usage(err.Error()))
		}
		p, err := calendar.ListCalendars(ctx(), d.Calendar, sess, *top, *page)
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, p)
	case "list":
		if hasHelp(args) {
			return writeHelp(d.Stdout, calendarHelp)
		}
		fsset := flag.NewFlagSet("calendar list", flag.ContinueOnError)
		fsset.SetOutput(d.Stderr)
		cal := fsset.String("calendar", "", "")
		start := fsset.String("start", "", "")
		end := fsset.String("end", "", "")
		top := fsset.Int("top", 0, "")
		page := fsset.String("page-token", "", "")
		if err := parseMixed(fsset, args); err != nil {
			return fail(d, domain.Usage(err.Error()))
		}
		p, err := calendar.ListEvents(ctx(), d.Calendar, sess, calendar.ListEventsQuery{
			Calendar: *cal, Start: *start, End: *end, Top: *top, PageToken: *page,
		})
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, p)
	case "get":
		if len(args) < 1 {
			return fail(d, domain.Usage("event id is required"))
		}
		ev, err := calendar.GetEvent(ctx(), d.Calendar, sess, args[0])
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, ev)
	case "create":
		return calendarCreate(args, d, sess, format)
	case "update":
		return calendarUpdate(args, d, sess, format)
	case "delete":
		return calendarDelete(args, d, sess, format)
	case "free":
		return calendarFree(args, d, sess, format)
	default:
		return fail(d, domain.Usagef("unknown calendar verb %q", verb))
	}
}

func calendarCreate(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("calendar create", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	subject := fsset.String("subject", "", "")
	start := fsset.String("start", "", "")
	end := fsset.String("end", "", "")
	loc := fsset.String("location", "", "")
	body := fsset.String("body", "", "")
	bodyFile := fsset.String("body-file", "", "")
	cal := fsset.String("calendar", "", "")
	when := fsset.String("when", "", "")
	until := fsset.String("until", "", "")
	dur := fsset.String("duration", "", "")
	dry := fsset.Bool("dry-run", false, "")
	var attendees []string
	fsset.Func("attendee", "", func(s string) error { attendees = append(attendees, s); return nil })
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	b, err := readBody(*body, *bodyFile, d.Stdin)
	if err != nil {
		return fail(d, err)
	}
	in := calendar.WriteInput{
		CalendarID: *cal, Subject: *subject, Start: *start, End: *end, Location: *loc, Body: b, Attendees: attendees,
		When: *when, Until: *until, Duration: *dur, Now: calendar.Clock(), TZ: calendarTZ(d, sess, *cal), DryRun: *dry,
	}
	id, err := calendar.Create(ctx(), d.Calendar, sess, &in)
	if err != nil {
		return fail(d, err)
	}
	out := map[string]any{"subject": *subject, "start": in.Start, "end": in.End, "timezone": in.TZ, "when": *when}
	if *dry {
		out["dry_run"] = true
		return success(d, format, out)
	}
	out["id"] = id
	return success(d, format, out)
}

func calendarUpdate(args []string, d Deps, sess domain.Session, format string) int {
	if len(args) < 1 {
		return fail(d, domain.Usage("event id is required"))
	}
	id, args := args[0], args[1:]
	fsset := flag.NewFlagSet("calendar update", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	subject := fsset.String("subject", "", "")
	start := fsset.String("start", "", "")
	end := fsset.String("end", "", "")
	loc := fsset.String("location", "", "")
	body := fsset.String("body", "", "")
	when := fsset.String("when", "", "")
	until := fsset.String("until", "", "")
	dur := fsset.String("duration", "", "")
	dry := fsset.Bool("dry-run", false, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	if err := calendar.Update(ctx(), d.Calendar, sess, calendar.WriteInput{
		ID: id, Subject: *subject, Start: *start, End: *end, Location: *loc, Body: *body,
		When: *when, Until: *until, Duration: *dur, Now: calendar.Clock(), TZ: calendarTZ(d, sess, ""), DryRun: *dry,
	}); err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, map[string]any{"dry_run": true, "id": id})
	}
	return success(d, format, map[string]any{"id": id})
}

func calendarDelete(args []string, d Deps, sess domain.Session, format string) int {
	if len(args) < 1 {
		return fail(d, domain.Usage("event id is required"))
	}
	fsset := flag.NewFlagSet("calendar delete", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	dry := fsset.Bool("dry-run", false, "")
	id := args[0]
	_ = parseMixed(fsset, args[1:])
	if err := calendar.Delete(ctx(), d.Calendar, sess, id, *dry); err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, map[string]any{"dry_run": true, "id": id})
	}
	return success(d, format, map[string]any{"id": id, "deleted": true})
}

func calendarTZ(d Deps, sess domain.Session, id string) string {
	p, err := calendar.ListCalendars(ctx(), d.Calendar, sess, 20, "")
	if err != nil || len(p.Items) == 0 {
		return "America/Chicago"
	}
	for _, c := range p.Items {
		if id != "" && (c.ID == id || c.Name == id) && c.Timezone != "" {
			return c.Timezone
		}
		if id == "" && c.IsDefault && c.Timezone != "" {
			return c.Timezone
		}
	}
	if p.Items[0].Timezone != "" {
		return p.Items[0].Timezone
	}
	return "America/Chicago"
}

func calendarFree(args []string, d Deps, sess domain.Session, format string) int {
	if hasHelp(args) {
		return writeHelp(d.Stdout, calendarHelp)
	}
	fsset := flag.NewFlagSet("calendar free", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	when := fsset.String("when", "", "")
	dur := fsset.String("duration", "", "")
	hours := fsset.String("hours", "", "")
	cal := fsset.String("calendar", "", "")
	top := fsset.Int("top", 0, "")
	dry := fsset.Bool("dry-run", false, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	if *dry {
		return fail(d, domain.Usage("dry-run does not apply to calendar free"))
	}
	p, err := calendar.Free(ctx(), d.Calendar, sess, calendar.FreeQuery{
		When: *when, Duration: *dur, Hours: *hours, Calendar: *cal, Top: *top,
		Now: calendar.Clock(), Location: calendarTZ(d, sess, *cal),
	})
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, p)
}
