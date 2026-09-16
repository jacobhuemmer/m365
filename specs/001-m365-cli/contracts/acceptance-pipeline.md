# Acceptance pipeline contract

m365 uses `unclebob/Acceptance-Pipeline-Specification`. Feature files are
product behavior; generated Go tests are projections.

## Inputs

```text
features/auth/*.feature
features/mail/*.feature
features/teams/*.feature
features/attachments/*.feature
features/cli/*.feature
```

Feature files MUST use product language (session, mail, chat, attachment,
dry-run, exit class). They MUST NOT name Go, Cobra, kiota, HTTP paths, or
MIME internals. Parameters for ids, limits, paths, and sizes. Fake Graph
and fake secret store by default.

## Normal flow

```text
gherkin-parser <feature> build/acceptance/ir/<feature>.json
gherkin-ir-dry-checker build/acceptance/ir/<feature>.json build/acceptance/dry/<feature>.json
acceptance-entrypoint-generator build/acceptance/ir/<feature>.json acceptance/generated/<feature>_acceptance_test.go
go test ./acceptance/generated
```

`scripts/acceptance.sh` owns this for developers and CI.

## Generated tests

- Live under `acceptance/generated/`; not edited by hand.
- Use `testing`.
- Load IR deterministically.
- Dispatch through `acceptance/runtime`.
- Fail if a step handler is missing.
- Do not replace unit tests.

## Runtime

- Fresh world per scenario/example.
- Fake clock and fake Graph.
- No sleeps to hide flakes.
- Interactive login steps use a fake browser/callback, not a real gesture, except documented manual quickstart.

## Mutation

`scripts/acceptance-mutation.sh` runs `gherkin-mutator` with
`acceptance/runner`. Classify `test_success`, `test_failure`,
`infrastructure_error`. Reports under `build/acceptance-mutation/`.
Hardening workflow: before release and after acceptance-contract changes.

Pin APS tooling to a reviewed upstream commit SHA (no release tags).
