# m365 Mail Namespace

This file owns Outlook list, read, thread, send, and reply. Shared CLI rules
are owned by [cli-contract.md](cli-contract.md). Attachments are owned by
[attachments.md](attachments.md). Auth and consent are owned by
[auth.md](auth.md).

## User Scenarios & Testing *(mandatory)*

### User Story 2 — List and read Outlook mail (Priority: P1)

The signed-in user lists inbox or a named folder, optionally unread-only or
search-filtered, with an explicit result limit. The user reads one message
as text and reads a conversation thread oldest-first, including the user's
own replies. Messages with documents show attachment metadata (name, size,
type) without dumping file bytes.

**Why this priority**: Reading mail is the primary job of v1 after sign-in.

**Independent Test**: With a usable mail session from login (no second
gesture) and synthetic mailbox fixtures, list the inbox with default limit,
list unread, search, get one message, and read a thread with and without
bodies. Assert empty folder success, unknown id exit `6`, attachment
metadata without bytes, and the applied limit in the result.

**Acceptance Scenarios**:

1. **Given** a usable mail session and an inbox with messages, **When** the
   user runs `mail list` with no `--top`, **Then** stdout returns at most the
   default limit, names that limit, and exits `0`.
2. **Given** an empty inbox or empty named folder, **When** the user lists it,
   **Then** the command exits `0` with an empty item list, not an error.
3. **Given** `--unread` or `--search`, **When** the user lists mail, **Then**
   only matching messages are returned, still bounded by `--top`.
4. **Given** a known message id, **When** the user runs `mail get`, **Then**
   stdout includes subject, participants, timestamp, body as text, and
   attachment metadata without file bytes.
5. **Given** a conversation with several messages including the user's own
   replies, **When** the user runs `mail thread` with `--bodies`, **Then**
   messages appear oldest-first, include the user's replies, and include
   bodies. Without `--bodies`, metadata is returned without bodies.
6. **Given** an unknown message id, **When** the user runs `mail get` or
   `mail thread`, **Then** the command exits `6`.
7. **Given** a message with no attachments, **When** the user gets it or lists
   its attachments, **Then** the attachment list is empty and the command
   exits `0`.

### User Story 4 — Send and reply to mail with dry-run (Priority: P2)

The user sends new mail or replies (sender only, or all recipients). Dry-run
shows the exact intended send without sending. A real send happens only when
the command runs without dry-run.

**Why this priority**: Sending is valuable but unsafe without dry-run, and it
depends on being able to read a message first.

**Independent Test**: Dry-run a send and a reply against fixtures; assert no
message is created. Run the same commands without dry-run; assert a message
is created and stdout reports success with a message identity. Omit required
fields and assert exit `3`.

**Acceptance Scenarios**:

1. **Given** required `--to`, `--subject`, and body, **When** the user runs
   `mail send --dry-run`, **Then** stdout shows the intended recipients,
   subject, body, and any attachment names and sizes, exits `0`, and no new
   mail appears in Sent Items.
2. **Given** the same inputs without `--dry-run`, **When** the user runs
   `mail send`, **Then** the message is sent and stdout reports success with
   a message identity.
3. **Given** a known message id and a body, **When** the user runs
   `mail reply --dry-run`, **Then** only the original sender is the intended
   recipient. With `--all`, all original recipients are intended. Neither
   dry-run creates a reply.
4. **Given** missing `--to`, missing subject, or missing body, **When** the
   user runs `mail send`, **Then** the command exits `3` and does not send.
5. **Given** an unknown reply target, **When** the user runs `mail reply`,
   **Then** the command exits `6` and does not send.

## Functional Requirements

- **FR-023**: `mail list` MUST list messages in inbox by default, or in a
  named folder, with optional unread-only and search filters, bounded by
  `--top` and continued with `--page-token`. Well-known folder names MUST
  include `inbox`, `sentitems`, `drafts`, and `all`.
- **FR-024**: `mail get` MUST return one message as text (subject,
  participants, timestamp, body text) plus attachment metadata without file
  bytes.
- **FR-025**: `mail thread` MUST return the conversation containing the given
  message, oldest-first, including the signed-in user's own replies.
  `--bodies` MUST include body text; without `--bodies`, metadata only.
- **FR-026**: An empty inbox, folder, search, or unread filter MUST exit `0`
  with an empty item list.
- **FR-027**: An unknown message id on get, thread, reply, or attachment
  operations MUST exit `6`. An unknown folder name MUST exit `6`. A missing
  required id or flag MUST exit `3`.
- **FR-028**: `mail send` MUST accept one or more `--to` recipients, a
  `--subject`, a body, optional `--cc`, optional local attachments, optional
  `--html`, and `--dry-run`.
- **FR-029**: `mail reply` MUST reply to the sender of the given message.
  `--all` MUST reply to all original recipients. Reply MUST keep the thread
  subject. Reply MUST accept a body, optional local attachments, optional
  `--html`, and `--dry-run`.
- **FR-030**: `--dry-run` on send or reply MUST show the intended send and
  MUST NOT create a message.
- **FR-031**: Body text MUST be accepted from a flag or from a file path,
  including stdin.
- **FR-032**: `mail list`, `mail get`, and `mail thread` MUST include
  attachment metadata on each message. A dedicated mail attachments-list
  verb MUST return the same metadata for one message. Zero attachments is
  an empty list, not an error.
- **FR-033**: Mail attachment save MUST live under the mail namespace and
  MUST follow [attachments.md](attachments.md).
- **FR-034**: v1 mail MUST NOT move, delete, create or delete folders,
  manage rules or categories, open calendar, operate a shared mailbox, or
  treat OneDrive or SharePoint as a file picker.
