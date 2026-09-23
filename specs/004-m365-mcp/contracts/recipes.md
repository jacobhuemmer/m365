# Recipes

Static text. No session. No Graph. No tokens, live mailbox content, or file bytes. Same body for `m365_help` `topic` and `prompts/get`.

Lookup topics: `mail-search`, `teams-find`, `calendar`, `files`. Write topics: `mail-write`, `teams-write`. Prompt descriptions for write topics are “Write recipe …”, not “Lookup recipe …”.

Teams find uses `teams find` / `send --to` (005).

## `mail-search`

MUST include:

- `mail list --folder all --search 'from:ajay'`
- `mail get`
- `mail thread`

MAY also name `subject:` and default `--top`. Then `m365_run` with namespace `mail` verb `list` flags `folder=all`, `search=from:ajay`.

## `teams-find`

MUST include:

- How to turn a person name into a 1:1 (`teams find Ajay`).
- How to turn a topic into a group (`teams find --group NOC`).
- Several matches MUST NOT send.
- Dry-run before notify (`write_opt_in` false / `--dry-run`).

MUST NOT tell the agent to invent a Graph people-search tool.

## `calendar`

MUST include:

- `calendar list`
- `calendar free` for tomorrow
- `calendar create --when 'tomorrow at 1:30 pm'` with dry-run

## `files`

MUST include:

- `files list`
- `files get`
- `files download` with a destination path (`--out`)
- dry-run upload

MUST NOT put file bytes in the recipe text.

## `mail-write`

MUST include:

- paragraphs via `<p>`, lists (`<ul>` / `<ol>`), and `<a href>`
- `mail send --html --body '…' --dry-run`
- `mail reply --html --body '…' --dry-run`
- MCP examples use `flags.body` string (not `body-file=-`)

MUST NOT:

- use a markdown heading as the mail body
- set `write_opt_in` true in the happy-path example

## `teams-write`

MUST include:

- `--html` with real HTML, or `--format md` with the documented subset
- `teams send --to Ajay --format md --text '…' --dry-run`
- MCP examples use `flags.text` (not `text-file=-`)

MUST NOT:

- post one run-on `--text` string with embedded markdown left unconverted
- mention `mentions[]`, Adaptive Cards, or Graph beta markdown
