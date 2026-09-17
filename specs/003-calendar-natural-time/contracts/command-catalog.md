# CLI command catalog (natural time and free)

Extends 002 calendar. Mail, teams, files, and `chat` are unchanged.

Exit classes unchanged. JSON default. Help exits 0 without a session.

Default duration: 30 minutes. Free default day: tomorrow (calendar TZ). Free `--top` default 5, max 20. Working hours default 09:00–17:00 local. Slot grid: 30 minutes.

## calendar (additive)

| Command | Args / flags | Defaults / notes |
| --- | --- | --- |
| `calendar create` | `--subject` (required), `--when`, `--until`, `--duration`, `--calendar`, `--dry-run`, still accepts `--start`/`--end` as a full RFC3339 when-phrase | `--when` is the primary start; `--when` + `--start` → 3; default duration 30 minutes |
| `calendar update` | `EVENT_ID`, optional `--when`/`--until`/`--duration` plus existing fields, `--dry-run` | same phrase rules when changing time |
| `calendar free` | `--when` (day phrase), `--duration`, `--hours`, `--calendar`, `--top` | `--when` default `tomorrow`; `--dry-run` → 3; empty day → 0 |

### `--when` (create)

Must resolve with a clock time. Examples in help: `tomorrow at 1:30 pm`, `today 9am`, `Friday at 9am`, `2026-09-18 13:30`.

### `--when` (free)

Day only is valid: `tomorrow`, `today`, `Friday`, or a calendar date. Default `tomorrow`.

### Dry-run create JSON

MUST include `dry_run: true`, `when` (original phrase), `start`, `end`, and timezone name. MUST NOT create.

### Free JSON

List page: `limit`, `count`, `items` of `{start, end, duration_minutes, calendar_id}`. Optional `window` / working-hours fields. No event bodies. `next_page` unused (single day). Fully booked: `count` 0, exit 0.
