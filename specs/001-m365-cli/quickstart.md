# Quickstart validation: m365 CLI v1

This is a run guide, not the test suite. Schemas: [contracts/](contracts/). Domain: [data-model.md](data-model.md). Catalog: [contracts/command-catalog.md](contracts/command-catalog.md).

## Prerequisites

- Go 1.25+
- macOS (Keychain for real login when the token fits; otherwise `~/.local/state/m365/session.json` mode `0600`)
- Client ID and Tenant ID supplied at runtime only:

```sh
export M365_CLIENT_ID='...'   # do not commit
export M365_TENANT_ID='...'   # do not commit
```

Or `~/.config/m365/config.json` with `client_id` and `tenant_id` keys. Never put values in the repo.

## Build

After implement creates the module:

```sh
make verify
go build -o bin/m365 ./cmd/m365
```

CI MUST run `make verify`.

## Automated first checks (fakes; no live tenant)

These are the first validation scenarios. They run against the fake Graph and a fake secret store.

1. **Status signed-out** — `m365 auth status` exits `0`; JSON has `signed_in: false`; no token fields.
2. **Mail list empty** — signed in with `fake-mail`; `m365 mail list` exits `0`; `count` is 0; `limit` is 10; stdout is one JSON object.
3. **Dry-run send with attach** — local synthetic file under `testdata/`; `m365 mail send --dry-run --to user@example.com --subject t --body b --attach testdata/note.txt` exits `0`; stdout lists name and size; fake Graph has no new sent item.
4. **Save refuse without overwrite** — destination exists; `m365 mail save-attachment msg-1 att-1 --out /tmp/exists.txt` exits `3`; file unchanged.

Unit: `make unit`. Acceptance: `make acceptance`. Interactive `auth login` is not required for these.

## Interactive login (manual; not CI)

```sh
./bin/m365 auth login
./bin/m365 auth status
./bin/m365 mail list --top 10
```

Expect a browser gesture, then a usable session, then an inbox page without a second login. Do not paste tokens. Live timing for SC-014 is optional and manual; CI uses fakes with a 15s deadline.

## Stop

If `auth status` is not usable, stop and run `auth login` in a real terminal. Do not loop login in agents.
