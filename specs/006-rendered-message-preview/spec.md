# Feature Specification: Rendered Message Preview and Format Checks

**Feature Branch**: `006-rendered-message-preview`

**Created**: 2026-09-27

**Status**: Draft

**Input**: User description: "~/.work/m365/SDO-563-rendered-message-preview/design.md" (SDO-563: preview the exact rendered message and block badly formatted sends)

## Specification Contract

### Objective

Mason, or an agent through the CLI or MCP, can trust m365 to send an email, reply, or Teams message that will not look odd or broken to the recipient. A dry-run shows exactly what will be sent. A body with a format problem is stopped before it is sent. In the terminal, a preview box shows roughly how the message will look.

### Goals

- A dry-run of `mail send`, `mail reply`, or `teams send` returns the exact content type and content that a real send then delivers, not the caller's raw input.
- Mail, replies, and Teams messages share one body model: plain text, `--format md`, or `--html`, all delivered as HTML.
- Every rendered body and mail subject is checked against a fixed set of format rules. A dry-run lists the problems; a real send with any problem fails as a usage error and sends nothing.
- `--preview` draws a text box of the message in the terminal.

### Non-goals

- Renaming or changing the meaning of `--dry-run`, on any verb or in the MCP write gate.
- Calendar event bodies.
- Tables, `style`, and `class` in any body mode.
- An override flag that sends a message despite format problems.
- Fetching the original message when previewing a reply.
- Three known side bugs, tracked separately and started after this feature: reply/Teams dropping attachments (SDO-564), `--human` errors printed as JSON (SDO-565), and the acceptance runner passing unmatched steps (SDO-566).

### Verification Strategy

Verification MUST cover, for each of `mail send`, `mail reply`, and `teams send`: dry-run `rendered` equals what the send delivers for the same input; each format rule with a passing and a failing case; clean plain and markdown input never tripping a rule, for both mail and Teams; a real send with problems exiting `3` with no delivery; the preview box at a fixed width, with and without problems; and `--preview --json` exiting `3`. New acceptance scenarios for preview, dry-run `rendered`, and a blocked send MUST prove that every step matched a handler. Fixtures MUST be synthetic.

## Document Map

This feature extends the `mail` and `teams` namespaces from [001-m365-cli](../001-m365-cli/spec.md) and the MCP write path from [004-m365-mcp](../004-m365-mcp/spec.md). It keeps their error schema and exit classes. The approved design, the rendering probe, and its results live in the SDO-563 workstream (`~/.work/m365/SDO-563-rendered-message-preview/`: `design.md`, `probe/results.md`).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Dry-run shows exactly what will be sent (Priority: P1)

Mason or an agent runs `mail send`, `mail reply`, or `teams send` with `--dry-run` and gets back the exact body that a real send would deliver: its content type and its content, after plain text or markdown has been turned into HTML. Plain mail goes out as HTML too, so markdown-looking text and paragraph breaks are handled the same way for mail, replies, and Teams.

**Why this priority**: Today a dry-run echoes the raw input, so it can look fine while the delivered message does not. Every other part of this feature depends on the preview and the send sharing one rendered body.

**Independent Test**: For each of the three commands and each body mode, run the command with `--dry-run` and without it against a fake service. The `rendered` content from the dry-run equals, character for character, the body the fake service received.

**Acceptance Scenarios**:

1. **Given** a plain-text mail body with two paragraphs, **When** the user runs `mail send --dry-run`, **Then** the JSON includes `rendered` with `content_type` `html` and the body as two paragraphs, and nothing is sent.
2. **Given** the same input without `--dry-run`, **When** the user sends, **Then** the delivered body has content type HTML and content identical to the dry-run's `rendered.content`.
3. **Given** a reply comment, **When** the user runs `mail reply --dry-run`, **Then** `rendered.content` is exactly the comment string the real reply sends, and `rendered.content_type` is `html`.
4. **Given** a Teams message with two paragraphs, **When** the user runs `teams send --dry-run`, **Then** `rendered.content` is one paragraph with the two parts separated by a blank line, exactly as the real send delivers it.
5. **Given** a clean body, **When** any of the three commands runs with `--dry-run`, **Then** the JSON includes `format_problems` as an empty list.

---

### User Story 2 - Badly formatted messages are stopped before they are sent (Priority: P1)

Every rendered body, and every mail subject, is checked against a fixed set of format rules. A dry-run lists every problem found. A real send with any problem fails with a usage error that says how many problems there are and suggests `--preview`, and nothing is sent. There is no flag to send anyway; a rule that misfires is fixed, not bypassed.

**Why this priority**: Agents write `--html` bodies with broken tags, leftover markdown, and literal `\n`. Today these go out unchecked and the only safeguard is guidance text. This is the part that makes sends trustworthy.

**Independent Test**: For each rule, send a body that breaks it with `--dry-run` (problem listed, exit `0`) and without it (exit `3`, fake service receives nothing). Send a clean body without `--dry-run` and confirm it is delivered.

**Acceptance Scenarios**:

1. **Given** an `--html` body containing an unclosed tag, **When** the user runs `mail send --dry-run`, **Then** the command exits `0` and `format_problems` lists one problem naming the broken-HTML rule and the tag.
2. **Given** the same body without `--dry-run`, **When** the user sends, **Then** the command exits `3` with the usage error shape, a message giving the problem count and the problem, a hint to run with `--preview`, and no message is sent.
3. **Given** a mail subject containing `**urgent**`, **When** the user sends, **Then** the send is refused with a leftover-markdown problem for the subject.
4. **Given** plain or `--format md` input with no mistakes, **When** it is rendered for mail or Teams, **Then** no format problem is reported.
5. **Given** a Teams send through MCP with write opt-in and a body with problems, **When** the tool runs, **Then** the result is an MCP error of the usage class and nothing is sent.
6. **Given** markdown syntax inside a code span or code block, **When** the body is checked, **Then** it is not reported as leftover markdown.

---

### User Story 3 - See the message in the terminal before sending (Priority: P2)

Mason adds `--preview` to `mail send`, `mail reply`, or `teams send` and sees a text box in the terminal: the message headers, a rule, and the body approximated as text, followed by any format problems, one per line. `--preview` never sends.

**Why this priority**: JSON `rendered` content is exact but hard for a person to read. The box lets Mason check a message at a glance. It depends on Story 1's rendered body and Story 2's problem list.

**Independent Test**: Run `--preview` for each command at a fixed width with a clean body and with a problem body, and compare the output to expected text exactly. Confirm nothing is sent.

**Acceptance Scenarios**:

1. **Given** a mail with To, Cc, Subject, and an attachment, **When** the user runs `mail send --preview`, **Then** the box shows To, Cc, Subject, and Attachments lines, a rule, then the body; nothing is sent; exit `0`.
2. **Given** a reply with reply-all, **When** the user runs `mail reply --preview`, **Then** the header reads "Reply to message <id>" with "(reply all)", and the original message is not fetched.
3. **Given** a Teams send to a chat or `--to` name, **When** the user runs `teams send --preview`, **Then** the header shows the chat or name and any attachments.
4. **Given** a body with paragraphs, a bulleted list, a numbered list, a heading, a link, and a code block, **When** it is previewed, **Then** paragraphs are separated by a blank line, bullets show as `•`, numbered items keep their numbers, headings are underlined, links show as `label (url)`, and code is indented.
5. **Given** a body with format problems, **When** it is previewed, **Then** the problems are listed below the box, one per line, and the command exits `0`.
6. **Given** `--preview --json`, **When** the command runs, **Then** it exits `3` with a usage error.
7. **Given** output is not a terminal, **When** `--preview` runs, **Then** the box is wrapped at 80 columns.

---

### User Story 4 - Write mail in markdown (Priority: P2)

Mason or an agent writes a mail or reply body in the same markdown subset Teams already accepts, with `--format md`, and it arrives as formatted HTML instead of literal `**` and `#`.

**Why this priority**: Agents naturally write markdown. With one body model, markdown mail becomes safe to send. It is useful on its own but less critical than the preview and the checks.

**Independent Test**: Send `**bold**`, a heading, a list, and a link with `mail send --format md --dry-run` and confirm the rendered HTML has the matching elements and no leftover markdown.

**Acceptance Scenarios**:

1. **Given** a markdown body, **When** the user runs `mail send --format md --dry-run`, **Then** `rendered.content` contains bold, heading, list, and link elements, and `format_problems` is empty.
2. **Given** the same body on `mail reply --format md`, **When** it runs, **Then** the reply comment is rendered the same way.
3. **Given** markdown prose containing `<` or `&` outside code, **When** it is rendered, **Then** those characters are escaped and shown literally, and raw HTML in markdown shows as text.
4. **Given** `--format` with any value other than `md`, **When** any of the three commands runs, **Then** it exits `3`.

---

### Edge Cases

- An empty body, a body that is only whitespace, or an empty mail subject: a usage error raised before dry-run, as for missing input today. This is not a format problem.
- `--preview` with `--dry-run`: allowed; `--preview` already implies dry-run.
- `--preview` passed through MCP: refused at the MCP boundary as a usage error, like `mcp serve --human`. MCP stays JSON-only.
- MCP secret redaction: through MCP, `rendered.content` is exact except where token-like text is replaced with `[redacted]`. The redaction stays.
- Reply body: the reply sends no content type field; `content_type: "html"` in `rendered` is reported because the comment renders as HTML.
- Teams paragraphs: joined by a blank line inside one paragraph; three or more line breaks in a row are still a format problem.
- Teams web shows a leading space after each line break. This is cosmetic and MUST NOT be reported as a problem.
- Code blocks on Teams for iOS lose their line structure, and a horizontal rule shows as a blank gap. These are known client limits, not format problems, and the body is not rewritten for them.
- A `--html` body that sent before this feature may now be refused. This is intended; the hint points to `--preview`.
- A terminal narrower than the headers: lines wrap to the terminal width.
- `--dry-run` on the other eight write verbs: unchanged.

## Requirements *(mandatory)*

### Functional Requirements

#### Rendering

- **FR-001**: `mail send`, `mail reply`, and `teams send` MUST each render the body once, before the dry-run decision. The dry-run result and the real send MUST use that same rendered body.
- **FR-002**: Real sends MUST deliver the rendered content type and content unchanged. For a reply, the rendered content MUST be the exact comment string sent.
- **FR-003**: Mail bodies MUST always be delivered as HTML. Plain mail input MUST be rendered to HTML the same way reply and Teams plain input are.
- **FR-004**: `mail send` and `mail reply` MUST accept `--format md`, using the same markdown subset and conversion as `teams send --format md`. The three body modes on all three commands are plain (default), `--format md`, and `--html`.
- **FR-005**: `--format` MUST reject any value other than `md`, and `--html` together with `--format md` MUST be rejected, both with exit `3`, on all three commands. This replaces the current Teams behaviour, where `--html` silently wins.
- **FR-006**: In `--format md`, `<` and `&` outside code MUST be escaped, so raw HTML written in markdown shows literally.
- **FR-007**: Rendering MUST take the target into account. For mail, each paragraph MUST be its own paragraph block. For Teams, a plain body MUST be one paragraph, with paragraphs separated by a blank line (two line breaks). In a Teams markdown body, adjacent paragraphs MUST be merged the same way. Headings, lists, code blocks, quotes and rules stay separate blocks.

#### Dry-run output

- **FR-008**: Dry-run JSON for the three commands MUST add `rendered` with `content_type` and `content`, and `format_problems` as a list of `{rule, detail}` objects, empty when clean. Existing dry-run fields MUST remain.
- **FR-009**: A dry-run MUST succeed (exit `0`) whether or not format problems are found.

#### Format rules

- **FR-010**: Every rendered body, in every mode, MUST be checked against these rules. Each violation is one format problem:
  1. A tag not on the allow-list: `p br h1 h2 h3 ul ol li pre code a strong b em i u s blockquote hr`.
  2. Any attribute other than `href` on `<a>`.
  3. Broken HTML: unclosed, mismatched, or stray end tags.
  4. Leftover markdown in visible text: `**bold**`, a line starting with `#`, `##`, or `###` followed by a space, `[label](url)`, or `` `code` ``. A line starting with `- ` is allowed. Text inside `<code>` or `<pre>` is exempt.
  5. Literal escape sequences `\n`, `\t`, or `\"` in the text.
  6. An empty paragraph, or three or more consecutive line breaks.
  7. An `href` whose scheme is not `http`, `https`, or `mailto`.
- **FR-011**: The mail subject MUST be checked against rules 4 and 5, and MUST be flagged if it contains a newline.
- **FR-012**: Clean plain and `--format md` input MUST never produce a format problem, for either mail or Teams.
- **FR-013**: A real send with one or more format problems MUST fail before anything is sent, with the existing usage error shape: class `usage`, exit `3`, message "N format problems: …" (listing them; "1 format problem: …" when there is one), and a hint "run with --preview to see them". There MUST be no override.
- **FR-014**: Through MCP, a real send with format problems MUST fail through the existing MCP error path with the same usage error. The MCP error schema MUST NOT change.

#### Preview

- **FR-015**: `mail send`, `mail reply`, and `teams send` MUST accept `--preview`. It MUST imply `--dry-run`, never send, and always draw the preview box, with or without `--human`.
- **FR-016**: The preview box MUST contain header lines, a rule, and the body approximated as text:
  - Mail: To, Cc, Subject, Attachments.
  - Reply: "Reply to message <id>", plus "(reply all)" when set. The original message MUST NOT be fetched.
  - Teams: the chat id or `--to` name, Attachments.
  - Body: a blank line between paragraphs, `•` for bullets, numbers for ordered lists, underlined headings, links as `label (url)`, indented code.
- **FR-017**: The box MUST be wrapped to the terminal width, or to 80 columns when output is not a terminal. It MUST NOT use colour.
- **FR-018**: Format problems MUST be listed below the box, one per line. The preview MUST exit `0` even when problems exist.
- **FR-019**: `--preview` together with `--json` MUST exit `3`. `--preview` through MCP MUST be refused at the MCP boundary as a usage error, whatever its value (including `false`).
- **FR-020**: `--help` for the three commands MUST list `--preview` and `--format md`.

#### Unchanged behavior

- **FR-021**: `--dry-run` MUST keep its name and meaning on all 11 write verbs and in the MCP write gate. The error JSON schema and exit classes from 001 and 004 MUST NOT change.
- **FR-022**: MCP secret redaction MUST continue to apply to `rendered` content.
- **FR-023**: Tests and fixtures MUST use synthetic messages. Live mailbox or chat content MUST NOT be stored in the repository.

### Key Entities

- **Rendered body**: What will be delivered: a content type (`html`) and content. For a reply, the content is the comment string.
- **Format problem**: One rule violation: the rule name and a detail saying where or what (for example, the tag or the text).
- **Target**: Mail or Teams. Decides how paragraphs are rendered.
- **Body mode**: Plain (default), markdown (`--format md`), or HTML (`--html`).
- **Preview**: Header lines, a rule, the body as wrapped text, and the list of format problems.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For 100% of tested inputs across the three commands and three body modes, the dry-run `rendered` content matches the delivered body character for character (outside MCP redaction).
- **SC-002**: 0 messages with a format problem are delivered: every failing case for every rule is refused with exit `3` and no delivery.
- **SC-003**: 0 format problems are reported for clean plain or markdown input, across all test fixtures for mail and Teams.
- **SC-004**: Markdown mail sent with `--format md` arrives with no literal `**`, `#`, or `[label](url)` in the visible text.
- **SC-005**: A person can check a message's headers, body, and problems from a single `--preview` screen, without reading JSON.
- **SC-006**: Every new acceptance scenario fails if any of its steps has no matching handler.
- **SC-007**: `--dry-run` behavior, the error schema, and exit classes of all other verbs are unchanged, confirmed by the existing suite passing.
- **SC-008**: No output contains access tokens, refresh tokens, authorization codes, or client secrets.

## Assumptions

- The allow-list comes from a live probe of Outlook (web, Mac, iOS) and Teams (web, Mac, iOS) on 2026-09-27. It is observed, not published by Microsoft; a future client change may make a passing body look odd. The probe script can be re-run.
- Blockquote stays allowed: Outlook shows it as an indent without a bar, which is readable.
- Code blocks stay allowed despite Teams for iOS flattening them; the limit is recorded, not linted.
- MCP callers whose `--html` bodies used to send may now get usage errors. This is intended.
- The markdown subset used by `teams send --format md` is reused, with three changes: `<` and `&` are escaped outside code (FR-006); inline `` `code` `` becomes a code span, so ordinary markdown doesn't trip the leftover-markdown rule; and Teams paragraphs are merged as in FR-007.
- `--preview` is terminal output for people; agents use `--dry-run` JSON.
- Client ID, Tenant ID, and secrets MUST NOT appear in this specification, source, tests, logs, or fixtures.
