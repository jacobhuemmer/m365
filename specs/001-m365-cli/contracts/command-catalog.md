# CLI command catalog

Public interface: `m365 <namespace> <verb> [flags]`. `auth` is core. `chat` is an alias of `teams` (same verbs and flags). JSON is default stdout. `--human` selects human-readable stdout. `--json` selects JSON. `--json` and `--human` together → exit `3`. `--help` exits `0` without a session.

Global flags (every command): `--json`, `--human`, `--help`. Optional `--verbose` / `--debug` MUST redact secrets.

Exit classes: `0` success; `3` usage/validation/config; `4` auth/consent; `5` service; `6` not-found.

Attachment caps (help MUST show): 10 MiB per file, 10 files per send/reply.

## Core

| Command | Args / flags | Default | Empty | Notes |
| --- | --- | --- | --- | --- |
| `m365 --help` | | | n/a | Lists auth, mail, teams; output modes; exit classes |
| `auth login` | | | n/a | Interactive browser; missing config → 3 |
| `auth status` | | | signed out → 0, `signed_in=false` | No secrets |
| `auth logout` | | | already signed out → 0 | Deletes Keychain item and `0600` session file |

## mail

| Command | Args / flags | Defaults |
| --- | --- | --- |
| `mail list` | `--folder`, `--unread`, `--search`, `--top`, `--page-token` | folder=`inbox`; top=`10` |
| `mail get` | `MESSAGE_ID` | body as text |
| `mail thread` | `MESSAGE_ID`, `--bodies` | oldest-first; bodies off |
| `mail send` | `--to` (repeat), `--subject`, `--body` / `--body-file`, `--cc` (repeat), `--attach` (repeat), `--html`, `--dry-run` | |
| `mail reply` | `MESSAGE_ID`, `--body` / `--body-file`, `--all`, `--attach`, `--html`, `--dry-run` | reply sender only |
| `mail attachments` | `MESSAGE_ID` | metadata list |
| `mail save-attachment` | `MESSAGE_ID`, `ATTACHMENT`, `--out`, `--overwrite` | refuse existing file |

Well-known folders: `inbox`, `sentitems`, `drafts`, `all`. Unknown folder → 6. Max `--top` 50.

## teams (`chat` alias)

| Command | Args / flags | Defaults |
| --- | --- | --- |
| `teams list` | `--top`, `--page-token` | top=`20` |
| `teams get` | `CHAT_ID` | members included |
| `teams messages` | `CHAT_ID`, `--top`, `--page-token`, `--include-system` | top=`20`; system off |
| `teams send` | `CHAT_ID`, `--text` / `--text-file`, `--attach`, `--html`, `--format md`, `--dry-run` | |
| `teams watch` | `--since`, `--chat` (repeat), `--top` | top=`20`; JSON lines |
| `teams attachments` | `CHAT_ID`, `MESSAGE_ID` | metadata list |
| `teams save-attachment` | `CHAT_ID`, `MESSAGE_ID`, `ATTACHMENT`, `--out`, `--overwrite` | refuse existing file |

`--body-file` / `--text-file` may be `-` (stdin).

Per-command empty/validation/auth/service/not-found: [cli-contract.md](../cli-contract.md#command-states).
