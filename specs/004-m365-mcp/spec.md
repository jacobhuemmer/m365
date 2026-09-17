# Feature Specification: m365 MCP Serve

**Feature Branch**: `004-m365-mcp`

**Created**: 2026-09-16

**Status**: Draft

**Input**: User description: "Implement M365 MCP as a compact dedicated server that wraps the existing signed-in-user m365 CLI, like kata mcp serve, not a generic Graph tool dump."

## Specification Contract

### Objective

Agents talk to Mason's Microsoft 365 the same way the terminal does: one signed-in user, the same namespaces and verbs, JSON results, dry-run on writes. MCP is a second door into that CLI, not a second Microsoft 365 client.

### Goals

- Expose `m365 mcp serve` over stdio so Grok, Claude, Cursor, Codex, and OpenCode can call mail, teams, calendar, and files without shelling out.
- Keep the tool list tiny (kata-shaped: a handful of tools, not one tool per Graph API).
- Reuse the existing delegated session, consent isolation, dry-run, limits, and exit classes.
- Never print tokens. Never become a generic Graph escape hatch.

### Non-goals

- Microsoft's Enterprise MCP (`mcp.svc.cloud.microsoft/enterprise`) — Entra directory search, not mail/calendar.
- Microsoft Agent365 remote Calendar/Mail MCP — tenant-hosted, not this local PKCE app.
- Third-party Graph MCP packages (dozens of tools, their own token files).
- Putting m365 behind lazy-mcp (that bus is Datadog, Notion, Atlassian, Context7).
- App-only credentials, a plugin marketplace, or one MCP tool per CLI flag.
- Changing mail, teams, calendar, or files command names for humans at a terminal.

### Verification Strategy

Verification MUST cover: `tools/list` returns the compact catalog; status works signed-in and signed-out; a read tool returns the same JSON shape as the CLI; a write tool with dry-run does not send; a write without dry-run is refused unless the caller opts in; missing consent is auth-class, not service-class; secrets never appear on the MCP wire. Fixtures MUST be synthetic.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Agent checks session and lists compact tools (Priority: P1)

An agent starts `m365 mcp serve` on stdio, lists tools, and calls status. Status reports signed-in, session usable, and per-namespace consent (mail, teams, calendar, files) with no tokens. If the session is missing, status says so and does not hang on a browser.

**Why this priority**: Agents cannot guess whether they may call mail or calendar until status exists. The compact catalog is the whole point of this server.

**Independent Test**: Drive stdio JSON-RPC against `m365 mcp serve`. `tools/list` returns the documented tool names only. Status with a usable session reports consent flags. Status with no session reports not signed in and does not open a browser.

**Acceptance Scenarios**:

1. **Given** `m365 mcp serve` on stdio, **When** the client lists tools, **Then** the catalog is the compact set in this spec (status, help, run) and not one tool per Graph endpoint.
2. **Given** a usable session, **When** the client calls status, **Then** the result names signed in, session usable, and consent for mail, teams, calendar, and files, with no secrets.
3. **Given** no session, **When** the client calls status, **Then** the result reports not signed in, and no interactive login starts.

---

### User Story 2 - Agent reads mail, teams, calendar, or files through run (Priority: P1)

The agent calls one run tool with namespace, verb, and flags. The result is the same JSON the CLI would print for that command. Limits still apply. Unknown namespace or verb is a usage failure.

**Why this priority**: Read is the first useful MCP surface. Writes wait until run matches the CLI.

**Independent Test**: With a fake Graph behind the CLI, call run for `mail list`, `teams list`, `calendar list`, and `files list` at default limits. Assert JSON matches CLI stdout for the same flags, and that `mail nope` is usage.

**Acceptance Scenarios**:

1. **Given** mail consent, **When** the agent runs namespace `mail` verb `list` with default flags, **Then** the tool result is the CLI's JSON list (bounded `--top`) and not a different schema.
2. **Given** calendar consent, **When** the agent runs `calendar` `list` or `calendar` `free`, **Then** the result is the CLI JSON for those verbs.
3. **Given** an unknown verb, **When** the agent calls run, **Then** the tool returns a usage error and does not call Microsoft 365.
4. **Given** teams consent missing, **When** the agent runs `teams` `list`, **Then** the error is auth-class, not service-class.

---

### User Story 3 - Writes stay dry-run unless the caller opts in (Priority: P2)

Send, reply, create, update, delete, upload, and move through MCP default to dry-run. A real write requires an explicit opt-in flag on that call. Help names this rule.

**Why this priority**: MCP callers are agents. Accidental send is worse than at a terminal. Dry-run is already the CLI safety rail.

**Independent Test**: Call run for `mail send` without opt-in; assert no send and a dry-run payload. Call with opt-in false/omitted vs true. Assert real send only when opt-in is true.

**Acceptance Scenarios**:

1. **Given** a mail send through MCP with no write opt-in, **When** the tool runs, **Then** the result is a dry-run preview and no message is sent.
2. **Given** write opt-in true and valid send fields, **When** the tool runs, **Then** the message is sent the same way `m365 mail send` without `--dry-run` would.
3. **Given** `calendar create` through MCP, **When** opt-in is omitted, **Then** no event is created.

---

### User Story 4 - Help without a session (Priority: P3)

The agent asks help for a namespace or verb and gets the same names, flags, and limits as `m365 <ns> --help`, without needing a session.

**Why this priority**: Discovery completes the compact catalog. Reads already deliver value.

**Independent Test**: Call help with no session for `mail` and `calendar create`. Assert flags and limits appear and exit is success.

**Acceptance Scenarios**:

1. **Given** no session, **When** the agent calls help for `calendar`, **Then** the result names verbs including list, create, and free, and names phrase examples if that CLI help does.
2. **Given** no session, **When** the agent lists MCP tools, **Then** each tool has a short description sufficient to choose status vs help vs run.

---

### Edge Cases

- MCP client connected while the Keychain session expires mid-call (auth-class).
- Run with extra unknown flags (usage).
- Run with `--top` above the CLI maximum (usage, not silent cap).
- File download through MCP: bytes still go to a user-named path, never onto the MCP wire as file contents.
- Interactive `auth login` is not an MCP tool (browser gesture stays a human terminal command).
- Two MCP clients must not require two Microsoft 365 apps; they share the one CLI session.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The server MUST be invoked as `m365 mcp serve` (stdio JSON-RPC). It MUST NOT be a second binary with a second token store.
- **FR-002**: The MCP catalog MUST be exactly three tools: `m365_status`, `m365_help`, `m365_run`. The server MUST NOT register one tool per Graph URL or per CLI flag.
- **FR-003**: `m365_status` MUST report the same facts as `m365 auth status` (signed in, session usable, account display name, per-namespace consent) and MUST NOT print tokens.
- **FR-004**: `m365_help` MUST accept optional namespace and verb and MUST return the CLI help text for that surface. It MUST work with no session.
- **FR-005**: `m365_run` MUST accept namespace, verb, and a flag map. It MUST execute the equivalent CLI command and return that command's JSON stdout on success.
- **FR-006**: `m365_run` MUST map CLI exit classes to MCP errors without collapsing them: usage, auth, service, not-found remain distinguishable.
- **FR-007**: Workload writes (mail send/reply, teams send, calendar create/update/delete, files upload/delete/move) through `m365_run` MUST dry-run unless the caller sets write opt-in true on that call.
- **FR-008**: `auth login` and `auth logout` MUST NOT be MCP tools. Login remains an interactive human gesture on the terminal.
- **FR-009**: MCP MUST NOT call Microsoft 365 except through the existing CLI namespaces. No generic Graph GET/PATCH tool.
- **FR-010**: File bytes MUST NOT appear in MCP tool results. Download still writes only to a named local path.
- **FR-011**: The MCP server MUST NOT log access tokens, refresh tokens, authorization codes, or client secrets.
- **FR-012**: Enabling MCP MUST NOT change the human CLI: `m365 mail list` and friends stay as specified in 001–003.
- **FR-013**: Agents SHOULD register this server as a dedicated MCP (like kata), not as a fifth lazy-mcp backend, and not as Microsoft's Enterprise MCP.

### Key Entities

- **MCP catalog**: The three tools and their descriptions.
- **Run request**: namespace, verb, flag map, write opt-in.
- **Run result**: CLI JSON payload, or a classed error (usage, auth, service, not-found).
- **Session**: The same Keychain session the CLI already uses.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `tools/list` returns exactly three tools.
- **SC-002**: Status for a usable session reports consent flags for mail, teams, calendar, and files with no secrets, in under 2 seconds locally.
- **SC-003**: A default `mail list` through MCP returns the same item count and ids as `m365 mail list` with the same flags.
- **SC-004**: 100% of send/create/delete calls without write opt-in leave the mailbox/calendar/drive unchanged.
- **SC-005**: Missing teams consent on a teams run is reported as auth failure, not a generic tool error.
- **SC-006**: A person at a terminal can still run every `m365` command unchanged after MCP is added.
- **SC-007**: No MCP response or log contains an access token, refresh token, authorization code, or client secret.

## Assumptions

- This is a second interface to the existing CLI, not a new Graph adapter. 001–003 remain the source of verbs and flags.
- Compact catalog is intentional. Kata uses four tools for a whole script library; m365 uses three for one signed-in user.
- Write opt-in default is false (dry-run). This is stricter than the terminal, where `--dry-run` is opt-in. MCP callers are agents.
- Registration snippet for agents: command `m365`, args `["mcp", "serve"]`, same PATH as the stowed CLI. No `npx`.
- lazy-mcp stays Datadog, Notion, Atlassian, Context7. kata stays `kata mcp serve`. m365 is a third dedicated server.
- Microsoft Enterprise MCP and Agent365 remote servers are out of scope even if the tenant could enable them.
- `auth login` stays human-only. Agents that see `silent_token.ok` false tell Mason to run `m365 auth login` in a terminal.
- Constitution v1 listed MCP as outside this repository. This spec explicitly adds `m365 mcp serve` as an in-repo interface of the same binary.
