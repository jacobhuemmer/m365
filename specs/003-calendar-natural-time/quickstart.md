# Quickstart validation: natural calendar times and free slots

Catalog: [contracts/command-catalog.md](contracts/command-catalog.md). Domain: [data-model.md](data-model.md). Fake: [contracts/fake-graph.md](contracts/fake-graph.md). Calendar verbs from 002 still apply.

## Prerequisites

- Existing m365 CLI with 002 calendar namespace
- `M365_FAKE=1` for CI; no live tenant in automated checks
- Client ID / Tenant ID at runtime only (never in the repo)

## Automated first checks (fakes)

Freeze now in tests to `2026-09-16T12:00:00-05:00`. Do not copy live events.

1. **Phrase dry-run** — `m365 calendar create --dry-run --subject Sync --when 'tomorrow at 1:30 pm'` exits `0`; JSON has `when`, `start` on 2026-09-17 13:30 CT, `end` 30 minutes later, `dry_run` true; fake has no new event.
2. **Ambiguous phrase** — `--when '1:30'` exits `3`; no event.
3. **Free tomorrow** — `m365 calendar free` exits `0`; `limit` 5; items do not overlap 09:00–10:00 or 13:00–14:00 CT on 2026-09-17.
4. **Fully booked** — fixture day with no gap exits `0` with `count` 0.
5. **Free dry-run** — `m365 calendar free --dry-run` exits `3`.
6. **Mail/teams unchanged** — `m365 mail list --help` and `m365 teams list --help` still name v1 flags.

Unit: `make unit`. Acceptance: `make acceptance`.

## Interactive (manual; not CI)

```sh
./bin/m365 calendar create --dry-run --subject Sync --when 'tomorrow at 1:30 pm'
./bin/m365 calendar free
```

Expect the resolved local Thursday 1:30–2:00 PM (or whatever "tomorrow" is) and a short free list. Do not paste live event bodies.

## Stop

If `auth status` is not usable, run `auth login` in a real terminal. Do not loop login in agents.
