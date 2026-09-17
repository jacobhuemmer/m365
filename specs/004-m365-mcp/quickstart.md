# Quickstart validation: m365 MCP Serve

This is a run guide, not the test suite. Catalog: [contracts/mcp-catalog.md](contracts/mcp-catalog.md). Run request: [contracts/json-run-request.schema.json](contracts/json-run-request.schema.json). Domain: [data-model.md](data-model.md). CLI checks remain in `specs/001-m365-cli/quickstart.md` and later namespace quickstarts.

## Prerequisites

- Go 1.25+
- Existing CLI module (this feature extends it)
- Client ID and Tenant ID at runtime only (never in the repo)

```sh
export M365_CLIENT_ID='...'   # do not commit
export M365_TENANT_ID='...'   # do not commit
```

## Build

```sh
make verify
go build -o bin/m365 ./cmd/m365
```

CI MUST run `make verify`. No live tenant in CI.

## Automated first checks (fakes)

Use `M365_FAKE=1` and synthetic fixtures, or the in-memory MCP client against fake `Deps`. Do not call a live mailbox.

1. **Catalog** — `tools/list` returns exactly `m365_status`, `m365_help`, `m365_run`.
2. **Status signed-out** — `m365_status` succeeds; `signed_in` false; no browser; no token fields.
3. **Status signed-in** — after fake login, status names consent for mail, teams, calendar, and files; finishes under 2s locally (SC-002).
4. **Read parity** — `m365_run` namespace `mail` verb `list` (default flags) returns the same `count` and item `id`s as `m365 mail list` (SC-003). Repeat for `teams list`, `calendar list`/`free`, `files list` with consent.
5. **Unknown verb** — `mail` `nope` is `isError` with `class=usage`; no Graph call.
6. **Missing consent** — teams run without teams consent is `class=auth`, not `service` (SC-005).
7. **Write gate** — `mail send` without `write_opt_in` returns `dry_run` true and does not send. Same for `calendar create`. With `write_opt_in` true, send matches CLI without `--dry-run`.
8. **Help** — `m365_help` with no session names calendar verbs including list, create, and free.
9. **Recipes** — no session: help with no topic names the four topics; `topic=mail-search` includes `from:ajay`; `topic=teams-find` says not to send on several matches; `prompts/list` is those four names; `tools/list` stays three tools (SC-008, SC-009).
10. **Human CLI unchanged** — `m365 mail list --help` still names v1 flags and limits.

Unit: `make unit`. Acceptance: `make acceptance`.

## Interactive (manual; not CI)

```sh
./bin/m365 auth login
./bin/m365 auth status
./bin/m365 mcp --help
```

Register `command: m365`, `args: ["mcp", "serve"]` per [contracts/agent-registration.md](contracts/agent-registration.md). Do not paste tokens or live mail bodies into chat or git.

## Stop

If status is not usable, stop and run `auth login` in a real terminal. Do not loop login through MCP.
