# Tasks: Natural Calendar Times and Free Slots

**Input**: Design documents from `/specs/003-calendar-natural-time/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md, constitution.md, existing 002 calendar namespace

**Tests**: MANDATORY per Constitution I. Frozen `now` in tests. RED commit (failing tests + fixtures only) then GREEN. EX-I-001 does not apply. No live Graph in unit tests. No live event bodies in the repo. Do **not** edit `internal/adapters/cli/mail.go` or `teams.go`.

**Organization**: Setup + Foundational block stories. US1–US3 map to spec User Stories 1–3. MVP = Setup + Foundational + US1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallel (different files, no incomplete deps)
- **[Story]**: `[US1]`–`[US3]` on user-story phases only
- Exact file paths required

## Path Conventions

`internal/app/calendar/`, `internal/adapters/cli/calendar.go`, `internal/adapters/graph/httpcalendar.go`, `internal/adapters/graph/fakecal.go`, `features/calendar/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Feature files live in the existing module. No new go.mod.

- [X] T001 Confirm `internal/app/calendar/` and `features/calendar/` exist from 002; do not recreate the module or Makefile

**Checkpoint**: `make unit` still passes.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain types, closed phrase parser with injected now+location, fake calendar timezone. Blocks US1–US3.

**⚠️ CRITICAL**: No user story work until this phase is complete. TDD: RED before GREEN.

- [X] T002 [P] Write RED tests in internal/domain/when_test.go then add WhenPhrase and ResolvedInterval in internal/domain/when.go quoting data-model: "Start MUST be before end." "Default end = start + 30 minutes when duration and until are omitted." "Duration MUST be a positive whole number of minutes, at most 8 hours."
- [X] T003 Write RED table tests in internal/app/calendar/when_test.go (frozen now `2026-09-16T12:00:00-05:00`, location `America/Chicago`) for valid phrases `tomorrow at 1:30 pm`, `today 9am`, `Friday at 9am`, `2026-09-18 13:30`, RFC3339-as-when; invalid `1:30`, `Friday`, `next week`, `after lunch`; implement ResolveWhen/ResolveDay/ParseDuration in internal/app/calendar/when.go (stdlib only; no NLP deps)
- [X] T004 [P] Extend fake calendar `cal-1` with `timeZone` `America/Chicago` in internal/adapters/graph/fakecal.go per contracts/fake-graph.md
- [X] T005 [P] Write RED tests in internal/adapters/graph/calendar_tz_test.go then return IANA timezone from GET calendar in internal/adapters/graph/httpcalendar.go (Windows name mapped in the adapter; unknown name is usage/service with a hint, not silent UTC)

**Checkpoint**: Phrase tests fail for unimplemented cases then pass; `fake-both` still denies calendar; mail/teams tests green.

---

## Phase 3: User Story 1 — Create an event with a language directive (Priority: P1) 🎯 MVP

**Goal**: `calendar create --when 'tomorrow at 1:30 pm'` without a machine timestamp. Default duration 30 minutes. Ambiguous phrases exit 3. Past start exits 3. Dry-run does not create.

**Independent Test**: Frozen now; dry-run `--when 'tomorrow at 1:30 pm' --subject Sync` shows 2026-09-17 13:30 CT for 30 minutes and no POST; real create returns id; `--when '1:30'` exits 3; `--when` and `--start` together exit 3.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 1 (MANDATORY)

- [X] T006 [P] [US1] Write Gherkin in features/calendar/when-create.feature for dry-run phrase, duration 1 hour, until 2:00 pm, ambiguous phrase exit 3, past `today at 9am` at noon exit 3; generate failing acceptance tests via scripts/acceptance.sh
- [X] T007 [P] [US1] Write RED tests in internal/app/calendar/create_when_test.go quoting "A resolved start that is already in the past in the calendar's local timezone MUST exit `3`" and default thirty minutes
- [X] T008 [P] [US1] Write RED CLI tests in internal/adapters/cli/calendar_when_test.go for `--when`, `--duration`, `--until`, `--when`+`--start` → 3, dry-run JSON `when`/`start`/`end`/`timezone`
- [X] T009 [US1] RED commit: git add only failing tests and fixtures from T006–T008; no production code

### Implementation for User Story 1

- [X] T010 [US1] Extend create/update use cases in internal/app/calendar/ports.go to resolve `--when`/`--until`/`--duration` with injected now and calendar location; RFC3339 as entire when-phrase still allowed
- [X] T011 [US1] Register `--when`, `--until`, `--duration` on create/update in internal/adapters/cli/calendar.go; help examples MUST lead with `tomorrow at 1:30 pm`; do not remove `--start`/`--end`
- [X] T012 [US1] Dry-run create MUST NOT call CreateEvent HTTP (assert in internal/adapters/graph/calendar_write_test.go / use-case tests)
- [X] T013 [US1] Implement acceptance/steps/calendar_when_steps.go until scripts/acceptance.sh passes features/calendar/when-create.feature
- [X] T014 [US1] After GREEN T010–T013, commit production code separately from T009; do not edit locked RED tests

**Checkpoint**: US1 independently testable on fakes; 002 RFC3339 create still works.

---

## Phase 4: User Story 2 — List the next free times tomorrow (Priority: P1)

**Goal**: `calendar free` lists bounded non-overlapping slots in working hours for a named day (default tomorrow). Fully booked → exit 0 empty. `--dry-run` → 3.

**Independent Test**: Fake tomorrow with busy 09:00–10:00 and 13:00–14:00 CT; `calendar free` default duration 30 minutes, `--top` 5; slots inside 09:00–17:00, no overlap; wall-to-wall busy day count 0 exit 0; `--dry-run` exit 3.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 2 (MANDATORY)

- [X] T015 [P] [US2] Write Gherkin in features/calendar/free.feature for default tomorrow, duration 1 hour, `--top 3`, empty booked day, `--dry-run` exit 3; generate failing acceptance tests via scripts/acceptance.sh
- [X] T016 [P] [US2] Write RED tests in internal/app/calendar/free_test.go quoting data-model: "Default `--top` = 5. Maximum = 20. Values `<1` or `>20` are usage, not silent caps." "Empty list (fully booked) is success." "If the day is today, start MUST be after now." Default working hours 09:00–17:00, 30-minute grid
- [X] T017 [P] [US2] Write RED CLI tests in internal/adapters/cli/calendar_free_test.go for `calendar free`, `--when tomorrow`, `--duration`, `--hours`, `--top 21` → 3, `--dry-run` → 3
- [X] T018 [US2] RED commit: failing tests from T015–T017 only

### Implementation for User Story 2

- [X] T019 [P] [US2] Implement FreeSlot/WorkingHours/BusyInterval helpers and Free use case in internal/app/calendar/free.go (no Graph URLs)
- [X] T020 [US2] Load one local day's calendarView as busy in internal/adapters/graph/httpcalendar.go; add synthetic busy tomorrow + fully-booked day in internal/adapters/graph/fakecal.go per contracts/fake-graph.md; MUST NOT call getSchedule or `/users/`
- [X] T021 [US2] Register `calendar free` (`--when`, `--duration`, `--hours`, `--calendar`, `--top`) in internal/adapters/cli/calendar.go
- [X] T022 [US2] Implement acceptance/steps/calendar_free_steps.go until scripts/acceptance.sh passes features/calendar/free.feature
- [X] T023 [US2] After GREEN T019–T022, commit production code separately from T018

**Checkpoint**: US2 independently testable; US1 still green; mail/teams unchanged.

---

## Phase 5: User Story 3 — Confirm the resolved time before writing (Priority: P2)

**Goal**: Dry-run stdout (JSON and human) shows original phrase and resolved local start/end. Real create get matches dry-run times.

**Independent Test**: Dry-run `--when 'tomorrow at 1:30 pm'` JSON includes `when`, `start`, `end`, `timezone`; human output includes the phrase and 1:30 PM; real create then get matches that window.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 3 (MANDATORY)

- [X] T024 [P] [US3] Extend features/calendar/when-create.feature (or features/calendar/when-confirm.feature) for human-mode dry-run showing phrase + local clocks; generate failing acceptance if new scenarios
- [X] T025 [P] [US3] Write RED CLI tests in internal/adapters/cli/calendar_when_test.go for `--human` dry-run containing original phrase and resolved clocks; JSON schema fields per contracts/json-create.schema.json
- [X] T026 [US3] RED commit: failing tests from T024–T025 only

### Implementation for User Story 3

- [X] T027 [US3] Include `when`, `start`, `end`, `timezone` on dry-run and real create JSON in internal/adapters/cli/calendar.go; humanize resolved local date and clocks
- [X] T028 [US3] After GREEN T027, commit production code separately from T026

**Checkpoint**: US3 independently testable on top of US1.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Help, perf, mail/teams regression, quickstart. No live tenant in CI.

- [X] T029 [P] Extend internal/adapters/cli/help.go calendar help with phrase examples, default duration 30 minutes, free default tomorrow, working hours 09:00–17:00, free `--top` 5/max 20, 30-minute grid; tests in internal/adapters/cli/help_calendar_files_test.go or help_when_test.go
- [X] T030 [P] CLI test internal/adapters/cli/perf_free_test.go: `calendar free` against fake finishes under 15s; phrase dry-run under 5s and does not create
- [X] T031 [P] Assert mail list --help and teams list --help still name v1 defaults (existing help_test.go)
- [X] T032 Run make verify on fakes; execute specs/003-calendar-natural-time/quickstart.md automated scenarios

**Checkpoint**: CI `make verify` without a live tenant. No NLP, getSchedule, or mail/teams flag changes.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup**: start immediately
- **Foundational**: depends on Setup; BLOCKS stories
- **US1 and US2**: after Foundational; MAY proceed in parallel (when.go vs free.go)
- **US3**: after US1
- **Polish**: after desired stories

### User Story Dependencies

- **US1 (P1) MVP**: after Foundational
- **US2 (P1)**: after Foundational (uses ResolveDay from T003)
- **US3 (P2)**: after US1

### Within Each User Story

- RED tests fail for the expected reason and RED-commit before implementation
- Domain/parser before CLI wiring
- `make unit` at story checkpoint

### Parallel Opportunities

- T002, T004, T005 after T001
- T006–T008 (US1 tests)
- T015–T017 (US2 tests) parallel with US1 tests after Foundational
- T029–T031 in Polish

---

## Parallel Example: User Story 1

```bash
Task: "Gherkin features/calendar/when-create.feature"
Task: "RED tests internal/app/calendar/create_when_test.go"
Task: "RED CLI tests internal/adapters/cli/calendar_when_test.go"
```

## Parallel Example: US1 vs US2

```bash
Task: "US1 phrase create"
Task: "US2 free slots (free.go + fakecal busy day)"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Setup + Foundational (parser + timezone)
2. US1 `--when` create
3. **STOP**: Independent Test US1 on fakes
4. Demo `calendar create --dry-run --when 'tomorrow at 1:30 pm'`

### Incremental Delivery

1. Foundational → frozen parser
2. US1 → phrase create MVP
3. US2 → `calendar free`
4. US3 → dry-run confirmation fields
5. Polish → `make verify`

---

## Task notes

- Do not edit mail.go or teams.go.
- Closed phrase set only; `next week` stays invalid.
- RED then GREEN commits: T009/T014, T018/T023, T026/T028.
- Fake tomorrow busy 09:00–10:00 and 13:00–14:00 CT per contracts/fake-graph.md.

## Notes

- [P] = different files, no incomplete dependencies
- [US1]–[US3] map to spec User Stories 1–3
- No MCP, getSchedule, NLP libraries, or `--guard`
- Client ID / Tenant ID never in source, tests, logs, or fixtures
