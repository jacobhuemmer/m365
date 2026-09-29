# Feature Specification: Email reply attachments

**Branch**: SDO-564-attachment-delivery
**Created**: 2026-09-28
**Status**: Implemented; automated validation passed; independent review and controlled recipient checks pending
**Input**: User requests actual document attachments on email replies, focusing on mail.

## User Scenarios & Testing
### User Story 1 — Reply with documents (P1)
The user attaches local documents to a reply or reply-all and expects the recipient to receive them alongside the reply.
**Independent test**: Submit both reply variants with synthetic files; observe all names and bytes in the outbound request and verify recipient delivery in an authorized controlled check.
**Acceptance scenarios**:
1. Valid reply/reply-all with files carries every requested file and the authored reply content.
2. A missing/unreadable file prevents sending and reports failure.
3. A service rejection reports failure without claiming delivery or retrying without files.
4. A reply without files retains existing body/thread/recipient behavior.
5. Dry-run lists names/sizes and sends/uploads nothing.
### Edge Cases
Multiple files, repeated base names, changed/unreadable paths, existing count/size validation, service/auth errors, plain/HTML/markdown content and both reply selectors.

## Requirements
- FR-001: Reply/reply-all MUST carry every requested valid local attachment or report failure; no silent omissions or text-only fallback.
- FR-002: Reply content, quoted thread and original recipient selection MUST retain existing semantics.
- FR-003: File read failure MUST occur before sending, use existing usage class, and report no success.
- FR-004: Service failures MUST retain existing error classes and report no successful send.
- FR-005: No-file replies, new-mail attachments, previews, validation ceilings and output schemas MUST retain existing behavior.
- FR-006: HTTP payload tests MUST capture both reply variants, attachment names and exact bytes; public clients MUST propagate outcomes.
### Key Entities
Outbound file (path/name/size), reply (message ID, reply-all selector, rendered content), send outcome (success/preview/failure).

## Success Criteria
- Every attachment name and byte sequence survives all synthetic delivery regressions.
- File read failures send zero messages; valid replies make one send request.
- No service error results in success or automatic resend without files.
- Controlled recipient verification confirms attachments, quoted thread and reply-all recipients before any production delivery claim.

## Assumptions
Existing authentication and local count/size validation are reused. Service limits can be lower than local validation ceilings and are reported as failures. Teams is deferred; this feature alone does not close all SDO-564 criteria. Live sends require separate explicit authorization.
