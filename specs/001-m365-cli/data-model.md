# Phase 1 Data Model: m365 CLI v1

Domain types for `internal/domain`. No Graph SDK fields, URLs, HTTP statuses, or token strings.

## Session

**Fields**: signed in (bool), session usable (bool), account display name, mail consented (bool), teams consented (bool).

**Invariants**:
- Session MUST NOT contain access tokens, refresh tokens, authorization codes, or client secrets.
- Signed out ⇒ session usable is false and both consent flags are false.
- `auth status` while signed out is a successful observation of this state.

**Transitions**:
- `signed_out → signed_in` after successful interactive login (session usable true; consent flags reflect granted namespaces).
- `signed_in → expired` when silent reuse fails; workload commands then fail auth (exit 4).
- `signed_in|expired → signed_out` after logout (secret store empty).

## Consent

**Fields**: mail (bool), teams (bool).

**Rules**:
- Mail use cases MAY run when mail is true even if teams is false.
- Teams use cases MAY run when teams is true even if mail is false.
- Missing flag on a workload command is an auth failure, not a service failure.

## MailMessage

**Fields**: id, conversation id, subject, from (name, address), to[], cc[], received time, is read, has attachments, body text (optional), attachments[] (metadata only).

**Rules**:
- Body present only when requested (`mail get`, `mail thread --bodies`).
- Attachment list empty is valid.
- Id unknown ⇒ not-found.

## MailThread

**Fields**: conversation id, items[] of MailMessage, count.

**Rules**:
- Oldest first.
- Includes the signed-in user's own replies.
- Without `--bodies`, items omit body text.

## Chat

**Fields**: id, topic, type, members[] (name, address), last message preview (optional).

**Rules**:
- Unknown id ⇒ not-found.

## ChatMessage

**Fields**: id, chat id, sender, created time, text, attachments[] (metadata only).

**Rules**:
- Page oldest-first.
- System/membership events omitted unless `--include-system`.
- Unknown id in a known chat ⇒ not-found.

## Attachment

**Fields**: id, display name, size (bytes), content type, parent (mail message id, or chat id + message id).

**Outbound extra**: local filesystem path.

**Rules**:
- Metadata never includes file bytes.
- Outbound path MUST be a local file (not a URL).
- Outbound size MUST be `1..10485760` (10 MiB). Zero-byte files are invalid.
- Outbound count per send/reply MUST be `1..10` when any attach flags are present; zero attach flags is valid.
- Two outbound files MAY share a base name if paths differ.
- Two inbound attachments sharing a display name MUST be saved by id; save-by-name is usage failure.
- Parent namespace MUST match the command namespace.

## ResultLimit

**Fields**: applied top (int), count (int), next page token (opaque string, empty if none).

**Rules**:
- Mail list default top = 10. Teams list/messages/watch default top = 20.
- Maximum top = 50. Values `<1` or `>50` are usage failures, not silent caps.
- Token is not a service URL. Domain and CLI treat it as an opaque string.

## CommandResult

**Fields**: stdout payload, stderr diagnostic, exit class (`0`, `3`, `4`, `5`, `6`).

**Rules**:
- Success data on stdout only.
- Failure: one diagnostic on stderr (JSON object in JSON mode; plain text in human mode).
- Watch success in JSON mode: zero or more JSON objects as lines on stdout.

## WatchCheckpoint

**Fields**: map of chat id → last emitted message id and created time.

**Rules**:
- No tokens, bodies, or bytes.
- File mode `0600`.
- Advances only through emitted events.
- `--since` seeds the first run when no checkpoint exists.

## MailDeltaChange and MailDeltaPage

**MailDeltaChange fields**: mail message, opaque revision, removed flag.

**MailDeltaPage fields**: changes[], next token for the current round, terminal
delta token for the next invocation.

**Rules**:
- A non-removed change MUST have a non-empty revision. Graph partial updates
  without `changeKey` (for example read-state-only `{id, isRead}` items) are
  not changes; the adapter skips them so the round can complete.
- A page has either a next token or a terminal delta token, never both.
- Tokens remain internal to the Graph adapter, application use case, and
  protected state. They MUST NOT appear in command output.
- Removed changes do not emit `mail.changed` events.

## MailWatchState

**Fields**: terminal cursor, map of message id to latest revision, revision ids
in oldest-first order.

**Rules**:
- State is isolated by normalized account and folder.
- At most 5,000 revisions are retained; oldest ids are pruned first.
- Missing state is empty. Malformed or unreadable state is an error.
- The state directory is mode `0700`; the file and same-directory temporary
  are mode `0600`; replacement is atomic after the temporary is synced.
- State contains no bodies, attachment bytes, or credentials.

## MailWatchEvent

**Fields**: event (`mail.changed`), message id, conversation id, received time,
subject, sender.

**Rules**:
- One event represents the newest changed message in one conversation.
- Events are sorted by received time ascending, then message id.
- Events contain no body, attachments, revision, or Graph cursor.
- JSON mode writes one event per line; no changes is successful empty stdout.

## DryRunSend

**Fields**: destination (recipients or chat id), subject (mail), body/text, attachments[] of {name, size}, html/md flags.

**Rules**:
- Observing a dry-run MUST NOT create a remote message or upload bytes.

## Validation summary

| Rule | Failure class |
| --- | --- |
| Missing Client ID/Tenant ID config | usage (3) |
| `--top` out of range | usage (3) |
| Missing required flags/ids/body/path | usage (3) |
| Attach missing/unreadable/empty/URL/oversize/too many | usage (3) |
| Save path omitted, unwritable, exists without `--overwrite`, ambiguous name | usage (3) |
| No session, expired session, missing namespace consent | auth (4) |
| Microsoft 365 throttle or other service rejection of a usable session | service (5) |
| Unknown folder, message, chat, or attachment | not-found (6) |
