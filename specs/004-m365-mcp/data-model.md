# Phase 1 Data Model: m365 MCP Serve

MCP is a second door into the existing CLI. Domain types below are MCP-facing. Workload entities (messages, events, drive items, Session) stay as specified in 001–003. No Graph SDK fields, URLs, HTTP statuses, or token strings.

## MCPCatalog

**Fields**: exactly three tools: `m365_status`, `m365_help`, `m365_run`. Each has a short description.

**Rules**:
- `tools/list` MUST return these names only (SC-001).
- No tool per Graph URL, per CLI verb, or per flag.
- No `auth login` / `auth logout` tools (FR-008).
- Descriptions MUST be enough to choose status vs help vs run.

## SessionSnapshot

Reuses `domain.Session` via `auth status`.

**Fields**: signed in, session usable, account display name, namespaces `{mail, teams, calendar, files}` booleans.

**Rules**:
- MUST NOT contain access tokens, refresh tokens, authorization codes, or client secrets.
- Signed out is success: `signed_in` false, no browser, no hang.
- Missing or expired session on a workload run is auth-class, not a new MCP class.
- Same Keychain / `0600` session file as the CLI.

## RunRequest

**Fields**:
- `namespace` (string, required)
- `verb` (string, required)
- `args` (string array, optional positionals)
- `flags` (object, optional)
- `write_opt_in` (bool, optional, default false)

**Rules**:
- `chat` is an alias of `teams`.
- Flag keys are CLI long names without `--`.
- Flag values: string, number, boolean, or array of strings.
- Unknown namespace, verb, or flag → usage.
- `auth` `login` / `auth` `logout` / namespace `mcp` → usage; hint to use a human terminal for login.
- `--body-file` or `--text-file` equal to `-` → usage.
- Write verbs (mail send/reply, teams send, calendar create/update/delete, files upload/create-folder/delete/move): if `write_opt_in` is not true, dispatch MUST pass `--dry-run`.
- `write_opt_in` true still honors an explicit `dry-run` true (no mutation).
- Reads ignore `write_opt_in`.

## RunResult

**Fields**: CLI stdout payload (JSON text), or a classed error.

**Success**:
- Text is the same JSON the CLI would print for that command (help: CLI help text).
- File download / save-attachment: path metadata only; never file bytes.

**Failure** (distinguishable classes, FR-006):

| Class | Meaning |
| --- | --- |
| `usage` | bad namespace/verb/flags, `--top` above max, login via MCP |
| `auth` | no session, expired session, missing namespace consent |
| `service` | Microsoft 365 / Graph failure |
| `not_found` | unknown message, chat, event, item |

## WriteGate

**Fields**: verb is a workload write (bool), opt-in (bool).

**Transitions**:
- write + opt-in false/omitted → dry-run preview, mailbox/calendar/drive unchanged (SC-004).
- write + opt-in true → same as CLI without `--dry-run` (unless flags also dry-run).
- not a write → no gate.

## Validation summary

| Rule | Failure class |
| --- | --- |
| Unknown tool name | MCP protocol error (not a CLI class) |
| Missing namespace or verb on run | usage |
| Unknown flag or `--top` above CLI max | usage (not silent cap) |
| `auth login` / `logout` / `mcp` via run | usage |
| `--body-file` `-` | usage |
| No session or missing consent | auth |
| Graph/service error | service |
| Unknown id | not-found |
| Write without opt-in | success dry-run, no mutation |
