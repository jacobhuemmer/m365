# Implementation Plan: m365 MCP Serve

**Branch**: `main` (setup-plan JSON reported `004-m365-mcp`; working tree is `main`) | **Date**: 2026-09-16 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/004-m365-mcp/spec.md`

**Note**: Design only. It does not add application source, `tasks.md`, or CI workflow files. It MUST NOT rewrite `specs/001-m365-cli/`, `specs/002-calendar-files/`, or `specs/003-calendar-natural-time/`.

## Summary

Add `m365 mcp serve` to the existing Go 1.25+ CLI so agents speak JSON-RPC on stdio to the same signed-in session the terminal uses. The catalog is three tools (`m365_status`, `m365_help`, `m365_run`). Each call runs in-process `cli.Run` with captured stdout/stderr (MCP already owns process stdio). Workload writes default to `--dry-run` unless `write_opt_in` is true. No second binary, no second token store, no generic Graph tool, no lazy-mcp. Mail, teams, calendar, and files command names stay unchanged.

## Technical Context

**Language/Version**: Go 1.25+ module `github.com/masonhuemmer/m365` (exists).

**Primary Dependencies**: Existing CLI stack (stdlib `flag`, `net/http`, `encoding/json`, `testing`; `golang.org/x/oauth2` and `go-keyring` stay in auth/keychain adapters). New: official `github.com/modelcontextprotocol/go-sdk` (`mcp` package, pin a v1 tag at implement; v1.8.0 observed 2026-09-16). No Cobra, no Graph SDK, no MSAL, no mark3labs MCP, no SSE/HTTP transport.

**Storage**: Unchanged token store (Keychain then `0600` session.json). MCP MUST NOT create another secret file. Download/save still write only to user-named paths.

**Testing**: Go `testing`; SDK `NewInMemoryTransports`; existing fake Graph and fake store; APS Gherkin under `features/mcp/`. Unit tests MUST NOT call live Graph. Synthetic fixtures only. Optional acceptance spawn of `m365 mcp serve` with `M365_FAKE=1`.

**Target Platform**: macOS local stdio MCP for Grok, Claude, Cursor, Codex, OpenCode. Dedicated server like `kata mcp serve`, not a lazy-mcp backend.

**Project Type**: Same single-module local CLI; MCP is a second interface of that binary.

**Performance Goals**: SC-002 — `m365_status` for a usable fake session finishes in under 2 seconds (CI: in-memory MCP + fake store). Default `mail list` through `m365_run` stays within the existing 15s fake-Graph list budget.

**Constraints**: Exit classes `0/3/4/5/6` remain on the CLI; MCP maps them to tool `isError` plus `{class,message,hint}` without collapsing. JSON stdout default on the CLI path. No secrets on the MCP wire, logs, fixtures, or this plan. Write opt-in default false. Compact catalog of three tools. No Enterprise MCP, no Agent365 remote, no plugin marketplace.

**Scale/Scope**: One namespace (`mcp`) with one verb (`serve`); three MCP tools wrapping existing 001–003 verbs.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Test-Driven Development (NON-NEGOTIABLE). PASS.** Every MCP slice starts RED (Gherkin and/or unit/MCP client tests). RED is a separate commit of failing tests and fixtures only. GREEN is the smallest production change. EX-I-001 does not apply.
- [x] **II. Clean Code. PASS.** New files SHOULD stay under 250 lines; over 250 needs a task note; over 500 MUST split. `cmd/m365` remains wiring. Split serve vs dispatch vs tests.
- [x] **III. Smallest Sufficient Design. PASS.** This spec adds MCP; it is no longer a forbidden pre-build. Three tools, in-process `Run`, no Graph-from-MCP, no HTTP transport, no per-verb tools. Official MCP SDK is the protocol seam only.
- [x] **IV. Testing Standards. PASS.** In-memory MCP + ports/fakes; no live Graph in unit tests. Contract tests cover `tools/list`, status, run JSON parity, dry-run writes, auth-class vs service-class. Acceptance Gherkin → APS → `acceptance/generated`. Synthetic fixtures. Assert observable JSON and `isError`, not SDK internals.
- [x] **V. CLI Experience Consistency. PASS.** Human CLI verbs/flags/exits unchanged (FR-012). `m365 mcp --help` exits 0 with no session and names serve, the three tools, and write opt-in. `mcp serve` stdout is JSON-RPC (the machine protocol); diagnostics on stderr; `--human` on serve is usage so JSON-RPC is not mixed with pretty text. `--help` remains the human surface.
- [x] **VI. Performance. PASS.** SC-002: metric = wall-clock for `m365_status`; threshold = 2s; CI = in-memory MCP + fake store. List runs still send `$top` through existing CLI limits. No silent mailbox dump.
- [x] **VII. Domain Isolation. PASS.** `mcp` is not a Graph workload. Do not add Graph ports under `internal/app/mcp`. Do not import mail/teams/calendar/files from each other. Graph URLs stay in `internal/adapters/graph`. Registration seam: `run.go` `case "mcp"`. Do not edit `mail.go`/`teams.go`/`calendar.go`/`files.go`.
- [x] **VIII. Secrets, Auth, Least Privilege. PASS.** Same delegated PKCE session. No new scopes. No login via MCP. Redact before MCP content. Fixtures use fake tokens only. Two MCP clients share the one CLI session store.
- [x] **Engineering Constraints. PASS.** Same Makefile gates. No surrounding refactors in bug fixes. Limits remain visible in CLI help that `m365_help` returns.
- [x] **Development Workflow. PASS.** Specify (done) → this Constitution Check → RED commit → GREEN commit → optional REFACTOR → review → `make verify`.

No new constitution exceptions. EX-I-001 remains historical for 001 T097–T119 only.

**Post-design re-check (after Phase 1): PASS.** `research.md`, `data-model.md`, `contracts/`, and `quickstart.md` keep Graph at adapters, MCP as a CLI door, classed tool errors, write opt-in dry-run, fake Graph as CI boundary, and no secrets in artifacts.

## Project Structure

### Documentation (this feature)

```text
specs/004-m365-mcp/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── mcp-catalog.md
│   ├── json-run-request.schema.json
│   ├── json-tool-error.schema.json
│   └── agent-registration.md
└── tasks.md                 # Future /speckit-tasks, not created by this plan
```

### Source Code (repository root)

```text
cmd/m365/                    # unchanged wiring (same Deps, including mcp serve)
internal/adapters/cli/
├── run.go                   # case "mcp" only
├── help.go                  # additive root/mcp help
├── mcp.go                   # NewMCPServer, serve, three tools
├── mcp_dispatch.go          # flag map → argv, write gate
├── mcp_test.go
└── mcp_perf_test.go         # status < 2s on fake
features/mcp/                # Gherkin
acceptance/steps/            # MCP stdio or in-process steps
```

**Structure Decision**: Same module and `cli` package (avoids mcp↔cli import cycle). No new Graph adapter. No `internal/app/mcp`.

## Implementation Workflow

1. Gherkin under `features/mcp/`.
2. APS generate; RED unit tests with `NewInMemoryTransports` + fake store/Graph; confirm expected failure.
3. RED commit (failing tests + fixtures only).
4. Smallest GREEN (`go get` SDK, serve, dispatch); GREEN commit; no locked RED-test edits.
5. Isolation: MCP path MUST NOT call Graph except through existing namespace `Run`.
6. `make verify` on fakes. Do not run live Graph in CI.

## Make Targets

Reuse the existing Makefile. Add a CLI/MCP perf test for `m365_status` under 2s against the fake (same pattern as list perf tests).

## Files To Add At Implement Time (not this command)

- `mcp.go`, `mcp_dispatch.go`, tests listed above.
- Additive root help line for `mcp`.
- Gherkin + step stubs.
- `go.mod` require of `github.com/modelcontextprotocol/go-sdk` (v1 tag).

## Out of This Plan

- Rewriting 001–003 specs or mail/teams/calendar/files command packages.
- `/speckit-tasks` and `/speckit-implement`.
- lazy-mcp, Microsoft Enterprise MCP, Agent365 remote MCP.
- HTTP/SSE MCP, resources, prompts, sampling.
- Interactive login from MCP.

## Risks And Tradeoffs

- **stdio ownership**: MUST capture CLI stdout or JSON-RPC breaks. Tests use in-memory transports, not the process stdio of `go test`.
- **SDK `error` return**: collapsing classes; handlers MUST return `CallToolResult{IsError: true}` instead.
- **`--body-file -`**: no stdin to steal; refuse with usage.
- **Write flag map `dry-run: false`**: MUST NOT defeat the MCP write gate.

## Complexity Tracking

No new constitution exceptions. EX-I-001 is not used for this feature.
