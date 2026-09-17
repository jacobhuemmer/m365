# Phase 0 Research: Natural Calendar Times and Free Slots

## Decision: Closed stdlib parser, not an NLP library

**Rationale**: Spec FR-004 is a closed phrase set. Constitution III forbids a generic assistant. `time`, `strings`, and small explicit rules in `internal/app/calendar/when.go` are enough. Inject `now time.Time` and `loc *time.Location`.

**Alternatives considered**:
- `github.com/araddon/dateparse` or `when`: extra dependency, accepts phrasing the spec rejects (`next week`, bare `1:30`).
- Shell out to `date`: not testable with a frozen now inside the CLI, and is what Mason asked to stop doing.

## Decision: Phrase resolution in the calendar app package

**Rationale**: Timezone belongs to the calendar, not the OS. CLI passes raw strings. Domain types (`WhenPhrase`, `ResolvedInterval`) have no Graph URLs. Use cases call `ResolveWhen(phrase, now, loc)`.

**Alternatives considered**:
- Parse in the CLI adapter: puts behavior in wiring (constitution II).
- Parse in the Graph adapter: couples English to HTTP.

## Decision: Busy intervals from one-day calendarView

**Rationale**: 002 already uses calendarView so list stays inside a window (FR-016 of 002 / FR-012 of 003). Free search loads that local day's occurrences, including recurring instances, then subtracts them from the working-hours grid in `free.go`. No `/users/{other}/calendar/getSchedule`.

**Alternatives considered**:
- `getSchedule` / mailbox-wide free/busy: out of spec (other people).
- Client-side series expansion: forbidden dump risk.

**Sources**: Microsoft Graph calendarView (observed 2026-09-17): https://learn.microsoft.com/en-us/graph/api/calendar-list-calendarview

## Decision: Timezone from the calendar, then mailbox settings

**Rationale**: Spec: target timezone is the named calendar's timezone; if missing, the mailbox timezone. Graph `calendar.timeZone` or `mailboxSettings.timeZone` may be Windows names (`Central Standard Time`). Adapter maps to IANA (`America/Chicago`) and returns the IANA name on results. Tests use a fake calendar with `America/Chicago`.

**Alternatives considered**:
- Always `time.Local` (machine TZ): wrong when the mailbox is CT and the agent host is not.
- Always UTC: contradicts "local date and clock time" on dry-run.

## Decision: Flags `--when`, `--until`, `--duration`; RFC3339 still valid as a when-phrase

**Rationale**: Help leads with language. Spec allows a machine timestamp as the entire when-phrase so agents are not blocked. `--when` and `--start` together → exit 3. Default duration 30 minutes when `--until` and `--duration` omitted.

**Alternatives considered**:
- Replace `--start`/`--end`: breaks 002 callers and live tests that already pass RFC3339.

## Decision: Free search is local interval arithmetic

**Rationale**: Working hours 09:00–17:00, 30-minute grid, duration default 30 minutes, `--top` default 5 max 20. Walk the grid, skip overlaps and past slots if the day is today. Fully booked → empty list exit 0. `--dry-run` on free → usage 3.

**Alternatives considered**:
- Graph `findMeetingTimes`: other-attendee scheduling assistant; out of spec.
