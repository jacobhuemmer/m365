# Quickstart validation: teams find and send --to

Catalog: [contracts/command-catalog.md](contracts/command-catalog.md). Domain: [data-model.md](data-model.md). Fake: [contracts/fake-graph.md](contracts/fake-graph.md). v1 teams checks remain in `specs/001-m365-cli/quickstart.md`.

## Prerequisites

- Go 1.25+
- Existing CLI (001 teams)
- Client ID / Tenant ID at runtime only (never in the repo)

## Build

```sh
make verify
go build -o bin/m365 ./cmd/m365
```

CI MUST run `make verify`. No live tenant in CI.

## Automated first checks (fakes)

`M365_FAKE=1` or in-process fake Graph. Synthetic names only.

1. **Unique 1:1 off page one** — fake has Ajay 1:1 not in first `teams list` page; `m365 teams find Ajay` exits 0 with that one chat; no group that merely includes Ajay.
2. **Several Ajays** — two 1:1s; find returns both (bounded `--top`); `teams send --to Ajay --text t` exits 3, no send.
3. **Group topic** — `teams find --group NOC` returns NOC-Dev group, not 1:1s.
4. **Dry-run --to** — unique Ajay; `teams send --to Ajay --text ping --dry-run` has `dry_run` true, chat id, no new message.
5. **Send by id unchanged** — `teams send chat-1 --dry-run --text hi` still works.
6. **Mail/calendar/files unchanged** — their `--help` still names v1/002/003 flags.

Unit: `make unit`. Acceptance: `make acceptance`.

## Interactive (manual; not CI)

```sh
./bin/m365 auth login
./bin/m365 teams find Ajay
./bin/m365 teams send --to Ajay --text ping --dry-run
```

Do not paste live member lists or tokens into chat or git.

## Stop

If status is not usable, `m365 auth login` in a terminal. Find never sends. Zero `--to` matches must not create a chat.
