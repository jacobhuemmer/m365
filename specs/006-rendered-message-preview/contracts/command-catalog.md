# Command Catalog Changes

Only the three commands below change. All other verbs, including `--dry-run` on the other eight write verbs, are unchanged.

## Flags

| Command | New / changed flags |
|---------|--------------------|
| `mail send` | `--format md` (new), `--preview` (new). Plain body now delivered as HTML |
| `mail reply` | `--format md` (new), `--preview` (new) |
| `teams send` (`chat send`) | `--preview` (new). `--format` now rejects values other than `md` |

`--preview` is a boolean flag (added to `boolFlags`). It implies `--dry-run`.

## Usage errors (class `usage`, exit 3)

| Condition | Message |
|-----------|---------|
| `--format` value other than `md` | `unsupported --format "<v>"; only md` |
| `--html` with `--format md` | `use --html or --format md, not both` |
| `--preview` with `--json` | `use only one of --preview or --json` |
| `preview` flag through MCP `m365_run` | `preview is a terminal flag`, hint `use dry-run; its JSON has rendered and format_problems` |
| Body renders to nothing (whitespace only) | same message as the existing missing-body error for that command |
| Real send with format problems | see [format-rules.md](format-rules.md#blocked-send-error) |

## Outputs

| Invocation | Stdout | Exit |
|------------|--------|------|
| `--dry-run` (JSON default or `--json`) | Existing dry-run object plus `rendered` and `format_problems` ([schema](json-dry-run.schema.json)) | 0 |
| `--dry-run --human` | Same object as indented JSON (unchanged `--human` behaviour) | 0 |
| `--preview` (with or without `--human`, `--dry-run`) | Preview box ([preview.md](preview.md)) | 0 |
| Real send, clean | Existing `{"id": …, "sent": true}` | 0 |
| Real send, problems | Nothing on stdout; error JSON on stderr | 3 |

## Help

`--help` for `mail send`, `mail reply` and `teams send` lists `--format md` (markdown subset: headings, bold, lists, links, inline code, fenced code) and `--preview` (show the message as text; never sends; not with `--json`). Help exits 0 without a session.

## Delivered body (Graph adapter)

| Command | Graph field | Value |
|---------|-------------|-------|
| `mail send` | `message.body` | `{"contentType": "HTML", "content": rendered.content}` |
| `mail reply` | `comment` | `rendered.content` |
| `teams send` | `body` | `{"contentType": "html", "content": rendered.content}` |
