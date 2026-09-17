# Feature Specification: Natural Teams Find and DM

**Feature Branch**: `005-teams-natural-find`

**Created**: 2026-09-16

**Status**: Draft

**Input**: User description: "Make it natural to DM people in Teams and to find the group chat with someone. An agent can navigate, but finding the chat to notify should be simple: DM Ajay, or find the group chat with X."

## Specification Contract

### Objective

Mason or a local agent names a person or a group in ordinary language and gets the right Teams chat to notify, without paging through chat ids. Sending still uses dry-run. Ambiguous names are listed, not guessed.

### Goals

- Turn “Ajay” into the 1:1 chat with Ajay, or a clear “no such chat / several matches” result.
- Turn “the group with X” or a topic fragment into matching group chats.
- Let send target a person name when that name resolves to exactly one chat.
- Stop relying on “search the 50 most recent chats on the client.”
- Leave mail, calendar, and files unchanged. Reuse Teams consent, JSON/human output, dry-run, and exit classes.

### Non-goals

- Team **channels** (this feature is chats: 1:1 and group).
- Creating a brand-new 1:1 when none exists (find existing chats only).
- Meetings, calls, or a general people-directory product.
- Overnight watchers, kata wrappers, or MCP servers in this repository.
- Changing how `teams send` works when a chat id is already known.

### Verification Strategy

Verification MUST cover unique 1:1 resolve, unique group resolve, zero matches, several matches (no send), send `--to` with dry-run, and that a matching 1:1 is still found when it is not on the first `teams list` page. Fixtures MUST be synthetic. Live chat bodies MUST NOT be stored in the repository.

## Document Map

This file extends the `teams` namespace from [001-m365-cli](../001-m365-cli/teams.md). It does not rewrite 001. Calendar natural time (003) and MCP (004) are separate features.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Find the chat with a person (Priority: P1)

Mason or an agent asks for the chat with a person by name (`Ajay`). The command returns the 1:1 chat when exactly one person-and-chat pair matches. It does not send. It does not pick a group chat just because Ajay is a member.

**Why this priority**: This is the lookup agents already improvise by scanning recent chats. It must work even when that 1:1 is not among the most recently listed chats.

**Independent Test**: With synthetic chats including a 1:1 with Ajay Kumar not on the first list page, and a group that also includes Ajay, `teams find Ajay` returns that one 1:1 with id and members. Two Ajays return both candidates and do not pick. No Ajay returns an empty list, exit `0`.

**Acceptance Scenarios**:

1. **Given** exactly one 1:1 whose other member is Ajay, **When** the user runs `teams find Ajay`, **Then** the command exits `0` with that one chat (id, type 1:1, members) and does not list group chats Ajay is in.
2. **Given** that 1:1 is not among the first page of `teams list`, **When** the user runs `teams find Ajay`, **Then** the same 1:1 is still returned.
3. **Given** two people named Ajay each with a 1:1, **When** the user runs `teams find Ajay`, **Then** the command exits `0` with both candidates (bounded by `--top`) and does not send.
4. **Given** no member or topic matching Ajay, **When** the user runs `teams find Ajay`, **Then** the command exits `0` with an empty list (not a service error).
5. **Given** missing Teams consent, **When** find runs, **Then** the command exits `4`.

---

### User Story 2 - Find a group chat by people or topic (Priority: P1)

Mason or an agent finds a group chat by topic fragment or by naming people in it (`the group with Ajay`, `NOC`). Group search does not hide behind 1:1 preference.

**Why this priority**: Notify-the-group is the other half of “find the chat.” Independently testable from 1:1 find.

**Independent Test**: Synthetic group titled “NOC-Dev” and another group whose members include Ajay and Jarrett. Find with topic `NOC` returns the titled group. Find with `--group` and `Ajay` returns groups that include Ajay, not the 1:1.

**Acceptance Scenarios**:

1. **Given** a group whose topic contains `NOC`, **When** the user runs `teams find NOC` with group intent, **Then** that group is listed with id, topic, and members.
2. **Given** `--group` (or an equivalent group-intent phrase), **When** the user runs `teams find Ajay`, **Then** results are group chats that include Ajay, not the 1:1.
3. **Given** several matching groups, **When** find runs, **Then** at most `--top` are returned, each with id, and none is auto-selected for send.
4. **Given** no matching group, **When** find runs with group intent, **Then** the command exits `0` with an empty list.

---

### User Story 3 - Send a DM by name after a unique resolve (Priority: P2)

When a person name resolves to exactly one 1:1, send may take `--to Ajay` instead of a chat id. Dry-run shows the resolved person, chat id, and text, and does not send. Several matches or zero matches refuse to send.

**Why this priority**: This is the natural DM. It depends on unique find working.

**Independent Test**: Unique Ajay 1:1: dry-run `--to Ajay` shows that chat and no message is created; real send delivers there. Two Ajays: send `--to Ajay` exits `3` and lists candidates. Unknown name: send `--to` exits `6` or `3` with no send.

**Acceptance Scenarios**:

1. **Given** exactly one 1:1 with Ajay, **When** the user runs `teams send --to Ajay --text "ping" --dry-run`, **Then** stdout shows Ajay, the chat id, and the text, and no message is created.
2. **Given** the same unique resolve without `--dry-run`, **When** the user sends, **Then** the message is sent to that 1:1 and stdout reports a message identity.
3. **Given** two matching Ajays, **When** the user sends `--to Ajay`, **Then** the command exits `3`, lists the candidate chats, and does not send.
4. **Given** no matching chat, **When** the user sends `--to Ajay`, **Then** the command does not send (empty-or-not-found class as specified in Assumptions) and does not create a new chat.
5. **Given** a known chat id, **When** the user sends with that id as today, **Then** behavior is unchanged from 001.

---

### Edge Cases

- Display name vs email vs first name only (`Ajay` vs `Ajay Kumar` vs `ajay@…`).
- Person in the org with no existing 1:1 (do not create; empty or not-found).
- Query that matches both a 1:1 and groups (default: 1:1 only unless group intent is set).
- Case and extra spaces in the query.
- `--top` omitted (documented default) and `--top` above maximum (exit `3`).
- Find while `teams list` first page does not include the match.
- Send `--to` together with an explicit chat id (usage error: pick one).
- Missing text on send (still exit `3`).
- Missing Teams consent (exit `4`).
- Mail, calendar, and files commands unchanged.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: This feature MUST extend the existing `teams` namespace (alias `chat`). It MUST NOT add a new top-level namespace and MUST NOT change mail, calendar, or files command names, flags, output shape, or exit classes.
- **FR-002**: `teams find` MUST accept a query string (person name, email, or topic fragment) and return matching chats the signed-in user belongs to: id, type, topic, members, last-message preview when available.
- **FR-003**: A query that looks like a single person (one name or one email, no group intent) MUST prefer 1:1 chats with that person. Group chats that merely include them MUST NOT appear in that default result.
- **FR-004**: Group intent (`--group`, or a query that names a group/topic rather than a single person) MUST return group chats matching topic and/or members, not 1:1s.
- **FR-005**: Zero matches MUST exit `0` with an empty list for `teams find`. Several matches MUST all be returned, bounded by `--top`, with no implicit winner.
- **FR-006**: Find MUST NOT miss a matching chat solely because it was not on the first page of `teams list`. If an internal search ceiling is hit without uniqueness, the command MUST say the search was incomplete rather than claim there is no chat.
- **FR-007**: Default `--top` for find is `10`; maximum is `20`. Omitting `--top` applies the default. Above-max `--top` exits `3`.
- **FR-008**: `teams send` MUST accept `--to <query>` as an alternative to a chat id. `--to` and a chat id together MUST exit `3`.
- **FR-009**: `--to` MUST resolve with the same rules as `teams find` for a single-person query. Exactly one 1:1 → that chat. Several matches → exit `3` with candidates, no send. Zero matches → no send and no new chat created.
- **FR-010**: Send with `--to` MUST support `--dry-run`. Dry-run MUST show the resolved person or topic, chat id, and text, and MUST NOT send.
- **FR-011**: Existing send-by-chat-id behavior from 001 MUST remain.
- **FR-012**: `--help` for `teams find` and `teams send` MUST show name-based examples (`Ajay`, `--group NOC`). Help exits `0` without a session.
- **FR-013**: Missing Teams consent MUST exit `4`. Microsoft 365 service errors MUST exit `5`. Find never sends.
- **FR-014**: Tests and fixtures MUST use synthetic people and chats. Live member lists and messages MUST NOT be stored in the repository.
- **FR-015**: This feature MUST NOT create chats, MUST NOT search team channels, and MUST NOT become a tenant-wide people browser.

### Key Entities

- **FindQuery**: The language the user typed (`Ajay`, `NOC`, an email).
- **ChatMatch**: A chat the user belongs to: id, type (1:1 or group), topic, members (name and email), last-message preview.
- **ResolveResult**: Unique match, several matches, none, or incomplete search (ceiling hit).
- **Person**: Display name and email of a chat member. Used for matching, not a separate directory dump.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A unique 1:1 is found by first name or full name without the caller supplying a chat id, even when that chat is not on the first `teams list` page.
- **SC-002**: Default person find does not return group chats that merely include that person.
- **SC-003**: Group-intent find returns groups by topic or members and does not auto-send.
- **SC-004**: Two people with the same first name never cause a silent send; send `--to` exits as usage and lists candidates.
- **SC-005**: Dry-run send `--to` with a unique match shows the resolved chat and creates no message.
- **SC-006**: `teams send` with a chat id still works exactly as in 001.
- **SC-007**: Mail, calendar, and files command names, flags, output shape, and exit classes remain unchanged.
- **SC-008**: After a usable session exists, a unique `teams find` completes in under 15 seconds on an ordinary connection.
- **SC-009**: No command prints access tokens, refresh tokens, authorization codes, or client secrets.
- **SC-010**: JSON stdout for find is exactly one parseable JSON value with no mixed-in diagnostics.

## Assumptions

- This feature sits on 001 Teams. List, get, messages, send-by-id, watch, attachments, JSON default, and exit classes `0/3/4/5/6` still apply.
- “Person query” = one token that is a given name, full name, or email. “Group intent” = `--group` flag, or a query with more than one person name, or a query that is clearly a topic (help examples: `--group`, `NOC`).
- Matching is case-insensitive. Extra spaces are ignored.
- Prefer exact display-name or email match, then prefix, then substring, among chats the user belongs to.
- Find does not create a 1:1 if none exists. That is a later feature.
- Send `--to` with zero matches exits `6` (not-found). Several matches exit `3` (usage: pick a chat id).
- Default `--top` 10, max 20 for find results. This limit is on **results**, not an excuse to scan only one list page.
- If paging all of the user’s chats would exceed a documented internal ceiling, find MUST report incomplete search (not “no matches”) when no unique match was already found.
- `chat` remains an alias for `teams`; `chat find` and `chat send --to` work the same.
- Client ID, Tenant ID, and secrets MUST NOT appear in this specification, source, tests, logs, or fixtures.
- Kata `m365-chat find=` is a consumer, not the spec. This CLI feature is what that helper should call later.
