# Tasks: Rendered Message Preview and Format Checks

**Input**: Design documents from `/specs/006-rendered-message-preview/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ (command-catalog.md, format-rules.md, preview.md, json-dry-run.schema.json), quickstart.md, `.specify/memory/constitution.md`, approved design `~/.work/m365/SDO-563-rendered-message-preview/design.md`

**Tests**: MANDATORY per Constitution I. Every slice is: SCAFFOLD commit (if needed) → write failing tests → run them and confirm every failure is an assertion failure (no compile, import, or fixture errors) → RED commit with tests and fixtures only → smallest GREEN change → GREEN commit. Locked tests are never edited afterwards without human approval. Exact-string assertions only in new tests (no `strings.Contains`). Synthetic bodies only; no live Graph, mailbox, or chat content.

**Scaffold convention (approved in T003)**: If a new test needs a symbol that doesn't exist yet, a SCAFFOLD commit comes before the RED commit. It adds only types, fields and function signatures whose bodies return zero values: no behaviour and no tests. The RED tests then compile and fail on their assertions, not on compilation (Constitution I).

**Do not**: touch calendar or files packages; fix SDO-564 (reply/Teams attachments), SDO-565 (`--human` errors) or SDO-566 (runner passes unmatched steps — only the new feature files get a guard); add an override flag; rename `--dry-run`; add colour; fetch the original message for a reply preview; bump the `go` directive in go.mod.

**Toolchain**: `export PATH=/opt/homebrew/bin:$PATH` (Go 1.27.1). Gate: `make verify`.

**Organization**: Setup + Foundational block all stories. US1–US4 map to spec User Stories 1–4. MVP = Setup + Foundational + US1 + US2 (both P1; the feature's promise needs both).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallel (different files, no incomplete deps)
- **[Story]**: `[US1]`–`[US4]` on user-story phases only
- Exact file paths required

## Path Conventions

`internal/domain/msgbody/` (new core package), `internal/app/mail/`, `internal/app/teams/`, `internal/adapters/graph/`, `internal/adapters/cli/`, `acceptance/runtime/`, `acceptance/steps/`, `features/mail/`, `features/teams/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Branch, baseline, and the human approval needed before any locked test moves.

- [X] T001 Create branch `006-rendered-message-preview` from `main` (378698e) with `git switch -c 006-rendered-message-preview`; never stage the pre-existing unrelated changes under `.specify/` and `.claude/`
- [X] T002 Record the baseline in the Notes section of specs/006-rendered-message-preview/tasks.md: result of `make verify` at 378698e, and binary size from `go build -o /tmp/m365-base ./cmd/m365 && stat -f %z /tmp/m365-base`
- [X] T003 First, find the tests affected by delivering plain mail as HTML: in a scratch change (never committed), make internal/adapters/graph/httpsend.go send mail as `contentType: HTML` with `plainTextToHTML(in.Body)`, run `go test ./...`, note every failing test by name, then `git restore` the file. Then STOP and request explicit human approval (Constitution I) for the items below, and record the approval text and date in Notes. If approval is refused, stop and record the blocker.
  - **(a)** Move internal/adapters/graph/mdhtml_test.go to internal/domain/msgbody/markdown_test.go. Change only the package line and the renamed function calls; no expectation changes.
  - **(b)** In US1's RED commit, change `TestHTTPTeamsSendTextHTMLMDPayloads` and `TestHTTPMailReplyKeepsQuotedThread` in internal/adapters/graph/httpsend_test.go. They will pass a pre-rendered `Rendered` value and assert it is sent unchanged. Drop the "html wins over md" case, because the CLI now rejects that flag pair (US4).
  - **(c)** Each existing test named by the scratch run above.
  - **(d)** The scaffold convention in this file's header.

**Checkpoint**: Branch exists, baseline recorded, approval recorded.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The pure `msgbody` package: types, rendering for both targets, and the `Problem` type. The Graph adapter calls it with today's output so every locked test passes unchanged. This blocks every story.

**⚠️ CRITICAL**: No user story work until this phase is complete.

### Refactor and scaffold

- [ ] T004 REFACTOR, no behaviour change:
  - `git mv internal/adapters/graph/mdhtml.go internal/domain/msgbody/markdown.go`.
  - `git mv internal/adapters/graph/mdhtml_test.go internal/domain/msgbody/markdown_test.go`, applying only T003(a).
  - Export `PlainTextToHTML` and `MDSubsetToHTML`, and call them from internal/adapters/graph/httpsend.go.
  - Run `make unit` (green), then REFACTOR commit.
- [ ] T005 SCAFFOLD, then SCAFFOLD commit:
  - internal/domain/msgbody/render.go: `Mode` (`Plain`, `Markdown`, `HTML`), `Target` (`Mail`, `Teams`), `Rendered{ContentType string \`json:"content_type"\`; Content string \`json:"content"\`}`, and `Render` returning `Rendered{}`.
  - internal/domain/msgbody/problem.go: `Problem{Rule, Detail}` (JSON `rule`, `detail`), `Problems`, the rule-id constants `tag-not-allowed attribute-not-allowed broken-html leftover-markdown literal-escape extra-blank-lines link-scheme newline-in-subject`, and `Err()` returning `nil`.

### Tests (RED)

- [ ] T006 [P] Write RED exact-output table tests in internal/domain/msgbody/render_test.go for `Render(mode, target, src) Rendered`. Every result has `ContentType == "html"`. Cases:
  - Plain/Mail: `"a\n\nb\nc"` → `"<p>a</p>\n<p>b<br>c</p>"`.
  - Plain/Teams: the same input → `"<p>a<br><br>b<br>c</p>"`.
  - Plain: `"a\n \n\t\nb"` → two paragraphs, not `<br>` runs (research R4).
  - Plain escaping: `"x < y & z"` → `"<p>x &lt; y &amp; z</p>"`.
  - Plain: `"   "` → `Content == ""`.
  - Markdown/Mail: `#`/`##`/`###` headings, `**bold**`, `-`/`*`/`1.` lists, `[docs](https://example.com)`, fenced code, and inline `` `x` `` → `<code>x</code>` (research R3).
  - Markdown escaping: `"a <b> & c"` → `"<p>a &lt;b&gt; &amp; c</p>"`; raw HTML shows literally (FR-006).
  - Markdown/Teams: adjacent paragraphs merge into one `<p>` joined by `<br><br>`, while a list or heading between them stays a separate block (FR-007, research R5).
  - HTML mode: returned byte-for-byte for both targets.
- [ ] T007 [P] Write RED tests in internal/domain/msgbody/problem_test.go for `Problems` (marshals `[]`, never `null`, when empty) and `Problems.Err()`:
  - `Class` is `usage`.
  - Message for one problem: `1 format problem: broken-html: unclosed <p>`.
  - Message for two: `2 format problems: broken-html: unclosed <p>; leftover-markdown: **bold**`.
  - Hint: `run with --preview to see them` (contracts/format-rules.md "Blocked send error").
- [ ] T008 Run `go test ./internal/domain/msgbody/...` and confirm every failure is an assertion failure, with no compile or import errors. RED commit containing only T006–T007

### Implementation (GREEN)

- [ ] T009 Implement `Render` in internal/domain/msgbody/render.go. Paragraph layout depends on the target only (data-model.md)
- [ ] T010 Split internal/domain/msgbody/markdown.go into plain.go (plain text → blocks, whitespace-only lines treated as blank) and markdown.go:
  - Escape `<` and `&` outside code.
  - Convert inline code.
  - Return blocks so render.go can join Teams paragraphs.
  - Keep `PlainTextToHTML` and `MDSubsetToHTML` exported with their current signatures; the moved locked tests call them.
  - Keep each file under 250 lines. Imports: stdlib only.
- [ ] T011 Implement `Problems` JSON (empty → `[]`) and `Err()` returning `*domain.Error` in internal/domain/msgbody/problem.go
- [ ] T012 In internal/adapters/graph/httpsend.go, call `msgbody.Render(mode, msgbody.Mail, text)` for mail reply **and** Teams. The Mail target reproduces today's `<p>` blocks, and the locked adapter inputs contain no `<`, `&` or backticks, so every locked test passes unchanged. Keep the adapter's "`--html` wins" choice until US1. **No test edits**
- [ ] T013 Run `make unit`, then GREEN commit T009–T012

**Checkpoint**: `msgbody` renders both targets; `make unit` green; no user-visible behaviour change yet.

---

## Phase 3: User Story 1 — Dry-run shows exactly what will be sent (Priority: P1) 🎯 MVP

**Goal**: The three commands render once in the app layer before the dry-run branch. Dry-run JSON carries `rendered` and `format_problems` (empty for now). The adapter sends `Rendered` unchanged. Mail plain text is always HTML.

**Independent Test**: For each command and each body mode, the dry-run `rendered.content` equals the payload captured by `captureJSONServer` for the same input. The mail `contentType` is `HTML`, the Teams one is `html`, and a reply sends only `comment`.

### Scaffold and tests for User Story 1 (MANDATORY)

- [ ] T014 [US1] SCAFFOLD, then SCAFFOLD commit:
  - Add `Rendered msgbody.Rendered` to `mail.SendInput`, `mail.ReplyInput` (internal/app/mail/ports.go) and `teams.SendInput` (internal/app/teams/ports.go), plus `MD bool` on the two mail inputs. Nothing reads or sets them yet.
  - Add `World.Sent int` and `Matches(text string) bool` returning `false` in acceptance/runtime/runtime.go.
- [ ] T015 [P] [US1] Write RED tests in internal/app/mail/rendered_test.go (new file; do not edit the locked send_test.go):
  - `Send` and `Reply` dry-runs return `rendered` = `msgbody.Rendered{ContentType: "html", Content: "<p>b</p>"}` for body `b`, and `format_problems` = empty `msgbody.Problems`. Existing keys `dry_run to subject body attachments` stay unchanged.
  - A real send passes a store input whose `Rendered` equals the dry-run value.
  - A body of `"   "` gives the same usage error as a missing body (research R8).
- [ ] T016 [P] [US1] Write RED tests in internal/app/teams/rendered_test.go:
  - `SendMapped` dry-run returns `rendered` with `"<p>a<br><br>b</p>"` for text `"a\n\nb"`.
  - The same holds on the `--to Alice` path.
  - A real send's store input has the identical `Rendered`.
  - Whitespace-only text is a usage error.
- [ ] T017 [P] [US1] Write RED parity tests in internal/adapters/graph/rendered_parity_test.go. For `mail send`, `mail reply` and `teams send` × plain / markdown / html modes:
  - Call the app use case with `DryRun: true`, then with `DryRun: false` against an `HTTPClient`/`HTTPTeams` backed by `captureJSONServer`.
  - Assert `rendered.content` equals `message.body.content` (with `contentType == "HTML"`), the `comment` string (and no `message` key), or `body.content` (with `contentType == "html"`), respectively.
  - In the same RED commit, apply the T003(b)-approved edits to internal/adapters/graph/httpsend_test.go.
- [ ] T018 [P] [US1] Write RED CLI tests in internal/adapters/cli/rendered_test.go:
  - `mail send --to user@example.com --subject t --body b --dry-run`, `mail reply msg-1 --body b --dry-run` and `teams send chat-1 --text b --dry-run` print JSON with `rendered` = `{"content_type":"html","content":"<p>b</p>"}` and `format_problems` = `[]`.
  - Without `--dry-run`, `Memory.Sent[0]`'s `Rendered.Content` is `<p>b</p>`.
- [ ] T019 [P] [US1] Write a RED test in internal/adapters/cli/mcp_rendered_test.go: `m365_run` for `mail send`, without write opt-in and with body `access_token=abc123`, returns dry-run JSON whose `rendered.content` is exactly `<p>[redacted]` and contains no `abc123` (FR-022, SC-008). Redaction consumes up to the next `"`, so the closing `</p>` is swallowed; that is existing redaction behaviour and stays
- [ ] T020 [P] [US1] Write Gherkin in features/mail/rendered.feature and features/teams/rendered.feature. Scenarios run a dry-run with single-token bodies (the `I run` step splits on spaces) and assert `stdout JSON "rendered.content" is "<p>b</p>"` and `stdout JSON "format_problems" is []`
- [ ] T021 [P] [US1] Write RED guard test acceptance/steps/new_features_test.go. It parses each file in a list (start with features/mail/rendered.feature and features/teams/rendered.feature; US2 and US3 append theirs) using `runtime.Parse`, and fails naming any step for which `runtime.Matches(text)` is false (research R9)
- [ ] T022 [US1] Confirm every failure in T015–T021 is an assertion failure, with no compile errors, then RED commit with tests and features only

### Implementation for User Story 1

- [ ] T023 [US1] Implement `Matches` in acceptance/runtime/runtime.go using the same prefix rule as `dispatch`. `dispatch` behaviour stays unchanged (SDO-566 is separate)
- [ ] T024 [US1] In internal/app/mail/ports.go:
  - After `ValidateOutbound`, set `Rendered = msgbody.Render(mode, msgbody.Mail, body)`. An empty `Rendered.Content` is the existing missing-body usage error.
  - Add `rendered` and `format_problems` (`msgbody.Problems{}`) to `dryPayload`.
  - Pass `Rendered` to the store.
- [ ] T025 [US1] In internal/app/teams/ports.go, render with target `msgbody.Teams` after `--to` resolution and `ValidateOutbound`, and add `rendered` and `format_problems` to the dry-run map. Keep the file under 250 lines
- [ ] T026 [US1] In internal/adapters/graph/httpsend.go:
  - Mail sends `{"contentType":"HTML","content":in.Rendered.Content}`.
  - Reply sends `{"comment": in.Rendered.Content}`.
  - Teams sends `{"contentType":"html","content":in.Rendered.Content}`.
  - Remove the `msgbody.Render` calls from T012 and the `HTML`/`MD` branching.
- [ ] T027 [US1] In acceptance/steps/rendered_steps.go, add the handler `stdout JSON "<dotted.path>" is <json>` (exact JSON value comparison). Set `w.Sent` from the memory adapter in the `I run` handler in acceptance/steps/registry.go
- [ ] T028 [US1] Run `make unit` and `make acceptance`, then GREEN commit T023–T027

**Checkpoint**: Parity holds for all 9 command × mode pairs; existing suites green.

---

## Phase 4: User Story 2 — Badly formatted messages are stopped before they are sent (Priority: P1)

**Goal**: Lint the rendered body (rules 1–7) and the mail subject. Dry-run lists the problems; a real send with problems is a usage error (exit 3) and sends nothing, in the CLI and through MCP.

**Independent Test**: Each rule has a failing case that is listed on dry-run (exit 0) and blocked on send (exit 3, `Memory.Sent` empty). Clean plain and markdown fixtures produce zero problems for both targets.

### Scaffold and tests for User Story 2 (MANDATORY)

- [ ] T029 [US2] SCAFFOLD: add `Lint(content string) Problems` in internal/domain/msgbody/lint_html.go and `LintSubject(subject string) Problems` in internal/domain/msgbody/lint_text.go, both returning `nil`; SCAFFOLD commit
- [ ] T030 [P] [US2] Write RED table tests in internal/domain/msgbody/lint_html_test.go, one pass and one fail case per rule, asserting exact `Problems` (contracts/format-rules.md):
  - Rule 1: `<table>`, `<span>`, `<div>` → `tag-not-allowed`. Every allow-listed tag passes: "`p br h1 h2 h3 ul ol li pre code a strong b em i u s blockquote hr`".
  - Rule 2: `style`/`class` on any tag, and any attribute other than `href` on `a`, → `attribute-not-allowed`.
  - Rule 3: `broken-html`, with details `unclosed <p>`, `</em> closes <strong>` and `stray </p>`. `<br>`, `<br/>` and `<hr>` need no end tag.
  - Rule 6: `extra-blank-lines`, with details `empty <p>` and `3 <br> in a row`, including `<br> <br> <br>`. Exactly `<br><br>` passes.
  - Rule 7: `link-scheme` for `javascript:`, `data:`, a relative href and an empty href. `http`, `https` and `mailto` pass.
  - Tag and attribute names are compared case-insensitively.
- [ ] T031 [P] [US2] Write RED table tests in internal/domain/msgbody/lint_text_test.go:
  - Rule 4 `leftover-markdown`: `**x**`, a line starting with `# `/`## `/`### `, `[l](u)` and `` `x` `` fail. A line starting with `- ` passes. The same text inside `<code>` or `<pre>` passes.
  - Rule 5 `literal-escape`: `\n`, `\t` and `\"` fail outside code and pass inside code.
  - Details are cut to 40 characters.
  - `LintSubject`: `**urgent**` → `leftover-markdown` with detail `subject: **urgent**`; `a\nb` → `newline-in-subject`.
  - Subject problems are listed before body problems.
- [ ] T032 [P] [US2] Write the invariant test internal/domain/msgbody/clean_test.go over fixtures internal/domain/msgbody/testdata/clean/*.txt (plain) and *.md (markdown). Fixtures cover paragraphs, line breaks, whitespace-only lines, `<`/`&` in prose, lists, headings, links, inline and fenced code, and a line starting with `- `. For both `Mail` and `Teams`, `Lint(Render(...).Content)` must be empty. It passes against the scaffold; it guards the GREEN implementation
- [ ] T033 [P] [US2] Write RED app tests in internal/app/mail/format_check_test.go and internal/app/teams/format_check_test.go:
  - An `--html` body `<p>hi` on dry-run → `format_problems` = `[{broken-html, unclosed <p>}]` and no error.
  - The same body sent for real → `*domain.Error` with class usage, message `1 format problem: broken-html: unclosed <p>` and the hint, and the store received nothing.
  - A mail subject `**urgent**` is blocked the same way.
- [ ] T034 [P] [US2] Write RED CLI and MCP tests:
  - internal/adapters/cli/format_check_test.go: `mail send ... --html --body <p>hi` → exit 3, stderr exactly `{"class":"usage","message":"1 format problem: broken-html: unclosed <p>","hint":"run with --preview to see them"}`, `Memory.Sent` empty.
  - internal/adapters/cli/mcp_format_test.go: `m365_run` with write opt-in and `{"html":true,"body":"<p>hi"}` → `IsError` of the usage class, nothing sent. Without opt-in → dry-run JSON listing `broken-html`.
- [ ] T035 [P] [US2] Write Gherkin in features/mail/format-check.feature and features/teams/format-check.feature:
  - A blocked send (`--html --body <p>hi`) → `exit code 3` and `nothing was sent`.
  - A dry-run → `stdout JSON "format_problems.0.rule" is "broken-html"`.
  - Append both files to the list in acceptance/steps/new_features_test.go.
- [ ] T036 [US2] Confirm every failure in T030–T035 is an assertion failure, with no compile errors, then RED commit with tests, fixtures and features only

### Implementation for User Story 2

- [ ] T037 [US2] Add `golang.org/x/net` (v0.59.0, or the newest version whose `go` directive ≤ 1.25.0) with `go get`. Confirm go.mod's `go 1.25.0` is unchanged, and record the version in Notes (research R13)
- [ ] T038 [US2] Implement `Lint` and rules 1, 2, 3, 6 and 7 in internal/domain/msgbody/lint_html.go using `html.Tokenizer` with an explicit open-tag stack, not `html.Parse` (research R7). `Lint` also calls the text rules from lint_text.go
- [ ] T039 [US2] Implement rules 4 and 5 over visible, unescaped text tokens outside `code`/`pre`, plus `LintSubject`, in internal/domain/msgbody/lint_text.go. Keep each file under 250 lines
- [ ] T040 [US2] Wire linting into internal/app/mail/ports.go and internal/app/teams/ports.go in the order from research R8. Dry-run returns the problems; otherwise, if there are problems, return `problems.Err()` before calling the store. Subject linting applies to `mail send` only
- [ ] T041 [US2] In acceptance/steps/rendered_steps.go, add the `nothing was sent` step (asserts `w.Sent == 0`) and array-index support in the dotted JSON path (`format_problems.0.rule`)
- [ ] T042 [US2] Run `make unit` and `make acceptance`, then GREEN commit T037–T041

**Checkpoint**: The MVP is complete. Dry-run is exact and badly formatted sends are blocked in the CLI and MCP.

---

## Phase 5: User Story 3 — See the message in the terminal before sending (Priority: P2)

**Goal**: `--preview` on the three commands draws the box described in contracts/preview.md, lists problems below it, never sends, and is rejected with `--json` and through MCP.

**Independent Test**: Golden-text tests at width 80 for mail, reply and Teams, with and without problems, match exactly, and nothing is sent. `--preview --json` exits 3.

### Scaffold and tests for User Story 3 (MANDATORY)

- [ ] T043 [US3] SCAFFOLD: add `Text(content string, width int) string` returning `""` in internal/domain/msgbody/text.go, and `drawPreview(headers []string, body string, problems msgbody.Problems, width int) string` returning `""` and `termWidth(w io.Writer) int` returning `0` in internal/adapters/cli/preview.go; SCAFFOLD commit
- [ ] T044 [P] [US3] Write RED table tests in internal/domain/msgbody/text_test.go for `Text`, covering every row of the "Body approximation" table in contracts/preview.md:
  - Paragraphs and `<br><br>` → a blank line.
  - `•` bullets and numbered lists.
  - Headings underlined with `=` (h1) or `-` (h2, h3).
  - Links → `label (url)`, or just `url` when the label equals the url.
  - `pre` indented 4 spaces and not wrapped.
  - `blockquote` → `> `; `hr` → `─`; entities decoded.
  - Word-wrap at `width`, splitting over-long words.
- [ ] T045 [P] [US3] Write RED golden tests in internal/adapters/cli/preview_test.go:
  - `drawPreview` reproduces the contracts/preview.md golden example exactly at width 40.
  - Width is clamped to at least 20.
  - Run through `Run` with buffer stdout (width 80), compare against internal/adapters/cli/testdata/preview/{mail,mail-problems,reply,reply-all,teams,teams-to,teams-problems}.golden.
  - The header lines match the table in contracts/preview.md: Cc and Attachments are omitted when empty; the reply header is `Reply to message msg-1 (reply all)`; with `--to`, the Teams header is `To: Alice (chat chat-1)`.
  - `Memory.Sent` is empty in every case.
- [ ] T046 [P] [US3] Write RED CLI tests in internal/adapters/cli/preview_flags_test.go:
  - `--preview --json` → exit 3 with message `use only one of --preview or --json`.
  - `--preview --human` and `--preview --dry-run` → the same box, exit 0.
  - A body with problems → problems listed after the box, exit 0.
  - `--help` for `mail send`, `mail reply` and `teams send` lists `--preview`.
- [ ] T047 [P] [US3] Write a RED MCP test in internal/adapters/cli/mcp_preview_test.go: `m365_run` with flags `{"preview": true}` and with `{"preview": false}` → usage error, message `preview is a terminal flag`, hint `use dry-run; its JSON has rendered and format_problems`
- [ ] T048 [P] [US3] Write Gherkin in features/mail/preview.feature and features/teams/preview.feature:
  - `--preview` succeeds, `stdout contains "Reply to message msg-1"` / `stdout contains "Chat: chat-1"`, and `nothing was sent`.
  - `--preview --json` → `exit code 3`.
  - Append both files to acceptance/steps/new_features_test.go.
- [ ] T049 [US3] Confirm every failure in T044–T048 is an assertion failure, with no compile errors, then RED commit with tests, goldens and features only

### Implementation for User Story 3

- [ ] T050 [US3] Implement `Text` in internal/domain/msgbody/text.go using the `x/net/html` tokenizer, keeping it under 250 lines
- [ ] T051 [US3] Add `golang.org/x/term` (v0.46.0, or the newest with `go` directive ≤ 1.25.0) and record it in Notes. Implement internal/adapters/cli/preview.go:
  - `termWidth`: terminal columns when `w` is an `*os.File` that is a terminal, else 80; minimum 20.
  - `drawPreview`, drawn with `┌─┐│├┤└┘`.
  - Header builders for mail, reply and Teams, reading `chat_id`, `rendered` and `format_problems` from the dry-run payload.
  - No colour.
- [ ] T052 [US3] Create internal/adapters/cli/bodyflags.go with the shared `--preview` handling (sets `DryRun`), and wire it into `mailSend` and `mailReply` in internal/adapters/cli/mail.go and `teamsSend` in internal/adapters/cli/teams.go. When `--preview` is set, write the box instead of calling `success`. Add only flag and call lines to mail.go (already 270 lines; refactor note in plan.md)
- [ ] T053 [US3] In internal/adapters/cli/run.go, add `"--preview": true` to `boolFlags` and return the usage error `use only one of --preview or --json` when both appear, next to the existing `--json`/`--human` check
- [ ] T054 [US3] In `buildRunArgs` in internal/adapters/cli/mcp_dispatch.go, reject a `preview` key in the flags map, whatever its value, with `&domain.Error{Class: domain.ClassUsage, Message: "preview is a terminal flag", Hint: "use dry-run; its JSON has rendered and format_problems"}`
- [ ] T055 [US3] Add `--preview` (show the message as text; never sends; not with `--json`) to `mailSendHelp`, `mailReplyHelp` and `teamsSendHelp` in internal/adapters/cli/help.go
- [ ] T056 [US3] Add the `stdout contains "<text>"` step to acceptance/steps/rendered_steps.go if no existing handler matches it
- [ ] T057 [US3] Run `make unit` and `make acceptance`, then GREEN commit T050–T056

**Checkpoint**: The preview works for all three commands; JSON and MCP paths are unchanged apart from the preview rejection.

---

## Phase 6: User Story 4 — Write mail in markdown (Priority: P2)

**Goal**: `mail send` and `mail reply` accept `--format md`. On all three commands, `--format` rejects any value other than `md`, and `--html` with `--format md` is a usage error.

**Independent Test**: `mail send --format md --dry-run` with a heading, bold, a list, a link and inline code renders the exact HTML with no problems. `--format html` exits 3.

No scaffold: these tests go through `Run`, and the `MD` field they use was added in T014.

### Tests for User Story 4 (MANDATORY)

- [ ] T058 [P] [US4] Write RED CLI tests in internal/adapters/cli/format_md_test.go:
  - `mail send --format md --body-file testdata/md/basic.md --dry-run` and `mail reply msg-1 --format md --body-file testdata/md/basic.md --dry-run` → exact `rendered.content` for internal/adapters/cli/testdata/md/basic.md, and `format_problems` = `[]`.
  - `--format html` on each of the three commands → exit 3, message `unsupported --format "html"; only md`.
  - `--html --format md` → exit 3, message `use --html or --format md, not both`.
  - `--help` for all three commands lists `--format md`.
- [ ] T059 [P] [US4] Write RED app tests in internal/app/mail/format_md_test.go: with `MD: true`, `Send` and `Reply` render with `msgbody.Markdown` for the `Mail` target
- [ ] T060 [US4] Confirm every failure in T058–T059 is an assertion failure, with no compile errors, then RED commit with tests and fixtures only

### Implementation for User Story 4

- [ ] T061 [US4] Add `bodyMode(html bool, format string) (msgbody.Mode, error)` in internal/adapters/cli/bodyflags.go, returning the two usage errors above (research R6)
- [ ] T062 [US4] Add a `--format` flag to `mailSend` and `mailReply` in internal/adapters/cli/mail.go, and use `bodyMode` there and in `teamsSend` in internal/adapters/cli/teams.go. Set `MD` on the app inputs, and select the mode in internal/app/mail/ports.go
- [ ] T063 [US4] Add `--format md` (markdown subset: headings, bold, lists, links, inline code, fenced code) to `mailSendHelp` and `mailReplyHelp` in internal/adapters/cli/help.go, and mention inline code in `teamsSendHelp`
- [ ] T064 [US4] Run `make unit`, then GREEN commit T061–T063

**Checkpoint**: The three body modes behave the same on all three commands.

---

## Phase 7: Polish & Cross-Cutting Concerns

- [ ] T065 [P] Write the RED test `TestLargeBodyDeadline` (256 KiB synthetic markdown body: Render + Lint + Text finish within 1 s) and `BenchmarkRenderLintText` in internal/domain/msgbody/perf_test.go. Run `go test -bench BenchmarkRenderLintText ./internal/domain/msgbody` and record ns/op in Notes. A result of 100 ms or more is a defect and blocks completion (Constitution VI)
- [ ] T066 [P] Build the branch binary (`go build -o /tmp/m365-new ./cmd/m365 && stat -f %z /tmp/m365-new`), compare it with the T002 baseline, and record both in Notes. Growth over 1.5 MiB is a defect and blocks completion (Constitution VI)
- [ ] T067 [P] Additive docs:
  - docs/m365.md: add `--format md` and `--preview` to the `mail send` and `mail reply` rows, `--preview` to the `teams send` row, and a sentence that dry-run JSON includes `rendered` and `format_problems` and that problems block a real send.
  - README.md: one line on `--preview` and on format problems blocking a send.
- [ ] T068 [P] In internal/adapters/cli/mcp_write_recipes.go, add text only: inline code in the md subset, and "a send with format problems fails with a usage error; dry-run shows `rendered` and `format_problems`". Run the locked recipe tests (`go test ./internal/adapters/cli -run Recipe`) to confirm they still pass
- [ ] T069 Check file sizes with `wc -l` on every touched .go file. Split any new file over 250 lines; record a note in Notes for internal/adapters/cli/mail.go (pre-existing 270). A REFACTOR commit is allowed only while the suite is green
- [ ] T070 Run the manual scenarios 1–8 in specs/006-rendered-message-preview/quickstart.md with `M365_FAKE=1`, and record the outcomes in Notes
- [ ] T071 Run `make verify` (fmt, vet, unit, race, coverage, gosec, govulncheck, acceptance, crap). All gates must pass; record the result in Notes

---

## Dependencies & Execution Order

- **Setup (T001–T003)** → **Foundational (T004–T013)** → **US1 (T014–T028)** → **US2 (T029–T042)** → **US3 (T043–T057)** and **US4 (T058–T064)** → **Polish (T065–T071)**.
- T003 (human approval) blocks T004 and T017.
- In every phase: REFACTOR/SCAFFOLD commit → tests → RED commit → implementation → GREEN commit.
- US2 depends on US1: `format_problems` and the render call sit in the app use case that US1 creates.
- US3 depends on US1 (`rendered`) and US2 (problems shown below the box).
- US4 depends on US1 only. US3 and US4 both edit internal/adapters/cli/bodyflags.go, mail.go, teams.go and help.go, so run them one after the other (US3 first, or US4 first; neither needs the other).

## Parallel Examples

- **Foundational**: T006 and T007 are different test files and can be written together.
- **US1**: after T014, T015–T021 are all different files.
- **US2**: after T029, T030, T031 and T032 (msgbody), T033 (app), T034 (CLI/MCP) and T035 (Gherkin) run together. T038 and T039 are different files and can run in parallel after T037.
- **US3**: after T043, T044–T048 run together. T050 (msgbody) and T054 (MCP) are independent of T051–T053.
- **US4**: T058 and T059 run together.
- **Polish**: T065, T066, T067 and T068 run together.

## Implementation Strategy

1. **MVP (P1)**: Setup → Foundational → US1 → US2. Stop and validate: parity tests, per-rule tables, the clean-input invariant, blocked sends in the CLI and MCP, and `make verify`. This delivers the ticket's core promise: dry-run is the truth, and bad bodies never go out.
2. **Increment 2**: US3 (`--preview`) for people at the terminal.
3. **Increment 3**: US4 (`--format md` for mail).
4. **Polish**: performance and binary-size measurements, docs, recipe text, and the final `make verify`.

## Notes

- Baseline (T002), 2026-09-27, branch `006-rendered-message-preview` at 378698e:
  - Binary size: 14263618 bytes (`/tmp/m365-base`).
  - `make verify`: fmt, vet, unit, race, coverage, govulncheck, acceptance and crap pass. `gosec` needed `sh scripts/install-tools.sh` first (`.tools/` was missing).
  - Pre-existing toolchain issue: `gosec` v2.25.0 built with local Go 1.27.1 crashes before analysing anything (`internal error: package "context" without types was imported from "command-line-arguments"`). Built and run with `GOTOOLCHAIN=go1.26.6` (the module's toolchain line), it passes with 0 issues. Gate runs in this feature use `GOTOOLCHAIN=go1.26.6` for gosec. Fixing the tool setup is out of scope.
- Tests affected by HTML mail, and locked-test approval (T003):
  - Scratch run (mail sent as `contentType: HTML` with `plainTextToHTML`, then `git restore`): `go test ./...` had no failures, so item (c) names no tests.
  - Approval: 2026-09-27, the user chose "Approve all" for (a), (b), (c: none) and (d) in the session prompt.
- Dependency versions (T037, T051):
- Benchmark and binary size (T065, T066):
- File-size notes (T069):
- Quickstart outcomes (T070):
- `make verify` result (T071):
