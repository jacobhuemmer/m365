# Tasks: Calendar and Files Namespaces

**Input**: Design documents from `/specs/002-calendar-files/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md, constitution.md, existing v1 CLI (`specs/001-m365-cli/`)

**Tests**: Tests are MANDATORY per Constitution Principle I. Every behavior slice includes Gherkin under `features/` when user-observable AND a focused RED unit/use-case/CLI test. Confirm each test fails for the expected reason (not compile/import). Commit Gherkin + RED tests as a locked RED Checkpoint BEFORE matching implementation. Do not edit a locked RED test without human approval. EX-I-001 does **not** apply: this feature MUST use separate RED then GREEN commits. Unit tests use fakes/ports: no live Microsoft Graph, no real Keychain, no live calendar bodies or OneDrive bytes. Synthetic files only under `testdata/`. Acceptance: Gherkin → APS IR → generated tests; generated files are not hand-edited.

**Organization**: Setup and Foundational block all stories. User stories US1–US6 map to spec User Stories 1–6. MVP = Setup + Foundational + US1. Do **not** edit `internal/adapters/cli/mail.go` or `internal/adapters/cli/teams.go`. Do **not** rewrite `specs/001-m365-cli/` except the permitted additive `namespaces.calendar` / `namespaces.files` keys on the v1 JSON stdout auth schema.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: `[US1]`–`[US6]` on user-story phases only
- Include exact file paths in descriptions

## Path Conventions

Existing module at repository root: `cmd/m365/`, `internal/{domain,app/{auth,mail,teams,calendar,files},adapters/{cli,graph,fs}}`, `features/{calendar,files}/`, `acceptance/`, `testdata/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Package directories for calendar and files. Module, Makefile, and APS already exist from 001.

- [X] T001 Create directory skeleton `internal/app/calendar/`, `internal/app/files/`, `internal/adapters/graph/` (new files only), `features/calendar/`, `features/files/`, `acceptance/steps/` placeholders with `doc.go` only where a package would not compile; do not recreate go.mod or Makefile

**Checkpoint**: Existing `make unit` still passes; new dirs exist; mail/teams sources unmodified.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Additive session consent, fake tokens, ResultLimit defaults, CLI registration seam, OAuth scopes, isolation tests. Blocks all user stories.

**⚠️ CRITICAL**: No user story work until this phase is complete. TDD: RED tests committed before implementation.

- [X] T002 [P] Write RED tests in internal/domain/session_test.go then add `CalendarConsented` and `FilesConsented` on Session in internal/domain/session.go; SignedOut MUST set all four consent flags false
- [X] T003 [P] Write RED tests in internal/app/auth/login_test.go then add Calendar/Files bools on auth.Blob in internal/app/auth/ports.go; Status MUST map them; FakeLogin in internal/adapters/graph/oauth.go MUST accept calendar and files flags without changing mail/teams meaning of existing two-arg helpers (add FakeLoginAll or extra params without breaking 001 call sites)
- [X] T004 [P] Write RED tests in internal/domain/result_test.go then extend ResultLimit defaults in internal/domain/result.go quoting data-model: "Calendar calendars list default top = 20." "Calendar event list default top = 10." "Files list default top = 20." "Maximum top = 50. Values `<1` or `>50` are usage failures, not silent caps."
- [X] T005 [P] Write RED tests in internal/adapters/graph/fake_test.go then add tokens `fake-calendar`, `fake-files`, `fake-all` per contracts/fake-graph.md in internal/adapters/graph/fake.go (split fakecalendar.go / fakefiles.go if fake.go would exceed 250 lines); `fake-both` MUST remain mail+teams only
- [X] T006 [P] Write RED tests in internal/adapters/cli/auth_test.go then add `namespaces.calendar` and `namespaces.files` on auth status JSON in internal/adapters/cli/auth.go; `mail` and `teams` keys MUST remain; no token fields
- [X] T007 [P] Add `Calendars.ReadWrite` and `Files.ReadWrite` to PKCEConfig scopes in internal/adapters/graph/oauth.go; MUST NOT add `Calendars.ReadWrite.Shared` or `Files.ReadWrite.All`; tests in internal/adapters/graph/oauth_test.go
- [X] T008 Write RED tests in internal/app/calendar/isolation_test.go that fail if this package imports `internal/app/mail`, `internal/app/teams`, or `internal/app/files`, and in internal/app/files/isolation_test.go that fail if this package imports `internal/app/mail`, `internal/app/teams`, or `internal/app/calendar`
- [X] T009 Register `calendar` and `files` namespaces in internal/adapters/cli/run.go (unknown verb → usage until story wiring); extend rootHelp in internal/adapters/cli/help.go to list calendar and files; `chat` MUST still route only to teams
- [X] T010 Wire calendar/files stores on Deps in internal/adapters/cli/run.go and cmd/m365/main.go as nil-safe until stories implement them — wiring only, no behavior in main
- [X] T011 [P] Additive-only edit of namespaces properties in specs/001-m365-cli/contracts/json-stdout.schema.json to allow `calendar` and `files` booleans; do not change mail/teams command schemas

**Checkpoint**: `make unit` green; `m365 --help` lists calendar and files; `auth status` JSON has four namespace keys; `fake-both` still denies calendar/files routes with exit 4. User stories may start.

---

## Phase 3: User Story 1 — List calendars and read events (Priority: P1) 🎯 MVP

**Goal**: List own calendars; list events on a named/default calendar inside a visible window with bounded `--top`; get one event as text.

**Independent Test**: With `fake-calendar` and synthetic fixtures, list calendars, list events with default window and default limit, list a named calendar, get one event. Assert empty window exit 0, unknown id exit 6, applied limit, occurrences only inside the window, mail list/teams list unchanged.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change.

### Tests for User Story 1 (MANDATORY)

- [X] T012 [P] [US1] Write Gherkin in features/calendar/list-get.feature for calendars list, default window event list, named calendar, get fields, empty window exit 0, unknown id exit 6; generate failing acceptance/generated/list-get_acceptance_test.go via scripts/acceptance.sh
- [X] T013 [P] [US1] Write RED tests in internal/app/calendar/list_test.go quoting "Calendar event list default top = 10." "Calendar calendars list default top = 20." "Default: now through now+7 days (clock injected)." and "Maximum top = 50. Values `<1` or `>50` are usage failures, not silent caps."
- [X] T014 [P] [US1] Write RED CLI tests in internal/adapters/cli/calendar_read_test.go for `calendar calendars|list|get` JSON `limit`/`count`/`next_page` and no secrets
- [X] T015 [US1] RED commit: git add only failing tests and fixtures from T012–T014; commit with no production code; record commit id in task notes

### Implementation for User Story 1

- [X] T016 [P] [US1] Implement Calendar, CalendarEvent, EventWindow in internal/domain/calendar.go with tests in internal/domain/calendar_test.go quoting "Start ≥ end ⇒ usage." and window default now through now+7 days
- [X] T017 [US1] Implement calendar ports and use cases in internal/app/calendar/ports.go, internal/app/calendar/list.go, internal/app/calendar/get.go with injected Now clock; Graph URLs stay out of this package
- [X] T018 [US1] Map calendars + calendarView HTTP in internal/adapters/graph/httpcalendar.go with tests in internal/adapters/graph/calendar_test.go that MUST call NewFakeServer (`httptest`) per contracts/fake-graph.md (`cal-1`, `ev-1`, `ev-occ-1` only in window). Memory-only tests do not satisfy this task
- [X] T019 [US1] Register `calendar calendars|list|get` flags (`--calendar`, `--start`, `--end`, `--top`, `--page-token`) in internal/adapters/cli/calendar.go (depends on T017, T018)
- [X] T020 [US1] Implement acceptance/steps/calendar_read_steps.go until scripts/acceptance.sh passes features/calendar/list-get.feature (depends on T019)
- [X] T021 [US1] After GREEN T016–T020, commit production code separately from the T015 RED commit; do not edit locked RED tests

**Checkpoint**: US1 independently testable on fakes; `make unit` green; mail.go/teams.go untouched.

---

## Phase 4: User Story 2 — Browse OneDrive folders and item metadata (Priority: P1)

**Goal**: Show drive root; list folder children with `--top` and `next_page`; get item metadata without file bytes.

**Independent Test**: With `fake-files`, show root, list default folder, follow page token, get file and folder. Assert empty folder exit 0, unknown id exit 6, no bytes on stdout/stderr, mail/teams attachment commands unchanged.

**RED Checkpoint**: commit failing tests and fixtures only, then GREEN.

### Tests for User Story 2 (MANDATORY)

- [X] T022 [P] [US2] Write Gherkin in features/files/list-get.feature for root, list children, next_page, get metadata, empty folder exit 0, unknown id exit 6, no bytes; generate failing acceptance/generated/list-get_acceptance_test.go via scripts/acceptance.sh (feature file name may be `files-list-get.feature` if calendar already used list-get)
- [X] T023 [P] [US2] Write RED tests in internal/app/files/list_test.go quoting "Files list default top = 20." "Maximum top = 50. Values `<1` or `>50` are usage failures, not silent caps."
- [X] T024 [P] [US2] Write RED CLI tests in internal/adapters/cli/files_read_test.go for `files root|list|get` JSON limit/count/next_page and no file bytes
- [X] T025 [US2] RED commit: git add only failing tests and fixtures from T022–T024; no production code

### Implementation for User Story 2

- [X] T026 [P] [US2] Implement DriveRoot and DriveItem in internal/domain/drive.go with tests in internal/domain/drive_test.go (metadata only; no bytes)
- [X] T027 [US2] Implement files ports and use cases in internal/app/files/ports.go, internal/app/files/list.go, internal/app/files/get.go — MUST NOT import calendar/mail/teams; Graph URLs stay out
- [X] T028 [US2] Map `/me/drive` root/children/item HTTP in internal/adapters/graph/httpfiles.go with tests in internal/adapters/graph/files_test.go via NewFakeServer per contracts/fake-graph.md (`root`, `folder-1`, `file-1`). Opaque `next_page` (not a Graph URL)
- [X] T029 [US2] Register `files root|list|get` (`--folder`, `--top`, `--page-token`) in internal/adapters/cli/files.go
- [X] T030 [US2] Implement acceptance/steps/files_read_steps.go until scripts/acceptance.sh passes the US2 feature (depends on T029)
- [X] T031 [US2] After GREEN T026–T030, commit production code separately from T025; do not edit locked RED tests

**Checkpoint**: US2 independently testable; US1 still green; mail/teams attachment CLI unchanged.

---

## Phase 5: User Story 3 — Create, update, and delete events with dry-run (Priority: P2)

**Goal**: Create/update/delete events. `--dry-run` shows intent and MUST NOT mutate.

**Independent Test**: Dry-run create (no event), real create (id), dry-run update/delete then real, missing subject/start/end exit 3, unknown id exit 6, mail send/reply unchanged.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 3 (MANDATORY)

- [X] T032 [P] [US3] Write Gherkin in features/calendar/write.feature for dry-run create (no event), real create id, update/delete dry-run then real, missing fields exit 3, unknown id exit 6; generate failing acceptance tests via scripts/acceptance.sh
- [X] T033 [P] [US3] Write RED tests in internal/app/calendar/write_test.go asserting dry-run does not call create/update/delete ports and quoting "Start ≥ end ⇒ usage."
- [X] T034 [P] [US3] Write RED CLI tests in internal/adapters/cli/calendar_write_test.go for `--dry-run` JSON (`dry_run: true`) vs created `id`
- [X] T035 [US3] RED commit: failing tests and fixtures from T032–T034 only

### Implementation for User Story 3

- [X] T036 [US3] Implement create/update/delete use cases in internal/app/calendar/write.go (dry-run short-circuit; required subject/start/end on create)
- [X] T037 [US3] Map POST/PATCH/DELETE HTTP in internal/adapters/graph/httpcalendar_write.go with tests in internal/adapters/graph/calendar_write_test.go via NewFakeServer — dry-run MUST NOT reach HTTP
- [X] T038 [US3] Register `calendar create|update|delete` (`--subject`, `--start`, `--end`, `--location`, `--body`/`--body-file`, `--attendee`, `--calendar`, `--dry-run`) in internal/adapters/cli/calendar.go
- [X] T039 [US3] Implement acceptance/steps/calendar_write_steps.go until scripts/acceptance.sh passes features/calendar/write.feature
- [X] T040 [US3] After GREEN T036–T039, commit production code separately from T035

**Checkpoint**: US3 independently testable; US1 list/get still green.

---

## Phase 6: User Story 4 — Download, upload, and organize OneDrive items (Priority: P2)

**Goal**: Download to user path (overwrite rules); upload; create-folder; delete; move/rename. Writes support `--dry-run`.

**Independent Test**: Download new path; refuse existing without overwrite; overwrite when asked; dry-run upload (no remote); real upload; create-folder; dry-run delete then real; move/rename; missing `--out` exit 3 no write; no bytes on stdout; mail/teams save-attachment unchanged.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 4 (MANDATORY)

- [X] T041 [P] [US4] Write Gherkin in features/files/write.feature for download path, refuse-without-overwrite, overwrite, dry-run upload, real upload, create-folder, delete dry-run/real, move, missing out exit 3; generate failing acceptance tests via scripts/acceptance.sh
- [X] T042 [P] [US4] Write RED tests in internal/app/files/write_test.go quoting "Size MUST be `1..104857600` (100 MiB). Zero-byte files are invalid." and download: destination exists without overwrite → exit 3 file unchanged
- [X] T043 [P] [US4] Write RED CLI tests in internal/adapters/cli/files_write_test.go using testdata/note.txt for upload dry-run metadata without bytes on stdout
- [X] T044 [US4] RED commit: failing tests and fixtures from T041–T043 only

### Implementation for User Story 4

- [X] T045 [US4] Implement download/upload/mkdir/delete/move use cases in internal/app/files/write.go reusing internal/adapters/fs/fs.go for local paths; dry-run MUST NOT upload or delete
- [X] T046 [US4] Map download content, simple PUT (≤4 MiB), createUploadSession (4 MiB–100 MiB), mkdir, delete, move HTTP in internal/adapters/graph/httpfiles_write.go with tests in internal/adapters/graph/files_write_test.go via NewFakeServer — dry-run MUST NOT reach HTTP; never `/sites` or other-drive URLs
- [X] T047 [US4] Register `files download|upload|create-folder|delete|move` (`--out`, `--overwrite`, `--file`, `--folder`, `--name`, `--dry-run`) in internal/adapters/cli/files.go
- [X] T048 [US4] Implement acceptance/steps/files_write_steps.go until scripts/acceptance.sh passes features/files/write.feature
- [X] T049 [US4] After GREEN T045–T048, commit production code separately from T044

**Checkpoint**: US4 independently testable; US2 list/get still green.

---

## Phase 7: User Story 5 — Independent calendar and files consent (Priority: P3)

**Goal**: Four independent consent flags. Missing calendar does not break files/mail/teams and vice versa. Own calendars and own drive only.

**Independent Test**: `fake-both` → calendar/files exit 4, mail/teams succeed. `fake-calendar` → calendar list succeeds, files exit 4. `fake-files` reverse. `auth status` shows four flags, no secrets. Fake MUST NOT succeed on `/users/{other}/calendars` or `/sites`.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 5 (MANDATORY)

- [X] T050 [P] [US5] Write Gherkin in features/auth/consent-calendar-files.feature for fake-both vs fake-calendar vs fake-files vs fake-all; generate failing acceptance tests via scripts/acceptance.sh
- [X] T051 [P] [US5] Write RED tests in internal/app/auth/consent_test.go that missing calendar or files consent is exit 4 not 5
- [X] T052 [P] [US5] Write RED tests in internal/adapters/graph/consent_scope_test.go that fake calendar routes reject Shared/other-user paths and files routes reject `/sites`
- [X] T053 [US5] RED commit: failing tests from T050–T052 only

### Implementation for User Story 5

- [X] T054 [US5] Enforce Require(sess, calendar/files) in calendar and files use cases in internal/app/calendar/ports.go and internal/app/files/ports.go
- [X] T055 [US5] Fake Graph 403 on calendar/files without matching token; 404/403 on `/users/` and `/sites` success paths in internal/adapters/graph/fake.go (or fakecalendar.go/fakefiles.go)
- [X] T056 [US5] Implement acceptance/steps/consent_calendar_files_steps.go until scripts/acceptance.sh passes features/auth/consent-calendar-files.feature
- [X] T057 [US5] After GREEN T054–T056, commit production code separately from T053

**Checkpoint**: US5 independently testable; mail/teams consent isolation from 001 still green.

---

## Phase 8: User Story 6 — Discover calendar and files without changing mail or teams (Priority: P3)

**Goal**: Help names new namespaces, verbs, limits, window, dry-run, overwrite. Mail/teams help, flags, exits unchanged. `chat` still aliases teams only.

**Independent Test**: `--help` with no session exit 0 for root, calendar, files. `mail list --help` and `teams list --help` match v1 flags/limits. `chat list` aliases teams list.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 6 (MANDATORY)

- [X] T058 [P] [US6] Write Gherkin in features/cli/calendar-files-help.feature for root/calendar/files help and mail/teams help regression; generate failing acceptance tests via scripts/acceptance.sh
- [X] T059 [P] [US6] Write RED CLI tests in internal/adapters/cli/help_calendar_files_test.go asserting calendar/files help names `--top` defaults 10/20, max 50, 7-day window, 100 MiB upload, `--dry-run`, and that mail list / teams list help text still includes v1 defaults 10/20
- [X] T060 [P] [US6] Extend internal/adapters/cli/chat_alias_test.go so `chat` still only reaches teams verbs (not calendar/files)
- [X] T061 [US6] RED commit: failing tests from T058–T060 only

### Implementation for User Story 6

- [X] T062 [US6] Extend internal/adapters/cli/help.go with calendar and files help strings (window, limits, dry-run, upload cap); do not change mailListHelp/teamsListHelp semantics
- [X] T063 [US6] Implement acceptance/steps/calendar_files_help_steps.go until scripts/acceptance.sh passes features/cli/calendar-files-help.feature
- [X] T064 [US6] After GREEN T062–T063, commit production code separately from T061

**Checkpoint**: US6 independently testable signed-out; `make unit` green.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Perf gate, secret redaction, quickstart on fakes, verify. No live tenant in CI.

- [X] T065 [P] CLI test internal/adapters/cli/perf_calendar_files_test.go: calendar list and files list against fake Graph finish in under 15s (SC-013)
- [X] T066 [P] Extend internal/adapters/cli/redact_test.go so calendar/files verbose output never prints tokens
- [X] T067 Run make verify green on fakes (fmt vet unit race coverage gosec govulncheck acceptance crap) with no live tenant
- [X] T068 Execute specs/002-calendar-files/quickstart.md automated scenarios against fakes (auth keys, calendar list window, dry-run create, files list, download refuse-without-overwrite, mail/teams help unchanged)

**Checkpoint**: CI can run `make verify` without a live tenant. No MCP, SharePoint, Shared calendar, or `graph GET` tasks exist.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1–US6 (Phases 3–8)**: Depend on Foundational
  - Sequential default: US1 → US2 → US3 → US4 → US5 → US6
  - After Foundational, US1 (calendar read) and US2 (files read) MAY proceed in parallel (different packages; isolation tests T008)
  - US3 depends on US1; US4 depends on US2; US5 depends on US1+US2; US6 can start after CLI seam (Foundational) but should follow verbs existing
- **Polish (Phase 9)**: Depends on desired stories being complete

### User Story Dependencies

- **US1 (P1) MVP**: After Foundational only
- **US2 (P1)**: After Foundational; MUST NOT import calendar packages
- **US3 (P2)**: After US1
- **US4 (P2)**: After US2
- **US5 (P3)**: After US1 and US2
- **US6 (P3)**: After verbs exist (practically after US3/US4 so write help is true)

### Within Each User Story

- Gherkin + RED unit/CLI tests MUST fail for the expected reason and be RED-committed before implementation
- Domain before use cases before Graph adapter before CLI wiring before acceptance steps
- `make unit` at story checkpoint

### Parallel Opportunities

- T002–T007, T011 after T001
- T012–T014 (US1 tests)
- T022–T024 (US2 tests) parallel with US1 tests after Foundational
- T032–T034, T041–T043, T050–T052, T058–T060
- T065–T066 in Polish

---

## Parallel Example: User Story 1

```bash
# After Foundational, launch US1 RED work in parallel:
Task: "Gherkin in features/calendar/list-get.feature"
Task: "RED tests in internal/app/calendar/list_test.go"
Task: "RED CLI tests in internal/adapters/cli/calendar_read_test.go"
```

## Parallel Example: User Story 1 vs 2

```bash
# After Foundational, calendar read and files read on different packages:
Task: "US1 Gherkin features/calendar/list-get.feature"
Task: "US2 Gherkin features/files/list-get.feature"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Independent Test for US1 against fakes; `make unit`
5. Demo `calendar calendars|list|get`

### Incremental Delivery

1. Setup + Foundational → seam + consent flags
2. US1 → calendar read MVP
3. US2 → files read (parallel-capable)
4. US3–US4 → writes with dry-run
5. US5–US6 → consent isolation and help
6. Polish → `make verify` on fakes

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Then: Developer A US1/US3; Developer B US2/US4; US5/US6 integrate at auth/help without mail↔teams↔calendar↔files imports

---

## Task notes

- Do not edit internal/adapters/cli/mail.go or teams.go.
- `fake-both` remains mail+teams only (contracts/fake-graph.md).
- RED then GREEN git commits are required for US1–US6 (T015/T021, T025/T031, T035/T040, T044/T049, T053/T057, T061/T064).
- Live event bodies and OneDrive bytes MUST NOT enter testdata/ or git.

## Notes

- [P] = different files, no incomplete dependencies
- [US1]–[US6] map to spec User Stories 1–6
- No MCP, SharePoint, Shared calendars, `graph GET`, `--guard`, Jenkins, Docker, or Helm tasks
- Client ID / Tenant ID values never in source, tests, logs, or fixtures
- Commit RED separately from GREEN; refactor only on GREEN
