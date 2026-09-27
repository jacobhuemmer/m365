# Implementation Plan: Rendered Message Preview and Format Checks

**Branch**: `main` (setup-plan JSON reported `006-rendered-message-preview`; working tree is `main`, branch at implement time) | **Date**: 2026-09-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/006-rendered-message-preview/spec.md`; approved design `~/.work/m365/SDO-563-rendered-message-preview/design.md` (based on `main@378698e`).

**Note**: Design only. No application source or `tasks.md`. Calendar and files packages are not touched. The three side bugs (SDO-564/565/566) are not fixed here.

## Summary

Move body rendering (plain → HTML, markdown subset → HTML) out of the Graph adapter into a new pure core package `internal/domain/msgbody`. The `mail` and `teams` app use cases render once, lint the result (and the mail subject), then either return a dry-run payload carrying `rendered` + `format_problems`, fail with a `usage` error when problems exist, or hand the already-rendered body to the store. Adapters send `Rendered` as-is. Mail always goes out as HTML; mail and reply gain `--format md`. The CLI adds `--preview` (implies dry-run, draws a text box from `msgbody.Text`), rejected with `--json` and at the MCP boundary. Lint uses `golang.org/x/net/html`'s tokenizer; terminal width comes from `golang.org/x/term`.

## Technical Context

**Language/Version**: Go, module `github.com/masonhuemmer/m365` (`go 1.25.0`, toolchain `go1.26.6`; Go 1.27.1 installed at `/opt/homebrew/bin/go`).

**Primary Dependencies**: Existing stdlib + `flag`. New: `golang.org/x/net/html` (v0.59.0 at plan time) for the lint tokenizer and preview text; `golang.org/x/term` (v0.46.0) for terminal width. `golang.org/x/sys` is already an indirect dependency. No markdown library, no terminal UI library.

**Storage**: N/A. No new files at runtime.

**Testing**: Go `testing`; `captureJSONServer` for Graph payloads; memory adapter (`graph.Memory.Sent`) for CLI tests; APS Gherkin under `features/mail/` and `features/teams/`; exact-string assertions and golden text for the preview.

**Target Platform**: macOS CLI and the in-process MCP server (`m365 mcp serve`).

**Project Type**: Single-module local CLI.

**Performance Goals**: No Graph call is added. Rendering, linting and preview text of a 256 KiB body complete in under 100 ms (test with a generous 1 s deadline to stay deterministic; benchmark reports the real number). Binary size grows by at most 1.5 MiB (see Make Targets).

**Constraints**: Exit classes `0/3/4/5/6` and error JSON schema unchanged; JSON stays default; MCP stays JSON-only; `--dry-run` unchanged on all 11 write verbs; no colour; no override flag; no Graph fetch for reply preview.

**Scale/Scope**: 3 commands (`mail send`, `mail reply`, `teams send`); 2 new flags (`--preview`, mail `--format md`); 7 body lint rules + subject rules; 1 new core package.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. TDD. PASS.** Each slice is SCAFFOLD (zero-value signatures only, so RED fails on assertions, not compilation) → locked RED → GREEN; moving the converter is a separate REFACTOR commit while green. Each slice's RED commit covers: rendering parity, each lint rule (pass+fail), clean-input invariant, blocked send, preview golden text, `--preview --json`, MCP rejection, step-match check. GREEN follows in its own commit.
- [x] **II. Clean Code. PASS with note.** New package split by responsibility (`plain.go`, `markdown.go`, `render.go`, `lint_html.go`, `lint_text.go`, `text.go`), each under 250 lines. **Refactoring note:** `internal/adapters/cli/mail.go` is already 270 lines at 378698e; this feature adds only flag lines there and puts shared body-flag parsing and preview drawing in new `cli/bodyflags.go` and `cli/preview.go`. It stays under 500; splitting it is not part of this feature. The existing `graph/mdhtml.go` (221 lines) is moved, not grown in place.
- [x] **III. Smallest Sufficient Design. PASS.** One pure package, no plugin rules or config. The tokenizer dependency replaces a hand-rolled parser (design decision 9). No override flag. Inline-code support and whitespace-line normalisation are added only because the clean-input invariant (FR-012) requires them (research R3, R4).
- [x] **IV. Testing. PASS.** Pure unit tests for `msgbody`; app tests with memory store; adapter parity tests with `captureJSONServer`; CLI golden tests at width 80; acceptance scenarios with a step-match guard (research R9). Synthetic bodies only.
- [x] **V. CLI Consistency. PASS.** Stdout/stderr unchanged; `--preview` is human output on stdout; errors keep the JSON shape; help lists new flags; `--format` rejects unknown values.
- [x] **VI. Performance. PASS.** Metrics named above: body processing time (256 KiB < 100 ms, measured by `BenchmarkRenderLintText`) and binary size (≤ +1.5 MiB vs 378698e, measured by `go build -o` and `stat`). Graph call volume unchanged or lower (blocked sends).
- [x] **VII. Isolation. PASS.** `msgbody` is a core package under `internal/domain/`, imports only stdlib and `x/net/html`. `mail` and `teams` both import it and still not each other (isolation tests unchanged). Graph URLs and content-type casing stay in the adapter.
- [x] **VIII. Secrets. PASS.** No scope changes. MCP redaction still applies to `rendered`. Fixtures synthetic.
- [x] **Engineering Constraints. PASS.** `make verify` (fmt, vet, unit, race, coverage, gosec, govulncheck, acceptance, crap) covers the new dependencies.
- [x] **Workflow. PASS.** Specify (done) → this check → RED → GREEN → REFACTOR → `make verify`.

No new constitution exceptions.

**Post-design re-check (after Phase 1): PASS.** Contracts keep the error schema; `rendered`/`format_problems` are additive dry-run fields; preview contract is terminal-only; data model has no I/O in `msgbody`.

## Project Structure

### Documentation (this feature)

```text
specs/006-rendered-message-preview/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── command-catalog.md
│   ├── format-rules.md
│   ├── preview.md
│   └── json-dry-run.schema.json
├── checklists/requirements.md
└── tasks.md                 # /speckit-tasks, not this command
```

### Source Code (implement time)

```text
internal/domain/msgbody/plain.go         # plainTextToHTML (moved), whitespace-line normalisation
internal/domain/msgbody/markdown.go      # mdSubsetToHTML (moved) + escaping + inline code, returns blocks
internal/domain/msgbody/render.go        # Mode, Target, Rendered, Render(); Teams paragraph join
internal/domain/msgbody/lint_html.go     # rules 1,2,3,6,7 via html.Tokenizer + tag stack
internal/domain/msgbody/lint_text.go     # rules 4,5 on visible text; LintSubject
internal/domain/msgbody/problem.go       # Problem, rule ids, Error() builder (usage + hint)
internal/domain/msgbody/text.go          # Text(content, width): body approximated as text
internal/domain/msgbody/*_test.go        # exact-output tables, clean-input invariant, benchmark
internal/app/mail/ports.go               # SendInput/ReplyInput: MD, Rendered; render+lint before dry-run
internal/app/teams/ports.go              # SendInput.Rendered; render+lint before dry-run
internal/adapters/graph/httpsend.go      # send in.Rendered as-is; conversions removed
internal/adapters/graph/mdhtml.go        # deleted (moved); mdhtml_test.go moved to msgbody
internal/adapters/graph/rendered_parity_test.go  # dry-run rendered == captured payload, per command
internal/adapters/cli/bodyflags.go       # --format validation, --html/--format conflict, --preview
internal/adapters/cli/preview.go         # header lines, box, wrap, problems list; termWidth
internal/adapters/cli/mail.go, teams.go  # wire flags; call preview writer when --preview
internal/adapters/cli/run.go             # boolFlags += --preview; --preview + --json usage error
internal/adapters/cli/mcp_dispatch.go    # reject `preview` flag at MCP boundary
internal/adapters/cli/help.go            # --preview and --format md in three help texts
internal/adapters/cli/testdata/preview/*.golden
acceptance/runtime/runtime.go            # World.Sent; exported Matches(text) (dispatch unchanged)
acceptance/steps/rendered_steps.go       # new step handlers
acceptance/steps/new_features_test.go    # every step in the new features matches a handler
features/{mail,teams}/rendered.feature
features/{mail,teams}/format-check.feature
features/{mail,teams}/preview.feature
```

**Structure Decision**: Same single module. One new core package; changes elsewhere are confined to mail/teams app use cases, the Graph send adapter, CLI wiring, MCP dispatch, and the acceptance harness.

## Graph mapping (adapter only)

- `POST /me/sendMail`: `body.contentType` is always `"HTML"`, `body.content` = `in.Rendered.Content`.
- `POST /me/messages/{id}/reply|replyAll`: `{"comment": in.Rendered.Content}`.
- `POST /me/chats/{id}/messages`: `body.contentType` `"html"`, `body.content` = `in.Rendered.Content`.
- Attachments behaviour unchanged (SDO-564 is separate).

## Make Targets

Reuse the Makefile; `make verify` is the gate. Measurements for Principle VI:

- `go test ./internal/domain/msgbody -run TestLargeBodyDeadline` (256 KiB, 1 s) and `go test -bench BenchmarkRenderLintText ./internal/domain/msgbody` (record ns/op in task notes; target < 100 ms).
- Binary size: `go build -o /tmp/m365-base ./cmd/m365` at 378698e and on the branch, compare with `stat -f %z`; fail if growth > 1.5 MiB.

## Out of This Plan

- SDO-564 (reply/Teams attachments), SDO-565 (`--human` errors), SDO-566 (acceptance runner passes unmatched steps: only the new features get a guard).
- Calendar bodies; golden tests for the converters beyond exact-output tables; `--note-to-self` client check.
- MCP recipe text changes beyond what `--preview` rejection needs.

## Risks And Tradeoffs

- **Allow-list is observed, not published** (design Open risks). Re-run `probe/probe.sh`; fix rules, never bypass.
- **Existing `--html` bodies may now be refused.** Intended; hint points to `--preview`.
- **Plain text with a literal backslash sequence** (for example a Windows path `C:\new`) trips rule 5, and plain mode has no code span to exempt it. Accepted per design; the detail names the text so the caller can switch to `--format md` with a code span.
- **Teams web leading space after `<br>`** is not lintable and not flagged.
- **`--html` + `--format md` together now fail** (research R6) where Teams used to let `--html` win silently.

## Complexity Tracking

No new constitution exceptions.
