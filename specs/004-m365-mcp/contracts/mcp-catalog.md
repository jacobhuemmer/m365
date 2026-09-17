# MCP catalog

Public interface: `m365 mcp serve` on stdio JSON-RPC. Human CLI remains `m365 <namespace> <verb> [flags]` as in 001–003. This catalog is the MCP door, not a Graph API dump.

`tools/list` MUST return exactly these three names (SC-001).

Transport: stdio only. No SSE, no Streamable HTTP.

## Tools

| Name | Description (MUST convey) | Session |
| --- | --- | --- |
| `m365_status` | Signed-in, session usable, per-namespace consent. No tokens. Does not open a browser. | Optional |
| `m365_help` | CLI help for a namespace or verb. Names flags, limits, and write opt-in. | None |
| `m365_run` | Run one CLI namespace+verb with a flag map. Returns that command's JSON. Writes dry-run unless `write_opt_in` is true. | Required for workloads |

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

Omitted namespace → root help (MUST mention `mcp`, write opt-in, and the three tools). Namespace only → that namespace help. Namespace+verb → verb help. Exit success without a session. Text is CLI help text, not JSON.

## `m365_run`

**Input**: [json-run-request.schema.json](json-run-request.schema.json).

**Output (success)**: one text content item = CLI JSON stdout for the equivalent command (help verbs: help text). File download/save: path metadata only.

**Output (failure)**: `isError` true; one text content item matching [json-tool-error.schema.json](json-tool-error.schema.json). Classes MUST remain `usage` | `auth` | `service` | `not_found`.

### Write gate

Verbs that MUST dry-run unless `write_opt_in` is true: `mail send`, `mail reply`, `teams send`, `calendar create`, `calendar update`, `calendar delete`, `files upload`, `files create-folder`, `files delete`, `files move`.

### Forbidden via run

`auth login`, `auth logout`, namespace `mcp` → `usage` with hint to run `m365 auth login` in a terminal.

### Alias

`chat` → `teams` (same as CLI).

## CLI surface (additive)

| Command | Behavior |
| --- | --- |
| `m365 --help` | Also lists `mcp`. Mail/teams/calendar/files lines stay. |
| `m365 mcp --help` / `m365 mcp serve --help` | Exit 0, no session. Names stdio, three tools, write opt-in default false. |
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

Redact access tokens, refresh tokens, authorization codes, and client secrets in every MCP content item and log line. File bytes MUST NOT appear.
