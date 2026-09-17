# Tasks: m365 MCP Serve

**Input**: Design documents from `/specs/004-m365-mcp/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md, constitution.md, existing CLI namespaces 001–003

**Tests**: MANDATORY per Constitution I. Spec Verification Strategy requires `tools/list`, named recipes, status, JSON parity, write dry-run, auth-class vs service-class, and no secrets on the wire. RED commit (failing tests + fixtures only) then GREEN. EX-I-001 does not apply. No live Graph in unit tests. Synthetic fixtures only. Do **not** edit `internal/adapters/cli/mail.go`, `teams.go`, `calendar.go`, or `files.go`. Do **not** rewrite `specs/001-m365-cli/`, `002-calendar-files/`, or `003-calendar-natural-time/`. Do **not** implement `teams find` (005).

**Organization**: US1–US4 (T001–T042) are DONE. Remaining work is US5 lookup recipes (T043+). MVP for the remainder = US5.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallel (different files, no incomplete deps)
- **[Story]**: `[US1]`–`[US5]` on user-story phases only
- Exact file paths required

## Path Conventions

`internal/adapters/cli/mcp.go`, `internal/adapters/cli/mcp_dispatch.go`, `internal/adapters/cli/mcp_recipes.go`, `internal/adapters/cli/run.go`, `internal/adapters/cli/help.go`, `features/mcp/`, `acceptance/steps/`

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

## Phase 8: User Story 5 — Lookup recipes live in the server (Priority: P1)

**Goal**: Help and named MCP prompts include mail-search, teams-find, calendar, and files recipes so an agent needs no skill file. `tools/list` stays exactly three tools.

**Independent Test**: No session. Help with no topic and with topics `mail-search`, `teams-find`, `calendar`, `files` returns the examples in `contracts/recipes.md`. `prompts/list` has those four names. `tools/list` is still three tools.

**RED Checkpoint**: failing tests only, then GREEN.

### Tests for User Story 5 (MANDATORY)

- [X] T043 [P] [US5] Write Gherkin in features/mcp/recipes.feature for help with no topic (four topic names), topic `mail-search` includes `from:ajay`, topic `teams-find` says not to send on several matches, `prompts/list` four names, `tools/list` still three; generate failing acceptance tests via scripts/acceptance.sh
- [X] T044 [P] [US5] Write RED tests in internal/adapters/cli/mcp_recipes_test.go quoting data-model: "`prompts/list` MUST return the four recipe names (SC-009)." "`tools/list` MUST return the three tool names only (SC-001)." "Same body from `prompts/get` and from `m365_help` with that `topic`." "MUST NOT include tokens, live mailbox content, or file bytes."
- [X] T045 [P] [US5] Write RED tests in internal/adapters/cli/mcp_help_test.go for `topic=mail-search` including `mail list --folder all --search 'from:ajay'`, `mail get`, `mail thread`; `topic=teams-find` including several matches MUST NOT send and `teams list` (not `teams find`); `topic=calendar` including `calendar list`, `calendar free`, `calendar create --when 'tomorrow at 1:30 pm'`; `topic=files` including `files list`, `files download`, `--out`, dry-run upload; unknown topic → usage; `topic` and `namespace` both set → usage
- [X] T046 [US5] RED commit: git add only failing tests and fixtures from T043–T045; no production code

### Implementation for User Story 5

- [X] T047 [P] [US5] Add static recipe bodies in internal/adapters/cli/mcp_recipes.go per contracts/recipes.md; no Graph, no session, no `teams find`
- [X] T048 [US5] Extend helpIn with `topic` in internal/adapters/cli/mcp.go; empty help names three tools and four topics; `topic` returns recipe text; register four prompts via `Server.AddPrompt`; `m365_run` Description MUST point at the four recipe topics (FR-015)
- [X] T049 [US5] Name the four topics in internal/adapters/cli/mcp.go `mcpHelp` (and root mcp help if needed)
- [X] T050 [US5] Implement acceptance/steps/mcp_recipes_steps.go until scripts/acceptance.sh passes features/mcp/recipes.feature
- [X] T051 [US5] After GREEN T047–T050, commit production code separately from T046; do not edit locked RED tests

**Checkpoint**: US5 independently testable with no session; US1–US4 still green; still exactly three tools.

---

## Phase 9: US5 Polish

- [X] T052 [P] Assert `m365 mcp --help` names the four recipe topics in internal/adapters/cli/mcp_cli_test.go
- [X] T053 Run make unit (and make verify if gosec is clean enough); execute specs/004-m365-mcp/quickstart.md recipe checks; no secrets in recipe text (SC-007, SC-008, SC-009)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup–Polish (Phases 1–7)**: DONE (T001–T042)
- **US5 (Phase 8)**: remaining; depends on existing serve/help/run
- **US5 Polish (Phase 9)**: after US5

### User Story Dependencies

- **US1–US4**: complete
- **US5 (P1)**: remaining remainder-MVP — recipes/prompts; MUST NOT implement 005 `teams find`

### Within Each User Story

- RED tests fail for the expected reason and RED-commit before implementation
- Dispatch/server before tool handlers
- `make unit` at story checkpoint

### Parallel Opportunities

- T043–T045 (US5 tests)
- T047 recipes text parallel with test files once RED is locked
- T052 in US5 Polish

---

## Parallel Example: User Story 5

```bash
Task: "Gherkin features/mcp/recipes.feature"
Task: "RED tests internal/adapters/cli/mcp_recipes_test.go"
Task: "RED tests internal/adapters/cli/mcp_help_test.go topic cases"
```

---

## Implementation Strategy

### Remainder MVP (User Story 5)

1. RED T043–T046
2. GREEN recipes + prompts + help topics
3. **STOP**: Independent Test US5 (no session, three tools still)

### Incremental Delivery

US1–US4 already shipped. US5 adds recipes without a fourth tool.

---

## Notes

- [P] tasks = different files, no incomplete dependencies
- [Story] label maps to spec US1–US5
- Quote data-model constraints in RED tests; do not invent Graph URLs or `teams find` in MCP
- Commit RED then GREEN separately; do not edit locked RED tests
- Avoid: fourth tool, skill-file dependency, live mailbox text in recipes
