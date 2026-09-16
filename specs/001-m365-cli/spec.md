# Feature Specification: m365 CLI v1

**Feature Branch**: `001-m365-cli`

**Created**: 2026-09-16

**Status**: Draft

**Input**: User description: "Specify v1 of the m365 CLI as one feature (short name m365-cli): a local command-line tool for one signed-in Microsoft 365 user to manage Outlook mail and Teams messages, replacing the existing local m365 command surface used by Mason and local agents, without replacing overnight incident monitoring or kata wrappers."

## Specification Contract

### Objective

m365 is a local command-line tool for one signed-in Microsoft 365 user. A person at a terminal and a local automation agent invoke the same commands. v1 lets that user sign in, inspect session status, list and read Outlook mail, list and read Teams chats, send and reply with an explicit dry-run, attach and save documents, and watch Teams for new messages.

### Goals

- Replace the existing local m365 command surface for auth, mail, and Teams (today invoked as `chat`) with a stable, testable contract.
- Keep result data on stdout and diagnostics on stderr, with parseable JSON for agents and a human-readable mode for people.
- Keep mail and Teams independently usable, including independent consent.
- Treat documents as first-class: attach local files on send, show attachment metadata on read, and save a named attachment only to a path the user names.
- Keep every list bounded and every limit visible. Never dump a full mailbox or chat silently.

### Non-goals

- Calendar, OneDrive as a drive, Planner, or a generic Microsoft 365 escape hatch.
- App-only or daemon credentials, shared-mailbox administration, or passwords on the command line.
- GUI, TUI, or web app.
- Overnight incident monitor, LaunchAgent, kata wrappers, or MCP servers (they remain outside this repository).
- Outbound send policy such as `--guard`, allowlists, or hourly caps (dry-run is the v1 safety rail).
- Language, module layout, SDK, token-storage backend, or CI tooling (plan).
- Recording live mailbox or chat content, including real attachments, in the repository.
- Printing file bytes on the terminal or in JSON stdout.

### Verification Strategy

Verification MUST cover login, status, and logout; mail list/get/thread/send/reply; Teams list/get/messages/send/watch; attachment attach/inspect/save; JSON parseability; secret redaction; visible limits; namespace consent isolation; empty and not-found states; and dry-run versus real send. Fixtures MUST be synthetic. Live mailbox and chat documents MUST NOT be stored in the repository.

## Document Map

This file is the overview and index. Canonical user-story, requirement, and success-criterion bodies live in the linked files; each ID appears in exactly one file.

| File | Canonical scope | Primary IDs |
| --- | --- | --- |
| [cli-contract.md](cli-contract.md) | Command shape, streams, formats, exit classes, help, limits, alias, namespace seam | US-010, FR-001..FR-013 |
| [auth.md](auth.md) | Delegated sign-in, status, logout, session reuse, consent isolation | US-001, US-009, FR-014..FR-022 |
| [mail.md](mail.md) | Outlook list, read, thread, send, reply | US-002, US-004, FR-023..FR-034 |
| [teams.md](teams.md) | Teams chats, messages, send, watch | US-003, US-005, US-008, FR-035..FR-047 |
| [attachments.md](attachments.md) | Outbound attach, inbound metadata, save to a local path | US-006, US-007, FR-048..FR-059 |
| [checklists/requirements.md](checklists/requirements.md) | Specification quality checklist | Review gate |

## User Scenarios & Testing *(mandatory)*

Canonical acceptance scenarios are split by contract boundary:

- **US-001 Sign in, inspect status, and log out** (P1): [auth.md](auth.md#user-story-1--sign-in-inspect-status-and-log-out-priority-p1)
- **US-002 List and read Outlook mail** (P1): [mail.md](mail.md#user-story-2--list-and-read-outlook-mail-priority-p1)
- **US-003 List Teams chats and read messages** (P1): [teams.md](teams.md#user-story-3--list-teams-chats-and-read-messages-priority-p1)
- **US-004 Send and reply to mail with dry-run** (P2): [mail.md](mail.md#user-story-4--send-and-reply-to-mail-with-dry-run-priority-p2)
- **US-005 Send a Teams chat message with dry-run** (P2): [teams.md](teams.md#user-story-5--send-a-teams-chat-message-with-dry-run-priority-p2)
- **US-006 Attach local documents on send or reply** (P2): [attachments.md](attachments.md#user-story-6--attach-local-documents-on-send-or-reply-priority-p2)
- **US-007 Save a named attachment to a local path** (P2): [attachments.md](attachments.md#user-story-7--save-a-named-attachment-to-a-local-path-priority-p2)
- **US-008 Watch Teams for new messages** (P2): [teams.md](teams.md#user-story-8--watch-teams-for-new-messages-priority-p2)
- **US-009 Use mail and Teams with independent consent** (P3): [auth.md](auth.md#user-story-9--use-mail-and-teams-with-independent-consent-priority-p3)
- **US-010 Discover commands via help** (P3): [cli-contract.md](cli-contract.md#user-story-10--discover-commands-via-help-priority-p3)

### Edge Cases

- Expired or missing session on a workload command versus `auth status` while signed out.
- Consent granted for mail but denied for Teams, and the reverse.
- Empty inbox, empty named folder, empty chat list, and a chat with no messages.
- Unknown or invalid message, chat, folder, or attachment identifier.
- List command with no `--top` (MUST apply the documented default, never unbounded).
- `--top` above the documented maximum (MUST fail validation, MUST NOT silently cap).
- Microsoft 365 service error or throttle while a session is usable.
- Dry-run versus a real send, including dry-run with attachments.
- Missing, unreadable, empty, oversize, or too-many local files on attach.
- Two outbound files with the same base name; two inbound attachments with the same display name.
- Save with no destination path; destination exists without overwrite; unwritable destination.
- Message or chat with zero attachments.
- Attachment save or list invoked on the wrong namespace for that message.
- Watch with no new messages; watch when more events exist than the visible limit.

## Requirements *(mandatory)*

### Functional Requirements

Canonical functional requirements are split by owned behavior:

- **CLI contract**: [cli-contract.md](cli-contract.md#functional-requirements)
- **Auth and consent**: [auth.md](auth.md#functional-requirements)
- **Mail**: [mail.md](mail.md#functional-requirements)
- **Teams**: [teams.md](teams.md#functional-requirements)
- **Attachments**: [attachments.md](attachments.md#functional-requirements)

### Key Entities *(include if feature involves data)*

- **Session**: The saved sign-in of one Microsoft 365 user. Attributes: whether the user is signed in, whether a silent session is usable, account display name, and which namespaces have consent. MUST NOT include access tokens, refresh tokens, authorization codes, or client secrets.
- **MailMessage**: One Outlook message. Attributes: id, conversation id, subject, from, to, cc, received time, read flag, body text when requested, and attachment metadata (not file bytes).
- **MailThread**: Ordered set of MailMessage items in one conversation, oldest first, including the signed-in user's own replies.
- **Chat**: One Teams chat. Attributes: id, topic, type, members, and last-message preview when listing.
- **ChatMessage**: One message in a Chat. Attributes: id, chat id, sender, created time, text, and attachment metadata (not file bytes).
- **Attachment**: A document belonging to one MailMessage or ChatMessage. Attributes: id, display name, size, content type, and parent message identity. Outbound attachments additionally have a local filesystem path.
- **CommandResult**: What a command emits: stdout payload, stderr diagnostics, and exit class.
- **ResultLimit**: The applied page size, the count returned, whether more results exist, and an opaque next-page token when they do. The token is not a service URL.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From a signed-out state, after one interactive login succeeds, `auth status` reports a usable session and `mail list` with the default limit returns an inbox page without a second sign-in gesture.
- **SC-002**: For every non-watch command, stdout in JSON mode is exactly one parseable JSON value with no log or diagnostic lines mixed in. Watch stdout in JSON mode is JSON lines: each non-empty line is one parseable JSON object, and stderr is not mixed into those lines.
- **SC-003**: No command, including verbose or debug modes, prints an access token, refresh token, authorization code, or client secret on stdout or stderr.
- **SC-004**: Every list command applies a documented result limit. The applied limit is visible in command output or `--help`. Omitting `--top` never returns an unbounded mailbox or chat dump.
- **SC-005**: With mail consent and without Teams consent, mail list and get succeed and every Teams command exits with the auth class, not the service class. With Teams consent and without mail consent, the reverse is true.
- **SC-006**: An empty inbox, empty folder, empty chat list, empty message list, or watch with no new events exits success with an empty result, not an error.
- **SC-007**: An unknown message, chat, folder, or attachment identity exits with the not-found class, distinct from auth and from service error.
- **SC-008**: Dry-run on mail send, mail reply, and Teams send shows the intended destination, body, and attachment names and sizes, and does not create a message the user can later read in Sent Items or in the chat.
- **SC-009**: Dry-run with attachments does not upload file bytes; no new remote attachment appears on a message.
- **SC-010**: Attachment save writes only to the user-supplied path and reports that path. Omitting the path is a usage failure and writes no file anywhere.
- **SC-011**: An oversize file or too many attachments is rejected with the usage/config class before send. No partial message is created.
- **SC-012**: Save to a path that already exists, without an overwrite flag, refuses and leaves the existing file unchanged.
- **SC-013**: `--help` for send, reply, and attach names the maximum attachment size and the maximum attachment count.
- **SC-014**: After a usable session exists, a default-limit inbox list completes in under 15 seconds on an ordinary connection. A default-limit chat list completes in under 15 seconds on an ordinary connection.
- **SC-015**: Removing or never enabling the Teams namespace does not change mail command names, flags, output shape, or exit classes. Removing or never enabling mail does not change Teams command names, flags, output shape, or exit classes.

## Assumptions

- The product command name is `m365`. The v1 workload namespaces are `mail` and `teams`. `auth` is shared core behavior, not a third workload.
- JSON is the default stdout format so local agents can parse it. `--human` selects human-readable stdout. `--json` is accepted and selects JSON.
- Exit classes: `0` success; `3` usage, validation, and configuration; `4` authentication and consent; `5` Microsoft 365 service error; `6` not-found. These MUST NOT collapse into a single generic failure.
- Default `--top` is `10` for `mail list` and `20` for `teams list`, `teams messages`, and `teams watch`. The maximum `--top` is `50`. A value above `50` is a validation error, not a silent cap.
- Pagination uses `--top` plus an opaque `--page-token` taken from a previous result's `next_page` field. Continuing a list MUST NOT require a raw service URL command.
- `chat` is a compatibility alias for the `teams` namespace. It is not a second domain. New documentation uses `teams`.
- Existing destination files are refused unless `--overwrite` is present.
- Maximum outbound attachment size is 10 MiB per file. Maximum outbound attachment count is 10 per send or reply. Empty (zero-byte) files are rejected as validation errors.
- Mail send/reply bodies and Teams send text MAY be supplied by a flag or by a file path, including stdin. Default content is ordinary text. Mail MAY accept `--html` to mark the body as HTML. Teams MAY accept `--format md` to mark structured text. This spec does not define a conversion algorithm.
- Watch emits new messages the signed-in user should see: every message in a 1:1 chat, @mentions, and extra chats named with `--chat`. Watch MAY persist a local checkpoint so a later invocation without `--since` emits only newer messages. Checkpoint storage is chosen at plan time. Watch MUST NOT start overnight agents.
- The tool is configured for one organizational app registration. Client ID, Tenant ID, and secrets MUST NOT appear in this specification, source, tests, logs, or fixtures.
- Outbound send policy (`--guard`, allowlists, hourly caps, automated tags) is out of v1. Dry-run is the v1 confirmation rail.
- This feature replaces the local m365 CLI command surface only. It does not replace origin/dotfiles overnight monitoring or catalog kata wrappers.

## Change Control

When updating this specification set, keep each requirement ID in exactly one canonical file, update this index if ownership changes, and update [checklists/requirements.md](checklists/requirements.md) with the validation pass.
