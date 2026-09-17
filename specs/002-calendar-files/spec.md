# Feature Specification: Calendar and Files Namespaces

**Feature Branch**: `002-calendar-files`

**Created**: 2026-09-16

**Status**: Draft

**Input**: User description: "Add calendar and files (OneDrive) namespaces to the existing m365 CLI for one signed-in Microsoft 365 user — the same person and local agents who already use mail and teams. List and read first, bounded writes with dry-run, independent consent, no change to mail or teams."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List calendars and read events (Priority: P1)

Mason or a local agent, already able to sign in, lists the signed-in user's own calendars, lists upcoming events on a named calendar (default: the user's default calendar) inside a visible time window with a bounded page size, and gets one event by id (subject, start, end, location, attendees, body as text, organizer).

**Why this priority**: Read is the first useful calendar surface. Writes wait until listing and get are trustworthy.

**Independent Test**: With a usable session that includes calendar consent and synthetic calendar fixtures, list calendars, list events with the default window and default limit, list events on a named calendar, and get one event. Assert empty calendar success, unknown id exit `6`, the applied limit in the result, occurrences only inside the requested window, and no change to `mail list` or `teams list` names, flags, output shape, or exit classes.

**Acceptance Scenarios**:

1. **Given** a usable session with calendar consent and at least one own calendar, **When** the user runs `calendar calendars`, **Then** the command exits `0` and lists the user's own calendars, each with id and name, bounded by `--top`.
2. **Given** a default calendar with events this week, **When** the user runs `calendar list` with no calendar name, **Then** the command lists upcoming events on the default calendar inside the documented default time window, applies the default `--top`, and shows that window and limit.
3. **Given** a named own calendar, **When** the user runs `calendar list` for that calendar, **Then** events come from that calendar only, not from other calendars.
4. **Given** a known event id, **When** the user runs `calendar get`, **Then** stdout includes subject, start, end, location, attendees, body as text, and organizer, and does not print secrets.
5. **Given** an empty calendar in the requested window, **When** the user runs `calendar list`, **Then** the command exits `0` with an empty event list.
6. **Given** an unknown event id or unknown calendar, **When** the user runs `calendar get` or `calendar list` for that identity, **Then** the command exits `6`.

---

### User Story 2 - Browse OneDrive folders and item metadata (Priority: P1)

Mason or a local agent shows the signed-in user's own OneDrive root, lists children of a folder with a bounded `--top` and a next-page token when more items exist, and gets item metadata (name, id, size, folder vs file, last modified, browser link) without printing file bytes.

**Why this priority**: Folder navigation is the first useful files surface. Download and upload wait until list and get are trustworthy.

**Independent Test**: With a usable session that includes files consent and synthetic drive fixtures, show the root, list a folder at the default limit, follow `next_page` with `--page-token`, and get one file and one folder. Assert empty folder success, unknown id exit `6`, no file bytes on stdout or stderr, and no change to mail or teams attachment commands.

**Acceptance Scenarios**:

1. **Given** a usable session with files consent, **When** the user runs `files root`, **Then** the command exits `0` and describes the user's own drive root (id and name) without listing every descendant.
2. **Given** a folder with children, **When** the user runs `files list` (default: root folder), **Then** the command lists that folder's children bounded by `--top` and includes `next_page` when more children exist.
3. **Given** a previous `next_page` token, **When** the user runs `files list --page-token`, **Then** the command returns the next page and does not require a raw service URL.
4. **Given** a known file or folder id, **When** the user runs `files get`, **Then** stdout includes name, id, size, whether it is a folder or a file, last modified time, and a browser link, and includes no file bytes.
5. **Given** an empty folder, **When** the user runs `files list`, **Then** the command exits `0` with an empty item list.
6. **Given** an unknown item id, **When** the user runs `files get` or `files list` for that identity, **Then** the command exits `6`.

---

### User Story 3 - Create, update, and delete events with dry-run (Priority: P2)

Mason or a local agent creates, updates, and deletes an event on an own calendar. Every write supports `--dry-run`. Dry-run shows the intended change and does not mutate. Create and update accept subject, start, end, and optional location, body, and attendees. Delete names an event id.

**Why this priority**: Writes are valuable only after the user can see calendars and events. Dry-run is the confirmation rail.

**Independent Test**: With calendar consent and synthetic fixtures, dry-run create, real create, dry-run update, real update, dry-run delete, and real delete. Assert dry-run leaves the calendar unchanged, missing required fields exit `3`, unknown id exit `6`, and mail send/reply behavior is unchanged.

**Acceptance Scenarios**:

1. **Given** a usable calendar session, **When** the user runs `calendar create --dry-run` with subject, start, and end, **Then** stdout shows the intended event and no event is created.
2. **Given** the same inputs without `--dry-run`, **When** the user runs `calendar create`, **Then** an event is created and stdout reports its id.
3. **Given** a known event, **When** the user runs `calendar update --dry-run` and then `calendar update` without `--dry-run`, **Then** dry-run does not change the event and the real update does.
4. **Given** a known event, **When** the user runs `calendar delete --dry-run` and then `calendar delete` without `--dry-run`, **Then** dry-run does not delete and the real delete does.
5. **Given** missing subject, start, or end on create, **When** the user runs `calendar create`, **Then** the command exits `3` and does not create.
6. **Given** an unknown event id, **When** the user runs `calendar update` or `calendar delete`, **Then** the command exits `6` and does not mutate.

---

### User Story 4 - Download, upload, and organize OneDrive items with dry-run (Priority: P2)

Mason or a local agent downloads a named file to a user-supplied local path (no overwrite unless asked), uploads a local file into a named folder, creates a folder, deletes an item, and moves or renames an item. Writes that can destroy or create remote items support `--dry-run` and do not mutate on dry-run.

**Why this priority**: File transfer and organize follow browse. Download reuses the same destination discipline as mail and teams attachment save.

**Independent Test**: With files consent, synthetic files, and a temp directory, download to a named path, refuse an existing destination without overwrite, overwrite when asked, dry-run upload (no remote file), real upload, create-folder, dry-run delete, real delete, and move/rename. Assert no file bytes on stdout, missing destination path exits `3` and writes no file, and mail/teams save-attachment behavior is unchanged.

**Acceptance Scenarios**:

1. **Given** a known file and a destination path that does not exist, **When** the user runs `files download`, **Then** the file is written only to that path, stdout reports the path, and stdout contains no file bytes.
2. **Given** a destination that already exists, **When** the user runs `files download` without `--overwrite`, **Then** the command exits `3` and the existing file is unchanged.
3. **Given** `--overwrite`, **When** the user runs `files download` to an existing path, **Then** the file is replaced.
4. **Given** a local file and a named folder, **When** the user runs `files upload --dry-run`, **Then** stdout shows the intended name and size and no remote item is created.
5. **Given** the same upload without `--dry-run`, **When** the user runs `files upload`, **Then** a remote file is created in that folder.
6. **Given** a parent folder, **When** the user runs `files create-folder`, **Then** a folder is created and stdout reports its id.
7. **Given** a known item, **When** the user runs `files delete --dry-run` then `files delete`, **Then** dry-run does not delete and the real delete does.
8. **Given** a known item, **When** the user runs `files move` with a new parent and/or a new name, **Then** the item is moved or renamed as requested.
9. **Given** no destination path on download, **When** the user runs `files download`, **Then** the command exits `3` and writes no file anywhere.

---

### User Story 5 - Independent calendar and files consent (Priority: P3)

Mail, teams, calendar, and files consent are independent. `auth status` reports calendar and files the same way it already reports mail and teams. Missing calendar consent does not break files, mail, or teams, and vice versa. Shared or other-people's calendars and drives are out of this product even if the sign-in could reach them.

**Why this priority**: Isolation is required for a complete product, but P1 reads already deliver value when all consents are present.

**Independent Test**: With a session that has mail and teams consent but not calendar, calendar commands exit `4` and mail/teams commands still succeed. Repeat with files missing. With calendar consent only, calendar list succeeds and files commands exit `4`. `auth status` includes per-namespace consent for mail, teams, calendar, and files without printing secrets.

**Acceptance Scenarios**:

1. **Given** a usable session, **When** the user runs `auth status`, **Then** stdout reports signed in, session usable, and consent flags for mail, teams, calendar, and files, and does not print tokens.
2. **Given** mail and teams consent without calendar consent, **When** the user runs `calendar list`, **Then** the command exits `4`, not `5`, and `mail list` still succeeds.
3. **Given** files consent without calendar consent, **When** the user runs `files list`, **Then** the command succeeds and `calendar list` exits `4`.
4. **Given** only own-calendar consent, **When** the user lists calendars, **Then** only the signed-in user's own calendars appear; shared and delegate calendars are not listed as a successful product result.
5. **Given** only own-drive consent, **When** the user runs `files root` or `files list`, **Then** the command operates on the signed-in user's own OneDrive only, not another person's drive or a SharePoint library.

---

### User Story 6 - Discover calendar and files without changing mail or teams (Priority: P3)

`--help` names the new namespaces, verbs, limits, time window, dry-run, and download overwrite rules. Adding calendar and files does not change mail or teams command names, flags, output shape, or exit classes. `chat` remains a teams alias only.

**Why this priority**: Discovery and non-regression complete the feature after the workloads work.

**Independent Test**: Invoke top-level, `calendar`, and `files` `--help` with no session (exit `0`). Re-run `mail list --help` and `teams list --help` and assert the same names, flags, limits, and exit classes as before this feature.

**Acceptance Scenarios**:

1. **Given** no session, **When** the user runs `m365 --help`, **Then** help lists `auth`, `mail`, `teams`, `calendar`, and `files`, and still names JSON, human mode, and exit classes.
2. **Given** no session, **When** the user runs `calendar --help` or `files --help`, **Then** help names the verbs, `--top` defaults and maximum, dry-run on writes, and (for calendar list) the time window flags.
3. **Given** no session, **When** the user runs `mail list --help` and `teams list --help`, **Then** help text, flags, and limits match the existing mail and teams contract.
4. **Given** `chat list`, **When** the user runs it, **Then** it still aliases `teams list` and is not a files or calendar command.

---

### Edge Cases

- Expired or missing session on a calendar or files command versus `auth status` while signed out.
- Calendar consent without files consent, and the reverse; either without mail or teams.
- Empty calendar list, empty event window, empty folder.
- Unknown calendar, event, folder, or file identity.
- List with no `--top` (documented default, never unbounded).
- `--top` above the documented maximum (validation failure, not a silent cap).
- Event list over a recurring series: only occurrences that fall in the requested window; the command MUST NOT dump the entire series.
- Start after end, missing timezone on event times, or missing required create fields.
- Download with no destination path; destination exists without `--overwrite`; unwritable destination.
- Upload of a missing, unreadable, empty, or oversize local file.
- Dry-run versus real create, update, delete, upload, and move.
- Microsoft 365 service error or throttle while a session is usable.
- Extra organization-wide or shared-mailbox powers that happen to exist on the sign-in MUST NOT be exercised by these commands.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The user-visible command shape MUST remain `m365 <namespace> <verb> [flags]`. This feature MUST add workload namespaces `calendar` and `files`. `files` means the signed-in user's own OneDrive as a drive. `auth` remains shared core. `chat` remains a compatibility alias for `teams` only.
- **FR-002**: Calendar MUST provide: list calendars; list events; get event; create event; update event; delete event. Files MUST provide: show drive root; list folder children; get item metadata; download file; upload file; create folder; delete item; move or rename item. Verb names MAY be clarified in help; coverage MUST match.
- **FR-003**: Calendar and files commands MUST reuse the existing CLI contract: result data on stdout, diagnostics on stderr, JSON default, `--human` for people, `--json` for JSON, and exit classes `0` / `3` / `4` / `5` / `6` unchanged.
- **FR-004**: Adding these namespaces MUST NOT change mail or teams command names, flags, output shape, or exit classes. This feature MUST NOT rewrite the v1 mail/teams specification.
- **FR-005**: Every list command MUST apply an explicit ResultLimit. Omitting `--top` MUST apply the documented default. A `--top` above the maximum MUST fail with exit `3` and MUST NOT silently cap. When more results exist, stdout MUST include an opaque `next_page` token used with `--page-token`, not a raw service URL. The applied limit MUST be visible in output or `--help`.
- **FR-006**: Create, update, delete, upload, and move MUST support `--dry-run`. Dry-run MUST show the intended change and MUST NOT mutate. A real write is a separate invocation without `--dry-run`. The tool MUST NOT require a stored dry-run receipt.
- **FR-007**: Commands, including verbose or debug output, MUST NOT print access tokens, refresh tokens, authorization codes, or client secrets.
- **FR-008**: `--help` MUST name the namespace, required inputs, output modes, list limits, the event time window (calendar list), and dry-run on writes. Help MUST exit `0` without a session.
- **FR-009**: `auth status` MUST report per-namespace consent for mail, teams, calendar, and files in the same style as today's mail and teams flags. Existing mail and teams consent fields MUST remain.
- **FR-010**: Mail, teams, calendar, and files consent MUST be independent. Missing calendar consent MUST NOT break files, mail, or teams commands. Missing files consent MUST NOT break calendar, mail, or teams. A workload command without its consent MUST exit `4`, not `5`.
- **FR-011**: Calendar commands MUST operate only on the signed-in user's own calendars. Shared and delegate calendars are out of scope. Files commands MUST operate only on the signed-in user's own OneDrive. SharePoint libraries, other users' drives, and tenant-wide file search are out of scope. If the sign-in also has broader powers, these commands MUST NOT call them and MUST NOT require them.
- **FR-012**: `calendar calendars` MUST list the user's own calendars bounded by `--top`.
- **FR-013**: `calendar list` MUST list events on a named calendar, defaulting to the user's default calendar, inside a visible time window, bounded by `--top`. The window MUST appear in command output or `--help`. Default window is recorded in Assumptions.
- **FR-014**: `calendar get` MUST return one event as text: subject, start, end, location, attendees, body text, organizer.
- **FR-015**: An empty calendar list or empty event list MUST exit `0` with an empty list. An unknown calendar or event id MUST exit `6`. A missing required id or flag MUST exit `3`.
- **FR-016**: Event list MUST include only occurrences that fall in the requested window. It MUST NOT silently expand an entire recurring series beyond that window.
- **FR-017**: `calendar create` MUST accept subject, start, and end, and MAY accept location, body, and attendees. `calendar update` MUST accept an event id and at least one field to change. `calendar delete` MUST require an event id.
- **FR-018**: `files root` MUST describe the user's own drive root without dumping the whole tree.
- **FR-019**: `files list` MUST list children of a folder (default: root folder) bounded by `--top` and continued with `--page-token`.
- **FR-020**: `files get` MUST return item metadata: name, id, size, folder vs file, last modified, browser link. It MUST NOT print file bytes.
- **FR-021**: An empty folder list MUST exit `0` with an empty list. An unknown item or folder id MUST exit `6`. A missing required id or flag MUST exit `3`.
- **FR-022**: `files download` MUST require the parent item identity and a destination path. Success MUST write only to that path and MUST report that path on stdout. Omitting the path MUST exit `3` and write no file. If the destination exists, download MUST refuse with exit `3` unless `--overwrite` is present. `--overwrite` MUST replace the file.
- **FR-023**: `files upload` MUST accept one local file and a destination folder. Missing, unreadable, empty, or oversize files MUST exit `3` and MUST NOT create a remote item. `--dry-run` MUST list name and size and MUST NOT upload.
- **FR-024**: `files create-folder` MUST create a folder in a named parent (default: root) and report the new id.
- **FR-025**: `files delete` MUST require an item id and MUST support `--dry-run` before a real delete.
- **FR-026**: `files move` MUST move and/or rename an item using a new parent folder, a new name, or both. At least one of new parent or new name MUST be present or the command exits `3`.
- **FR-027**: File bytes MUST NOT be written to stdout or stderr on root, list, get, upload, or download. Download writes bytes only to the user-named path.
- **FR-028**: Tests and fixtures MUST use synthetic calendars, events, folders, and files. Live event bodies and live OneDrive file bytes MUST NOT be stored in the repository.
- **FR-029**: v1 of this feature MUST NOT include room or equipment booking, a scheduling-assistant product, conferencing as a product, category management, calendar sharing or permissions admin, group calendars, mailbox-wide free/busy search, sharing-link admin, SharePoint site libraries, version-history restore, Office co-authoring, thumbnail bytes, tenant-wide search, app-only credentials, passwords on the CLI, a generic Microsoft 365 escape hatch, a plugin marketplace, overnight monitors, kata wrappers, or MCP servers in this repository.

### Key Entities *(include if feature involves data)*

- **Session**: The saved sign-in of one Microsoft 365 user. Attributes: signed in, session usable, account display name, and consent flags for mail, teams, calendar, and files. MUST NOT include tokens, codes, or client secrets.
- **Calendar**: One of the signed-in user's own calendars. Attributes: id, name, and whether it is the default calendar.
- **CalendarEvent**: One event or one occurrence in a window. Attributes: id, calendar id, subject, start, end, location, attendees, organizer, body text when requested. Recurring series are represented as occurrences inside the requested window, not as an unbounded expansion.
- **DriveRoot**: The signed-in user's own OneDrive root. Attributes: id, name.
- **DriveItem**: A file or folder in that drive. Attributes: id, name, size, folder vs file, last modified, parent, browser link. MUST NOT include file bytes in command output.
- **CommandResult**: stdout payload, stderr diagnostics, exit class.
- **ResultLimit**: applied page size, count, whether more results exist, opaque next-page token (not a service URL).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After one interactive login succeeds, `auth status` reports a usable session and consent for calendar and for files (true or false as granted) without a second sign-in gesture.
- **SC-002**: With calendar consent, a default-limit `calendar list` for the default window returns at most that many events, shows the window and limit, and never dumps an entire mailbox calendar or an entire recurring series.
- **SC-003**: With files consent, a default-limit `files list` of a folder returns at most that many children and never dumps the entire drive.
- **SC-004**: Dry-run create of an event shows subject, start, and end and does not create an event the user can later get by a new id.
- **SC-005**: Dry-run upload shows the file name and size and does not create a remote item the user can later list.
- **SC-006**: After a dry-run create and a dry-run upload, the same commands without `--dry-run` perform the real write and return an id.
- **SC-007**: Mail and teams command names, flags, output shape, and exit classes remain unchanged. `chat` still aliases `teams` only.
- **SC-008**: With calendar consent and without files consent, calendar list succeeds and every files command exits with the auth class, not the service class. The reverse is true. Mail and teams keep working when only their own consent is present.
- **SC-009**: No command, including verbose or debug modes, prints an access token, refresh token, authorization code, or client secret.
- **SC-010**: For every calendar and files command, JSON stdout is exactly one parseable JSON value with no log or diagnostic lines mixed in.
- **SC-011**: An empty event window or empty folder exits success with an empty result. An unknown event or item id exits with the not-found class, distinct from auth and from service error.
- **SC-012**: Download writes only to the user-supplied path and reports that path. Omitting the path is a usage failure and writes no file. An existing destination without overwrite is refused and left unchanged.
- **SC-013**: After a usable session exists, a default-limit event list completes in under 15 seconds on an ordinary connection. A default-limit folder list completes in under 15 seconds on an ordinary connection.
- **SC-014**: No calendar or files command prints file bytes on the terminal or in JSON stdout.
- **SC-015**: Removing or never enabling calendar does not change files, mail, or teams command names, flags, output shape, or exit classes. Removing or never enabling files does not change calendar, mail, or teams.

## Assumptions

- This feature extends the existing m365 CLI. It does not replace v1 mail, teams, auth, or attachments.
- JSON remains the default stdout format. `--human` selects human-readable stdout.
- Exit classes remain `0` success; `3` usage, validation, and configuration; `4` authentication and consent; `5` Microsoft 365 service error; `6` not-found.
- Default `--top` is `20` for `calendar calendars` and `files list`, and `10` for `calendar list` (events). Maximum `--top` is `50`. Values above `50` or below `1` are validation errors, not silent caps.
- Default event window for `calendar list` is from now through seven days ahead. `--start` and `--end` override it. Both ends are inclusive of the instant given. The window MUST be visible in output or `--help`.
- Event start and end MUST include a timezone or a UTC instant. A start at or after end is a validation error.
- Pagination uses `--top` plus an opaque `--page-token` from `next_page`.
- Existing local destination files are refused unless `--overwrite` is present (same discipline as attachment save).
- Maximum upload size is 100 MiB per file. One local file per `files upload`. Zero-byte files are validation errors.
- `files list` without a folder identity lists the root folder. `files create-folder` without a parent creates in the root.
- `files move` accepts a new parent folder, a new display name, or both.
- Dry-run is the confirmation rail. The product does not persist a dry-run ticket that gates the next write.
- Calendar uses only the signed-in user's own calendars (the permission people already grant for their own calendar). Shared and delegate calendars stay out even if the sign-in also has them.
- Files uses only the signed-in user's own OneDrive (the permission people already grant for their own drive). SharePoint libraries, other users' drives, and organization-wide file access stay out even if the sign-in also has them.
- Client ID, Tenant ID, and secrets MUST NOT appear in this specification, source, tests, logs, or fixtures.
- Overnight incident monitor, kata wrappers, and MCP servers stay outside this repository.
