# Data Model: Rendered Message Preview and Format Checks

All types live in `internal/domain/msgbody` unless noted. No persistence.

## Mode

Body input mode, chosen from CLI flags.

| Value | Flag | Rendering |
|-------|------|-----------|
| `Plain` | (default) | Escape, blank lines → paragraphs, single newlines → `<br>` |
| `Markdown` | `--format md` | Markdown subset → HTML; `<` and `&` escaped outside code |
| `HTML` | `--html` | Content used as given |

Validation (CLI boundary): `--format` must be `md` or absent; `--html` and `--format md` are mutually exclusive.

## Target

`Mail` or `Teams`. Mail send and mail reply use `Mail`; Teams send uses `Teams`. Affects paragraph layout only (research R5).

## Rendered

| Field | JSON | Notes |
|-------|------|-------|
| `ContentType` | `content_type` | Always `"html"` |
| `Content` | `content` | Exact delivered body. For a reply, the exact `comment` string |

Invariant: the value returned in dry-run is the value passed to the store on a real send (FR-001, FR-002).

## Problem

| Field | JSON | Notes |
|-------|------|-------|
| `Rule` | `rule` | One of the rule ids in [format-rules.md](contracts/format-rules.md) |
| `Detail` | `detail` | Where or what: the tag, attribute, text excerpt, or `subject: …` |

`Problems` is an ordered list (document order; subject problems first). Empty list, never null, in JSON. `Problems.Err()` builds the usage error (research R12).

## Preview (CLI adapter, `internal/adapters/cli/preview.go`)

| Part | Source |
|------|--------|
| Header lines | Mail: To, Cc (omitted when empty), Subject, Attachments. Reply: `Reply to message <id>` + ` (reply all)`, Attachments. Teams: `Chat: <chat_id>` or `To: <name> (chat <chat_id>)`, Attachments |
| Body | `msgbody.Text(Rendered.Content, width-4)` |
| Problems | `Problems`, printed below the box |

## Changed app inputs

| Type | New fields |
|------|------------|
| `mail.SendInput` | `MD bool`, `Rendered msgbody.Rendered` |
| `mail.ReplyInput` | `MD bool`, `Rendered msgbody.Rendered` |
| `teams.SendInput` | `Rendered msgbody.Rendered` (already has `MD`) |

`Rendered` is set only by the app use case, never by the CLI.

## Flow (per send)

```text
parsed flags → required fields → (teams: resolve --to) → validate files
  → Render(mode, target, raw) → empty? usage error
  → Lint(content) [+ LintSubject for mail send]
  → dry-run?  yes → payload {…existing, rendered, format_problems}   (exit 0)
              no  → problems? usage error (exit 3, nothing sent)
                    else store.Send(Rendered)
```
