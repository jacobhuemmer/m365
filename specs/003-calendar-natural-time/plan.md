# Implementation Plan: Natural Calendar Times and Free Slots

**Branch**: `main` (setup-plan JSON reported `003-calendar-natural-time`; working tree is `main`) | **Date**: 2026-09-17 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/003-calendar-natural-time/spec.md`

**Note**: Design only. No application source, `tasks.md`, MCP, or CI files. MUST NOT rewrite `specs/001-m365-cli/` or `specs/002-calendar-files/` except additive calendar help/catalog notes if required at implement time.

## Summary

Extend the existing `calendar` namespace so create (and optional update) accepts a closed set of English when-phrases such as `tomorrow at 1:30 pm`, and add `calendar free` to list the next non-overlapping slots on the signed-in user's own calendar for a named day (default tomorrow). Parsing and slot math live in `internal/app/calendar` with an injected clock and timezone. Graph stays in the adapter: calendarView for that local day as busy intervals, calendar timezone from the calendar or mailbox settings. Mail, teams, and files MUST NOT change. Dry-run create MUST show the original phrase and the resolved local start/end. Tests freeze "now" and use synthetic events.

## Technical Context

**Language/Version**: Go 1.25+ module `github.com/masonhuemmer/m365` (exists).

**Primary Dependencies**: Stdlib only for phrase parsing (`time`, `strings`, `regexp` if needed). Existing HTTP Graph adapter, `flag`, `testing`, `httptest`. No NLP library, no `when`, no `dateparse`, no Graph SDK, no Cobra.

**Storage**: Unchanged token store. No new secret files. No persistence of parsed phrases.

**Testing**: Go `testing`; freeze `Now` in calendar use cases; `httptest` fake Graph with synthetic busy events; APS Gherkin under `features/calendar/`. No live Graph in unit tests. No live event bodies in the repo.

**Target Platform**: macOS CLI, one signed-in user or local agent.

**Project Type**: Same single-module local CLI.

**Performance Goals**: SC-008 — default `calendar free` for tomorrow under 15s (CI: fake Graph). Dry-run phrase create under 5s and MUST NOT call create HTTP.

**Constraints**: Exit classes `0/3/4/5/6`; JSON default; no secrets; own calendar only; closed phrase set; default duration 30 minutes; free `--top` default 5 max 20; working hours 09:00–17:00 local; 30-minute slot grid; `--dry-run` on free is usage (3); mail/teams/files unchanged.

**Scale/Scope**: One new verb (`calendar free`); additive flags on `calendar create`/`update` (`--when`, `--until`, `--duration`). Phrase parser is a closed grammar, not a general assistant.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. TDD. PASS.** RED Gherkin + unit/CLI tests with frozen now; RED commit then GREEN. EX-I-001 does not apply.
- [x] **II. Clean Code. PASS.** Parser, free-slot math, and CLI wiring in separate files under 250 lines. No util grab-bag named `timeutil`.
- [x] **III. Smallest Sufficient Design. PASS.** Closed phrase set from Assumptions, not an NLP engine. Do not pre-build "next week" or attendee free/busy.
- [x] **IV. Testing. PASS.** Injected clock and `*time.Location`. Fake Graph busy intervals. No live calendar dumps.
- [x] **V. CLI Consistency. PASS.** Same streams, JSON/human, exits, `--help`, dry-run on create. New `--when` is documented; `--start`/`--end` remain valid as a full when-phrase (RFC3339) so agents are not blocked.
- [x] **VI. Performance. PASS.** SC-008 metrics and fake-Graph CI deadline. Free search uses one day's calendarView (`$top` bounded), not an unbounded series dump.
- [x] **VII. Isolation. PASS.** Logic stays in `internal/app/calendar`. MUST NOT import mail, teams, or files. Graph URLs stay in the adapter. `chat` unchanged.
- [x] **VIII. Secrets. PASS.** Same delegated session. No new scopes required beyond 002 calendar. No secrets in plan, tests, or fixtures.
- [x] **Engineering Constraints. PASS.** Existing Makefile gates. Visible limits (free `--top`, working hours, grid, duration).
- [x] **Workflow. PASS.** Specify (done) → this check → RED → GREEN → `make verify`.

No new constitution exceptions.

**Post-design re-check (after Phase 1): PASS.** Parser is domain/app with frozen now; contracts lock `--when`/`calendar free`; fake Graph supplies timezone + one-day busy; mail/teams/files catalogs unchanged.

## Project Structure

### Documentation (this feature)

```text
specs/003-calendar-natural-time/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── command-catalog.md
│   ├── fake-graph.md
│   ├── json-create.schema.json
│   └── json-free-slot.schema.json
└── tasks.md                 # /speckit-tasks, not this command
```

### Source Code (implement time)

```text
internal/app/calendar/
├── when.go          # closed phrase → ResolvedInterval (injected now + location)
├── when_test.go
├── free.go          # working-hours grid minus busy intervals
├── free_test.go
└── ports.go         # extend create/update; add Free use case
internal/adapters/cli/calendar.go   # --when --until --duration; free verb
internal/adapters/graph/httpcalendar.go  # calendar timezone; day calendarView for free
features/calendar/when-create.feature
features/calendar/free.feature
```

**Structure Decision**: No new package. Phrase parsing is application logic next to calendar use cases. CLI only collects flags. Adapter only maps Graph timezone and calendarView.

## Implementation Workflow

1. RED: frozen-now table tests for the closed phrase set (valid + every listed invalid).
2. RED: free-slot tests (gaps, fully booked empty success, today ignores past, `--top`).
3. RED commit (tests/fixtures only).
4. GREEN: parser, free math, create/free wiring, fake timezone + busy day.
5. GREEN commit. Isolation tests still forbid calendar→mail/teams/files.
6. `make verify` on fakes.

## Graph mapping (adapter only)

- Calendar timezone: `timeZone` on `GET /me/calendars/{id}` or `GET /me/calendar`. If empty, `GET /me/mailboxSettings` `timeZone` (Windows or IANA name mapped in the adapter to `*time.Location`). Domain sees an IANA name string, not Graph types.
- Busy for a local day: `calendarView` with that day's local start/end (not `/me/events` without a window).
- Create still POST `/me/calendars/{id}/events` with resolved RFC3339 instants. Dry-run MUST NOT POST.

## Make Targets

Reuse existing Makefile. Add CLI tests: phrase dry-run does not hit Graph create; `calendar free` against fake finishes under 15s.

## Files To Add At Implement Time

- `when.go` / `free.go` and tests.
- CLI flags and `calendar free`.
- Fake Graph calendar `timeZone` and a synthetic tomorrow with mixed busy/free.
- Gherkin features.
- Do not edit `mail.go` or `teams.go`.

## Out of This Plan

- `/speckit-tasks`, `/speckit-implement`.
- Shared calendars, rooms, other-attendee free/busy, recurrence on create, all-day create, NLP.
- Rewriting 001/002 specs.

## Risks And Tradeoffs

- **Windows timezone names from mailboxSettings**: adapter maps common names to IANA; unknown name → usage/service with a hint, not a silent UTC.
- **Closed grammar vs convenience**: "next week" is invalid by spec; keep it invalid.
- **`--start`/`--end` vs `--when`**: both allowed; if `--when` is set, it is the start; `--start` RFC3339 counts as a when-phrase. If both `--when` and `--start` are set, exit `3`.

## Complexity Tracking

No new constitution exceptions.
