# MCP catalog

Public interface: `m365 mcp serve` on stdio JSON-RPC. Human CLI remains `m365 <namespace> <verb> [flags]` as in 001–003. This catalog is the MCP door, not a Graph API dump.

`tools/list` MUST return exactly these three names (SC-001).

Transport: stdio only. No SSE, no Streamable HTTP.

## Tools

| Name | Description (MUST convey) | Session |
| --- | --- | --- |
| `m365_status` | Signed-in, session usable, per-namespace consent. No tokens. Does not open a browser. | Optional |
| `m365_help` | CLI help for a namespace or verb, or a recipe `topic` (`mail-search`, `teams-find`, `calendar`, `files`, `mail-write`, `teams-write`). No session required. | None |
| `m365_run` | Run one CLI namespace+verb with a flag map. Returns that command's JSON. Writes dry-run unless `write_opt_in` is true. Lookup and write examples: help topics mail-search, teams-find, calendar, files, mail-write, teams-write. | Required for workloads |

Unknown tool name → MCP protocol error. Do not add `m365_login`, `m365_graph`, or per-verb tools.

## `m365_status`

**Input**: empty object.

**Output**: same JSON as `m365 auth status` (see `specs/001-m365-cli/contracts/json-stdout.schema.json` auth object). Keys: `signed_in`, `session_usable`, `account` (when signed in), `namespaces.mail|teams|calendar|files`. MUST NOT include token fields.

Signed out → success, `signed_in` false, no interactive login.

## `m365_help`

**Input**:

| Field | Type | Required |
| --- | --- | --- |
| `namespace` | string | no |
| `verb` | string | no |
| `topic` | string | no (`mail-search` \| `teams-find` \| `calendar` \| `files` \| `mail-write` \| `teams-write`) |

No args → overview: three tools, six recipe topics, write opt-in. `topic` set → recipe text from [recipes.md](recipes.md) (not CLI `--help`). `namespace` / `verb` → CLI help. `topic` and `namespace` together → usage. Unknown topic → usage. No session. Text, not JSON.

## Named recipes (MCP prompts)

`prompts/list` MUST return exactly: `mail-search`, `teams-find`, `calendar`, `files`, `mail-write`, `teams-write`. Lookup prompt descriptions are “Lookup recipe …”; write prompt descriptions are “Write recipe …”. `prompts/get` returns the same body as `m365_help` for that topic. MUST NOT add a fourth tool.

## `m365_run`

**Input**: [json-run-request.schema.json](json-run-request.schema.json).

**Output (success)**: one text content item = CLI JSON stdout for the equivalent command (help verbs: help text). File download/save: path metadata only.

**Output (failure)**: `isError` true; one text content item matching [json-tool-error.schema.json](json-tool-error.schema.json). Classes MUST remain `usage` | `auth` | `service` | `not_found`.

### Write gate

Verbs that MUST dry-run unless `write_opt_in` is true: `mail send`, `mail reply`, `teams send`, `calendar create`, `calendar update`, `calendar delete`, `files upload`, `files create-folder`, `files delete`, `files move`.

`mail watch`, including `mail watch --classify`, is read/classify behavior and
MUST NOT be added to the write gate. A classified event and its `actionable`
field are routing metadata only. A later `mail reply` remains a separate call
and MUST still be forced to dry-run unless `write_opt_in` is true.

The supported agent sequence is: consume `mail.response_classified`, read the
cited thread, compose text outside `m365`, invoke `mail reply` for a preview,
obtain approval, and only then make a distinct opted-in reply call. Neither an
`actionable` value nor the earlier watch invocation carries authorization into
that later call. Classification does not create an Outlook Draft item.

### Forbidden via run

`auth login`, `auth logout`, namespace `mcp` → `usage` with hint to run `m365 auth login` in a terminal.

### Alias

`chat` → `teams` (same as CLI).

## CLI surface (additive)

| Command | Behavior |
| --- | --- |
| `m365 --help` | Also lists `mcp`. Mail/teams/calendar/files lines stay. |
| `m365 mcp --help` / `m365 mcp serve --help` | Exit 0, no session. Names stdio, three tools, write opt-in default false, six recipe topics. |
| `m365 mcp serve` | JSON-RPC on stdio until stdin closes. `--human` → usage (3). |
| `m365 mail list` etc. | Unchanged (FR-012). |

## Error mapping

| CLI exit | MCP |
| --- | --- |
| 0 | `isError` false, stdout text |
| 3 | `isError` true, `class=usage` |
| 4 | `isError` true, `class=auth` |
| 5 | `isError` true, `class=service` |
| 6 | `isError` true, `class=not_found` |

Do not map these onto JSON-RPC error codes.

## Secrets

Redact access tokens, refresh tokens, authorization codes, client secrets,
TypeSafe API keys, and bearer values in every MCP content item and log line.
`TYPESAFE_API_KEY` and its alias `TYPESAFE_AI_TOKEN` are process environment only and MUST NOT be accepted in an
`m365_run` request. File bytes, classifier request state, and provider response
bodies MUST NOT appear.
