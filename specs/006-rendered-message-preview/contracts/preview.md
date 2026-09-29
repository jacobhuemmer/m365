# Preview Output Contract

`--preview` writes plain text to stdout. No colour, no ANSI escapes. Width `W` = terminal columns when stdout is a terminal, else 80; minimum 20. Every line of the box is exactly `W` columns wide; content is wrapped at word boundaries to `W - 4` (words longer than that are split).

## Layout

```text
┌──────────────────────────────────────────┐
│ <header line>                            │
│ …                                        │
├──────────────────────────────────────────┤
│ <body text>                              │
└──────────────────────────────────────────┘
<problems block, only when problems exist>
```

Problems block:

```text
2 format problems:
- broken-html: unclosed <p>
- leftover-markdown: **bold**
```

## Header lines

| Command | Lines |
|---------|-------|
| `mail send` | `To: a@example.com, b@example.com`; `Cc: …` (omitted when empty); `Subject: …`; `Attachments: note.txt (12 B)` (omitted when none) |
| `mail reply` | `Reply to message msg-1` or `Reply to message msg-1 (reply all)`; `Attachments: …` (omitted when none) |
| `teams send` | `Chat: chat-1`, or `To: Alice (chat chat-1)` when `--to` resolved; `Attachments: …` (omitted when none) |

## Body approximation (`msgbody.Text`)

| HTML | Text |
|------|------|
| `p`, and a `<br><br>` pair | Blank line between paragraphs |
| `br` | Line break |
| `h1`–`h3` | Heading text, then a line of `=` (h1) or `-` (h2, h3) of the same length |
| `ul > li` | `• item` |
| `ol > li` | `1. item`, `2. item`, … |
| `a href` | `label (url)`; just `url` when label equals url |
| `pre` | Each line indented 4 spaces, not wrapped |
| `code` inline | Text as-is |
| `strong b em i u s` | Text as-is |
| `blockquote` | Lines prefixed with `> ` |
| `hr` | A line of `─` |
| Entities | Decoded (`&lt;` → `<`) |

Disallowed or broken markup is shown best-effort as its text content; the problems list explains it.

## Golden example (W = 40)

Input: `m365 teams send chat-1 --format md --preview` with text (via `--text-file`; `⏎` marks a newline) `Hi **team**⏎⏎- one⏎- two⏎⏎See [docs](https://example.com)`.

```text
┌──────────────────────────────────────┐
│ Chat: chat-1                         │
├──────────────────────────────────────┤
│ Hi team                              │
│                                      │
│ • one                                │
│ • two                                │
│                                      │
│ See docs (https://example.com)       │
└──────────────────────────────────────┘
```
