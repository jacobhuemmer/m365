# Phase 0 Research: m365 MCP Serve

All Technical Context items are resolved. Client ID, Tenant ID, and token values MUST NOT appear here.

## Decision: Official Go MCP SDK over stdio, not a hand-rolled JSON-RPC loop

**Rationale**: Spec FR-001 is stdio JSON-RPC `tools/list` and `tools/call`. The official module `github.com/modelcontextprotocol/go-sdk` (`mcp` package) already implements initialize, tools, `StdioTransport`, `NewInMemoryTransports` for tests, and `CallToolResult.IsError`. Observed latest tagged module on pkg.go.dev: v1.8.0 (2026-09-16). Implement pins a v1 tag; do not vendor a second protocol stack.

This SDK is a **protocol adapter**, not a Graph SDK and not a generic provider marketplace. Constitution III forbids pre-building a Graph escape hatch; it does not forbid speaking MCP.

**Alternatives considered**:
- Hand-rolled JSON-RPC 2.0 on stdin/stdout: fewer dependencies; easy to miss initialize, pagination, and cancellation.
- `mark3labs/mcp-go`: community SDK; rejected in favor of the official module.
- Microsoft Enterprise MCP or Agent365 remote Calendar/Mail MCP: out of spec (non-goals).
- lazy-mcp backend: out of spec (FR-013).

**Sources** (observed 2026-09-16):
- https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp
- https://github.com/modelcontextprotocol/go-sdk
- https://modelcontextprotocol.io/specification/2025-06-18/server/tools

## Decision: In-process `cli.Run` with captured buffers, never exec

**Rationale**: MCP already owns process stdin/stdout. Execing `m365 mail list` on the same streams would corrupt JSON-RPC. Spec FR-001 forbids a second binary and a second token store. Each tool call clones `Deps` with `bytes.Buffer` stdout/stderr, builds argv, and calls `Run`. The Keychain/file session is the same `Store`. Two `m365 mcp serve` processes share that store on disk/Keychain, not in-memory.

**Alternatives considered**:
- `exec.Command("m365", ...)`: stdio collision; extra binary; harder fakes.
- Reimplement mail/calendar use cases inside MCP: second Microsoft 365 client; violates FR-009.

## Decision: Exactly three tools; `mcp` is a CLI namespace, not a Graph workload

**Rationale**: FR-002. Kata-shaped catalog: `m365_status`, `m365_help`, `m365_run`. `m365 mcp serve` matches `m365 <namespace> <verb>`. Do not add `internal/app/mcp` Graph ports. Do not register resources, prompts, sampling, or SSE/HTTP.

**Alternatives considered**:
- One MCP tool per CLI verb or per Graph URL: forbidden by FR-002/FR-009.
- Separate `m365-mcp` binary: forbidden by FR-001.

## Decision: Tool execution errors keep CLI classes via `isError`

**Rationale**: MCP distinguishes protocol errors (unknown tool, invalid JSON-RPC) from tool errors (`CallToolResult.IsError`). Returning a Go `error` from the handler becomes a collapsed protocol error and would violate FR-006. Success: `isError` false; one `text` content item whose text is exactly CLI JSON stdout (help is CLI help text). Failure: `isError` true; one `text` content item that is the CLI stderr JSON object `{class, message, hint}` from `specs/001-m365-cli/contracts/json-error.schema.json`. Redact secrets before the wire (FR-011, SC-007).

**Alternatives considered**:
- Map exit 3/4/5/6 onto JSON-RPC codes: collapses class into a numeric code agents do not share with the CLI.
- `structuredContent` only: spec requires the same JSON the CLI would print; text of stdout is that JSON. Do not require `outputSchema` (CLI payloads vary by verb).

## Decision: Write opt-in injects `--dry-run`; it cannot be bypassed by flags

**Rationale**: FR-007 / US3. MCP callers are agents. Default `write_opt_in` false/omitted. Write verbs: mail `send`/`reply`; teams `send`; calendar `create`/`update`/`delete`; files `upload`/`create-folder`/`delete`/`move`. For those verbs, unless `write_opt_in` is JSON `true`, dispatch MUST add `--dry-run` even if the flag map sets `dry-run` false. If `write_opt_in` is true and the flag map sets `dry-run` true, CLI dry-run wins (no mutation). Reads ignore `write_opt_in`. Local filesystem writes (`files download`, `save-attachment`) are not workload writes; they still require `--out` and MUST NOT put file bytes on the MCP wire (FR-010).

**Alternatives considered**:
- Same default as the terminal (real send unless `--dry-run`): rejected; spec is stricter for agents.
- A fourth tool `m365_write`: rejected; compact catalog is three tools.

## Decision: Flag map is CLI flags without `--`; no generic Graph

**Rationale**: FR-005. `namespace`, `verb`, optional `args` (positionals), optional `flags` object. Keys are long flag names (`to`, `top`, `page-token`, `when`). Values: string, number, boolean, or string array (repeatable `--to`/`--attach`/`--attendee`). Booleans: `true` emits the flag; `false` omits it. `chat` aliases `teams`. Unknown namespace/verb/flag → usage, no Graph. `auth login`, `auth logout`, and `mcp` are usage with a terminal-login hint. `--body-file`/`--text-file` value `-` is usage (MCP has no CLI stdin). `teams watch` is the existing one-shot watch; result text is the CLI JSON-lines stdout.

**Alternatives considered**:
- Pass a raw argv string: injection-prone; worse for schemas.
- Refuse `teams watch`: FR-005 says run the equivalent command; watch already returns and does not daemonize.

## Decision: Same `cli` package files, registration seam only

**Rationale**: `internal/adapters/mcp` importing `cli` while `cli.Run` dispatches to MCP is an import cycle. Keep serve/dispatch in `internal/adapters/cli` (`mcp.go`, `mcp_dispatch.go`). `run.go` gains `case "mcp"`. Do not edit `mail.go` / `teams.go` / `calendar.go` / `files.go`. `cmd/m365` stays wiring.

**Alternatives considered**:
- Extract a `Runner` port into `internal/app`: extra layer with no behavior.
- New `internal/app/mcp` use cases that call Graph: forbidden.

## Decision: Tests use in-memory MCP plus fakes; acceptance may spawn stdio

**Rationale**: Constitution IV. Unit/contract tests: official SDK `NewInMemoryTransports` + existing fake Graph and fake store. Do not call live Graph. Status SC-002: wall-clock under 2s against the fake store. Mail list SC-003: MCP JSON vs `cli.Run` JSON, same ids/count. Acceptance Gherkin under `features/mcp/` may spawn `m365 mcp serve` with `M365_FAKE=1` and speak JSON-RPC on stdio. Synthetic fixtures only.

**Alternatives considered**:
- Live tenant MCP in CI: forbidden.
- Skip Gherkin: violates constitution I/IV.
