# Phase 1 Data Model: Natural Calendar Times and Free Slots

Extends 002 calendar entities. No Graph SDK fields or URLs in domain types.

## WhenPhrase

**Fields**: original text (string).

**Rules**:
- Stored on dry-run and create JSON as the user typed it (trimmed).
- Empty when-phrase on create is usage unless `--start` RFC3339 is supplied as the entire when-phrase.

## ResolvedInterval

**Fields**: timezone name (IANA), local date, start instant, end instant, start clock (display), end clock (display).

**Rules**:
- Start MUST be before end.
- Start MUST be in the future relative to injected now in that timezone (create).
- Default end = start + 30 minutes when duration and until are omitted.
- Duration MUST be a positive whole number of minutes, at most 8 hours.

## DurationPhrase

**Fields**: original text, minutes (int).

**Rules**:
- Accepted: `N minutes`, `N hour`, `N hours` with N a positive integer.
- Invalid or non-integer N → usage.

## FreeSlot

**Fields**: calendar id, start, end, duration minutes.

**Rules**:
- No event bodies.
- Lies inside working hours for that local day.
- Does not overlap busy occurrences.
- Start is on the 30-minute grid (09:00, 09:30, …) unless documented otherwise.
- If the day is today, start MUST be after now.

## WorkingHours

**Fields**: start clock, end clock (end exclusive), timezone.

**Rules**:
- Default 09:00–17:00 local.
- `--hours start-end` overrides on the same day (example `9:00-17:00`).
- Start MUST be before end.

## BusyInterval

**Fields**: start, end.

**Rules**:
- From occurrences on the one named own calendar that overlap the local day.
- Tentative and busy block. Cancelled do not. All-day on that date blocks the whole working day.

## ResultLimit (free)

**Rules**:
- Default `--top` = 5. Maximum = 20. Values `<1` or `>20` are usage, not silent caps.
- Empty list (fully booked) is success.

## Validation summary

| Rule | Failure class |
| --- | --- |
| Ambiguous/unsupported when-phrase | usage (3) |
| Day without clock time on create | usage (3) |
| Clock without am/pm or 24-hour | usage (3) |
| Past resolved start | usage (3) |
| Start ≥ end; until on another date | usage (3) |
| Duration zero/negative/non-integer/>8h | usage (3) |
| `--when` and `--start` both set | usage (3) |
| `--dry-run` on `calendar free` | usage (3) |
| `--top` out of range on free | usage (3) |
| Missing calendar consent | auth (4) |
| Graph/service error | service (5) |
| Unknown calendar | not-found (6) |
| Fully booked day | success (0), empty items |
