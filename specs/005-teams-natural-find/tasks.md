# Tasks: Natural Teams Find and DM

**Input**: Design documents from `/specs/005-teams-natural-find/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md, constitution.md, existing 001 teams namespace

**Tests**: MANDATORY per Constitution I. RED commit (failing tests + fixtures only) then GREEN. EX-I-001 does not apply. No live Graph. No live member lists in the repo. Do **not** edit `internal/adapters/cli/mail.go`, `calendar.go`, `files.go`, or MCP files. Do **not** rewrite `specs/001-m365-cli/` except additive teams catalog/help notes. Do **not** implement chat create, channels, or MCP recipe updates.

**Organization**: Setup + Foundational block stories. US1–US3 map to spec User Stories 1–3. MVP = Setup + Foundational + US1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallel (different files, no incomplete deps)
- **[Story]**: `[US1]`–`[US3]` on user-story phases only
- Exact file paths required

## Path Conventions

`internal/app/teams/`, `internal/adapters/cli/teams.go`, `internal/adapters/graph/httpteams.go`, `internal/adapters/graph/memory.go`, `features/teams/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Existing module. No new go.mod.

- [X] T001 Confirm `internal/app/teams/` and `features/teams/` exist from 001; do not recreate the module or Makefile

**Checkpoint**: `make unit` still passes.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: FindQuery/ResolveResult, find `--top` 10/20, fake chats off list page one, ListChats paging + members. Blocks US1–US3.

**⚠️ CRITICAL**: No user story work until this phase is complete. TDD: RED before GREEN.

- [X] T002 [P] Write RED tests in internal/app/teams/find_query_test.go then parse FindQuery in internal/app/teams/find.go quoting data-model: "Person: one email, one given name, or `First Last`." "Group: `--group` set, or query matches `(?i)^(?:the )?group with (.+)$`." "Empty query → usage (3)."
- [X] T003 [P] Write RED tests in internal/domain/result_test.go then add DefaultFindTop=10 and MaxFindTop=20 in internal/domain/result.go quoting "Default `--top` = 10. Maximum = 20. Values `<1` or `>20` are usage, not silent caps." Find MUST NOT use list MaxTop 50.
- [X] T004 [P] Seed synthetic chats per contracts/fake-graph.md in internal/adapters/graph/memory.go: `chat-ajay` (Ajay Kumar 1:1, **not** in first 20 list items), `chat-ajay-b` (Ajay Singh), `chat-group-ajay` (group including Ajay), `chat-noc-dev` (topic NOC-Dev), fillers so list default page omits `chat-ajay`; keep `chat-1`/`chat-2` for 001
- [X] T005 Write RED tests in internal/adapters/graph/teams_page_test.go then page `GET /me/chats?$expand=members&$top=50` with nextLink in internal/adapters/graph/httpteams.go; map members to `domain.Person`; never `/users` or `/me/people`; scan ceiling 10 pages (500 chats) quoted from data-model ScanCeiling

**Checkpoint**: Query parser and fake seed in place; `teams list` first page still excludes `chat-ajay`; mail/calendar/files tests green.

---

## Phase 3: User Story 1 — Find the chat with a person (Priority: P1) 🎯 MVP

**Goal**: `teams find Ajay` returns the unique 1:1 even if it is not on the first list page. Groups that merely include Ajay are omitted. Two Ajays list both. Zero matches exit 0 empty. No send.

**Independent Test**: Fake: Ajay 1:1 off page one + group with Ajay. `teams find Ajay` → that 1:1 only. Two Ajays → both, no send. No Ajay → exit 0 empty. Missing consent → 4.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 1 (MANDATORY)

- [X] T006 [P] [US1] Write Gherkin in features/teams/find.feature for unique Ajay 1:1 off list page one, two Ajays, empty, missing consent exit 4, group with Ajay not in default person result; generate failing acceptance tests via scripts/acceptance.sh
- [X] T007 [P] [US1] Write RED tests in internal/app/teams/find_test.go quoting "Default person find: 1:1 only; groups that merely include the person MUST NOT appear." "None: find → count 0, incomplete false, exit 0." Matching: exact then prefix then substring; case-insensitive
- [X] T008 [P] [US1] Write RED CLI tests in internal/adapters/cli/teams_find_test.go for `teams find Ajay` JSON query/intent/limit/count/incomplete/items, `--top 21` → 3, empty query → 3
- [X] T009 [US1] RED commit: git add only failing tests and fixtures from T006–T008; no production code

### Implementation for User Story 1

- [X] T010 [US1] Implement Find in internal/app/teams/find.go: page ListChats until unique 1:1, result `--top`, or ceiling; stop early on unique; `incomplete` true if ceiling hit without unique decision
- [X] T011 [US1] Register `teams find` (`QUERY`, `--group`, `--top`) in internal/adapters/cli/teams.go; split teams.go if it would exceed 250 lines
- [X] T012 [US1] Implement acceptance/steps/teams_find_steps.go until scripts/acceptance.sh passes features/teams/find.feature
- [X] T013 [US1] After GREEN T010–T012, commit production code separately from T009; do not edit locked RED tests

**Checkpoint**: US1 independently testable on fakes; `teams list`/`send` by id unchanged.

---

## Phase 4: User Story 2 — Find a group chat by people or topic (Priority: P1)

**Goal**: `--group` or `group with X` returns groups by topic/members, not 1:1s. Several groups bounded by `--top`. Empty group search exit 0.

**Independent Test**: `--group NOC` returns NOC-Dev. `--group Ajay` returns groups including Ajay, not the 1:1. No matching group → 0 empty.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 2 (MANDATORY)

- [X] T014 [P] [US2] Extend features/teams/find.feature (or features/teams/find-group.feature) for `--group NOC`, `--group Ajay` (groups only), `find 'group with Ajay'`, empty group search; generate failing acceptance if new scenarios
- [X] T015 [P] [US2] Write RED tests in internal/app/teams/find_test.go quoting "Group find: groups only." intent `group`
- [X] T016 [US2] RED commit: failing tests from T014–T015 only

### Implementation for User Story 2

- [X] T017 [US2] Apply group intent in internal/app/teams/find.go: topic substring and/or members; never return 1:1s
- [X] T018 [US2] Wire `--group` in internal/adapters/cli/teams.go; help examples `teams find Ajay` and `teams find --group NOC`
- [X] T019 [US2] After GREEN T017–T018, commit production code separately from T016

**Checkpoint**: US2 independently testable; US1 still green.

---

## Phase 5: User Story 3 — Send a DM by name after a unique resolve (Priority: P2)

**Goal**: `teams send --to Ajay` after unique 1:1 resolve. Dry-run shows person, chat id, text, no send. Several matches exit 3. Zero matches exit 6, no create. Chat-id send unchanged.

**Independent Test**: Unique Ajay: dry-run `--to` no message; real send delivers. Two Ajays: `--to` exits 3. Unknown: 6, no new chat. `teams send chat-1` still works.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 3 (MANDATORY)

- [X] T020 [P] [US3] Write Gherkin in features/teams/send-to.feature for unique dry-run, unique send, several → 3, none → 6, `--to`+id → 3, send-by-id unchanged; generate failing acceptance tests via scripts/acceptance.sh
- [X] T021 [P] [US3] Write RED tests in internal/app/teams/send_to_test.go quoting data-model: "`--to` XOR chat id (both → usage 3)." "Several → usage 3, candidates, no send." "Zero (complete) → not-found 6, no create."
- [X] T022 [P] [US3] Write RED CLI tests in internal/adapters/cli/teams_send_to_test.go for `--to Ajay --dry-run` JSON `dry_run`/`to`/`chat_id`/`text` per contracts/json-send-to.schema.json
- [X] T023 [US3] RED commit: failing tests from T020–T022 only

### Implementation for User Story 3

- [X] T024 [US3] Extend SendInput with To in internal/app/teams/ports.go; resolve via Find (person intent) before send; dry-run includes `to` and `chat_id`
- [X] T025 [US3] Register `--to` on send in internal/adapters/cli/teams.go; `--to` and positional chat id together → 3
- [X] T026 [US3] Implement acceptance/steps/teams_send_to_steps.go until scripts/acceptance.sh passes features/teams/send-to.feature
- [X] T027 [US3] After GREEN T024–T026, commit production code separately from T023

**Checkpoint**: US3 independently testable; send-by-id 001 tests still green.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Help, alias, perf, isolation, quickstart. No live tenant. No MCP edits.

- [X] T028 [P] Extend teams help in internal/adapters/cli/help.go: find examples Ajay / `--group NOC`, send `--to`, find `--top` 10/20, scan incomplete; tests in internal/adapters/cli/help_test.go or teams_find_test.go
- [X] T029 [P] CLI tests that `chat find Ajay` aliases teams in internal/adapters/cli/chat_alias_test.go or teams_find_test.go
- [X] T030 [P] Perf test internal/adapters/cli/perf_find_test.go: unique `teams find Ajay` against fake finishes under 15s (SC-008)
- [X] T031 [P] Assert mail list --help, calendar --help, files --help unchanged in internal/adapters/cli/help_test.go (FR-001 / SC-007)
- [X] T032 Isolation: internal/app/teams MUST NOT import mail/calendar/files; no `/users` in httpteams.go
- [X] T033 Run make unit; execute specs/005-teams-natural-find/quickstart.md automated scenarios; no secrets (SC-009)

**Checkpoint**: CI fakes only. Find never sends. No new 1:1 created.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup**: start immediately
- **Foundational**: depends on Setup; BLOCKS stories
- **US1**: after Foundational — MVP
- **US2**: after Foundational (shares Find); MAY parallel with US1 after T010 exists
- **US3**: after US1 (send `--to` reuses person find)
- **Polish**: after desired stories

### User Story Dependencies

- **US1 (P1) MVP**: after Foundational
- **US2 (P1)**: after Foundational
- **US3 (P2)**: after US1

### Parallel Opportunities

- T002, T003, T004 after T001
- T006–T008 (US1 tests)
- T014–T015 (US2 tests) after Foundational
- T020–T022 (US3 tests) after US1 GREEN
- T028–T031 in Polish

---

## Parallel Example: User Story 1

```bash
Task: "Gherkin features/teams/find.feature"
Task: "RED tests internal/app/teams/find_test.go"
Task: "RED CLI tests internal/adapters/cli/teams_find_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Setup + Foundational (query parse, fake off-page Ajay, paged chats)
2. US1 `teams find Ajay`
3. **STOP**: Independent Test US1 on fakes

### Incremental Delivery

1. Foundational → seed + ceiling
2. US1 → person find MVP
3. US2 → `--group`
4. US3 → `send --to`
5. Polish → help, alias, 15s find

---

## Notes

- [P] tasks = different files, no incomplete dependencies
- Quote data-model constraints in RED tests
- Commit RED then GREEN separately
- Avoid: `/users`, chat create, MCP recipe edits, silent first-page-only search
