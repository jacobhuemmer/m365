# Live email reply attachment validation — 2026-09-28

## Authorization and scope
User explicitly requested: Proceed with live reply-attachment test. Built the local branch CLI at 6e4297d and created a unique self-addressed synthetic test thread. Sent one reply and one reply-all, each with two small local text documents. No other recipients, existing customer threads, Teams operations or new auth scopes.

## Results
- Both dry-runs showed two files and no formatting problems; both real sends returned success.
- Read-only verification queried Graph's actual Inbox folder, filtered to the synthetic conversation, with an explicit ten-message limit. Received three messages: seed, reply, reply-all.
- Each received reply contains reply-document.txt (59 bytes) and reply-notes.txt (43 bytes, including UTF-8 content).
- Downloaded each received attachment through /me/messages/{id}/attachments/{attachment-id}/$value. SHA-256 matched the local source for all four received files.
- Both delivered HTML bodies retain the original SDO564-SEED marker alongside the authored operation-specific reply text.
- Both received To lists contain only the signed-in test account; no CC recipients. This verifies recipients for this self-addressed case.

## Evidence
Live CLI: mail send to signed-in account; mail reply with two --attach flags, with and without --all; dry-run before each send. Synthetic test subject was unique; do not store mailbox IDs or live account details here. Read-only probe used existing process-local credential handling; no credential values were observed. Verification summary:
{
  "inbox_test_messages": 3,
  "results": [
    {
      "attachment_bytes_match": true,
      "attachment_names": [
        "reply-document.txt",
        "reply-notes.txt"
      ],
      "operation": "REPLY",
      "quoted_original_present": true,
      "received_in_inbox": true,
      "recipients_match_self_addressed_thread": true
    },
    {
      "attachment_bytes_match": true,
      "attachment_names": [
        "reply-document.txt",
        "reply-notes.txt"
      ],
      "operation": "REPLYALL",
      "quoted_original_present": true,
      "received_in_inbox": true,
      "recipients_match_self_addressed_thread": true
    }
  ]
}

## Limits
Self-addressed small-file scenario only. Multi-person reply-all, actual Outlook UI rendering, large payload limits and other file types were not exercised. Teams remains deferred. Synthetic test messages remain in the user's mailbox; no deletion performed. Scratch data is outside the repository and is temporary.
