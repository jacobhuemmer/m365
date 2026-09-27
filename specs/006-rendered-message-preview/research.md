# Research: Rendered Message Preview and Format Checks

Background research (code trace, Graph docs, live probe) lives in the SDO-563 workstream: `~/.work/m365/SDO-563-rendered-message-preview/research.md` (Q1–Q7) and `probe/results.md`. The approved design (decisions 1–16) settles scope. This file records the decisions made while planning against `main@378698e`. No NEEDS CLARIFICATION items remain.

## R1: Where rendering lives

- **Decision**: New core package `internal/domain/msgbody` holding `Render`, `Lint`, `LintSubject`, and `Text`. `plainTextToHTML` and `mdSubsetToHTML` move there from `internal/adapters/graph/mdhtml.go`.
- **Rationale**: Design decision 1. Both `app/mail` and `app/teams` need it, and namespaces must not import each other (VII), so it belongs in a core package. A sub-package keeps the `x/net/html` import out of the flat `internal/domain` package, and avoids the name `body`, which collides with local variables in CLI code.
- **Alternatives**: Put it in `internal/domain` (adds an HTML dependency to every domain import); keep conversion in the adapter and re-run it for dry-run (two code paths can drift, the bug this feature fixes).

## R2: How the rendered body reaches the adapter

- **Decision**: `mail.SendInput`, `mail.ReplyInput` and `teams.SendInput` gain `Rendered msgbody.Rendered`, set by the app use case after lint. `mail.SendInput`/`ReplyInput` gain `MD bool`. Adapters read only `Rendered` for the body. The memory adapter records the input as today, so CLI tests see the rendered body in `Memory.Sent`.
- **Rationale**: One value flows from render to dry-run to send. Keeping raw `Body`/`Text` preserves existing required-field checks and dry-run fields.
- **Alternatives**: Overwrite `Body` with HTML and set `HTML=true` (changes a field's meaning mid-flow); return a typed dry-run struct (changes JSON key order of existing output).

## R3: Markdown inline code

- **Decision**: The markdown converter turns `` `x` `` into `<code>x</code>` (content escaped, no bold/link processing inside).
- **Rationale**: Lint rule 4 flags leftover `` `code` ``. Without conversion, ordinary clean markdown with an inline code span would fail FR-012 (clean input never trips a rule). Inline code is also what the design's exemption ("send literal markdown in a code span") relies on.
- **Alternatives**: Drop the backtick check from rule 4 (weakens the design's rule); leave it (every agent body with inline code is blocked).

## R4: Plain text whitespace-only lines

- **Decision**: Before splitting paragraphs, `plainTextToHTML` treats lines containing only spaces or tabs as blank.
- **Rationale**: Today `"a\n \n \nb"` becomes `a<br> <br> <br>b`, which trips rule 6 (3+ consecutive `<br>`) from clean input, violating FR-012.
- **Alternatives**: Let rule 6 ignore breaks separated by whitespace text (then real spacing abuse in `--html` passes).

## R5: Teams paragraph join with markdown blocks

- **Decision**: The markdown converter produces a list of blocks. For the Teams target, each run of adjacent paragraph blocks is merged into one `<p>` joined by `<br><br>`; headings, lists, `pre`, `blockquote` and `hr` stay separate blocks. Plain Teams input is always one `<p>`. Mail keeps one `<p>` per paragraph. Blocks are separated by `\n` as today.
- **Rationale**: Design decision 15 (Teams shows no gap between `<p>` blocks; `<br><br>` inside one `<p>` shows the gap on Mac, web and iOS). A list or heading cannot sit inside `<p>` without breaking HTML (rule 3).
- **Alternatives**: Put the whole md body in one `<p>` (invalid nesting); insert `<br><br>` between all blocks (breaks rule 3 outside `<p>`, and iOS adds leading spaces to text outside `<p>` per re-probe P1).

## R6: Body mode flags

- **Decision**: Modes are plain (default), `--format md`, `--html`. `--format` accepts only `md`; any other value, or `--html` together with `--format md`, is a usage error (exit 3) on all three commands. Parsing lives in `cli/bodyflags.go`.
- **Rationale**: FR-005; the design says new code rejects anything other than `md`. Allowing both flags and silently picking one is the pattern the design lists to avoid.
- **Alternatives**: Keep Teams' "`--html` wins" (silent precedence).

## R7: Lint parsing

- **Decision**: Use `html.Tokenizer` from `golang.org/x/net/html` with an explicit open-tag stack, not `html.Parse`. Void elements are `br` and `hr` (self-closing form allowed). All other allow-listed elements need an explicit end tag. Visible text for rules 4 and 5 is the tokenizer's text tokens, unescaped, excluding text inside `code`/`pre`.
- **Rationale**: `html.Parse` silently repairs broken markup, so it cannot detect rule 3. The tokenizer reports tags and text exactly as written.
- **Alternatives**: Regex scan (fragile, constitution III warns); hand-rolled tokenizer (design decision 9 rejects).

## R8: Where each check runs and in what order

- **Decision**: In the app use case: auth → existing required-field checks (and Teams `--to` resolution) → `ValidateOutbound(files)` → render → empty-after-render check → lint body (+ subject for `mail send`) → dry-run branch (returns payload with problems) → if problems: `usage` error → store call.
- **Rationale**: Missing input stays a usage error before dry-run (design 12). A body that renders to nothing (whitespace only) is treated as a missing body, so a blank message is never sent. Dry-run always succeeds (FR-009).
- **Alternatives**: Lint in the CLI (MCP and CLI would diverge; app tests could not cover it).

## R9: Proving acceptance steps match (SDO-566 not fixed here)

- **Decision**: Add exported `runtime.Matches(text string) bool` using the same prefix rule as `dispatch`, without changing `dispatch`. A unit test in `acceptance/steps` parses the three new feature files with `runtime.Parse` and fails if any step does not match a registered handler. `runtime.World` gains `Sent int` so a step can assert nothing was delivered.
- **Rationale**: Design decision 10 requires the new scenarios to prove their steps match, while the general fix is SDO-566.
- **Alternatives**: Fix `dispatch` to fail on unmatched steps (that is SDO-566 and may break existing features).

## R10: Preview output and terminal width

- **Decision**: `cli/preview.go` builds header lines from the parsed flags and the dry-run payload (`chat_id` for Teams), takes `rendered` and `format_problems` from the payload, calls `msgbody.Text(content, inner)` for the body, and draws a box with Unicode box-drawing characters. Width = terminal columns when stdout is an `*os.File` that is a terminal (`x/term`), else 80; minimum 20. The drawing function takes width as a parameter so golden tests are fixed-width.
- **Rationale**: No change to `Deps`; tests write to buffers and get 80.
- **Alternatives**: Add a `Width` field to `Deps` (wiring for one value); ASCII-only borders (less readable; no requirement for ASCII).

## R11: `--preview` with `--json`, and through MCP

- **Decision**: `run.go` returns a usage error when `--json` and `--preview` both appear, next to the existing `--json`/`--human` check. `buildRunArgs` in `mcp_dispatch.go` rejects a `preview` key in the flag map with a usage error whose hint points to dry-run's `rendered` and `format_problems`.
- **Rationale**: FR-019; design decision 5. The MCP boundary is where `auth login` and stdin files are already refused.
- **Alternatives**: Let MCP draw the box into JSON text (breaks JSON-only).

## R12: Error for a blocked send

- **Decision**: `msgbody.Problems.Err()` returns `&domain.Error{Class: usage, Message: "N format problems: rule: detail; …", Hint: "run with --preview to see them"}`, with "1 format problem" for one.
- **Rationale**: FR-013 keeps the existing error shape; MCP wraps it as `IsError` unchanged (FR-014).

## R13: Dependency versions

- **Decision**: Add `golang.org/x/net` (v0.59.0) and `golang.org/x/term` (v0.46.0), or the newest versions whose `go` directive does not exceed the module's `go 1.25.0`. If `go get` would raise the directive, pin an older version and record it in task notes.
- **Rationale**: Keep the module's Go floor unchanged; `govulncheck` in `make verify` covers both.
