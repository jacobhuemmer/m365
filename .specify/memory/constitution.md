<!--
Sync Impact Report
- Version change: unratified template -> 1.0.0
- Modified principles: template placeholders -> eight initial project principles
- Added sections: Engineering Constraints; Development Workflow & Quality Gates; Next Actions
- Removed sections: none
- Follow-up TODOs: none
-->
# m365 Constitution

## Core Principles

### I. Test-Driven Development (NON-NEGOTIABLE)

- Every behavior change MUST begin with a test in RED. The author MUST run the test and confirm
  that it fails for the expected behavioral reason, not because of a compile, import, or fixture
  error.
- RED MUST be locked in a separate git commit containing only the failing test and fixtures
  required to run it. After that commit, the test MUST NOT be modified, weakened, skipped, or
  rewritten without explicit human approval.
- If a locked test is incorrect, work MUST stop and the author MUST request human direction. The
  author MUST NOT silently edit the test to fit an implementation.
- GREEN MUST be reached with the smallest production change that satisfies the locked test. Work
  MUST NOT be marked complete while a locked RED test fails unless missing human input or an
  external dependency blocks progress; the blocker MUST be recorded in the plan or task notes.
- Refactoring MUST occur only while the suite is GREEN.

Rationale: separate, immutable RED evidence makes intent inspectable and prevents tests from being
bent to accommodate a struggling implementation.

### II. Clean Code: Small, Single-Purpose Modules

- Source files SHOULD remain under 250 lines. A file over 250 lines MUST have a refactoring note in
  the plan or task notes. A file over 500 lines MUST be split before completion unless a documented
  human exception exists.
- Functions and types MUST perform one coherent responsibility and MUST be named for that
  responsibility.
- Utility grab-bags, vaguely named helpers, and shared mutable globals are forbidden. Immutable
  process configuration and composition-root wiring are permitted.
- CLI entrypoints, including `main` and command registration, MUST contain wiring only. Behavior
  MUST live in domain or application packages.

Rationale: small, purpose-specific modules make behavior easier to understand, test, replace, and
review without coupling unrelated Graph workloads.

### III. Smallest Sufficient Design

- Implementations MUST be the smallest design that satisfies the current specification and locked
  tests.
- Speculative abstractions, unused extension points, duplicated domain models, and wrapper layers
  that add no behavior are forbidden.
- Similar code SHOULD remain explicit until a third concrete use case demonstrates a stable
  abstraction.
- Validation MUST occur at system boundaries: CLI input, Graph responses, and filesystem or
  keychain access. Implementations MUST NOT handle states that cannot occur at those boundaries or
  within established domain invariants.
- The namespace registration contract required by the two v1 namespaces is not speculative.
  Implementations MUST provide only the seam needed to register and remove namespaces. They MUST
  NOT pre-build calendar, OneDrive, other workloads, a plugin marketplace, a generic provider SDK,
  or a configuration DSL without a later specification and locked tests.

Rationale: deliberate restraint minimizes maintenance cost while preserving the one extension seam
the product already requires.

### IV. Testing Standards: Right Boundary, Deterministic, Behavioral

- Unit tests MUST cover pure domain rules and application use cases. Graph, network, browser, and OS
  keychain dependencies MUST be replaced by fakes through application-owned ports. Unit tests MUST
  NOT call live Microsoft Graph.
- Contract and integration tests MUST cover CLI parsing, stdout, stderr, exit codes, and Graph
  request/response mapping through recorded synthetic fixtures or a fake Graph server.
- Acceptance tests MUST exercise user-observable login, mail, and teams CLI workflows generated
  from Gherkin features. Generated acceptance tests MUST remain separate from unit tests.
- The acceptance toolchain MUST follow an Acceptance-Pipeline-Specification flow of Gherkin to an
  intermediate representation, generated tests, an optional dry-check, and example mutation. A
  heavyweight BDD runner MUST NOT be introduced unless a locked test demonstrates the need.
- Tests MUST be deterministic. Time, randomness, and HTTP MUST be injected at a seam. Retries MUST
  NOT conceal flaky behavior.
- Tests MUST assert user-observable behavior or stable contracts such as exit codes, stdout/stderr
  schemas, or domain results. They MUST NOT assert private fields, log strings, or internal call
  counts when an observable outcome exists.
- The plan and quickstart MUST document the commands that run each test suite.

Rationale: tests provide durable confidence only when they observe the correct boundary, run
repeatably, and remain independent of live user data and external services.

### V. CLI Experience Consistency

- Every user-facing command MUST read from arguments, flags, or stdin; write result data to stdout;
  and write diagnostics to stderr.
- Every command MUST support human-readable output and machine-readable JSON output. JSON written
  to stdout MUST be parseable and MUST NOT contain logs or diagnostics.
- The CLI MUST define and document stable exit codes for success, usage or configuration errors,
  authentication failures, Graph/API failures, and not-found results. Once a failure class exists,
  commands MUST NOT collapse all failure classes into exit code `1`.
- Each command MUST specify and test empty, validation, authentication-failure, Graph-error, and
  happy-path behavior. A command with an unhandled required state is incomplete.
- Each command MUST expose `--help` that identifies its namespace, required flags, and output modes.
  New namespaces MUST reuse established flag, output, and error conventions unless the plan records
  a reason for a new pattern.
- Commands and verbose or debug output MUST NOT print access tokens, refresh tokens, authorization
  codes, or client secrets. Sensitive values MUST be redacted before output.
- Top-N limits, pagination, and sampling MUST be visible in command output or documentation. Silent
  truncation is forbidden.

Rationale: stable streams, formats, help, and failure contracts make the CLI predictable for both
people and automation.

### VI. Performance: Measurable Expectations

- A plan for any change affecting latency, memory, binary size, or Graph call volume MUST name the
  metric, threshold, and repeatable measurement method before implementation.
- Graph list and pagination operations MUST define an explicit page size or result limit. Silent
  full-mailbox and full-chat retrieval is forbidden.
- A measured regression against a documented performance target MUST be treated as a defect and
  MUST block completion unless covered by a written exception.

Rationale: explicit limits protect local resources, user time, and Graph quotas, while measurable
targets make performance claims reviewable.

### VII. Domain Isolation and Namespace Extensibility

- Each Graph workload MUST be a namespace with its own domain types and application use cases. The
  v1 namespaces are `mail` and `teams`.
- Namespace packages MUST NOT import one another. Shared authentication, configuration, Graph
  transport, output formatting, and error types MUST live in core packages.
- The CLI surface MUST follow `m365 <namespace> <verb> [flags]`. Adding or removing a namespace MUST
  NOT require edits inside any other namespace.
- Graph HTTP details MUST remain in an outbound adapter. Domain and application code MUST use ports
  owned near their consumers and expressed in workload language, such as list messages, send mail,
  or list chats. They MUST NOT expose Graph URLs, JSON payloads, or SDK types.
- Domain packages MUST NOT import CLI, Graph SDK, OS, or HTTP packages. Application/use-case
  packages MUST own their ports; CLI, Graph, and keychain adapters MUST implement those ports.
- Later namespaces require their own specifications. The architecture MUST provide the registration
  seam but MUST NOT implement future workload behavior in advance.

Rationale: dependency direction isolates business behavior from Microsoft and operating-system
details while allowing workloads to evolve independently.

### VIII. Secrets, Auth, and Least Privilege

- Delegated OAuth for a signed-in Microsoft 365 user MUST be the only v1 authentication path.
  Passwords MUST NOT be accepted through flags or stdin.
- Client identifiers, tenant identifiers, client secrets, access tokens, refresh tokens, and
  authorization codes MUST NOT be embedded in this constitution, source, tests, fixtures, logs,
  commits, or CI output.
- Tokens and secrets MUST be stored in the OS keychain or secret store, or in an equivalent
  `0600`-permission backend selected during planning. World-readable token files are forbidden.
- Graph scopes MUST be least-privilege and namespace-scoped. Enabling mail MUST NOT require Teams
  scopes, and enabling Teams MUST NOT require mail scopes.
- Authentication failures MUST have an exit code and stderr contract distinct from Graph/API
  failures.
- Fixtures and tests MUST use fake tokens and synthetic Graph bodies. Live mailbox or chat content
  MUST NOT be recorded in the repository.

Rationale: delegated access handles private user data; least privilege, secret isolation, and
synthetic fixtures reduce both exposure and blast radius.

## Engineering Constraints

- The implementation MUST follow clean architecture dependency direction: domain is independent;
  application packages own use cases and ports; adapters connect CLI, Graph, keychain, filesystem,
  and HTTP concerns; composition occurs at the CLI entrypoint.
- The implementation language and concrete tools MUST be selected in the plan, not in this
  constitution.
- A Makefile or language-appropriate equivalent MUST expose named blocking targets for formatting,
  linting or vetting, unit tests, concurrency or race checks where supported, coverage, security
  scanning, dependency vulnerability checks, acceptance tests, acceptance-example mutation, and
  complexity/CRAP or an equivalent risk metric.
- Quality gates MUST block completion unless a written human-approved exception identifies the
  governing principle, rationale, exact scope, and retirement action.
- A bug fix MUST NOT include a surrounding refactor. Unrelated cleanup MUST be a separate change
  and MUST begin with its own RED test when behavior is affected.
- Plans and user documentation MUST disclose every top-N limit, page boundary, and sampling rule.

## Development Workflow & Quality Gates

1. **Specify**: A specification MUST define observable behavior, boundaries, failure states, and
   acceptance scenarios without selecting unnecessary implementation details.
2. **Constitution Check**: Every plan MUST include a pass/fail check for every principle in this
   constitution. A failure MUST cite its written human-approved exception, including rationale,
   scope, and retirement action, before implementation begins.
3. **RED checkpoint**: The failing test and required fixtures MUST be committed separately after its
   expected failure has been observed. That commit MUST contain no production implementation.
4. **GREEN checkpoint**: The smallest implementation that passes the locked test and all applicable
   gates MUST be committed separately.
5. **REFACTOR checkpoint**: Refactoring MAY follow in a separate commit only while all tests remain
   GREEN. It MUST NOT include unrelated cleanup.
6. **Review**: Reviewers MUST verify architecture boundaries, observable CLI contracts, secret
   handling, performance limits, and all named quality gates. A change request enforcing this
   constitution MUST cite the relevant principle or section.
7. **Completion**: Work MUST NOT be complete until all applicable named gates pass, acceptance
   artifacts are synchronized, documented measurements meet their thresholds, and every exception
   or external blocker is recorded.

## Governance

- This constitution governs specifications, plans, tasks, implementation, reviews, and refactors.
  When another project document conflicts with it, this constitution MUST take precedence.
- Runtime agent guidance, including `AGENTS.md`, `CLAUDE.md`, and equivalent files, MUST NOT
  contradict this constitution. Such guidance MAY add stricter, non-conflicting procedures.
- An amendment MUST document its rationale, affected principles or sections, migration or retirement
  actions, and explicit human approval. The amended file MUST include a temporary Sync Impact Report
  comment listing the version change, modified principles, added or removed sections, and follow-up
  work; that comment SHOULD be removed before commit after human review.
- Constitution versions MUST use semantic versioning. MAJOR increments apply to incompatible
  governance changes, principle removals, or principle redefinitions. MINOR increments apply when a
  principle or section is added or materially expands obligations. PATCH increments apply only to
  clarifications, wording corrections, and other non-semantic refinements.
- Every plan MUST perform the Constitution Check defined above. Every code review MUST verify
  compliance and MUST reject an unexplained violation of a MUST requirement.
- Waiving any MUST requires written human approval in the plan or task notes naming the principle,
  rationale, scope, and retirement action. An exception MUST be narrow and MUST NOT silently become
  precedent.

**Version**: 1.0.0 | **Ratified**: 2026-09-16 | **Last Amended**: 2026-09-16

## Next Actions

- Specify the CLI, delegated OAuth login, mail namespace, and teams namespace with
  `/speckit-specify`.
- After specification, choose Go or Rust, Graph SDK or raw HTTP, the OAuth public-client flow, and
  the token storage backend with `/speckit-plan`.
