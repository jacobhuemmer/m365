# CLI command catalog (calendar and files)

Extends v1. Public interface remains `m365 <namespace> <verb> [flags]`. Mail, teams, and `chat` are unchanged; see `specs/001-m365-cli/contracts/command-catalog.md`.

Global flags and exit classes are unchanged. JSON default; `--human`; `--json` and `--human` together → 3. `--help` exits 0 without a session.

Upload cap (help MUST show): 100 MiB per file, one file per upload.

Event window (help MUST show): default now through +7 days; `--start` `--end`; `--top` default 10 (events) / 20 (calendars list and files list); max 50.

## Core (additive)

| Command | Change |
| --- | --- |
| `m365 --help` | Also lists `calendar` and `files`. Mail/teams lines stay. |
| `auth status` | `namespaces` also has `calendar` and `files` booleans. `mail` and `teams` remain. |

## calendar

| Command | Args / flags | Defaults |
| --- | --- | --- |
| `calendar calendars` | `--top`, `--page-token` | top=`20` |
| `calendar list` | `--calendar`, `--start`, `--end`, `--top`, `--page-token` | default calendar; window now→+7d; top=`10` |
| `calendar get` | `EVENT_ID` | body as text |
| `calendar create` | `--subject`, `--start`, `--end`, `--location`, `--body`/`--body-file`, `--attendee` (repeat), `--calendar`, `--dry-run` | default calendar |
| `calendar update` | `EVENT_ID`, same optional fields as create, `--dry-run` | at least one field required |
| `calendar delete` | `EVENT_ID`, `--dry-run` | |

Unknown calendar or event → 6. Empty list → 0. Create missing subject/start/end → 3.

## files

| Command | Args / flags | Defaults |
| --- | --- | --- |
| `files root` | | own drive root metadata |
| `files list` | `--folder`, `--top`, `--page-token` | folder=`root`; top=`20` |
| `files get` | `ITEM_ID` | metadata only, no bytes |
| `files download` | `ITEM_ID`, `--out`, `--overwrite` | refuse existing file |
| `files upload` | `--file`, `--folder`, `--dry-run` | folder=`root` |
| `files create-folder` | `--name`, `--folder`, `--dry-run` | folder=`root` |
| `files delete` | `ITEM_ID`, `--dry-run` | |
| `files move` | `ITEM_ID`, `--folder`, `--name`, `--dry-run` | at least one of `--folder` or `--name` |

Unknown item → 6. Empty folder → 0. Missing `--out` on download → 3, no write. `--file` may be a local path only.

## Command states

Every new verb MUST define happy-path, empty (where a list), validation, auth (4), service (5), and not-found (6) when an identity is required. Dry-run is validation-success with `dry_run: true` and no mutation.
