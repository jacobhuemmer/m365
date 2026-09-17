# Lookup recipes

Static text. No session. No Graph. No tokens, live mailbox content, or file bytes. Same body for `m365_help` `topic` and `prompts/get`.

Until 005 lands, Teams uses `teams list`, not `teams find`.

## `mail-search`

MUST include:

- `mail list --folder all --search 'from:ajay'`
- `mail get`
- `mail thread`

MAY also name `subject:` and default `--top`. Then `m365_run` with namespace `mail` verb `list` flags `folder=all`, `search=from:ajay`.

## `teams-find`

MUST include:

- How to turn a person name into a 1:1 (today: `teams list`, match members; later: `teams find` when 005 exists).
- How to turn a topic into a group.
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
