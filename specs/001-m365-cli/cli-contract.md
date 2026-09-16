# m365 CLI Contract

This file owns the user-visible command surface: shape, streams, formats, exit
classes, help, list limits, the `chat` alias, and the namespace seam. Auth
behavior is owned by [auth.md](auth.md). Mail verbs are owned by
[mail.md](mail.md). Teams verbs are owned by [teams.md](teams.md). Attachment
bytes and save rules are owned by [attachments.md](attachments.md).

## User Scenarios & Testing *(mandatory)*

### User Story 10 — Discover commands via help (Priority: P3)

A person or a local agent runs `m365 --help` and the `--help` of a namespace
or verb. Help names the namespace, required inputs, output modes, list
limits, and attachment size and count caps so the caller can invoke the
command without guessing.

**Why this priority**: Discovery is required for a complete CLI, but login,
read, and send already deliver the product. Help ships with those stories.

**Independent Test**: Invoke top-level, namespace, and verb `--help` with no
saved session and assert the required names, limits, and caps appear on
stdout with exit `0`.

**Acceptance Scenarios**:

1. **Given** no arguments beyond `--help`, **When** the user runs `m365 --help`,
   **Then** help lists the `auth`, `mail`, and `teams` surfaces, names JSON and
   human output modes, and names the stable exit classes.
2. **Given** a mail send help request, **When** the user runs
   `m365 mail send --help`, **Then** help names required inputs, `--dry-run`,
   output modes, and the attachment size and count caps.
3. **Given** a Teams send help request, **When** the user runs
   `m365 teams send --help`, **Then** help names required inputs, `--dry-run`,
   output modes, and the same attachment caps.
4. **Given** a list command help request, **When** the user runs
   `m365 mail list --help` or `m365 teams list --help`, **Then** help names
   `--top`, the default limit, the maximum limit, and `--page-token`.

## Functional Requirements

- **FR-001**: The user-visible command shape MUST be
  `m365 <namespace> <verb> [flags]`. v1 workload namespaces MUST be `mail` and
  `teams`. `auth` MUST be shared core behavior, not a workload namespace.
- **FR-002**: v1 MUST provide these verbs (names MAY be clarified in help
  text; coverage MUST match):
  - Core: `auth login`, `auth status`, `auth logout`, top-level `--help`
  - mail: `list`, `get`, `thread`, `send`, `reply`, list attachments, save
    attachment
  - teams: `list`, `get`, `messages`, `send`, `watch`, list attachments, save
    attachment
- **FR-003**: Every command MUST write result data to stdout and diagnostics
  to stderr. Commands MUST read from arguments, flags, or stdin. JSON on
  stdout MUST NOT contain log or diagnostic lines.
- **FR-004**: Every command MUST support machine-readable JSON and
  human-readable output. JSON MUST be the default. `--json` MUST select JSON.
  `--human` MUST select human-readable stdout. Non-watch JSON stdout MUST be
  exactly one parseable JSON value. `teams watch` JSON stdout MUST be JSON
  lines: each non-empty line one parseable JSON object.
- **FR-005**: Exit classes MUST be stable and documented: `0` success; `3`
  usage, validation, and configuration; `4` authentication and consent; `5`
  Microsoft 365 service error; `6` not-found. Once these classes exist,
  commands MUST NOT collapse them into a single generic failure.
- **FR-006**: Every command MUST define happy-path, empty, validation,
  auth-failure, and service-error behavior. Commands that take a message,
  chat, folder, or attachment identity MUST also define not-found behavior.
  A command with an unhandled required state is incomplete.
- **FR-007**: Every list or watch command MUST apply an explicit ResultLimit.
  Default and maximum `--top` values are recorded in [spec.md](spec.md)
  Assumptions. Omitting `--top` MUST apply the default and MUST NOT dump an
  unbounded mailbox or chat. A `--top` above the maximum MUST fail with exit
  `3` and MUST NOT silently cap. When more results exist, stdout MUST include
  an opaque `next_page` token; the next page MUST be requested with
  `--page-token`, not with a raw service URL. The applied limit MUST be
  visible in output or `--help`. Silent truncation is forbidden.
- **FR-008**: Commands, including verbose or debug output, MUST NOT print
  access tokens, refresh tokens, authorization codes, or client secrets.
  Those values MUST be redacted if a verbose mode exists.
- **FR-009**: `--help` MUST name the namespace (or core surface), required
  inputs, output modes, list limits, and attachment size and count caps where
  those caps apply. Help MUST exit `0` and MUST NOT require a signed-in
  session.
- **FR-010**: Adding, removing, or never enabling a namespace MUST NOT change
  another namespace's command names, flags, output shape, or exit classes.
  v1 MUST NOT specify calendar, OneDrive, or other future workload behavior.
  v1 MUST NOT invent a plugin marketplace or a generic provider surface.
- **FR-011**: `chat` MUST be accepted as a compatibility alias for the `teams`
  namespace with the same verbs and flags. `chat` is not a second domain.
  Canonical names in this specification and in new help text MUST be `teams`.
- **FR-012**: Mail `send`, mail `reply`, and Teams `send` MUST support
  `--dry-run`. Dry-run MUST show the intended send, including attachment
  names and sizes when files are named, and MUST NOT send.
- **FR-013**: In JSON mode, a failure MUST write one diagnostic JSON object to
  stderr with an error class, a human-readable message, and an optional hint,
  and MUST NOT include secrets. Human mode MUST write a plain-text diagnostic
  to stderr for the same classes.

## Command states

| Command | Empty | Validation | Auth (4) | Service (5) | Not-found (6) |
| --- | --- | --- | --- | --- | --- |
| `auth login` | n/a | missing configuration → 3 | login refused → 4 | Microsoft 365 unavailable → 5 | n/a |
| `auth status` | signed out is success with `signed_in=false` | n/a | n/a | Microsoft 365 unavailable while checking → 5 | n/a |
| `auth logout` | already signed out is success | n/a | n/a | n/a | n/a |
| `mail list` | no messages → 0, empty items | bad flags or `--top` out of range → 3 | no/expired session or no mail consent → 4 | yes | unknown folder → 6 |
| `mail get` / `thread` | n/a | missing id → 3 | 4 | yes | unknown message → 6 |
| `mail send` / `reply` | n/a | missing required fields or attach validation → 3 | 4 | yes | reply target unknown → 6 |
| `teams list` | no chats → 0, empty items | `--top` out of range → 3 | 4 | yes | n/a |
| `teams get` / `messages` | no messages → 0, empty items | missing id → 3 | 4 | yes | unknown chat or message → 6 |
| `teams send` | n/a | missing text or attach validation → 3 | 4 | yes | unknown chat → 6 |
| `teams watch` | no new events → 0, empty stdout in JSON | `--top` out of range → 3 | 4 | yes | unknown `--chat` id → 6 |
| attachment list | no attachments → 0, empty items | missing parent id → 3 | 4 | yes | unknown parent → 6 |
| attachment save | n/a | missing path, unwritable, exists without overwrite, ambiguous name → 3 | 4 | yes | unknown parent or attachment → 6 |

Happy path for each command is exit `0` with the documented stdout payload.
