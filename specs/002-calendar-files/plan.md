# Implementation Plan: Calendar and Files Namespaces

**Branch**: `main` (setup-plan JSON reported `002-calendar-files`; working tree is `main`) | **Date**: 2026-09-16 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-calendar-files/spec.md`

**Note**: This plan is design only. It does not add application source, `tasks.md`, MCP, or CI workflow files. It MUST NOT rewrite `specs/001-m365-cli/`.

## Summary

Add `calendar` and `files` namespaces to the existing Go 1.25+ `m365` CLI. `cmd/m365` stays wiring only. New domain types and use cases live in `internal/app/calendar` and `internal/app/files`; those packages MUST NOT import `mail` or `teams`, and mail/teams MUST NOT import them. A thin Graph HTTP adapter maps `/me/calendars`, calendarView, `/me/drive` to spec entities. Auth stays delegated PKCE; login requests calendar and files scopes in addition to mail and teams; consent flags are independent. Tests extend the in-process fake Graph. Acceptance adds Gherkin under `features/calendar/` and `features/files/`. Mail and teams command names, flags, output shape, and exit classes MUST NOT change. `chat` remains a teams alias only.

## Technical Context

**Language/Version**: Go 1.25+ module `github.com/masonhuemmer/m365` (already exists).

**Primary Dependencies**: Same as v1: stdlib `flag`, `net/http`, `encoding/json`, `os`, `testing`, `httptest`; `golang.org/x/oauth2` PKCE in the auth adapter only; `github.com/zalando/go-keyring` with `0600` file fallback. No Cobra, no Microsoft Graph SDK, no kiota, no MSAL.

**Storage**: Unchanged token store (Keychain then `0600` session.json). No new secret files. Download writes only to the user-named path via the existing fs port.

**Testing**: Go `testing`; `httptest` fake Graph; APS Gherkin → generated tests. Unit tests MUST NOT call live Graph. New fake tokens for calendar/files isolation; `fake-both` keeps mail+teams only so v1 tests stay valid.

**Target Platform**: macOS CLI for one signed-in user or a local agent.

**Project Type**: Same single-module local CLI.

**Performance Goals**: SC-013 — default-limit `calendar list` and `files list` finish in under 15 seconds after a usable session. CI measures against the fake Graph (15s deadline). Live timing is optional manual.

**Constraints**: Exit classes `0/3/4/5/6` unchanged; JSON stdout default; no secrets in output, fixtures, logs, or this plan; independent consent across mail, teams, calendar, files; calendar `--top` default 10 (events) / 20 (calendars), files list default 20, max 50; upload max 100 MiB; no file bytes on stdout; own calendars and own OneDrive only; no Shared/All graph capabilities; no MCP, SharePoint libraries, plugin marketplace, overnight monitor.

**Scale/Scope**: Two new namespaces, about 14 verbs, plus additive `auth status` consent keys. Registration seam in CLI `run.go` and core session only.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Test-Driven Development (NON-NEGOTIABLE). PASS.** Every calendar/files slice starts with RED (Gherkin and/or unit/CLI). RED is a separate git commit of failing tests and fixtures only. GREEN is the smallest production change. EX-I-001 does not apply here; this feature MUST use separate RED then GREEN commits.
- [x] **II. Clean Code. PASS.** New files SHOULD stay under 250 lines; over 250 needs a task note; over 500 MUST split. `cmd/m365` remains wiring. Split the Graph fake if calendar/files routes would push `fake.go` over 250.
- [x] **III. Smallest Sufficient Design. PASS.** Calendar and files are specified (002); they are no longer forbidden pre-builds. No plugin marketplace, no SharePoint, no generic Graph shell. Three similar lines beat a premature calendar+files abstraction. Do not merge calendar and files into one package.
- [x] **IV. Testing Standards. PASS.** Ports/fakes; no live Graph in unit tests. Fake Graph for mapping. Acceptance Gherkin → APS → `acceptance/generated`, not hand-edited. Inject clock for the default 7-day window. Synthetic fixtures only.
- [x] **V. CLI Experience Consistency. PASS.** Reuse v1 streams, JSON/human, exit classes, `--help`, dry-run, visible limits. New namespaces follow the same flag patterns unless this plan records a reason (none).
- [x] **VI. Performance. PASS.** SC-013: metric = wall-clock for default-limit event list and default-limit folder list; threshold = 15s; CI = test against fake Graph. List calls always send `$top`. Calendar list uses a time window (calendarView), never an unbounded series dump.
- [x] **VII. Domain Isolation. PASS.** `internal/app/calendar` and `internal/app/files` MUST NOT import each other or `mail`/`teams`. Mail and teams MUST NOT import calendar/files. Graph URLs stay in `internal/adapters/graph`. CLI `run.go` registers namespaces (seam). `chat` stays teams-only. Isolation tests in each new package.
- [x] **VIII. Secrets, Auth, Least Privilege. PASS.** Delegated PKCE only. Request `Calendars.ReadWrite` and `Files.ReadWrite` in addition to existing mail/teams scopes. Do not request or call `Calendars.ReadWrite.Shared` or `Files.ReadWrite.All`. Consent flags independent. No Client ID/Tenant ID/tokens in this plan, research, contracts, tests, or fixtures. Live event bodies and OneDrive bytes stay out of the repo.
- [x] **Engineering Constraints. PASS.** Same Makefile gates. No surrounding refactors in bug fixes. Top-N and window visible.
- [x] **Development Workflow. PASS.** Specify (done) → this Constitution Check → RED commit → GREEN commit → optional REFACTOR → review → `make verify`.

No new constitution exceptions. EX-I-001 remains historical for 001 T097–T119 only.

**Post-design re-check (after Phase 1): PASS.** `research.md`, `data-model.md`, `contracts/`, and `quickstart.md` keep Graph at adapters, domain free of SDK fields, fake Graph as CI boundary, own-calendar/own-drive only, and additive auth consent keys. Mail/teams catalogs unchanged.

## Project Structure

### Documentation (this feature)

```text
specs/002-calendar-files/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── command-catalog.md
│   ├── fake-graph.md
│   ├── json-event.schema.json
│   └── json-drive-item.schema.json
└── tasks.md                 # Future /speckit-tasks, not created by this plan
```

### Source Code (repository root)

```text
.
├── cmd/m365/                    # wiring: register calendar + files stores
├── internal/
│   ├── domain/                  # Calendar, CalendarEvent, DriveRoot, DriveItem
│   ├── app/
│   │   ├── auth/                # Session + Blob consent flags (core, additive)
│   │   ├── calendar/            # use cases + ports; MUST NOT import mail/teams/files
│   │   ├── files/               # use cases + ports; MUST NOT import mail/teams/calendar
│   │   ├── mail/                # unchanged
│   │   └── teams/               # unchanged
│   └── adapters/
│       ├── cli/                 # calendar.go, files.go; run.go seam; help; auth status keys
│       ├── graph/               # httpcalendar.go, httpfiles.go; fake calendar/files routes
│       └── fs/                  # reuse download/upload local paths
├── features/calendar/
├── features/files/
└── testdata/                    # synthetic files only (reuse note.txt)
```

**Structure Decision**: Same module. New packages beside mail/teams. Do not edit `internal/adapters/cli/mail.go` or `teams.go`. Allowed core edits: `run.go` registration, `help.go` root text, `auth` status namespaces, `domain.Session` / `auth.Blob`, OAuth scope list, fake Graph routes behind new tokens.

## Implementation Workflow

1. Gherkin under `features/calendar/` and `features/files/`.
2. APS generate; focused RED unit/CLI tests; confirm expected failure.
3. RED commit (failing tests + fixtures only).
4. Smallest GREEN; GREEN commit; no locked RED-test edits.
5. Isolation tests that fail if calendar/files import mail/teams or each other.
6. `make verify` on fakes. Do not run live Graph in CI.

## Graph mapping (adapter only)

Documented so implementers do not invent URLs in domain code. Domain speaks calendars, events, folders, items.

- Calendars: `GET /me/calendars` with `$top`. Never `/users/{other}/calendars`.
- Event list: `GET /me/calendar/calendarView` or `GET /me/calendars/{id}/calendarView` with `startDateTime` and `endDateTime`. Do not list `/me/events` without a window (that would expand series).
- Event get/create/update/delete: `/me/events/{id}` and `POST /me/calendars/{id}/events`. `Prefer: outlook.body-content-type="text"`.
- Drive: `GET /me/drive`, `GET /me/drive/root`, `GET /me/drive/items/{id}`, `GET /me/drive/items/{id}/children`. Never `/drives/{other}`, never `/sites`.
- Download: `GET .../content` into the fs port only.
- Upload ≤4 MiB: `PUT /me/drive/items/{parent}:/{name}:/content`. Upload 4 MiB–100 MiB: createUploadSession then chunked PUT in the adapter. Dry-run MUST NOT start a session or PUT bytes.

Opaque `next_page` encoding stays the v1 `p.` + base64url of `@odata.nextLink`, only when the host is Graph or the fake.

## Make Targets

Reuse the existing Makefile. Add a CLI perf test for calendar list and files list under 15s against the fake (same pattern as T100).

## Files To Add At Implement Time (not this command)

- Packages listed above.
- Gherkin features and step stubs.
- Additive auth JSON keys; do not remove mail/teams keys.
- Optional additive properties on `specs/001-m365-cli/contracts/json-stdout.schema.json` `namespaces` (`calendar`, `files`) — the only permitted 001-file edit, because auth is core. Do not change mail/teams command schemas.

## Out of This Plan

- Rewriting 001 spec, plan, mail.go, or teams.go.
- `/speckit-tasks` and `/speckit-implement`.
- Live calendar bodies or OneDrive bytes in git.
- MCP, SharePoint, shared calendars, Files.ReadWrite.All, room booking, `--guard`.

## Risks And Tradeoffs

- **calendarView vs events**: calendarView is required so list stays inside the window (FR-016).
- **Upload session**: Graph simple PUT caps at 4 MiB; spec allows 100 MiB, so the adapter MUST implement upload session for larger files. Dry-run never creates a session.
- **Additive auth JSON**: v1 tests that only check `signed_in` stay green; any test that exact-matches `namespaces` with two keys MUST be updated in this feature (core, not a mail/teams change).
- **fake-both**: MUST remain mail+teams only so 001 isolation fixtures do not silently gain calendar consent.

## Complexity Tracking

No new constitution exceptions. EX-I-001 is not used for this feature.
