# Phase 0 Research: Calendar and Files Namespaces

## Decision: New packages, not a rewrite of mail/teams

**Rationale**: Spec FR-004 and constitution VII. Calendar and files are later namespaces with their own specification. Registration happens at the CLI seam (`run.go`). Mail and teams packages stay independent.

**Alternatives considered**:
- Fold calendar into mail because both are Outlook: rejected; constitution forbids namespace packages importing one another and the spec forbids changing mail.
- One `internal/app/graphworkloads` bag: rejected; smallest sufficient design and isolation tests.

## Decision: Event list uses calendarView with an explicit window

**Rationale**: Spec FR-016 — list MUST show occurrences in the requested window only and MUST NOT dump an entire recurring series. Microsoft Graph `calendarView` returns occurrences in `startDateTime`/`endDateTime`. Default window: injected clock `now` through `now+7d` (spec Assumptions). `--start`/`--end` override. Times are RFC3339 with timezone or UTC.

**Alternatives considered**:
- `GET /me/events` without a window: returns series masters and unbounded follow-up; violates FR-016.
- Client-side expansion of `recurrence`: extra logic, easy to dump the series.

**Sources**: Microsoft Graph calendarView (observed 2026-09-16): https://learn.microsoft.com/en-us/graph/api/calendar-list-calendarview

## Decision: Own calendars and own drive only

**Rationale**: Spec FR-011. Adapter calls `/me/calendars`, `/me/calendar/calendarView`, `/me/events`, `/me/drive`, `/me/drive/items`. PKCE adds `Calendars.ReadWrite` and `Files.ReadWrite` only. Do not add `Calendars.ReadWrite.Shared` or `Files.ReadWrite.All`. Do not call `/users/{id}/calendars`, `/me/calendars/{id}/getSchedule` as a product, `/sites`, or `/drives/{other}`.

**Alternatives considered**:
- Request Shared/All "because Entra already has them": rejected; spec forbids calling or requiring them.

**Sources**: Graph application permissions vs delegated `/me` (observed 2026-09-16).

## Decision: Independent consent flags on Session

**Rationale**: Spec FR-009/FR-010. Extend `domain.Session` and `auth.Blob` with `CalendarConsented` and `FilesConsented`. Login parses granted scopes. `auth status` JSON `namespaces` gains `calendar` and `files` keys; `mail` and `teams` remain. Missing calendar consent → exit 4 on calendar verbs only.

**Alternatives considered**:
- Infer calendar from mail consent: rejected; independent consent is a P3 story and a success criterion.

## Decision: Fake tokens stay additive

**Rationale**: v1 `fake-both` is mail+teams. Changing it would break 001 isolation. New tokens: `fake-calendar`, `fake-files`, `fake-all` (all four). Calendar routes without calendar consent → 403 mapped to exit 4. Same for files.

**Alternatives considered**:
- Reuse `fake-both` for calendar: rejected; hides missing-consent bugs and changes v1 fixtures.

## Decision: Upload session above 4 MiB

**Rationale**: Graph simple upload max is 4 MiB. Spec max is 100 MiB. Adapter: PUT content if size ≤ 4 MiB; otherwise createUploadSession and chunk. Use cases validate 1..104857600 bytes before the adapter. Dry-run stops in the use case (no HTTP). Zero-byte files are usage (3).

**Alternatives considered**:
- Cap uploads at 4 MiB: contradicts spec Assumptions (100 MiB).
- Always upload session: extra round-trips for small files; allowed later if tests stay green, not required.

**Sources**: Graph drive item upload (observed 2026-09-16): https://learn.microsoft.com/en-us/graph/api/driveitem-put-content https://learn.microsoft.com/en-us/graph/api/driveitem-createuploadsession

## Decision: Download reuses the fs port

**Rationale**: Spec FR-022 matches attachment save: user `--out`, refuse existing without `--overwrite`, no bytes on stdout. Reuse `internal/adapters/fs`.

**Alternatives considered**:
- Print bytes to stdout: forbidden by FR-027.
- Default path in cwd: forbidden; missing path is exit 3.

## Decision: Injected clock for default window

**Rationale**: Constitution IV — time MUST be injected. Calendar list default `now`..`now+7d` uses a `Now func() time.Time` in the calendar use case; tests freeze time.

**Alternatives considered**:
- `time.Now()` in the use case: flaky window tests.

## Decision: APS Gherkin beside v1 features

**Rationale**: Same acceptance pipeline as 001. New files under `features/calendar/` and `features/files/`. Product language (calendar, event, folder, dry-run), not HTTP codes.

**Alternatives considered**:
- Skip Gherkin for 002: violates constitution I/IV.
