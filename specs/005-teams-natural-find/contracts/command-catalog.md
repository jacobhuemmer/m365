# CLI command catalog (teams find and send --to)

Extends 001 teams. Mail, calendar, files, MCP, and `chat` alias unchanged. `chat find` / `chat send --to` work.

Exit classes unchanged. JSON default. Help exits 0 without a session.

Find `--top` default 10, max 20 (results). Scan ceiling: 10 pages × 50 chats (help MUST mention incomplete).

## teams (additive)

| Command | Args / flags | Defaults / notes |
| --- | --- | --- |
| `teams find` | `QUERY`, `--group`, `--top` | person query unless `--group` or `group with X`; empty list exit 0 |
| `teams send` | existing id path **or** `--to QUERY`, plus existing `--text` / `--dry-run` / attach | `--to` and CHAT_ID together → 3 |

### Find JSON

`query`, `intent` (`person`\|`group`), `limit`, `count`, `incomplete`, `items` of Chat (id, type, topic, members, last_message). One JSON object. `incomplete` true means ceiling hit, not “no such chat.”

### Send `--to` dry-run JSON

MUST include `dry_run: true`, `to` (query), `chat_id`, `text`. MUST NOT send.

### Help examples (MUST appear)

- `teams find Ajay`
- `teams find --group NOC`
- `teams send --to Ajay --text ping --dry-run`
