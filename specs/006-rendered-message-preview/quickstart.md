# Quickstart: Rendered Message Preview and Format Checks

Validation guide. Contracts: [command-catalog](contracts/command-catalog.md), [format-rules](contracts/format-rules.md), [preview](contracts/preview.md), [dry-run schema](contracts/json-dry-run.schema.json). Types: [data-model](data-model.md).

## Prerequisites

- Go at `/opt/homebrew/bin/go` (1.27.1) on `PATH`.
- Repo tools for acceptance (`.tools/bin`, as `make acceptance` expects).
- No live account needed: every scenario below uses the fake store (`M365_FAKE=1`) or test fakes.

## Test suites

| Suite | Command | Covers |
|-------|---------|--------|
| Unit (render, lint, text, app, CLI golden, MCP) | `make unit` | FR-001–FR-020, per-rule pass/fail, clean-input invariant |
| Adapter parity | `go test ./internal/adapters/graph -run Rendered` | dry-run `rendered` == captured Graph payload, per command |
| Acceptance | `make acceptance` | preview, dry-run `rendered`, blocked send; step-match guard in `go test ./acceptance/steps` |
| Performance | `go test -bench BenchmarkRenderLintText ./internal/domain/msgbody` | 256 KiB body < 100 ms |
| Full gate | `make verify` | fmt, vet, unit, race, coverage, gosec, govulncheck, acceptance, crap |

## Manual scenarios (fake mode)

```sh
export M365_FAKE=1
go build -o /tmp/m365 ./cmd/m365
/tmp/m365 auth login
```

1. **Dry-run shows the delivered body**
   `printf 'Hello\n\nSecond paragraph' > /tmp/b.txt`
   `/tmp/m365 mail send --to a@example.com --subject Hi --body-file /tmp/b.txt --dry-run`
   → exit 0; `rendered` = `{"content_type":"html","content":"<p>Hello</p>\n<p>Second paragraph</p>"}`; `format_problems` = `[]`.

2. **Teams paragraphs join in one `<p>`**
   `/tmp/m365 teams send chat-1 --text-file /tmp/b.txt --dry-run`
   → `rendered.content` = `<p>Hello<br><br>Second paragraph</p>`.

3. **Markdown mail**

   ```sh
   printf '# Update\n\n**done** see `x`' > /tmp/m.md
   /tmp/m365 mail send --to a@example.com --subject Hi --body-file /tmp/m.md --format md --dry-run
   ```

   → `rendered.content` contains `<h1>Update</h1>`, `<strong>done</strong>`, `<code>x</code>`; no problems.

4. **Blocked send**
   `/tmp/m365 mail send --to a@example.com --subject Hi --html --body '<p>hi'`
   → exit 3; stderr `{"class":"usage","message":"1 format problem: broken-html: unclosed \u003cp\u003e","hint":"run with --preview to see them"}` (the error writer escapes `<` and `>`; the decoded message reads `unclosed <p>`); nothing sent.
   Same command with `--dry-run` → exit 0, `format_problems` lists `broken-html`.

5. **Subject check**
   `/tmp/m365 mail send --to a@example.com --subject '**urgent**' --body hi --dry-run`
   → `format_problems` has `{"rule":"leftover-markdown","detail":"subject: **urgent**"}`.

6. **Preview box**
   `/tmp/m365 mail reply msg-1 --all --body-file /tmp/b.txt --preview`
   → box with `Reply to message msg-1 (reply all)`, a rule, two paragraphs separated by a blank line; exit 0.
   `/tmp/m365 mail reply msg-1 --body hi --preview --json` → exit 3.
   `/tmp/m365 teams send chat-1 --text hi --preview | cat` → box is 80 columns wide.

7. **Flag validation**
   `/tmp/m365 teams send chat-1 --text hi --format html --dry-run` → exit 3.
   `/tmp/m365 mail send --to a@example.com --subject Hi --body hi --html --format md --dry-run` → exit 3.

8. **MCP**
   In an MCP session, `m365_run` for `mail send` with flags `{"preview": true}` → usage error. With write opt-in and `{"html": true, "body": "<p>hi"}` → `IsError`, usage class, nothing sent. Without opt-in → dry-run JSON with `rendered` and `format_problems`.

## Measurements to record in task notes

- Benchmark ns/op for `BenchmarkRenderLintText` (target < 100 ms).
- Binary size at 378698e vs branch (`go build -o … ./cmd/m365`, `stat -f %z`; growth ≤ 1.5 MiB).
