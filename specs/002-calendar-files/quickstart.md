# Quickstart validation: calendar and files

This is a run guide, not the test suite. Catalog: [contracts/command-catalog.md](contracts/command-catalog.md). Domain: [data-model.md](data-model.md). Fake: [contracts/fake-graph.md](contracts/fake-graph.md). v1 mail/teams checks remain in `specs/001-m365-cli/quickstart.md`.

## Prerequisites

- Go 1.25+
- Existing v1 CLI module (this feature extends it)
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

Use `M365_FAKE=1` and synthetic fixtures. Do not call a live mailbox or drive.

1. **Auth status keys** — after fake login with `fake-all`, `m365 auth status` exits `0`; `namespaces.mail`, `namespaces.teams`, `namespaces.calendar`, and `namespaces.files` are present; no token fields.
2. **Calendar list window** — `m365 calendar list` exits `0`; `limit` is 10; items are inside the default window; a recurring series does not dump extra occurrences.
3. **Calendar dry-run create** — `m365 calendar create --dry-run --subject t --start <rfc3339> --end <rfc3339>` exits `0`; `dry_run` true; fake has no new event.
4. **Files list empty/page** — `m365 files list` exits `0`; `limit` is 20; stdout is one JSON object; no file bytes.
5. **Download refuse without overwrite** — destination exists; `m365 files download file-1 --out /tmp/exists.txt` exits `3`; file unchanged.
6. **Mail/teams unchanged** — `m365 mail list --help` and `m365 teams list --help` still name the v1 flags and limits; `chat list` still aliases teams.

Unit: `make unit`. Acceptance: `make acceptance`.

## Interactive login (manual; not CI)

```sh
./bin/m365 auth login
./bin/m365 auth status
./bin/m365 calendar list --top 10
./bin/m365 files list --top 20
```

Expect calendar and files consent flags, a bounded event page, and a bounded folder page. Do not paste tokens or live file contents into chat or git.

## Stop

If `auth status` is not usable, stop and run `auth login` in a real terminal. Do not loop login in agents.
