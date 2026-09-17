# Phase 1 Data Model: Calendar and Files

Domain types for `internal/domain`. No Graph SDK fields, URLs, HTTP statuses, or token strings. Mail and teams entities are unchanged (see `specs/001-m365-cli/data-model.md`).

## Session (additive)

**Existing fields**: signed in, session usable, account display name, mail consented, teams consented.

**Added fields**: calendar consented (bool), files consented (bool).

**Invariants**:
- Session MUST NOT contain tokens, codes, or client secrets.
- Signed out ⇒ usable false and all four consent flags false.
- Mail/teams flags keep their v1 meaning.

## Calendar

**Fields**: id, name, is default (bool).

**Rules**:
- Unknown id or name ⇒ not-found.
- List is the signed-in user's own calendars only.

## CalendarEvent

**Fields**: id, calendar id, subject, start, end, location, organizer (Person), attendees[] (Person), body text (optional).

**Rules**:
- Start and end are timezone-aware instants (RFC3339 or UTC). Start ≥ end ⇒ usage.
- Body present on get; list MAY omit body.
- List items are occurrences inside the requested window, not an unbounded series.
- Unknown id ⇒ not-found.

## EventWindow

**Fields**: start, end (inclusive instants), source (`default` or `flags`).

**Rules**:
- Default: now through now+7 days (clock injected).
- `--start`/`--end` override; both visible in output or help.
- Window is required on list; never implied as "all future".

## DriveRoot

**Fields**: id, name.

**Rules**: Describes the signed-in user's own drive root only.

## DriveItem

**Fields**: id, name, size (bytes), is folder (bool), last modified, parent id (optional), browser link (optional).

**Rules**:
- Metadata never includes file bytes.
- Unknown id ⇒ not-found.
- Folder list children are direct children only.

## ResultLimit

Unchanged from v1: applied top, count, opaque next page token.

**Defaults for this feature**:
- Calendar calendars list default top = 20.
- Calendar event list default top = 10.
- Files list default top = 20.
- Maximum top = 50. Values `<1` or `>50` are usage failures, not silent caps.

## DryRunWrite

**Fields**: verb, target identity, intended fields (subject/times/name/size/parent), `dry_run: true`.

**Rules**:
- Observing a dry-run MUST NOT create, update, delete, upload, or move a remote item.
- MUST NOT start an upload session or write a local file.

## OutboundFile (upload)

**Fields**: local path, display name, size.

**Rules**:
- Path MUST be a local file (not a URL).
- Size MUST be `1..104857600` (100 MiB). Zero-byte files are invalid.
- One file per upload.

## Validation summary

| Rule | Failure class |
| --- | --- |
| `--top` out of range | usage (3) |
| Missing required flags/ids/path/name/subject/start/end | usage (3) |
| Start ≥ end, missing timezone | usage (3) |
| Upload missing/unreadable/empty/URL/oversize | usage (3) |
| Download path omitted, unwritable, exists without `--overwrite` | usage (3) |
| Move with neither new parent nor new name | usage (3) |
| No session, expired session, missing namespace consent | auth (4) |
| Microsoft 365 throttle or other service rejection | service (5) |
| Unknown calendar, event, folder, or item | not-found (6) |
