# m365 Auth and Consent

This file owns delegated sign-in, session status, logout, and independent
mail versus Teams consent. Command streams and exit classes are owned by
[cli-contract.md](cli-contract.md). This file specifies what the user
observes, not how sign-in is implemented.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Sign in, inspect status, and log out (Priority: P1)

The mailbox owner signs in with an interactive Microsoft 365 user gesture,
then sees whether a silent session is usable. Later commands reuse that
session until logout or expiry. Logout forgets the saved session. "Not signed
in" is distinct from "signed in, but Microsoft 365 rejected the request."

**Why this priority**: No mail or Teams work is possible without a session
and without distinguishable failure classes.

**Independent Test**: From signed-out, run login (or a test double of the
user gesture), status, logout, and status again. Assert status reports a
usable session after login, logout clears it, and a workload command after
logout exits `4` (a stub command is enough; `mail list` reuse is US-002).
Assert a usable session plus a service rejection exits `5`, not `4`.

**Acceptance Scenarios**:

1. **Given** no saved session, **When** the user completes `auth login` with
   the interactive sign-in gesture, **Then** the command exits `0` and later
   `auth status` reports signed in with a usable silent session and no
   secrets.
2. **Given** a usable session, **When** the user runs `auth status`, **Then**
   stdout reports signed in, session usable, account display name, and
   per-namespace consent, and does not print tokens or secrets.
3. **Given** a usable session, **When** the user runs `auth logout`, **Then**
   the saved session is gone, `auth status` reports not signed in, and the
   next mail or Teams command exits `4`.
4. **Given** no saved session, **When** the user runs `auth status`, **Then**
   the command exits `0` and reports not signed in (this is not an auth
   failure of the status command).
5. **Given** a usable session, **When** a mail or Teams command is rejected
   by Microsoft 365, **Then** the command exits `5` and stderr names a
   service error, not an auth error.
6. **Given** an expired session, **When** the user runs a mail or Teams
   command, **Then** the command exits `4` and `auth status` reports that a
   silent session is not usable.

### User Story 9 — Use mail and Teams with independent consent (Priority: P3)

The user can grant mail access without Teams access, and Teams without mail.
Missing consent for one namespace MUST NOT break the other. Attachment
operations stay inside the namespace of the message they belong to.

**Why this priority**: Least privilege is required by the constitution, but
the product is already useful with both namespaces consented.

**Independent Test**: With only mail consent, run mail list (success) and
teams list (exit `4`). With only Teams consent, run the reverse. Attempting
to save a Teams attachment through a mail command fails as validation or
not-found in the mail namespace, not as a Teams call.

**Acceptance Scenarios**:

1. **Given** mail consent and no Teams consent, **When** the user runs
   `mail list` and `teams list`, **Then** mail list succeeds and teams list
   exits `4` with an auth-class diagnostic naming missing Teams access.
2. **Given** Teams consent and no mail consent, **When** the user runs
   `teams list` and `mail list`, **Then** teams list succeeds and mail list
   exits `4` with an auth-class diagnostic naming missing mail access.
3. **Given** a Teams message with an attachment, **When** the user invokes a
   mail attachment command against that identity, **Then** the command does
   not use Teams access and fails in the mail namespace (validation or
   not-found), leaving the Teams namespace unused.

## Functional Requirements

- **FR-014**: v1 authentication MUST be delegated sign-in of the interactive
  Microsoft 365 user. Login MUST use an interactive user gesture (browser or
  equivalent). Passwords MUST NOT be accepted as flags or stdin. App-only
  and shared-mailbox admin modes MUST NOT exist in v1.
- **FR-015**: After a successful login, later commands MUST reuse the saved
  session without repeating the gesture until logout or expiry.
- **FR-016**: `auth status` MUST report whether the user is signed in,
  whether a silent session is usable, the account display name, and which of
  `mail` and `teams` have consent. Status MUST NOT print secrets. Status
  while signed out MUST exit `0` and report not signed in.
- **FR-017**: `auth logout` MUST remove the saved session from the local
  secret store. Logout when already signed out MUST exit `0` and leave the
  user signed out.
- **FR-018**: Workload commands MUST distinguish: no session or expired
  session or missing namespace consent → exit `4`; usable session rejected
  by Microsoft 365 → exit `5`. `auth login` refused by the user or by
  Microsoft 365 identity MUST exit `4`. Missing deployment configuration
  MUST exit `3`.
- **FR-019**: Mail commands MUST NOT require Teams consent. Teams commands
  MUST NOT require mail consent. Enabling one namespace MUST NOT fail
  because the other namespace lacks consent.
- **FR-020**: The tool MUST be configured for one organizational app
  registration as deployment configuration. This specification MUST NOT
  contain Client ID, Tenant ID, or any secret.
- **FR-021**: `auth login` when a session already exists MUST exit `0` with
  a usable session (the user MAY complete another gesture; the observable
  result is a usable session).
- **FR-022**: Attachment operations MUST execute in the namespace of the
  parent message. A mail attachment command MUST NOT require Teams consent.
  A Teams attachment command MUST NOT require mail consent.
