# Feature Specification: Natural Calendar Times and Free Slots

**Feature Branch**: `003-calendar-natural-time`

**Created**: 2026-09-16

**Status**: Draft

**Input**: User description: "Create calendar events with language directives such as tomorrow at 1:30 pm instead of machine timestamps, and look up the next available times on my calendar for tomorrow."

## Specification Contract

### Objective

People and local agents schedule on the signed-in user's own calendar using ordinary English time phrases, and they can ask which times are still free on a named day (especially tomorrow) without reading the whole calendar themselves.

### Goals

- Let Mason create an event by saying when it starts in language, not by typing a full timestamp.
- Show the resolved local date and clock time on dry-run so a wrong parse is visible before anything is written.
- List the next free slots on his own calendar for a named day so he can pick a time that does not overlap existing events.
- Keep mail, teams, and files unchanged. Reuse calendar consent, dry-run, JSON/human output, and exit classes from the existing CLI.

### Non-goals

- Shared, delegate, or other people's calendars.
- Room or equipment booking, a scheduling-assistant product, or mailbox-wide free/busy of other people.
- Open-ended chat parsing ("sometime next week", "after lunch", "end of day").
- Recurring series creation, all-day events, and time-zone conversion as a user-facing product.
- Changing how mail, teams, or files commands work.
- Overnight monitors, kata wrappers, or MCP servers in this repository.

### Verification Strategy

Verification MUST cover phrase create with dry-run, refusal of ambiguous phrases, default duration, free-slot listing for tomorrow, empty-day (fully booked) success, and no change to mail or teams. Fixtures MUST be synthetic. Live event bodies MUST NOT be stored in the repository.

## Document Map

This file is the overview. It extends [002-calendar-files](../002-calendar-files/spec.md) for the calendar namespace only. Files/OneDrive is out of this feature.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create an event with a language directive (Priority: P1)

Mason or a local agent creates an event by saying when it starts in ordinary language, for example `tomorrow at 1:30 pm`. They do not type a machine timestamp. Dry-run shows the original phrase and the resolved local date, weekday, and clock time. A real create without `--dry-run` writes that resolved event. Duration defaults to thirty minutes unless they name an end or a length.

**Why this priority**: This is the reason to schedule from a terminal or an agent. Timestamp flags are the thing Mason does not want to type.

**Independent Test**: With calendar consent and a frozen "now", dry-run create with `tomorrow at 1:30 pm` and assert stdout shows tomorrow's local date at 1:30 PM for thirty minutes and that no event is created. Real create then returns an id whose get matches that window. An ambiguous phrase such as `1:30` with no day and no am/pm exits `3` and creates nothing.

**Acceptance Scenarios**:

1. **Given** a usable calendar session and a known local "now", **When** the user runs `calendar create --dry-run` with subject `Sync` and when `tomorrow at 1:30 pm`, **Then** stdout shows the original phrase, the resolved local date and 1:30 PM, an end thirty minutes later, and no event is created.
2. **Given** the same inputs without `--dry-run`, **When** the user runs `calendar create`, **Then** an event is created on the default calendar and stdout reports its id.
3. **Given** `tomorrow at 1:30 pm` and duration `1 hour`, **When** the user dry-runs create, **Then** the shown end is 2:30 PM local on that same date.
4. **Given** `tomorrow at 1:30 pm` and until `2:00 pm`, **When** the user dry-runs create, **Then** the shown window is 1:30 PM to 2:00 PM local.
5. **Given** an ambiguous or unsupported phrase (`Friday`, `1:30`, `next week`, `after lunch`), **When** the user runs create, **Then** the command exits `3`, names that the when-phrase could not be used, and creates nothing.
6. **Given** a start that has already passed in the calendar's local time (`today at 9am` when it is already afternoon), **When** the user runs create, **Then** the command exits `3` and creates nothing.

---

### User Story 2 - List the next free times tomorrow (Priority: P1)

Mason or a local agent asks for the next available times on his own default calendar for tomorrow. The command returns a bounded list of start–end slots that fit a meeting length (default thirty minutes), fall inside working hours, and do not overlap existing events. It does not create an event.

**Why this priority**: Finding a gap is the other half of natural scheduling. It is useful even if create still used timestamps.

**Independent Test**: With synthetic events that occupy part of tomorrow's working hours, run `calendar free` for tomorrow at the default duration and default limit. Assert returned slots sit in tomorrow's working hours, do not overlap those events, are at most `--top` long, and that a fully booked working day exits `0` with an empty list.

**Acceptance Scenarios**:

1. **Given** a default calendar with some busy periods tomorrow and some open gaps, **When** the user runs `calendar free` with when `tomorrow`, **Then** the command exits `0` and lists the next free slots of the default duration inside tomorrow's working hours, bounded by `--top`.
2. **Given** no `--when`, **When** the user runs `calendar free`, **Then** the day is tomorrow in the calendar's local timezone.
3. **Given** duration `1 hour`, **When** the user runs `calendar free` for tomorrow, **Then** every listed slot is one hour long and still does not overlap existing events.
4. **Given** a working day with no remaining gap of the requested length, **When** the user runs `calendar free`, **Then** the command exits `0` with an empty list (this is not not-found).
5. **Given** `--top 3`, **When** more than three slots exist, **Then** at most three are returned and the applied limit is visible.
6. **Given** `calendar free --dry-run`, **When** the user runs it, **Then** the command is rejected as usage (`3`) because free does not write; dry-run does not apply.

---

### User Story 3 - Confirm the resolved time before writing (Priority: P2)

On dry-run create, the user sees both the language they typed and the exact local start and end that would be written. JSON output includes those resolved times so an agent can show or reuse them without guessing.

**Why this priority**: Phrase parsing is only safe if the resolved time is inspectable. It depends on create already accepting phrases.

**Independent Test**: Dry-run create with `tomorrow at 1:30 pm`. Assert human and JSON output both include the original phrase and the resolved local start and end, and that a following `calendar get` of a real create matches those resolved times.

**Acceptance Scenarios**:

1. **Given** a dry-run create with `tomorrow at 1:30 pm`, **When** the user reads stdout, **Then** they can see the phrase they typed and the resolved local date, start clock time, and end clock time without opening the calendar.
2. **Given** JSON mode, **When** the same dry-run runs, **Then** the single JSON value includes the original phrase and the resolved start and end.
3. **Given** a real create after a matching dry-run, **When** the user gets the new event, **Then** its start and end match the dry-run resolution.

---

### Edge Cases

- Phrase with a day but no clock time (`tomorrow`, `Friday`).
- Phrase with a clock time but no day and no am/pm (`1:30`).
- Supported weekday name (`Friday at 9am`) when that weekday is today and the clock time has already passed.
- `today` during working hours versus after working hours.
- Start at or after end (`tomorrow at 2pm` until `1pm`).
- Duration that is not a whole number of minutes, zero, or negative.
- `--when` and `--until` on different calendar dates.
- Free search on a day that is fully booked inside working hours.
- Free search when "now" is on the requested day (only remaining future gaps).
- Unknown calendar name on free or create.
- Missing calendar consent (exit `4`, not `5`).
- `--top` omitted (documented default) and `--top` above the maximum (exit `3`).
- Recurring events that occupy part of tomorrow: those occurrences count as busy for free search; the command still MUST NOT dump the whole series.
- Mail and teams commands remain unchanged.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: This feature MUST extend the existing `calendar` namespace. It MUST NOT add a new top-level namespace and MUST NOT change mail, teams, or files command names, flags, output shape, or exit classes.
- **FR-002**: `calendar create` MUST accept a when-phrase as the primary way to name the start. The user MUST NOT be required to type a machine timestamp. Help examples MUST use phrases such as `tomorrow at 1:30 pm`.
- **FR-003**: A when-phrase MUST resolve in the target calendar's local timezone (default calendar unless a calendar is named). Dry-run and result output MUST show that local date and clock time.
- **FR-004**: Accepted when-phrases are the closed set in Assumptions (relative days, next weekday plus a clock time, optional calendar date plus a clock time). Clock time MUST include am/pm or use a 24-hour clock. Any other wording MUST exit `3` and MUST NOT create.
- **FR-005**: If duration is omitted and until is omitted, create MUST use a default duration of thirty minutes. The user MAY supply a duration phrase (`30 minutes`, `1 hour`) or an until-phrase that resolves on the same local date.
- **FR-006**: Until, if present, MUST be a clock time or a full when-phrase on the same local date as start. Start at or after end MUST exit `3`.
- **FR-007**: A resolved start that is already in the past in the calendar's local timezone MUST exit `3` and MUST NOT create.
- **FR-008**: Dry-run create MUST print the original when-phrase and the resolved local start and end, and MUST NOT create. Real create is a separate invocation without `--dry-run`.
- **FR-009**: JSON create output (dry-run and real) MUST include the original phrase and the resolved start and end as one parseable JSON value with no log lines.
- **FR-010**: `calendar update` MAY accept the same when-phrase and duration/until rules when the user is changing the time. Update still requires an event id.
- **FR-011**: `calendar free` MUST list free slots on the signed-in user's own named calendar (default: default calendar) for a named day. Default day is tomorrow in that calendar's local timezone when `--when` is omitted.
- **FR-012**: `calendar free` MUST treat existing events on that calendar as busy, including occurrences of recurring events that fall in the day. It MUST NOT inspect other people's calendars or shared calendars.
- **FR-013**: Free slots MUST lie inside working hours for that local day, MUST be at least the requested duration long (default thirty minutes), MUST not overlap busy times, and MUST start on the documented slot grid. Default working hours and grid are in Assumptions.
- **FR-014**: If the requested day is today, free search MUST ignore gaps that have already ended. It MUST NOT propose a slot whose start is in the past.
- **FR-015**: `calendar free` MUST apply an explicit ResultLimit. Default `--top` is `5`. Maximum `--top` is `20`. Omitting `--top` applies the default. Above-max `--top` exits `3`. A fully booked day exits `0` with an empty list.
- **FR-016**: `calendar free` MUST NOT create, update, or delete events. `--dry-run` on free is a usage error (exit `3`).
- **FR-017**: `--help` for `calendar create` and `calendar free` MUST show phrase examples, default duration, default free day (tomorrow), working hours, and limits. Help exits `0` without a session.
- **FR-018**: Missing calendar consent MUST exit `4`. Unknown calendar MUST exit `6`. Ambiguous or unsupported phrases MUST exit `3`. Microsoft 365 service errors MUST exit `5`.
- **FR-019**: Tests and fixtures MUST use a frozen "now" and synthetic events. Live calendar content MUST NOT be stored in the repository.
- **FR-020**: This feature MUST NOT include room booking, other-attendee free/busy, all-day events, recurrence rules on create, time-zone conversion as a user-facing feature, or a generic natural-language assistant.

### Key Entities *(include if feature involves data)*

- **WhenPhrase**: The language the user typed to name a day and clock time (for example `tomorrow at 1:30 pm`). Stored on dry-run and create results as the original text.
- **ResolvedInterval**: The local start and end after a phrase is understood: local date, start clock time, end clock time, timezone name of the calendar. This is what would be written.
- **DurationPhrase**: Optional length (`30 minutes`, `1 hour`). Default thirty minutes.
- **FreeSlot**: One available interval on a day: start, end, duration, calendar id. Does not include event bodies.
- **WorkingHours**: Inclusive local start clock time and exclusive local end clock time used only by free search.
- **CalendarEvent**: Unchanged from 002. Busy for free search when an occurrence overlaps the day.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A person can dry-run create an event using only `tomorrow at 1:30 pm` (plus a subject) and see the resolved local date and 1:30 PM without typing a timestamp.
- **SC-002**: In 100% of successful dry-run creates, stdout shows both the original phrase and the resolved local start and end.
- **SC-003**: Ambiguous phrases (`1:30`, `Friday`, `next week`) never create an event; they fail as usage errors.
- **SC-004**: `calendar free` with no day argument returns slots for tomorrow's working hours, at most five by default, none overlapping existing events on that calendar.
- **SC-005**: A fully booked working day returns success with an empty slot list, not a not-found or service error.
- **SC-006**: Default-duration create without an until-phrase is thirty minutes long, visible on dry-run.
- **SC-007**: Mail and teams command names, flags, output shape, and exit classes remain unchanged.
- **SC-008**: After a usable session exists, a default `calendar free` for tomorrow completes in under 15 seconds on an ordinary connection. A dry-run phrase create completes in under 5 seconds and does not create an event.
- **SC-009**: No command prints access tokens, refresh tokens, authorization codes, or client secrets.
- **SC-010**: JSON stdout for create dry-run and for free is exactly one parseable JSON value with no mixed-in diagnostics.

## Assumptions

- This feature sits on `002-calendar-files`. Calendar list, get, create, update, delete, own-calendar-only, dry-run, JSON default, and exit classes `0/3/4/5/6` still apply. This spec changes how create (and optional update) names times, and adds `calendar free`.
- Target timezone is the named calendar's timezone. If the calendar has no timezone, use the signed-in mailbox's timezone. Commands MUST show which local timezone they used.
- Default duration is 30 minutes.
- Default working hours for free search are 09:00 to 17:00 in that local timezone. `--hours` MAY override as `start-end` clock times on the same day (for example `9:00-17:00`).
- Free slots start on a 30-minute grid (09:00, 09:30, …) unless the duration itself forces a different alignment; the grid is documented in `--help`.
- Default `--top` for `calendar free` is `5`; maximum is `20`.
- Default day for `calendar free` is tomorrow relative to "now" in the calendar timezone.
- Accepted when-phrases (closed set):
  - `today` or `tomorrow` plus a clock time: `tomorrow at 1:30 pm`, `today 9am`, `tomorrow at 13:30`
  - Next occurrence of a weekday plus a clock time: `Friday at 9am` (if that clock time today has passed, the next week's that weekday)
  - A calendar date plus a clock time: `18 September 2026 at 1:30 pm` or `2026-09-18 13:30`
- Clock times MUST include `am`/`pm` or use 24-hour `HH:mm`. Bare `1:30` is invalid.
- A day word without a clock time is invalid for create. It is valid for `calendar free` (`tomorrow`, `Friday`, `today`).
- Duration phrases: `N minutes` or `N hour`/`N hours` with N a positive integer. Maximum duration is 8 hours.
- Until-phrase is a clock time on the start's local date, or a full when-phrase that resolves to the same local date.
- Create still requires a subject. Default calendar unless a calendar is named.
- "Busy" is any existing event occurrence on that one calendar that overlaps the candidate slot. Tentative and "busy" events both block. Cancelled occurrences do not. All-day events on that date block the whole working day.
- Free search does not book. The user runs `calendar create` separately with a when-phrase.
- A machine timestamp, if supplied as the entire when-phrase, MAY be accepted so agents that already have an instant are not blocked. Help and human examples MUST still lead with language directives. Users are never required to type one.
- Client ID, Tenant ID, and secrets MUST NOT appear in this specification, source, tests, logs, or fixtures.
