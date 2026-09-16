# Phase 0 Research: m365 CLI v1

All Technical Context items are resolved. Client ID and Tenant ID values MUST NOT appear here.

## Decision: Go 1.25+ single module

**Rationale**: Locked by the plan prompt. The repo has no `go.mod` yet. Go 1.25 matches the constitution's language-agnostic gates once tools are chosen (`gosec` currently expects 1.25+). Intended module path: `github.com/masonhuemmer/m365`.

**Alternatives considered**:
- Rust: allowed only if the plan prompt is edited; architecture and gates stay the same with Rust equivalents.
- Older Go: rejected; user required 1.25+.

**Sources**:
- Go 1.25 release notes: https://go.dev/doc/go1.25
- gosec: https://github.com/securego/gosec

## Decision: CLI parser is stdlib `flag.FlagSet` plus a small registry

**Rationale**: The Go `flag` package documents `FlagSet` for subcommands. v1 has two workload namespaces, core `auth`, global `--json`/`--human`, and `--help`. That is a small, closed tree. A framework is not required. The CLI adapter owns routing, including the `chat` → `teams` alias, so namespace packages never import each other.

**Alternatives considered**:
- Cobra / urfave/cli / Kong: richer help generators; rejected as framework religion for ~15 verbs (constitution III).
- Hand-rolled argv with no `flag`: rejected; `FlagSet` already parses `--flag`, `--flag=x`, and `--help` (`ErrHelp`).

**Sources**:
- Go `flag` package (FlagSet subcommands): https://pkg.go.dev/flag (observed 2026-09-16)

## Decision: Thin Graph HTTP adapter, not the official SDK

**Rationale**: `msgraph-sdk-go` / kiota types would leak into callers. Constitution VII forbids Graph URLs, JSON payloads, and SDK types in domain and use cases. `net/http` plus `encoding/json` in `internal/adapters/graph` maps to domain entities. Adapter encodes `@odata.nextLink` as an opaque `next_page` token; stdout never contains a service URL.

**Alternatives considered**:
- Official Graph SDK: rejected because types and request builders would cross the adapter boundary.
- Raw URL command (`graph GET`): out of spec.

**Adapter mapping (adapter-only; not domain)**:
- Mail list/get: `GET /me/messages` or `GET /me/mailFolders/{id}/messages` with `$top`, `$search`/`$filter`, `Prefer: outlook.body-content-type="text"` when bodies are requested. Least-privilege delegated permission: `Mail.Read` (list/get/thread/save); `Mail.Send` (send/reply).
- Teams list/get/messages/send: chats and chat messages. Least-privilege: `Chat.Read`; `ChatMessage.Send`.
- Attachments: metadata on message payloads; save via the parent namespace attachment content endpoint. Never print bytes.
- Errors: 401/insufficient session → exit 4; 403 missing namespace consent → exit 4; 404 → exit 6; 429/5xx → exit 5.

**Sources**:
- List messages: https://learn.microsoft.com/en-us/graph/api/user-list-messages?view=graph-rest-1.0 (default page size 10; `$top`; next page via `@odata.nextLink`; do not extract `$skip` to invent paging) (observed 2026-09-16)

## Decision: OAuth public-client PKCE with localhost callback

**Rationale**: Spec requires interactive delegated sign-in without specifying the grant. Microsoft identity platform documents authorization code + PKCE for desktop/native apps and recommends `http://localhost` for system-browser public clients. Public clients MUST NOT use a client secret. `golang.org/x/oauth2` provides `GenerateVerifier`, `S256ChallengeOption`, `VerifierOption`, and `Exchange`. Device-code and WAM are not added unless loopback cannot work on macOS.

**Flow**:
1. Load Client ID and Tenant ID from environment (`M365_CLIENT_ID`, `M365_TENANT_ID`) or `~/.config/m365/config.json` (keys `client_id`, `tenant_id`). Env overrides file. Missing either → exit 3. Values never committed.
2. Bind `http://127.0.0.1:<ephemeral>/` as redirect URI (must match the app registration's localhost redirect).
3. Open the system browser to the authorize URL with PKCE S256, `access_type=offline`, scopes for mail and teams plus `openid` `offline_access` `profile`.
4. Exchange the code with the verifier; store tokens in Keychain when they fit, otherwise a `0600` session file; never print them.
5. Incremental consent: a denied Teams scope MUST NOT block mail, and vice versa. Missing namespace consent on a workload command → exit 4.

**Alternatives considered**:
- Device code: extra UX; deferred unless loopback fails.
- MSAL / confidential client secret: rejected (public CLI; secrets cannot live in the binary).
- Passwords on the command line: forbidden by spec.

**Sources**:
- Microsoft identity platform auth code flow (PKCE, localhost, public clients, no secret): https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-auth-code-flow (observed 2026-09-16)
- `golang.org/x/oauth2` PKCE: https://pkg.go.dev/golang.org/x/oauth2 (GenerateVerifier / S256ChallengeOption / VerifierOption)

## Decision: Tokens in macOS Keychain via `go-keyring`, with `0600` file fallback

**Rationale**: Constitution requires OS keychain or 0600 equivalent. v1 target is macOS. `github.com/zalando/go-keyring` stores generic password items in Keychain. Microsoft identity JWTs often exceed the Keychain item size `go-keyring` accepts (`data passed to Set was too big`). When `Set` fails for size, the same blob MUST be written to `$XDG_STATE_HOME/m365/session.json` defaulting to `~/.local/state/m365/session.json`, mode `0600`. Logout MUST delete the Keychain item and the file. World-readable token files are forbidden. Tests use a fake `SecretStore` port; they MUST NOT talk to the real Keychain in unit tests.

**Alternatives considered**:
- Keychain only: rejected after live login; Entra tokens do not fit `go-keyring` Set.
- `0600` file only: constitution-legal but weaker than Keychain when the blob fits.
- `keybase/go-keychain`: macOS-only API; still subject to item-size limits; `go-keyring` plus file fallback is smaller to fake.

## Decision: Watch checkpoint is a 0600 state file of high-watermarks

**Rationale**: Spec allows a local checkpoint so later `teams watch` without `--since` emits only newer events. Contents: opaque chat id → last emitted message id and created time. No tokens, no message bodies, no attachment bytes. Path: `$XDG_STATE_HOME/m365/watch.json` defaulting to `~/.local/state/m365/watch.json`, mode `0600`. Checkpoint advances only through emitted events (`--top` visible; no silent skip). Watch is a one-shot poll that writes JSON lines, not a daemon.

**Alternatives considered**:
- Keychain for checkpoint: wrong store (not a secret).
- Graph delta/websocket subscription: larger than v1; poll is sufficient.
- Overnight LaunchAgent: out of this repo.

## Decision: JSON stdout and stderr schemas

**Rationale**: Spec locks JSON default, one JSON value per non-watch command, JSON lines for watch, one error object on stderr. Schemas live in `contracts/`. List payloads include `limit`, `count`, `next_page`. Attachment objects never include bytes. Error `class` is `usage` | `auth` | `service` | `not_found` matching exits 3/4/5/6.

**Alternatives considered**:
- Default human on TTY: rejected; existing agents parse JSON; spec assumption is JSON default.
- HTTP status in error JSON: rejected; product classes only.

## Decision: Fake Graph is `httptest.Server` owned by tests

**Rationale**: Stdlib `httptest` is an in-process HTTP server. One fake covers mail, teams, attachments, 401/403/404/429, and independent consent (token or test header names granted namespaces). Unit tests inject the fake base URL into the Graph adapter. Acceptance uses the same fake. No live tenant in CI. Synthetic mailbox/chat documents only.

**Alternatives considered**:
- Pure in-memory ports without HTTP: used for domain/use-case unit tests; not enough for adapter mapping tests.
- Recorded live traffic: forbidden (live content MUST NOT enter the repo).

## Decision: APS acceptance pipeline, not Godog

**Rationale**: Constitution IV. Pin `unclebob/Acceptance-Pipeline-Specification` by commit SHA.

**Alternatives considered**:
- Godog/Cucumber runner: rejected unless a locked test proves need.
- Acceptance tests only: rejected; unit tests remain mandatory.

**Sources**:
- https://github.com/unclebob/Acceptance-Pipeline-Specification

## Decision: CRAP gate threshold 15

**Rationale**: Matches Frontdoor's `make crap` bar. `scripts/crap.sh` combines cyclomatic complexity and coverage; CRAP > 15 fails unless a later justification file exists. `cmd/m365` wiring may be uncovered; behavior stays in tested packages.

**Alternatives considered**:
- Coverage-only gate: weaker than constitution complexity/CRAP requirement.
