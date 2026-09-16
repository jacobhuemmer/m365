# Fake Graph fixture contract

Tests and acceptance own an in-process HTTP server (`net/http/httptest`) that
the Graph adapter calls. No live tenant. No real mailbox or chat documents in
the repository.

## Identity

The fake accepts `Authorization: Bearer <token>` where `<token>` is a
synthetic value such as `fake-mail`, `fake-teams`, `fake-both`, or
`fake-expired`. Tokens MUST NOT resemble real JWTs copied from production.

| Token | Mail | Teams | Notes |
| --- | --- | --- | --- |
| `fake-both` | yes | yes | default signed-in fixture |
| `fake-mail` | yes | no | Teams routes return insufficient-consent (mapped to exit 4) |
| `fake-teams` | no | yes | Mail routes return insufficient-consent |
| `fake-expired` | no | no | mapped to exit 4 |
| missing/unknown | no | no | mapped to exit 4 |

## Resources tests MAY assume

Synthetic, stable ids (examples, not secrets):

- Mail folder `inbox` with zero or more messages `msg-1`, `msg-2`.
- Conversation `conv-1` containing `msg-1` then `msg-2` (oldest first), including a reply from the signed-in user.
- Chat `chat-1` (1:1), `chat-2` (group). Message `cmsg-1` in `chat-1`.
- Attachment `att-1` on `msg-1` and `cmsg-1`: name `note.txt`, size 12, content type `text/plain`. Bytes are the twelve characters `synthetic-ok`. Never a live document.
- Empty inbox / empty chat list fixtures.
- Unknown ids `missing-id` → not-found mapping (exit 6).
- Throttle fixture → service mapping (exit 5).

## Behaviors the fake MUST implement

- Honor `$top` and return an opaque next-page token when more synthetic items exist. MUST NOT require clients to send a raw next-link URL.
- `Prefer: outlook.body-content-type=text` returns body as text when present.
- Send/reply without a dry-run flag in the adapter creates a synthetic sent item; dry-run never reaches the fake as a create (use cases short-circuit).
- Attachment list returns metadata only. Content download is used only by save; tests assert the destination file, not stdout bytes.
- 429 and 5xx map to service errors. 404 maps to not-found. Consent misses map to auth.

## Forbidden

- Recording live Graph transcripts into `testdata/`.
- Embedding Client ID, Tenant ID, or real tokens.
- Serving OneDrive/SharePoint drive browsers.
