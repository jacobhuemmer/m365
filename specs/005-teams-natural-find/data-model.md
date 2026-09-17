# Phase 1 Data Model: Natural Teams Find and DM

Extends 001 Teams. No Graph SDK fields or URLs in domain types. No live member lists in fixtures.

## FindQuery

**Fields**: raw text, intent (`person` | `group`), `--group` bool.

**Rules**:
- Trim; case-insensitive match later.
- Person: one email, one given name, or `First Last`.
- Group: `--group` set, or query matches `(?i)^(?:the )?group with (.+)$`.
- Empty query → usage (3).

## ChatMatch

Reuses `domain.Chat`: id, type (`oneOnOne` / `group` as today), topic, members (name, address), last_message preview.

**Rules**:
- User must belong to the chat.
- Default person find: 1:1 only; groups that merely include the person MUST NOT appear.
- Group find: groups only.

## ResolveResult

**Fields**: query, intent, limit, count, incomplete (bool), items ([]ChatMatch).

**States**:
- Unique: count 1, incomplete false (find may still list one item).
- Several: count 2..limit, no winner.
- None: find → count 0, incomplete false, exit 0.
- Incomplete: ceiling hit, no unique decision; `incomplete` true; MUST NOT claim no chat.

## Person (member)

**Fields**: name, address (email). Matching: exact display name or email, then prefix, then substring. Signed-in account excluded as the “other” 1:1 member when known.

## ResultLimit (find)

**Rules**: Default `--top` = 10. Maximum = 20. Values `<1` or `>20` are usage, not silent caps. This bounds **results**, not Graph pages scanned.

## ScanCeiling

**Rules**: At most 10 Graph pages of 50 chats (500). Stop early on unique person match. Exceeding without uniqueness → incomplete.

## Send `--to`

**Rules**:
- `--to` XOR chat id (both → usage 3).
- Resolve as person find.
- Unique 1:1 → existing send (dry-run shows person/topic, chat id, text).
- Several → usage 3, candidates, no send.
- Zero (complete) → not-found 6, no create.
- Incomplete resolve → usage 3 with incomplete, no send.

## Validation summary

| Rule | Failure class |
| --- | --- |
| Empty find query | usage (3) |
| `--top` out of 1..20 | usage (3) |
| `--to` and chat id | usage (3) |
| Several `--to` matches | usage (3) |
| Incomplete `--to` | usage (3) |
| Zero `--to` matches (complete) | not-found (6) |
| Missing Teams consent | auth (4) |
| Graph/service | service (5) |
| Find zero matches (complete) | success (0), empty items |
| Find never sends | n/a |
