# Fake Graph fixture contract (natural time)

Extends `specs/002-calendar-files/contracts/fake-graph.md`. Same `httptest` server. Token `fake-calendar` / `fake-all` required for calendar routes.

## Timezone

Synthetic default calendar `cal-1` reports `timeZone`: `America/Chicago`. Tests MUST freeze now in that location.

## Busy day for free search

Relative to frozen now `2026-09-16T12:00:00-05:00` (Wednesday):

- Tomorrow (`2026-09-17`) working hours 09:00–17:00 CT.
- Busy: `09:00–10:00` and `13:00–14:00` (synthetic ids `ev-busy-1`, `ev-busy-2`). No live bodies.
- Remaining 30-minute grid gaps MUST be computable without Graph slot APIs.
- A second fixture day with wall-to-wall busy 09:00–17:00 yields an empty free list.

## Phrase create

Dry-run MUST NOT reach POST `/me/events`. Real create still returns a synthetic id. Ambiguous phrases never hit the fake.

## Forbidden

- Live calendar dumps in `testdata/`.
- `/users/{other}/calendar/getSchedule` as success.
- Changing `fake-both` to grant calendar.
