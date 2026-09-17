# Tasks: m365 MCP Serve

**Input**: Design documents from `/specs/004-m365-mcp/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md, constitution.md, existing CLI namespaces 001–003

**Tests**: MANDATORY per Constitution I. Spec Verification Strategy requires `tools/list`, status signed-in/out, JSON parity, write dry-run, auth-class vs service-class, and no secrets on the wire. RED commit (failing tests + fixtures only) then GREEN. EX-I-001 does not apply. No live Graph in unit tests. Synthetic fixtures only. Do **not** edit `internal/adapters/cli/mail.go`, `teams.go`, `calendar.go`, or `files.go`. Do **not** rewrite `specs/001-m365-cli/`, `002-calendar-files/`, or `003-calendar-natural-time/`.

**Organization**: Setup + Foundational block stories. US1–US4 map to spec User Stories 1–4. MVP = Setup + Foundational + US1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallel (different files, no incomplete deps)
- **[Story]**: `[US1]`–`[US4]` on user-story phases only
- Exact file paths required

## Path Conventions

`internal/adapters/cli/mcp.go`, `internal/adapters/cli/mcp_dispatch.go`, `internal/adapters/cli/run.go`, `internal/adapters/cli/help.go`, `features/mcp/`, `acceptance/steps/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Feature lives in the existing module. No new binary, Makefile, or token store.

- [X] T001 Create `features/mcp/` and leave `acceptance/steps/` ready for MCP steps; do not recreate `go.mod` or Makefile; do not add `internal/app/mcp` or Graph ports

**Checkpoint**: `make unit` still passes; mail/teams/calendar/files sources unmodified.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Official MCP SDK, `m365 mcp serve` seam, in-process `Run` with captured buffers, three-tool server skeleton, flag-map argv helper. Blocks US1–US4.

**⚠️ CRITICAL**: No user story work until this phase is complete. TDD: RED before GREEN.

- [X] T002 Pin `github.com/modelcontextprotocol/go-sdk` v1 (observed v1.8.0) in go.mod / go.sum; import only `mcp` in adapters, never in `internal/domain` or `internal/app/*`
- [X] T003 Write RED tests in internal/adapters/cli/mcp_cli_test.go then register namespace `mcp` verb `serve` in internal/adapters/cli/run.go; unknown mcp verb → usage (3); `m365 mcp serve --human` → usage (3); `cmd/m365/main.go` stays wiring-only (same Deps)
- [X] T004 [P] Extend rootHelp in internal/adapters/cli/help.go to list `mcp` without changing mail/teams/calendar/files lines; tests in internal/adapters/cli/help_test.go or mcp_cli_test.go
- [X] T005 Write RED tests in internal/adapters/cli/mcp_test.go using `mcp.NewInMemoryTransports` that `tools/list` returns exactly `m365_status`, `m365_help`, `m365_run` quoting data-model: "`tools/list` MUST return these names only (SC-001)."; implement NewMCPServer in internal/adapters/cli/mcp.go registering those three tools (handlers may still fail until story GREEN)
- [X] T006 Implement serve loop in internal/adapters/cli/mcp.go: `Run(["m365","mcp","serve"])` calls `server.Run(ctx, &mcp.StdioTransport{})`; each tool call MUST clone Deps with `bytes.Buffer` stdout/stderr so MCP JSON-RPC owns process stdio
- [X] T007 [P] Implement FlagMapToArgs in internal/adapters/cli/mcp_dispatch.go with tests in internal/adapters/cli/mcp_dispatch_test.go quoting data-model: "Flag keys are CLI long names without `--`." "Flag values: string, number, boolean, or array of strings." "Booleans: `true` emits the flag; `false` omits it." Repeatable arrays become repeated `--to`/`--attach`/`--attendee`
- [X] T008 Write mcpHelp text in internal/adapters/cli/help.go for `m365 mcp --help` / `m365 mcp serve --help`: stdio JSON-RPC, three tool names, write opt-in default false, no session required, exit 0 (same file as T004; do after T004)

**Checkpoint**: In-memory client lists exactly three tools; `m365 --help` mentions mcp; `fake-both` mail/teams tests still green. User stories may start.

---

## Phase 3: User Story 1 — Agent checks session and lists compact tools (Priority: P1) 🎯 MVP

**Goal**: `tools/list` is the compact catalog. `m365_status` reports signed-in, session usable, and mail/teams/calendar/files consent with no tokens. Signed-out status does not open a browser.

**Independent Test**: Drive in-memory JSON-RPC against NewMCPServer. `tools/list` returns only the three documented names. Usable fake session → consent flags, no secrets, under 2s. No session → not signed in, no login/browser.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 1 (MANDATORY)

- [X] T009 [P] [US1] Write Gherkin in features/mcp/catalog-status.feature for tools/list compact catalog, status signed-in with four consent flags, status signed-out with no browser; generate failing acceptance tests via scripts/acceptance.sh
- [X] T010 [P] [US1] Write RED tests in internal/adapters/cli/mcp_status_test.go quoting data-model: "MUST NOT contain access tokens, refresh tokens, authorization codes, or client secrets." "Signed out is success: `signed_in` false, no browser, no hang." Assert JSON matches `m365 auth status` keys `signed_in`, `session_usable`, `namespaces.mail|teams|calendar|files`
- [X] T011 [US1] RED commit: git add only failing tests and fixtures from T009–T010; no production code

### Implementation for User Story 1

- [X] T012 [US1] Implement `m365_status` in internal/adapters/cli/mcp.go by running in-process `auth status` (JSON) into CallToolResult text; `isError` false on signed-out; MUST NOT call Login or open a browser
- [X] T013 [US1] Redact MCP content with internal/adapters/cli/redact.go before the wire (FR-011); assert no `token`/`bearer` substrings in mcp_status_test.go
- [X] T014 [US1] Implement acceptance/steps/mcp_status_steps.go (in-memory MCP + fake store/Graph, same pattern as acceptance/steps/registry.go) until scripts/acceptance.sh passes features/mcp/catalog-status.feature
- [X] T015 [US1] After GREEN T012–T014, commit production code separately from T011; do not edit locked RED tests

**Checkpoint**: US1 independently testable on fakes; human CLI `m365 auth status` unchanged.

---

## Phase 4: User Story 2 — Agent reads mail, teams, calendar, or files through run (Priority: P1)

**Goal**: One `m365_run` tool with namespace, verb, and flag map returns the same JSON the CLI would print. Limits apply. Unknown verb is usage. Missing consent is auth-class.

**Independent Test**: Fake Graph; run `mail list`, `teams list`, `calendar list`/`free`, `files list` at default limits; JSON item count and ids match `cli.Run` for the same flags. `mail nope` → usage, no Graph. Teams list without teams consent → `class=auth` not `service`.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 2 (MANDATORY)

- [X] T016 [P] [US2] Write Gherkin in features/mcp/run-read.feature for mail list JSON parity, calendar list/free, unknown verb usage, missing teams consent auth-class; generate failing acceptance tests via scripts/acceptance.sh
- [X] T017 [P] [US2] Write RED tests in internal/adapters/cli/mcp_run_test.go quoting data-model: "`chat` is an alias of `teams`." "Unknown namespace, verb, or flag → usage." Compare MCP text JSON to `cli.Run` stdout for `mail list` (same count and ids, SC-003)
- [X] T018 [P] [US2] Write RED tests in internal/adapters/cli/mcp_errors_test.go that `--top` above the CLI maximum is usage not a silent cap, and missing consent is `class=auth` with `isError` true per contracts/json-tool-error.schema.json (`usage` | `auth` | `service` | `not_found`)
- [X] T019 [US2] RED commit: failing tests from T016–T018 only

### Implementation for User Story 2

- [X] T020 [US2] Implement `m365_run` in internal/adapters/cli/mcp.go + mcp_dispatch.go: required `namespace` and `verb`; optional `args` and `flags` per contracts/json-run-request.schema.json; build argv and call `Run` with buffered Deps; success text = CLI stdout; failure `isError` true with stderr JSON class (do **not** return a Go error that collapses to protocol -32603)
- [X] T021 [US2] Refuse `auth login`, `auth logout`, and namespace `mcp` as usage with hint to run `m365 auth login` in a terminal quoting data-model: "`auth` `login` / `auth` `logout` / namespace `mcp` → usage; hint to use a human terminal for login." Refuse `--body-file`/`--text-file` value `-` as usage
- [X] T022 [US2] Map `files download` / `save-attachment` results as path metadata only quoting data-model: "File download / save-attachment: path metadata only; never file bytes." Test in internal/adapters/cli/mcp_run_test.go
- [X] T023 [US2] Implement acceptance/steps/mcp_run_steps.go until scripts/acceptance.sh passes features/mcp/run-read.feature
- [X] T024 [US2] After GREEN T020–T023, commit production code separately from T019; do not edit locked RED tests

**Checkpoint**: US2 independently testable; US1 still green; no Graph URLs in mcp.go.

---

## Phase 5: User Story 3 — Writes stay dry-run unless the caller opts in (Priority: P2)

**Goal**: Send/reply/create/update/delete/upload/move through MCP default to dry-run. Real write only when `write_opt_in` is true. Help names the rule (help text may land in US4; this story locks behavior).

**Independent Test**: `mail send` without opt-in → `dry_run` true, fake mailbox unchanged. `write_opt_in` true → send like CLI without `--dry-run`. `calendar create` with opt-in omitted → no event.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 3 (MANDATORY)

- [X] T025 [P] [US3] Write Gherkin in features/mcp/run-write.feature for mail send no opt-in (no send), opt-in true (sent), calendar create omitted opt-in (no event); generate failing acceptance tests via scripts/acceptance.sh
- [X] T026 [P] [US3] Write RED tests in internal/adapters/cli/mcp_write_test.go quoting data-model: "Write verbs (mail send/reply, teams send, calendar create/update/delete, files upload/create-folder/delete/move): if `write_opt_in` is not true, dispatch MUST pass `--dry-run`." "`write_opt_in` true still honors an explicit `dry-run` true (no mutation)." "`write_opt_in` false/omitted cannot be bypassed by `flags.dry-run` false."
- [X] T027 [US3] RED commit: failing tests from T025–T026 only

### Implementation for User Story 3

- [X] T028 [US3] Implement WriteGate in internal/adapters/cli/mcp_dispatch.go; reads ignore `write_opt_in`; inject `--dry-run` for write verbs unless opt-in JSON true
- [X] T029 [US3] Implement acceptance/steps/mcp_write_steps.go until scripts/acceptance.sh passes features/mcp/run-write.feature
- [X] T030 [US3] After GREEN T028–T029, commit production code separately from T027

**Checkpoint**: 100% of send/create/delete without opt-in leave fake mailbox/calendar/drive unchanged (SC-004).

---

## Phase 6: User Story 4 — Help without a session (Priority: P3)

**Goal**: `m365_help` returns CLI help for a namespace or verb with no session. Each MCP tool description is enough to choose status vs help vs run.

**Independent Test**: No session; help for `mail` and `calendar create` names flags/limits and exits success. `tools/list` descriptions distinguish the three tools. Calendar help names list, create, free, and phrase examples if CLI help does.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 4 (MANDATORY)

- [X] T031 [P] [US4] Write Gherkin in features/mcp/help.feature for help with no session on `calendar` (verbs list/create/free) and tool descriptions on tools/list; generate failing acceptance tests via scripts/acceptance.sh
- [X] T032 [P] [US4] Write RED tests in internal/adapters/cli/mcp_help_test.go: omitted namespace → root help including mcp and write opt-in; `namespace=calendar` names list/create/free; no Store.Get required
- [X] T033 [US4] RED commit: failing tests from T031–T032 only

### Implementation for User Story 4

- [X] T034 [US4] Implement `m365_help` in internal/adapters/cli/mcp.go by running in-process `--help` for the requested surface; tool Descriptions on NewMCPServer MUST suffice to choose status vs help vs run
- [X] T035 [US4] Implement acceptance/steps/mcp_help_steps.go until scripts/acceptance.sh passes features/mcp/help.feature
- [X] T036 [US4] After GREEN T034–T035, commit production code separately from T033

**Checkpoint**: Help works signed-out; US1–US3 still green.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Perf, secret scan, CLI regression, expired session, registration snippet, quickstart. No live tenant in CI.

- [X] T037 [P] CLI/MCP perf test in internal/adapters/cli/mcp_perf_test.go: `m365_status` against fake store finishes under 2 seconds (SC-002)
- [X] T038 [P] Expired-session run is `class=auth` in internal/adapters/cli/mcp_errors_test.go (no hang, no browser)
- [X] T039 [P] Assert `m365 mail list --help` and `m365 teams list --help` still name v1 flags/limits in internal/adapters/cli/help_test.go (FR-012)
- [X] T040 [P] Confirm contracts/agent-registration.md command `m365` args `["mcp","serve"]`; no Client ID/tokens in that file
- [X] T041 Split internal/adapters/cli/mcp.go / mcp_dispatch.go if either exceeds 250 lines (constitution II); over 500 MUST split before completion
- [X] T042 Run make verify on fakes; execute specs/004-m365-mcp/quickstart.md automated scenarios; no secrets in MCP responses or logs (SC-007)

**Checkpoint**: CI `make verify` without a live tenant. No lazy-mcp, no Enterprise MCP, no per-Graph tools.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: start immediately
- **Foundational (Phase 2)**: depends on Setup; BLOCKS all stories
- **US1 (Phase 3)**: after Foundational — MVP
- **US2 (Phase 4)**: after Foundational; uses serve + dispatch; independently testable from US1
- **US3 (Phase 5)**: after US2 (write gate is on `m365_run`)
- **US4 (Phase 6)**: after Foundational; MAY proceed in parallel with US2 if mcp.go conflicts are sequenced
- **Polish (Phase 7)**: after desired stories

### User Story Dependencies

- **US1 (P1) MVP**: after Foundational — catalog + status
- **US2 (P1)**: after Foundational — run reads (does not require US1 status beyond shared server)
- **US3 (P2)**: after US2
- **US4 (P3)**: after Foundational — help only

### Within Each User Story

- RED tests fail for the expected reason and RED-commit before implementation
- Dispatch/server before tool handlers
- `make unit` at story checkpoint

### Parallel Opportunities

- T004, T007 after T002
- T009–T010 (US1 tests)
- T016–T018 (US2 tests) after Foundational, parallel with US1 tests
- T025–T026 (US3 tests) after US2 RED shape exists
- T031–T032 (US4 tests) after Foundational
- T037–T040 in Polish

---

## Parallel Example: User Story 1

```bash
Task: "Gherkin features/mcp/catalog-status.feature"
Task: "RED tests internal/adapters/cli/mcp_status_test.go"
```

## Parallel Example: US1 vs US2 tests

```bash
Task: "US1 catalog/status tests"
Task: "US2 run-read tests (mcp_run_test.go + run-read.feature)"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Setup + Foundational (SDK, serve, three-tool skeleton, dispatch helper)
2. US1 `tools/list` + `m365_status`
3. **STOP**: Independent Test US1 on fakes
4. Demo: agent lists three tools and sees signed-out vs signed-in consent

### Incremental Delivery

1. Foundational → stdio server with compact catalog
2. US1 → status MVP
3. US2 → read via `m365_run`
4. US3 → write opt-in dry-run
5. US4 → help without a session
6. Polish → 2s status, `make verify`

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. After Foundational:
   - Developer A: US1 status
   - Developer B: US2 run reads (coordinate mcp.go)
   - Developer C: US4 help
3. US3 write gate after US2

---

## Notes

- [P] tasks = different files, no incomplete dependencies
- [Story] label maps to spec US1–US4
- Quote data-model constraints in RED tests; do not invent Graph URLs in MCP
- Commit RED then GREEN separately; do not edit locked RED tests
- Stop at any checkpoint to validate the story independently
- Avoid: second binary, lazy-mcp, one tool per Graph URL, collapsing error classes, file bytes on the wire
