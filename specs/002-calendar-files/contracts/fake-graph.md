# Fake Graph fixture contract (calendar and files)

Extends `specs/001-m365-cli/contracts/fake-graph.md`. Same `httptest` server. No live tenant. No live event bodies or OneDrive bytes in the repository.

## Identity

v1 tokens keep v1 meaning. `fake-both` is **mail + teams only** (no calendar, no files).

| Token | Mail | Teams | Calendar | Files |
| --- | --- | --- | --- | --- |
| `fake-both` | yes | yes | no | no |
| `fake-mail` | yes | no | no | no |
| `fake-teams` | no | yes | no | no |
| `fake-calendar` | no | no | yes | no |
| `fake-files` | no | no | no | yes |
| `fake-all` | yes | yes | yes | yes |
| `fake-expired` | no | no | no | no |

Calendar routes without calendar consent → 403 → exit 4. Files routes without files consent → 403 → exit 4.

## Resources tests MAY assume

Synthetic, stable ids:

- Calendar `cal-1` (default, name `Calendar`), `cal-2` (name `Work`).
- Event `ev-1` on `cal-1` inside the frozen default window: subject `Standup`, start/end RFC3339, organizer `user@example.com`.
- Recurring series `ev-series` with **one** occurrence `ev-occ-1` inside the default window; listing the default window MUST NOT return extra occurrences outside it.
- Drive root `root` / item `folder-1` (folder), file `file-1` name `note.txt`, size 12, bytes `synthetic-ok` (same as v1 testdata).
- Empty calendar window and empty folder fixtures.
- Unknown ids → not-found (exit 6).
- Throttle → service (exit 5).

## Behaviors the fake MUST implement

- Honor `$top` and calendarView `startDateTime`/`endDateTime`.
- Event list returns occurrences in the window only.
- `Prefer` text body on event get.
- Create/update/delete/upload/move without dry-run mutate the in-memory fake; dry-run never reaches those HTTP methods (use cases short-circuit).
- Download content is used only by `files download`; tests assert the destination file, not stdout bytes.
- Do not serve `/users/{other}/calendars`, `/sites`, or other-drive URLs as success.

## Forbidden

- Recording live calendar or OneDrive content into `testdata/`.
- Embedding Client ID, Tenant ID, or real tokens.
- Changing `fake-both` to grant calendar or files.
