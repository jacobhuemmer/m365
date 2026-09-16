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
