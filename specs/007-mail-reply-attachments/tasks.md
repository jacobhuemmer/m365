# Tasks: Email reply attachments

## Setup
- [x] T001 Obtain/configure authorized Git user.name and user.email; repository-local identity configured as Mason Huemmer <49791141+jacobhuemmer@users.noreply.github.com>.
- [x] T002 Read constitution/design and run baseline checks from Makefile in existing worktree.
## User Story 1 — Reply/replyAll delivery
- [x] T003 [US1] Add internal/adapters/graph/attachment_reply_test.go covering both selectors, names/Graph @odata.type/decoded bytes, exact rendered comment and no message.body, no-file parity, a later unreadable file after an earlier valid file (zero sends), and attachment-bearing Graph rejection on both endpoints (no success/fallback).
- [x] T004 [US1] Add internal/adapters/cli/attachment_reply_test.go for CLI/MCP real HTTP payload propagation, previews and failures; synthetic files/sessions only.
- [x] T005 [US1] Add features/attachments/reply-attachments.feature and acceptance/steps/reply_attach_steps.go using real HTTP adapter, regenerate acceptance artifacts.
- [x] T006 [US1] Run focused regressions; observe expected missing-attachments RED, commit tests/required fixtures only, lock tests.
- [x] T007 [US1] Modify internal/adapters/graph/httpsend.go with existing fileAttachments and optional message.attachments alongside comment; propagate errors before send, no text-only fallback.
- [x] T008 [US1] Update docs/m365.md and internal/adapters/cli/help.go for reply delivery and limits; no Teams changes or live delivery claims.
- [x] T009 [US1] Run focused tests, make verify and make acceptance-mutation, measure send counts, review diff and confirm locked tests unchanged; separate GREEN commit.
## Completion
- [ ] T010 Save code-review and independent critique, address hard findings, record checks/commit IDs in tasks.md and QRSPI notes.

## Dependencies
T001–T002 → T003–T005 → T006 RED → T007–T009 GREEN → T010. One vertical mail slice; shared source edits sequential.

## Manual checks pending
Controlled recipient checks in plan.md await explicit live-send authorization. The mail-only scope does not close the Teams portion of SDO-564.

## Automated evidence — 2026-09-28
- T002: Constitution, approved design/structure and TDD references read; baseline `make verify` and `make acceptance-mutation` passed at 9252a31.
- T003–T005: Real HTTP adapter exercised through public HTTPClient.Reply, CLI Run and MCP m365_run. Synthetic temporary files and fake sessions only. Dedicated Gherkin acceptance scenarios and generated artifacts synchronized.
- T006: Locked test-only RED commit `3e9f580`. `go test ./internal/adapters/graph ./internal/adapters/cli -run TestReplyAttachment` failed because both selectors omitted message.attachments; a later unreadable file returned sent after one request. `make acceptance` failed for the same missing-attachment behavior. No compile/import/fixture failures.
- T007–T008: Existing fileAttachments encoder reused; all reads precede the sole POST. Only optional message.attachments added alongside comment. No body/recipient overrides, new scopes or Teams edits. Help and contract disclose local limits and service rejection behavior.
- T009: Focused regressions, `make verify` and `make acceptance-mutation` passed. No gate-fix attempts required. No diff in locked tests/fixtures from RED commit. Measured zero requests on later file-read failure and previews; exactly one POST on accepted sends and Graph rejection, with no fallback/success on rejection. No-file request remains comment-only.
- Logs: ignored worktree build/baseline.log, build/baseline-mutation.log, build/reply-red.log, build/reply-acceptance-red.log, build/reply-green.log, build/reply-verify.log and build/reply-mutation.log.
- T010 remains pending independent review and critique. No live recipient checks performed; synthetic 202 proves request serialization only. Teams portion of SDO-564 remains deferred.
