# m365 Teams Namespace

This file owns Teams chat list, get, messages, send, and watch. Shared CLI
rules are owned by [cli-contract.md](cli-contract.md). Attachments are owned
by [attachments.md](attachments.md). Auth and consent are owned by
[auth.md](auth.md). Existing `chat` command names are the behavioral source;
the canonical namespace is `teams`.

## User Scenarios & Testing *(mandatory)*

### User Story 3 — List Teams chats and read messages (Priority: P1)

The signed-in user lists chats and reads messages in a chat, with explicit
result limits. Messages with documents show attachment metadata the same way
mail does.

**Why this priority**: Reading chats is the Teams counterpart of reading
mail and is required for v1.

**Independent Test**: With a usable Teams session and synthetic chat
fixtures, list chats with the default limit, get one chat with members, and
list messages. Assert empty list success, unknown chat exit `6`, attachment
metadata without bytes, and the applied limit in the result.

**Acceptance Scenarios**:

1. **Given** a usable Teams session and existing chats, **When** the user
   runs `teams list` with no `--top`, **Then** stdout returns at most the
   default limit, names that limit, and includes chat id, topic or member
   summary, type, and last-message preview.
2. **Given** no chats, **When** the user lists chats, **Then** the command
   exits `0` with an empty item list.
3. **Given** a known chat id, **When** the user runs `teams get`, **Then**
   stdout includes members and chat metadata.
4. **Given** a known chat id, **When** the user runs `teams messages` with a
   limit, **Then** stdout returns at most that many messages, oldest-first
   within the page, with text and attachment metadata without file bytes.
5. **Given** an unknown chat id, **When** the user runs `teams get` or
   `teams messages`, **Then** the command exits `6`.
6. **Given** a chat with no messages, **When** the user lists messages,
   **Then** the command exits `0` with an empty item list.

### User Story 5 — Send a Teams chat message with dry-run (Priority: P2)

The user sends a message to a chat. Dry-run shows the exact intended send
without sending. A real send happens only without dry-run.

**Why this priority**: Sending is valuable after the user can read the chat,
and dry-run is the v1 safety rail.

**Independent Test**: Dry-run a send against a known chat; assert no message
is created. Send without dry-run; assert the message is created and stdout
reports an identity. Missing text exits `3`. Unknown chat exits `6`.

**Acceptance Scenarios**:

1. **Given** a known chat id and text, **When** the user runs
   `teams send --dry-run`, **Then** stdout shows the chat, text, and any
   attachment names and sizes, exits `0`, and no new message appears in the
   chat.
2. **Given** the same inputs without `--dry-run`, **When** the user sends,
   **Then** the message is sent and stdout reports success with a message
   identity.
3. **Given** missing text, **When** the user sends, **Then** the command
   exits `3` and does not send.
4. **Given** an unknown chat id, **When** the user sends, **Then** the
   command exits `6` and does not send.

### User Story 8 — Watch Teams for new messages (Priority: P2)

The user or a local watcher consumes a stream of new messages the signed-in
user should see: 1:1 chats, extra named chats, and mentions. This story does
not define overnight agent spawn policy.

**Why this priority**: Watch is how local automation notices new chat mail,
but it is not required to list or send.

**Independent Test**: With synthetic events, run watch with no new messages
(empty success), with 1:1 and mention events (emitted), with `--chat` for an
extra chat, with `--since`, and with more events than `--top` (limit visible,
checkpoint only through emitted events). Assert JSON lines on stdout.

**Acceptance Scenarios**:

1. **Given** no new matching messages, **When** the user runs `teams watch`
   in JSON mode, **Then** the command exits `0` with empty stdout.
2. **Given** new 1:1 messages and an @mention in another chat, **When** the
   user runs `teams watch`, **Then** those events are emitted as JSON lines
   with chat id, message id, sender, text, time, and a reason.
3. **Given** `--chat` naming an extra group chat, **When** new messages
   arrive there, **Then** watch emits them even when they are not mentions.
4. **Given** more new events than `--top`, **When** watch runs, **Then** at
   most `--top` events are emitted, the applied limit is visible, and a later
   watch without `--since` continues after the last emitted event rather than
   skipping unseen events.
5. **Given** no session, **When** watch runs, **Then** the command exits `4`.
6. **Given** an unknown `--chat` id, **When** watch runs, **Then** the
   command exits `6`.

## Functional Requirements

- **FR-035**: `teams list` MUST list chats bounded by `--top` and continued
  with `--page-token`. Each item MUST include chat id, type, topic or member
  summary, members when available, and a last-message preview.
- **FR-036**: `teams get` MUST return one chat and its members.
- **FR-037**: `teams messages` MUST list messages in a chat bounded by
  `--top` and continued with `--page-token`, oldest-first within the page,
  with text and attachment metadata. `--include-system` MUST default off.
- **FR-038**: An empty chat list or empty message list MUST exit `0` with an
  empty item list.
- **FR-039**: An unknown chat id on get, messages, send, watch `--chat`, or
  attachment operations MUST exit `6`. An unknown message id in a known chat
  MUST exit `6`. A missing required id or flag MUST exit `3`.
- **FR-040**: `teams send` MUST accept a chat id, message text from a flag or
  file path including stdin, optional local attachments, optional `--html` or
  `--format md`, and `--dry-run`.
- **FR-041**: `--dry-run` on send MUST show the intended send and MUST NOT
  create a message.
- **FR-042**: `teams watch` MUST emit new messages the signed-in user should
  see: every message in a 1:1 chat, @mentions, and extra chats named with
  `--chat`. JSON mode MUST write one JSON object per event as a line on
  stdout. Each event MUST include chat id, message id, sender, text, time,
  and reason, plus attachment metadata when documents are present.
- **FR-043**: Watch with no new events MUST exit `0`. JSON mode MUST write no
  objects to stdout in that case.
- **FR-044**: Watch MUST accept `--since` as a start time when no checkpoint
  exists. Watch MAY persist a local checkpoint. The checkpoint MUST advance
  only through emitted events. Watch MUST apply `--top` (default and maximum
  in [spec.md](spec.md) Assumptions). Watch MUST NOT start overnight agents,
  LaunchAgents, or other watchers.
- **FR-045**: v1 Teams MUST NOT include meetings, calls, files as a drive,
  channel team-admin, or a generic Microsoft 365 escape hatch. Pagination
  MUST use `--top` and `--page-token` on these commands, not a raw URL
  command.
- **FR-046**: `teams messages` MUST include attachment metadata. A dedicated
  Teams attachments-list verb MUST return the same metadata for one message
  in a chat. Zero attachments is an empty list, not an error.
- **FR-047**: Teams attachment save MUST live under the teams namespace and
  MUST follow [attachments.md](attachments.md).
