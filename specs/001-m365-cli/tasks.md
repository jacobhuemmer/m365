# Tasks: m365 CLI v1

**Input**: Design documents from `/specs/001-m365-cli/`

**Prerequisites**: plan.md, spec.md, cli-contract.md, auth.md, mail.md, teams.md, attachments.md, research.md, data-model.md, contracts/, quickstart.md, constitution.md

**Tests**: Tests are MANDATORY per Constitution Principle I (Test-Driven Development). Every behavior slice includes Gherkin under `features/` when user-observable AND a focused RED unit/use-case/CLI test. Confirm each test fails for the expected reason (not compile/import). Commit Gherkin + RED tests as a locked RED Checkpoint BEFORE matching implementation. Do not edit a locked RED test without human approval. Unit tests use fakes/ports: no live Microsoft Graph, no real Keychain, no live mailbox fixtures. Synthetic files only under `testdata/`. Acceptance: Gherkin → APS IR → generated tests per contracts/acceptance-pipeline.md; generated files are not hand-edited.

**Organization**: Setup and Foundational block all stories. User stories US1–US10 map to spec US-001–US-010. MVP = Setup + Foundational + US1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: `[US1]`–`[US10]` on user-story phases only
- Include exact file paths in descriptions

## Path Conventions

Single Go module at repository root per plan.md: `cmd/m365/`, `internal/{config,domain,app/{auth,mail,teams},adapters/{cli,graph,keychain,fs,watchstate}}`, `features/`, `acceptance/{generated,runtime,steps,runner}`, `testdata/`, `Makefile`, `scripts/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Go module, directory skeleton, Makefile gates, tool install, gitignore, env placeholders.

- [X] T001 Initialize Go module `github.com/masonhuemmer/m365` with `go 1.25` in go.mod and go.sum; create directory skeleton `cmd/m365/`, `internal/config/`, `internal/domain/`, `internal/app/auth/`, `internal/app/mail/`, `internal/app/teams/`, `internal/adapters/cli/`, `internal/adapters/graph/`, `internal/adapters/keychain/`, `internal/adapters/fs/`, `internal/adapters/watchstate/`, `features/auth/`, `features/mail/`, `features/teams/`, `features/attachments/`, `features/cli/`, `acceptance/generated/`, `acceptance/runtime/`, `acceptance/steps/`, `acceptance/runner/`, `testdata/`, `scripts/` with placeholder `doc.go` only where a package would not compile
- [X] T002 [P] Create Makefile at repo root with targets `fmt`, `vet`, `unit` (`go test ./...` excluding `./acceptance/generated`), `race`, `coverage`, `gosec`, `govulncheck`, `acceptance`, `acceptance-mutation`, `crap`, `verify` aggregating those gates per plan.md Make Targets
- [X] T003 [P] Create scripts/install-tools.sh installing gosec@v2.25.0, govulncheck@v1.6.0, and APS tools from `unclebob/Acceptance-Pipeline-Specification` pinned to a reviewed commit SHA recorded in this task's notes; T010 MUST use that SHA's gherkin-parser, gherkin-ir-dry-checker, and gherkin-mutator (no in-repo substitute without a written exception)
- [X] T004 [P] Create .gitignore at repo root excluding binaries (`bin/`, `m365`), coverage.out, `build/acceptance/`, `build/acceptance-mutation/`, `.env`, key material, `acceptance/generated/` if regenerated locally, and OS junk
- [X] T005 [P] Create .env.example at repo root with empty placeholders only: `M365_CLIENT_ID=` and `M365_TENANT_ID=` — never real values

**Checkpoint**: `make fmt vet` is runnable on the skeleton; no Jenkins, Docker, MCP, or Helm files exist.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: APS pipeline, config, domain result/limit/errors, CLI shell, Graph fake, Keychain fake, fs port, `cmd/m365` wiring.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete. TDD applies: RED tests committed before implementation.

- [X] T006 [P] Write RED tests in acceptance/runtime/runtime_test.go then implement APS IR load, Background+Scenario expansion, fresh world state, regex step dispatch, and assertion vs infrastructure error classification in acceptance/runtime/runtime.go per contracts/acceptance-pipeline.md
- [X] T007 [P] Write RED tests in cmd/acceptance-entrypoint-generator/main_test.go then implement IR-to-`acceptance/generated/<feature>_acceptance_test.go` generation with generated-code headers and missing-handler failure in cmd/acceptance-entrypoint-generator/main.go
- [X] T008 [P] Write RED tests in acceptance/steps/registry_test.go then implement regex-capture handler registry and fake-clock/world helpers in acceptance/steps/registry.go
- [X] T009 [P] Write RED tests in acceptance/runner/main_test.go then implement APS mutation worker stdin/stdout protocol classifying `test_success`/`test_failure`/`infrastructure_error` and writing reports under build/acceptance-mutation/ in acceptance/runner/main.go
- [X] T010 Create scripts/acceptance.sh (gherkin-parser → build/acceptance/ir → gherkin-ir-dry-checker → generator → `go test ./acceptance/generated`) and scripts/acceptance-mutation.sh (gherkin-mutator through acceptance/runner) per contracts/acceptance-pipeline.md using the APS SHA from T003
- [X] T011 [P] Write RED tests in internal/config/config_test.go then implement env `M365_CLIENT_ID`/`M365_TENANT_ID` plus `~/.config/m365/config.json` keys `client_id`/`tenant_id` (env overrides file) in internal/config/config.go; missing client/tenant → exit 3; never log values
- [X] T012 [P] Write RED tests in internal/domain/result_test.go then implement CommandResult and ResultLimit in internal/domain/result.go quoting data-model: "Mail list default top = 10. Teams list/messages/watch default top = 20. Maximum top = 50. Values `<1` or `>50` are usage failures, not silent caps." Token is not a service URL
- [X] T013 [P] Write RED tests in internal/domain/errors_test.go then implement exit classes `0/3/4/5/6` and error types without SDK fields in internal/domain/errors.go mapping to stderr classes `usage`/`auth`/`service`/`not_found` per contracts/json-error.schema.json
- [X] T014 [P] Write RED tests in internal/adapters/cli/root_test.go then implement `flag.FlagSet` command registry, global `--json`/`--human` (both set → exit 3), stdout data vs stderr diagnostics, JSON error object on stderr, and `--help` exit 0 without a session in internal/adapters/cli/root.go
- [X] T015 [P] Write RED tests in internal/adapters/graph/fake_test.go then implement `httptest` fake Graph in internal/adapters/graph/fake.go per contracts/fake-graph.md covering tokens `fake-both`/`fake-mail`/`fake-teams`/`fake-expired`, 401/403/404/429, independent consent, and synthetic messages `msg-1`/`chat-1` — no live tenant
- [X] T016 [P] Write RED tests in internal/adapters/keychain/store_test.go then implement SecretStore port and in-memory fake in internal/adapters/keychain/fake.go and internal/adapters/keychain/store.go; unit tests MUST NOT call real Keychain
- [X] T099 Implement macOS Keychain `auth.Store` in internal/adapters/keychain/keyring.go using github.com/zalando/go-keyring; unit tests still use Fake; live wiring of Keychain-plus-0600-fallback is T126
- [X] T017 [P] Write RED tests in internal/adapters/fs/fs_test.go then implement local filesystem port (read attach path, write save path, refuse existing without overwrite) in internal/adapters/fs/fs.go quoting data-model: "If the destination exists, save MUST refuse with exit `3` unless `--overwrite` is present."
- [X] T018 Implement cmd/m365/main.go as wiring only: load config, construct keychain/graph/fs adapters (real vs test injected), register CLI, map domain errors to exit codes — no behavior in main; live token store is T126 (Keychain then 0600 file)

**Checkpoint**: Foundation ready — `m365 --help` exits 0, missing config on a command that needs it exits 3, fake Graph and fake Keychain are injectable, `make unit` passes. User stories may start.

---

## Phase 3: User Story 1 — Sign in, inspect status, and log out (Priority: P1) 🎯 MVP

**Goal**: The mailbox owner signs in with an interactive Microsoft 365 user gesture, then sees whether a silent session is usable. Later commands reuse that session until logout or expiry. Logout forgets the saved session. "Not signed in" is distinct from "signed in, but Microsoft 365 rejected the request."

**Independent Test**: From signed-out, run login (or a test double of the user gesture), status, logout, and status again. Assert status reports a usable session after login, logout clears it, and a workload command after logout exits `4` (stub ok; `mail list` reuse is US2). Assert a usable session plus a service rejection exits `5`, not `4`.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 1 (MANDATORY)

- [X] T019 [P] [US1] Write Gherkin in features/auth/login-status-logout.feature for login success, status signed-out exit 0 with `signed_in=false`, status usable without secrets, logout idempotent, expired session → exit 4, service rejection of usable session → exit 5; generate failing acceptance/generated/login-status-logout_acceptance_test.go via scripts/acceptance.sh
- [X] T020 [P] [US1] Write RED use-case tests in internal/app/auth/login_test.go for login (fake browser/callback), session reuse, logout clearing SecretStore, missing config → usage
- [X] T021 [P] [US1] Write RED CLI tests in internal/adapters/cli/auth_test.go for `auth login`/`status`/`logout` stdout JSON per contracts/json-stdout.schema.json and no token fields
- [X] T097 [US1] RED commit: git add only failing tests and fixtures from T019–T021; commit with no production code; record commit id in task notes

### Implementation for User Story 1

- [X] T022 [P] [US1] Implement Session in internal/domain/session.go with tests in internal/domain/session_test.go: signed in, session usable, account display name, mail consented, teams consented; Session MUST NOT contain access tokens, refresh tokens, authorization codes, or client secrets
- [X] T023 [US1] Implement auth ports and use cases (login with injected opener, status, logout) in internal/app/auth/ports.go and internal/app/auth/login.go using SecretStore fake — PKCE/localhost live only in the auth adapter, not domain
- [X] T024 [US1] Implement PKCE public-client adapter (empty client secret, `GenerateVerifier`/`S256ChallengeOption`/`VerifierOption`) in internal/adapters/graph/oauth.go with tests in internal/adapters/graph/oauth_test.go against a fake authorize/token server — never print codes or tokens
- [X] T101 [US1] Browser opener + localhost callback ports in internal/app/auth/ports.go; fake in tests; real opener in internal/adapters/graph/oauth.go (PKCE S256, no device-code)
- [X] T025 [US1] Register `auth login|status|logout` in internal/adapters/cli/auth.go wiring app/auth use cases (depends on T023, T024)
- [X] T026 [US1] Implement acceptance/steps/auth_steps.go until scripts/acceptance.sh passes features/auth/login-status-logout.feature (depends on T025)
- [X] T098 [US1] After GREEN T022–T026 and T101, commit production code separately from the T097 RED commit; do not edit locked RED tests

**Checkpoint**: US1 independently testable against fakes; `make unit` green; previous foundation still green. Workload after logout exits `4` without requiring `mail list`.

---

## Phase 4: User Story 2 — List and read Outlook mail (Priority: P1)

**Goal**: The signed-in user lists inbox or a named folder, optionally unread-only or search-filtered, with an explicit result limit. The user reads one message as text and reads a conversation thread oldest-first, including the user's own replies. Messages with documents show attachment metadata (name, size, type) without dumping file bytes.

**Independent Test**: With a usable mail session from login (no second gesture) and synthetic mailbox fixtures, list the inbox with default limit, list unread, search, get one message, and read a thread with and without bodies. Assert empty folder success, unknown id exit `6`, attachment metadata without bytes, and the applied limit in the result.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 2 (MANDATORY)

- [X] T027 [P] [US2] Write Gherkin in features/mail/list-get-thread.feature for default `--top` 10, unread, search, empty inbox exit 0, get as text, thread oldest-first with/without `--bodies`, unknown id exit 6, attachment metadata without bytes; generate failing acceptance/generated/list-get-thread_acceptance_test.go via scripts/acceptance.sh
- [X] T028 [P] [US2] Write RED tests in internal/app/mail/list_test.go and internal/app/mail/get_test.go for ResultLimit defaults ("Mail list default top = 10. Maximum top = 50. Values `<1` or `>50` are usage failures, not silent caps."), empty list success, not-found
- [X] T029 [P] [US2] Write RED CLI tests in internal/adapters/cli/mail_read_test.go for `mail list|get|thread` JSON `limit`/`count`/`next_page` and no attachment bytes on stdout
- [X] T102 [US2] RED commit: git add only failing tests and fixtures from T027–T029; commit with no production code; record commit id in task notes

### Implementation for User Story 2

- [X] T030 [P] [US2] Implement MailMessage and MailThread in internal/domain/mail.go with tests in internal/domain/mail_test.go (oldest-first thread including own replies; body optional)
- [X] T031 [US2] Implement mail ports (list/get/thread) next to the consumer in internal/app/mail/ports.go and use cases in internal/app/mail/list.go, internal/app/mail/get.go, internal/app/mail/thread.go — MUST NOT import `internal/app/teams`; Graph URLs stay out of this package
- [X] T032 [US2] Map list/get/thread HTTP+JSON to domain in internal/adapters/graph/mail.go with tests in internal/adapters/graph/mail_test.go that MUST call NewFakeServer (`httptest`) per contracts/fake-graph.md (`Prefer` text body; opaque `next_page`; never emit service URLs). Memory-only tests do not satisfy this task.
- [X] T033 [US2] Register `mail list|get|thread` flags (`--folder`, `--unread`, `--search`, `--top`, `--page-token`, `--bodies`) in internal/adapters/cli/mail.go (depends on T031, T032)
- [X] T034 [US2] Implement acceptance/steps/mail_read_steps.go until scripts/acceptance.sh passes features/mail/list-get-thread.feature (depends on T033)
- [X] T103 [US2] After GREEN T030–T034, commit production code separately from the T102 RED commit; do not edit locked RED tests

**Checkpoint**: US2 independently testable; US1 still green; `make unit` green.

---

## Phase 5: User Story 3 — List Teams chats and read messages (Priority: P1)

**Goal**: The signed-in user lists chats and reads messages in a chat, with explicit result limits. Messages with documents show attachment metadata the same way mail does.

**Independent Test**: With a usable Teams session and synthetic chat fixtures, list chats with the default limit, get one chat with members, and list messages. Assert empty list success, unknown chat exit `6`, attachment metadata without bytes, and the applied limit in the result.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 3 (MANDATORY)

- [X] T035 [P] [US3] Write Gherkin in features/teams/list-get-messages.feature for default `--top` 20, empty chat list exit 0, get+members, messages oldest-first, unknown chat exit 6, metadata without bytes; generate failing acceptance/generated/list-get-messages_acceptance_test.go via scripts/acceptance.sh
- [X] T036 [P] [US3] Write RED tests in internal/app/teams/list_test.go and internal/app/teams/messages_test.go quoting "Teams list/messages/watch default top = 20. Maximum top = 50. Values `<1` or `>50` are usage failures, not silent caps."
- [X] T037 [P] [US3] Write RED tests in internal/app/mail/isolation_test.go that fail if this package imports `internal/app/teams`, and in internal/app/teams/isolation_test.go that fail if this package imports `internal/app/mail`
- [X] T038 [P] [US3] Write RED CLI tests in internal/adapters/cli/teams_read_test.go for `teams list|get|messages` and `chat` alias routing to the same handlers (CLI adapter only, not a domain)
- [X] T104 [US3] RED commit: git add only failing tests and fixtures from T035–T038; commit with no production code; record commit id in task notes

### Implementation for User Story 3

- [X] T039 [P] [US3] Implement Chat and ChatMessage in internal/domain/teams.go with tests in internal/domain/teams_test.go
- [X] T040 [US3] Implement teams ports and use cases in internal/app/teams/ports.go, internal/app/teams/list.go, internal/app/teams/get.go, internal/app/teams/messages.go — MUST NOT import `internal/app/mail`; Graph URLs stay in the adapter
- [X] T041 [US3] Map chats/messages HTTP+JSON in internal/adapters/graph/teams.go with tests in internal/adapters/graph/teams_test.go that MUST call NewFakeServer (`httptest`) per contracts/fake-graph.md (`--include-system` default off). Memory-only tests do not satisfy this task.
- [X] T042 [US3] Register `teams list|get|messages` and `chat` alias in internal/adapters/cli/teams.go (alias is CLI routing only) (depends on T040, T041)
- [X] T043 [US3] Implement acceptance/steps/teams_read_steps.go until scripts/acceptance.sh passes features/teams/list-get-messages.feature (depends on T042)
- [X] T105 [US3] After GREEN T039–T043, commit production code separately from the T104 RED commit; do not edit locked RED tests

**Checkpoint**: US3 independently testable; mail packages still do not import teams; `make unit` green.

---

## Phase 6: User Story 4 — Send and reply to mail with dry-run (Priority: P2)

**Goal**: The user sends new mail or replies (sender only, or all recipients). Dry-run shows the exact intended send without sending. A real send happens only when the command runs without dry-run.

**Independent Test**: Dry-run a send and a reply against fixtures; assert no message is created. Run the same commands without dry-run; assert a message is created and stdout reports success with a message identity. Omit required fields and assert exit `3`.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 4 (MANDATORY)

- [X] T044 [P] [US4] Write Gherkin in features/mail/send-reply.feature for dry-run send/reply (no Sent Item), real send creates id, `--all` vs sender-only, missing `--to`/`--subject`/body → exit 3, unknown reply target → exit 6; generate failing acceptance/generated/send-reply_acceptance_test.go via scripts/acceptance.sh
- [X] T045 [P] [US4] Write RED tests in internal/app/mail/send_test.go asserting dry-run does not call send port and does not create a message
- [X] T046 [P] [US4] Write RED CLI tests in internal/adapters/cli/mail_send_test.go for `--dry-run` JSON (`dry_run: true`) vs sent `id`
- [X] T106 [US4] RED commit: git add only failing tests and fixtures from T044–T046; commit with no production code; record commit id in task notes

### Implementation for User Story 4

- [X] T047 [US4] Implement DryRunSend and mail send/reply use cases in internal/app/mail/send.go and internal/app/mail/reply.go (body from flag or file/stdin including `-`)
- [X] T048 [US4] Map send/reply HTTP in internal/adapters/graph/mail_send.go with tests in internal/adapters/graph/mail_send_test.go — dry-run MUST NOT reach create
- [X] T049 [US4] Register `mail send|reply` (`--to`, `--subject`, `--body`/`--body-file`, `--cc`, `--html`, `--dry-run`, `--all`) in internal/adapters/cli/mail.go
- [X] T050 [US4] Implement acceptance/steps/mail_send_steps.go until scripts/acceptance.sh passes features/mail/send-reply.feature (depends on T049)
- [X] T107 [US4] After GREEN T047–T050, commit production code separately from the T106 RED commit; do not edit locked RED tests

**Checkpoint**: US4 independently testable; dry-run never sends; `make unit` green.

---

## Phase 7: User Story 5 — Send a Teams chat message with dry-run (Priority: P2)

**Goal**: The user sends a message to a chat. Dry-run shows the exact intended send without sending. A real send happens only without dry-run.

**Independent Test**: Dry-run a send against a known chat; assert no message is created. Send without dry-run; assert the message is created and stdout reports an identity. Missing text exits `3`. Unknown chat exits `6`.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 5 (MANDATORY)

- [X] T051 [P] [US5] Write Gherkin in features/teams/send.feature for dry-run (no new chat message), real send id, missing text → 3, unknown chat → 6; generate failing acceptance/generated/send_acceptance_test.go via scripts/acceptance.sh
- [X] T052 [P] [US5] Write RED tests in internal/app/teams/send_test.go asserting dry-run does not call send port and does not create a message
- [X] T053 [P] [US5] Write RED CLI tests in internal/adapters/cli/teams_send_test.go for `--dry-run` vs sent `id`
- [X] T108 [US5] RED commit: git add only failing tests and fixtures from T051–T053; commit with no production code; record commit id in task notes

### Implementation for User Story 5

- [X] T054 [US5] Implement teams send use case in internal/app/teams/send.go (`--text`/`--text-file`, optional `--html`/`--format md`) without importing mail
- [X] T055 [US5] Map send HTTP in internal/adapters/graph/teams_send.go with tests in internal/adapters/graph/teams_send_test.go — dry-run MUST NOT reach create
- [X] T056 [US5] Register `teams send` in internal/adapters/cli/teams.go
- [X] T057 [US5] Implement acceptance/steps/teams_send_steps.go until scripts/acceptance.sh passes features/teams/send.feature (depends on T056)
- [X] T109 [US5] After GREEN T054–T057, commit production code separately from the T108 RED commit; do not edit locked RED tests

**Checkpoint**: US5 independently testable; dry-run never sends; `make unit` green.

---

## Phase 8: User Story 6 — Attach local documents on send or reply (Priority: P2)

**Goal**: The user attaches one or more local documents when sending or replying (mail and Teams). Dry-run lists the files that would be attached (name, size) and does not upload. Missing, unreadable, empty, oversize, and too-many files are validation failures, not service errors.

**Independent Test**: Dry-run send with two synthetic local files; assert names and sizes appear and no upload occurs. Repeat without dry-run and assert the sent message carries those attachments. Missing path, empty file, file over 10 MiB, and an 11th file each exit `3` with no send.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 6 (MANDATORY)

- [X] T058 [P] [US6] Write Gherkin in features/attachments/attach-send.feature for dry-run listing name+size with no upload, real send includes attachments, missing/unreadable/URL/empty/oversize/too-many → exit 3; generate failing acceptance/generated/attach-send_acceptance_test.go via scripts/acceptance.sh
- [X] T059 [P] [US6] Write RED tests in internal/app/mail/attach_test.go and internal/app/teams/attach_test.go quoting "Outbound size MUST be `1..10485760` (10 MiB). Zero-byte files are invalid." and "Outbound count per send/reply MUST be `1..10` when any attach flags are present"; assert dry-run does not upload and stdout has no file bytes
- [X] T060 [P] [US6] Add synthetic fixture testdata/note.txt (contents `synthetic-ok` only) and RED CLI tests in internal/adapters/cli/attach_test.go for `--attach` dry-run metadata without bytes on stdout
- [X] T110 [US6] RED commit: git add only failing tests and fixtures from T058–T060; commit with no production code; record commit id in task notes

### Implementation for User Story 6

- [X] T061 [P] [US6] Implement Attachment in internal/domain/attachment.go with tests in internal/domain/attachment_test.go (metadata only; local path outbound; two same base names allowed if paths differ)
- [X] T062 [US6] Validate attach paths via internal/adapters/fs/fs.go in mail/teams send use cases (internal/app/mail/send.go, internal/app/teams/send.go, internal/app/mail/reply.go) — remote URL rejected as usage; no silent skip
- [X] T063 [US6] Upload bytes only on non-dry-run in internal/adapters/graph/attach.go with tests in internal/adapters/graph/attach_test.go against the fake
- [X] T064 [US6] Register repeatable `--attach` on mail send/reply and teams send in internal/adapters/cli/mail.go and internal/adapters/cli/teams.go
- [X] T065 [US6] Implement acceptance/steps/attach_steps.go until scripts/acceptance.sh passes features/attachments/attach-send.feature (depends on T064)
- [X] T111 [US6] After GREEN T061–T065, commit production code separately from the T110 RED commit; do not edit locked RED tests

**Checkpoint**: US6 independently testable; dry-run does not upload; no bytes on stdout; `make unit` green.

---

## Phase 9: User Story 7 — Save a named attachment to a local path (Priority: P2)

**Goal**: The user names the message or chat message, the attachment, and a destination path. Success reports the path written. Omitting the path is a usage error. An existing local file is refused unless overwrite is explicit.

**Independent Test**: Save a synthetic attachment to a new path; assert the file exists and stdout reports that path. Omit the path (exit `3`, no write). Save to an existing path without overwrite (exit `3`, file unchanged). Save with overwrite (file replaced). Unknown attachment exits `6`.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 7 (MANDATORY)

- [X] T066 [P] [US7] Write Gherkin in features/attachments/save.feature for save to `--out` reporting path, omitted path → 3 no write, existing without `--overwrite` → 3 file unchanged, `--overwrite` replaces, unknown → 6, ambiguous display name → 3, no bytes on stdout; generate failing acceptance/generated/save_acceptance_test.go via scripts/acceptance.sh
- [X] T067 [P] [US7] Write RED tests in internal/app/mail/save_test.go and internal/app/teams/save_test.go quoting "If the destination exists, save MUST refuse with exit `3` unless `--overwrite` is present."
- [X] T068 [P] [US7] Write RED CLI tests in internal/adapters/cli/save_attachment_test.go for `mail save-attachment` and `teams save-attachment` — stdout JSON `path` only, never file bytes
- [X] T112 [US7] RED commit: git add only failing tests and fixtures from T066–T068; commit with no production code; record commit id in task notes

### Implementation for User Story 7

- [X] T069 [US7] Implement save use cases in internal/app/mail/save.go and internal/app/teams/save.go (parent namespace only; wrong-namespace id is not a successful cross-save)
- [X] T070 [US7] Download content only into the fs port in internal/adapters/graph/save.go with tests in internal/adapters/graph/save_test.go (metadata list verbs `mail attachments` / `teams attachments`)
- [X] T071 [US7] Register `mail attachments`, `mail save-attachment`, `teams attachments`, `teams save-attachment` (`--out`, `--overwrite`) in internal/adapters/cli/mail.go and internal/adapters/cli/teams.go
- [X] T072 [US7] Implement acceptance/steps/save_steps.go until scripts/acceptance.sh passes features/attachments/save.feature (depends on T071)
- [X] T113 [US7] After GREEN T069–T072, commit production code separately from the T112 RED commit; do not edit locked RED tests

**Checkpoint**: US7 independently testable; save writes only the given path; no bytes on stdout; `make unit` green.

---

## Phase 10: User Story 8 — Watch Teams for new messages (Priority: P2)

**Goal**: The user or a local watcher consumes a stream of new messages the signed-in user should see: 1:1 chats, extra named chats, and mentions. Watch is a one-shot JSON-lines poll plus a 0600 checkpoint, not a daemon.

**Independent Test**: With synthetic events, run watch with no new messages (empty success), with 1:1 and mention events (emitted), with `--chat` for an extra chat, with `--since`, and with more events than `--top` (limit visible, checkpoint only through emitted events). Assert JSON lines on stdout.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 8 (MANDATORY)

- [X] T073 [P] [US8] Write Gherkin in features/teams/watch.feature for empty watch (exit 0, empty stdout JSON), `one_to_one`/`mention`/`watched_chat` reasons, `--chat`, `--since`, `--top` visible without skipping unemitted events, unknown `--chat` → 6, no session → 4; generate failing acceptance/generated/watch_acceptance_test.go via scripts/acceptance.sh
- [X] T074 [P] [US8] Write RED tests in internal/app/teams/watch_test.go for one-shot poll, JSON lines schema contracts/json-watch-event.schema.json, checkpoint advances only through emitted events
- [X] T075 [P] [US8] Write RED tests in internal/adapters/watchstate/store_test.go for `$XDG_STATE_HOME/m365/watch.json` default `~/.local/state/m365/watch.json` mode `0600`, high-watermark ids only — never tokens or message bodies
- [X] T114 [US8] RED commit: git add only failing tests and fixtures from T073–T075; commit with no production code; record commit id in task notes

### Implementation for User Story 8

- [X] T076 [US8] Implement WatchCheckpoint in internal/domain/watch.go with tests in internal/domain/watch_test.go matching data-model WatchCheckpoint
- [X] T077 [US8] Implement watchstate file store in internal/adapters/watchstate/store.go (0600, no secrets)
- [X] T078 [US8] Implement watch use case in internal/app/teams/watch.go (poll, not a daemon; no overnight agent spawn)
- [X] T079 [US8] Register `teams watch` (`--since`, `--chat`, `--top`) JSON lines in internal/adapters/cli/teams.go
- [X] T080 [US8] Implement acceptance/steps/watch_steps.go until scripts/acceptance.sh passes features/teams/watch.feature (depends on T079)
- [X] T115 [US8] After GREEN T076–T080, commit production code separately from the T114 RED commit; do not edit locked RED tests

**Checkpoint**: US8 independently testable; watch is not a daemon; `make unit` green.

---

## Phase 11: User Story 9 — Use mail and Teams with independent consent (Priority: P3)

**Goal**: The user can grant mail access without Teams access, and Teams without mail. Missing consent for one namespace MUST NOT break the other. Attachment operations stay inside the namespace of the message they belong to.

**Independent Test**: With only mail consent, run mail list (success) and teams list (exit `4`). With only Teams consent, run the reverse. Attempting to save a Teams attachment through a mail command fails as validation or not-found in the mail namespace, not as a Teams call.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 9 (MANDATORY)

- [X] T081 [P] [US9] Write Gherkin in features/auth/consent-isolation.feature for `fake-mail` vs `fake-teams` tokens (mail list success / teams list exit 4 and reverse; mail save of a Teams id does not call Teams); generate failing acceptance/generated/consent-isolation_acceptance_test.go via scripts/acceptance.sh
- [X] T082 [P] [US9] Write RED tests in internal/app/auth/consent_test.go that missing namespace consent is exit 4 not 5
- [X] T083 [P] [US9] Extend internal/app/mail/isolation_test.go and internal/app/teams/isolation_test.go so mail save/list attachments never import teams and vice versa
- [X] T116 [US9] RED commit: git add only failing tests and fixtures from T081–T083; commit with no production code; record commit id in task notes

### Implementation for User Story 9

- [X] T084 [US9] Enforce consent flags on mail vs teams use cases in internal/app/mail/ports.go consumers and internal/app/teams/ports.go consumers using Session.mail/teams consented
- [X] T085 [US9] Map fake Graph 403 insufficient-consent to auth class in internal/adapters/graph/errors.go with tests in internal/adapters/graph/errors_test.go
- [X] T086 [US9] Implement acceptance/steps/consent_steps.go until scripts/acceptance.sh passes features/auth/consent-isolation.feature (depends on T084, T085)
- [X] T117 [US9] After GREEN T084–T086, commit production code separately from the T116 RED commit; do not edit locked RED tests

**Checkpoint**: US9 independently testable; mail-only and teams-only fixtures pass; `make unit` green.

---

## Phase 12: User Story 10 — Discover commands via help (Priority: P3)

**Goal**: A person or a local agent runs `m365 --help` and the `--help` of a namespace or verb. Help names the namespace, required inputs, output modes, list limits, and attachment size and count caps so the caller can invoke the command without guessing.

**Independent Test**: Invoke top-level, namespace, and verb `--help` with no saved session and assert the required names, limits, and caps appear on stdout with exit `0`.

**RED Checkpoint**: commit failing tests and required fixtures only, then GREEN with the smallest production change. Do not edit a locked RED test without human approval.

### Tests for User Story 10 (MANDATORY)

- [X] T087 [P] [US10] Write Gherkin in features/cli/help.feature for `m365 --help`, `mail send --help`, `teams send --help`, `mail list --help`, `teams list --help` signed-out exit 0 naming JSON/human, exit classes, `--top` defaults/max, attach 10 MiB and 10 files; generate failing acceptance/generated/help_acceptance_test.go via scripts/acceptance.sh
- [X] T088 [P] [US10] Write RED CLI tests in internal/adapters/cli/help_test.go asserting help text includes `--top` default 10 mail / 20 teams, max 50, and attach caps 10 MiB / 10 files with no session
- [X] T118 [US10] RED commit: git add only failing tests and fixtures from T087–T088; commit with no production code; record commit id in task notes

### Implementation for User Story 10

- [X] T089 [US10] Implement help text on each FlagSet in internal/adapters/cli/root.go, internal/adapters/cli/mail.go, and internal/adapters/cli/teams.go (caps and limits always visible)
- [X] T090 [US10] Implement acceptance/steps/help_steps.go until scripts/acceptance.sh passes features/cli/help.feature (depends on T089)
- [X] T119 [US10] After GREEN T089–T090, commit production code separately from the T118 RED commit; do not edit locked RED tests

**Checkpoint**: US10 independently testable signed-out; `make unit` green.

---

## Phase 13: Polish & Cross-Cutting Concerns

**Purpose**: Alias parity, secret redaction, verify/mutation/CRAP, quickstart on fakes. No live tenant in CI.

- [X] T091 [P] Write RED then implement `chat` alias parity tests in internal/adapters/cli/chat_alias_test.go (every teams verb reachable as `chat` with same flags/exits)
- [X] T092 [P] Write RED then implement secret redaction tests in internal/adapters/cli/redact_test.go (`--verbose`/`--debug` never prints access tokens, refresh tokens, authorization codes, or client secrets)
- [X] T093 Create scripts/crap.sh computing CRAP from coverage + complexity; CRAP > 15 fails unless justified in task notes; wire `make crap`
- [X] T094 Run make verify green on fakes (fmt vet unit race coverage gosec govulncheck acceptance crap) with no live tenant
- [X] T095 Run make acceptance-mutation after acceptance contracts exist; record APS SHA in task notes
- [X] T096 Execute specs/001-m365-cli/quickstart.md automated scenarios against fakes (status signed-out, mail list empty, dry-run send with attach, save refuse-without-overwrite)
- [X] T100 [P] CLI test internal/adapters/cli/perf_list_test.go: mail list and teams list against fake Graph finish in under 15s (SC-014)

**Checkpoint**: CI can run `make verify` without a live tenant. No MCP, calendar, OneDrive, `graph GET`, or `--guard` tasks exist.

---

## Phase 14: Live contract gaps (post-validation)

**Purpose**: Fix live-path defects vs spec: not-found exit class, list bodies, opaque `next_page`, live `teams watch`, watch mentions (FR-042), list/thread attachment metadata (FR-032/FR-053), Keychain-plus-0600 token store.

- [X] T120 Map Graph 404 and invalid-id 400 (`ErrorInvalidId` / itemNotFound) to exit 6 in internal/adapters/graph/errors.go with tests in internal/adapters/graph/errors_test.go; mail get unknown id MUST NOT be exit 5
- [X] T121 Strip body from mail list (and omit body in `$select`) in internal/adapters/graph/httpmail.go; test in internal/adapters/graph/mail_test.go that list items have empty body while get still returns body; attachment metadata on list/thread is T125
- [X] T122 Encode `@odata.nextLink` as opaque `next_page` (not a Graph URL) in internal/adapters/graph/httpmail.go; decode `--page-token` only when it maps to the Graph base; tests in internal/adapters/graph/mail_test.go
- [X] T123 Implement HTTPTeams.Watch in internal/adapters/graph/httpteams.go (one-shot poll of 1:1 chats, `--chat`, `--since`; unknown `--chat` → not-found); tests via NewFakeServer in internal/adapters/graph/teams_test.go; mention-reason events are T124
- [X] T124 [US8] RED then GREEN: `teams watch` MUST emit reason `mention` for @mentions of the signed-in user outside 1:1 and `--chat` (FR-042) in internal/adapters/graph/httpwatch.go with tests in internal/adapters/graph/teams_test.go via NewFakeServer; do not weaken features/teams/watch.feature
- [X] T125 [US2] RED then GREEN: `mail list` and `mail thread` MUST include attachment metadata (id, display name, size, content type) per FR-032/FR-053 without putting body text on list; `$expand=attachments` (or equivalent) in internal/adapters/graph/httpmail.go; tests in internal/adapters/graph/mail_test.go via NewFakeServer
- [X] T126 [US1] RED then GREEN: live cmd/m365 (not M365_FAKE=1) MUST try Keychain `auth.Store` first and, if `Set` fails because the blob is too large, persist in `keychain.FileStore` at `LiveSessionPath()` mode `0600`; logout MUST delete both; tests in internal/adapters/keychain without calling real Keychain; wire in cmd/m365/main.go

---

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1–US10 (Phases 3–12)**: Depend on Foundational
  - Sequential default: US1 → US2 → US3 → US4 → US5 → US6 → US7 → US8 → US9 → US10
  - After Foundational, US2 (mail) and US3 (teams) MAY proceed in parallel (different packages; isolation tests T037)
  - US4 depends on US2 send surface; US5 depends on US3; US6 depends on US4+US5; US7 depends on US2+US3; US8 depends on US3; US9 depends on US2+US3; US10 can start after CLI shell (Foundational) but should follow verbs existing
- **Polish (Phase 13)**: Depends on desired stories being complete
- **Live contract gaps (Phase 14)**: Depends on Polish; T124–T126 remain open after T120–T123

### User Story Dependencies

- **US1 (P1) MVP**: After Foundational only
- **US2 (P1)**: After Foundational; first useful workload after auth
- **US3 (P1)**: After Foundational; MUST NOT import mail packages
- **US4 (P2)**: After US2
- **US5 (P2)**: After US3
- **US6 (P2)**: After US4 and US5
- **US7 (P2)**: After US2 and US3
- **US8 (P2)**: After US3
- **US9 (P3)**: After US2 and US3
- **US10 (P3)**: After verbs exist (practically after US6/US7 so attach caps in send help are true)

### Within Each User Story

- Gherkin + RED unit/CLI tests MUST fail for the expected reason and be RED-committed before implementation
- Domain before use cases before Graph adapter before CLI wiring before acceptance steps
- `make unit` at story checkpoint

### Parallel Opportunities

- T002–T005 after T001
- T006–T009, T011–T017 after Setup (different files)
- T019–T021 (US1 tests)
- T027–T029 (US2 tests)
- T035–T038 (US3 tests) parallel with US2 tests after Foundational
- T044–T046, T051–T053, T058–T060, T066–T068, T073–T075, T081–T083, T087–T088
- T091–T092 in Polish

---

## Parallel Example: User Story 1

```bash
# After Foundational, launch US1 RED work in parallel:
Task: "Gherkin in features/auth/login-status-logout.feature"
Task: "RED tests in internal/app/auth/login_test.go"
Task: "RED CLI tests in internal/adapters/cli/auth_test.go"
```

## Parallel Example: User Story 2 vs 3

```bash
# After Foundational+US1, mail and teams read stories can proceed on different packages:
Task: "US2 Gherkin features/mail/list-get-thread.feature"
Task: "US3 Gherkin features/teams/list-get-messages.feature"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Independent Test for US1 against fakes; `make unit`
5. Demo `auth login|status|logout` (login gesture faked in tests)

### Incremental Delivery

1. Setup + Foundational → foundation ready
2. US1 → MVP auth
3. US2 → first useful workload (mail read)
4. US3 → teams read (no mail imports)
5. US4–US8 → send, attach, save, watch
6. US9–US10 → consent isolation and help
7. Polish → `make verify` on fakes

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Then: Developer A US1/US2/US4/US6-mail; Developer B US3/US5/US8; attach/save/consent/help integrate at boundaries without mail↔teams imports

---

## Task notes

- APS SHA: `accaa33d503340c56513ef387258f8da929ba902` (T003/T010).
- EX-I-001 (Principle I): T097–T119 US1–US10 RED/GREEN git checkpoints mixed in the first implement; history cannot be split without a rewrite. Scope is those git tasks only. Retirement: T124–T126 and later slices MUST use a RED commit then a GREEN commit. MUST NOT be cited as precedent. Human approval 2026-09-16 (analyze remediation C1).
- T099: Keychain adapter in `internal/adapters/keychain/keyring.go`; unit tests use Fake. Live Keychain-plus-0600-fallback is T126.
- T124: FR-042 mention-reason watch events in `httpwatch.go`; `TestHTTPWatchMentionReason`.
- T125: FR-032/FR-053 `$expand=attachments` on mail list/thread; `TestHTTPMailListAndThreadAttachmentMetadata`.
- T126: `keychain.Fallback` Keychain then `0600` session.json; live `cmd/m365` wired; unit tests use stubs (no real Keychain).
- T032/T041: `httptest` via `NewFakeServer` in `mail_test.go` / `teams_test.go`.
- T100: `internal/adapters/cli/perf_list_test.go`.
- T101: `BrowserOpener` / `LoopbackStart` in `internal/app/auth/ports.go`; `StartLoopback` in `internal/adapters/graph/oauth_loopback.go`.

## Notes

- [P] = different files, no incomplete dependencies
- [US1]–[US10] map to spec US-001–US-010
- No MCP, calendar, OneDrive, `graph GET`, `--guard`, Jenkins, Docker, or Helm tasks
- Client ID / Tenant ID values never in source, tests, logs, or fixtures
- Commit RED separately from GREEN; refactor only on GREEN
