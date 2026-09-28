# Implementation Plan: Email reply attachments

Branch: SDO-564-attachment-delivery. Design: ~/.work/m365/SDO-564-attachment-delivery/design.md.

## Summary
Reuse fileAttachments to populate message.attachments alongside the existing reply comment, only when files exist. Read all files before the single reply/replyAll POST. Leave body and recipients to existing Graph reply semantics. Exact combined JSON delivery remains a controlled live-check item; stable API/SDK evidence supports the candidate.

## Technical Context
Existing Go 1.25 / pinned toolchain 1.26.6; net/http/httptest, existing MCP SDK, no added dependencies/storage/scopes. Local file cap 10 MiB, count 10, unchanged. These are validation ceilings rather than service guarantees.

## Constitution Check
I: observed test-only locked RED then minimal GREEN, repository-local author identity confirmed before RED commit. II: focused files below 250 lines; no refactor. III: reuse existing encoder and payload builder, no extra abstraction. IV: public HTTP and CLI/MCP seams, deterministic synthetic files/fake sessions, no live calls. V: preserve streams/exits/schemas/help. VI: measure zero POSTs on local file failure and one POST on success; no extra calls for no-file replies. VII: Graph details stay in adapter; ports unchanged. VIII: synthetic data only, auth/scopes unchanged. All principles pass at design level; automated execution gates passed; Standards/Spec and independent review passed without hard findings. No waivers.

## Project Structure
Modify internal/adapters/graph/httpsend.go. Add focused attachment_reply_test.go under graph and cli. Add features/attachments/reply-attachments.feature and dedicated acceptance/steps/reply_attach_steps.go; regenerate acceptance tests. Clarify supported email reply attachments in docs/m365.md and relevant help, without claiming live verification.

## One Vertical Slice
1. Add HTTP/CLI/MCP regressions and Gherkin fixture handlers exercising real HTTP adapter. Confirm current requests lack attachments and fail for expected behavioral reason. Commit only tests/fixtures as locked RED. Do not change previously locked tests without user approval.
2. Minimal GREEN: call fileAttachments, propagate error, construct comment payload plus optional message.attachments; no message.body/recipient overrides. Read failures send nothing; Graph failures propagate. Update capability docs; keep Teams deferred.
3. Focused tests and make verify plus make acceptance-mutation. Save reviews and metrics, commit GREEN separately. All automated checks must pass; do not claim recipient delivery from synthetic 202 alone.

## Manual checks pending
- [ ] Authorized controlled reply and reply-all: recipient receives exact file names/bytes, authored reply, quoted thread and correct recipients. No live send authorization exists yet.

## Readiness
Git identity confirmed repository-locally: Mason Huemmer <49791141+jacobhuemmer@users.noreply.github.com>. Test-only RED locked in 3e9f580; minimal GREEN implemented and automated gates passed. No open local implementation design questions; live service behavior is explicitly pending verification.

## Results
Public HTTP, CLI and MCP attachment regressions and dedicated generated acceptance pass. Baseline and GREEN `make verify` plus `make acceptance-mutation` pass; no gate-fix attempts. Zero requests on file-read failure or previews; exactly one POST on success or service rejection. Locked tests/fixtures are unchanged. See tasks.md for commands, RED evidence and review status.

Review result: no required code fixes. One fixture-wording heuristic recorded in QRSPI review report; locked regression tests preserved. Full go test ./... passed after GREEN. Manual recipient checks remain pending.
