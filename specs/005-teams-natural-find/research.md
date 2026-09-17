# Phase 0 Research: Natural Teams Find and DM

All Technical Context items are resolved. Client ID, Tenant ID, and token values MUST NOT appear here.

## Decision: Page `/me/chats` with members; do not search `/users` or `/me/people`

**Rationale**: FR-015 forbids a tenant-wide people browser. FR-006 forbids missing a 1:1 only because it was off the first `teams list` page. Graph has no “chat with this display name” filter that is reliable without expanding members. The adapter pages `GET /me/chats?$expand=members` with `$top=50` until a unique match, `--top` results are filled, or the scan ceiling is hit. Matching is in `internal/app/teams` on member display name/email and group topic among chats the user already belongs to.

**Alternatives considered**:
- Client-side filter of one `teams list` page: fails SC-001 / FR-006.
- `GET /users` or `/me/people`: tenant people dump; violates FR-015.
- Create 1:1 if missing: forbidden by FR-015 / Assumptions.

**Sources** (observed 2026-09-16): Microsoft Graph list chats — https://learn.microsoft.com/en-us/graph/api/chat-list

## Decision: Documented scan ceiling 10 pages × 50 chats (500)

**Rationale**: Constitution VI forbids silent full-chat retrieval. Spec: if paging would exceed a documented ceiling without uniqueness, report **incomplete**, not “no matches.” Ceiling = 10 Graph pages of 50 = 500 chats. Result `--top` remains 10 default / 20 max (results, not scan size). JSON includes `incomplete: true` when the ceiling stopped the scan and no unique match was already decided. Unique match found earlier → stop paging (SC-008).

**Alternatives considered**:
- Unbounded nextLink until empty: quota and time risk; silent truncation forbidden if we stop without saying so.
- Ceiling 50 (one page): contradicts FR-006.

## Decision: Group intent is `--group` or a closed “group with …” phrase

**Rationale**: Assumptions. Default single token / email / `First Last` → 1:1 member match (FR-003). `--group` → group chats by topic substring and/or members (FR-004). Query matching `(?i)^(?:the )?group with (.+)$` → group intent + that person as member filter. Topic example `NOC` is `teams find --group NOC` (help). Bare `teams find NOC` is a person query (likely empty 1:1). Closed grammar, not NLP.

**Alternatives considered**:
- Guess all-caps tokens are topics: fragile (`AJAY`).
- Always search both 1:1 and groups: violates FR-003 / SC-002.

## Decision: Match rank exact, then prefix, then substring; case-insensitive

**Rationale**: Assumptions. Trim spaces. 1:1 “other member” excludes the signed-in account when Session.Account is set. Several same-rank matches → all returned (find) or usage 3 (send `--to`).

**Alternatives considered**:
- Fuzzy/Levenshtein: extra library, speculative.
- Org directory disambiguation: FR-015.

## Decision: Find is a new verb; send `--to` reuses Find

**Rationale**: FR-002 / FR-008. `teams find QUERY`. `teams send --to QUERY` runs the person-query find. Unique 1:1 → send path. Several → exit 3 + candidates. Zero → exit 6, no create. `--to` + chat id → 3. Chat-id send unchanged (FR-011). `chat` alias already maps to teams.

**Alternatives considered**:
- `teams dm Ajay`: extra verb; spec names `find` and `send --to`.
- Change 001 send positional: forbidden.

## Decision: Stay in `internal/app/teams`; no MCP/mail/calendar/files edits

**Rationale**: FR-001, 005 non-goals (MCP). Constitution VII. Additive help on teams find/send only. Fake Graph seeds extra synthetic chats (Ajay 1:1 past first list page, two Ajays, NOC-Dev group). Live member lists stay out of git.

**Alternatives considered**:
- Update 004 MCP teams-find recipe in this feature: spec non-goal; later 004 follow-up.
