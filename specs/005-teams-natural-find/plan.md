# Implementation Plan: Natural Teams Find and DM

**Branch**: `main` (setup-plan JSON reported `005-teams-natural-find`; working tree is `main`) | **Date**: 2026-09-16 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/005-teams-natural-find/spec.md`

**Note**: Design only. No application source, `tasks.md`, MCP, or CI files. MUST NOT rewrite `specs/001-m365-cli/` except additive teams help/catalog notes at implement time. MUST NOT edit mail, calendar, files, or MCP packages.

## Summary

Add `teams find` and `teams send --to <query>` on the existing Teams namespace. Person queries resolve 1:1 chats by member name/email across paged `/me/chats` (not one list page). `--group` (or `group with X`) resolves group chats by topic/members. Several matches never send. Zero find matches are empty success; zero `--to` matches are not-found and do not create a chat. Graph stays in the adapter; match/rank in `internal/app/teams`. Fake Graph seeds Ajay 1:1 off the first list page. Mail/calendar/files/MCP unchanged.

## Technical Context

**Language/Version**: Go 1.25+ module `github.com/masonhuemmer/m365` (exists).

**Primary Dependencies**: Stdlib (`strings`, `regexp` for the closed `group with` phrase). Existing Graph HTTP adapter, `flag`, `testing`, `httptest`. No NLP, no people-directory SDK, no Graph SDK, no Cobra.

**Storage**: Unchanged token store. No new files.

**Testing**: Go `testing`; fake Graph with synthetic chats/members; APS Gherkin under `features/teams/`. No live Graph or live member lists in unit tests/repo.

**Target Platform**: macOS CLI, one signed-in user or local agent.

**Project Type**: Same single-module local CLI.

**Performance Goals**: SC-008 — unique `teams find` under 15s after a usable session (CI: fake Graph, 15s deadline). Stop paging once unique 1:1 is found.

**Constraints**: Exit classes `0/3/4/5/6`; JSON default; no secrets; find `--top` default 10 max 20; scan ceiling 10×50 chats; no chat create; no channels; no `/users` search; `chat` alias works; send-by-id unchanged.

**Scale/Scope**: One new verb (`find`); additive `--to` and `--group` on existing send/find. Closed intent grammar.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. TDD. PASS.** RED Gherkin + unit/CLI; RED commit then GREEN. EX-I-001 does not apply.
- [x] **II. Clean Code. PASS.** Find/match in `find.go` under 250 lines; CLI wiring in `teams.go` (split if over 250).
- [x] **III. Smallest Sufficient Design. PASS.** Closed `--group` / `group with` grammar. No people directory, no auto-create 1:1, no MCP recipe edit in this feature.
- [x] **IV. Testing. PASS.** Fake chats; Ajay 1:1 not on first list page. No live bodies.
- [x] **V. CLI Consistency. PASS.** Same streams, JSON/human, exits, `--help`, dry-run. New flags documented. Find empty = 0; send `--to` none = 6.
- [x] **VI. Performance. PASS.** SC-008: wall-clock unique find; 15s; CI fake Graph. Graph list always `$top=50` per page; ceiling 500; no silent full dump.
- [x] **VII. Isolation. PASS.** Logic in `internal/app/teams` only. MUST NOT import mail/calendar/files. Graph URLs in adapter. `chat` alias unchanged.
- [x] **VIII. Secrets. PASS.** Same delegated Teams scopes. Synthetic names only (`Ajay Kumar`, not live tenants).
- [x] **Engineering Constraints. PASS.** Existing Makefile. Visible `--top` and incomplete flag.
- [x] **Workflow. PASS.** Specify (done) → this check → RED → GREEN → `make verify`.

No new constitution exceptions.

**Post-design re-check (after Phase 1): PASS.** Find pages chats with members; match in app layer; contracts lock find/`--to`; fake Graph holds off-page 1:1; mail/calendar/files/MCP catalogs unchanged.

## Project Structure

### Documentation (this feature)

```text
specs/005-teams-natural-find/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── command-catalog.md
│   ├── fake-graph.md
│   ├── json-find.schema.json
│   └── json-send-to.schema.json
└── tasks.md                 # /speckit-tasks, not this command
```

### Source Code (implement time)

```text
internal/app/teams/find.go          # Find + match/rank; scan via ListChats pages
internal/app/teams/find_test.go
internal/app/teams/ports.go         # SendInput.To; Find use case
internal/domain/result.go           # DefaultFindTop=10, MaxFindTop=20 (or find-local)
internal/adapters/cli/teams.go      # find verb; send --to --group
internal/adapters/graph/httpteams.go  # $expand=members when listing for find
internal/adapters/graph/memory.go   # extra synthetic chats
features/teams/find.feature
features/teams/send-to.feature
```

**Structure Decision**: Same module, teams package only.

## Graph mapping (adapter only)

- Scan: `GET /me/chats?$expand=members&$top=50` + `@odata.nextLink` up to 10 pages.
- Never `/users`, `/me/people`, `/teams`, channel paths.
- Never POST chat to create a 1:1.

## Make Targets

Reuse Makefile. CLI perf test: unique find against fake under 15s.

## Out of This Plan

- MCP recipe update (004).
- Creating 1:1s, channels, kata wrappers.
- Rewriting 001 teams send-by-id.

## Risks And Tradeoffs

- **NOC without `--group`**: person query; help must show `--group NOC`.
- **Incomplete vs empty**: ceiling + zero matches → `incomplete: true`, not “no chat.”
- **NormalizeTop max 50 vs find max 20**: find uses its own 1..20 check so list stays 50.

## Complexity Tracking

No new constitution exceptions.
