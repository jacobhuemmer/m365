# Reply attachment feasibility — 2026-09-28

## Stable first-party contract evidence
- v1.0 reply accepts comment plus message writable properties; comment and message.body cannot both be supplied. Mail.Send is sufficient. MIME explicitly allows embedded attachments. https://learn.microsoft.com/en-us/graph/api/message-reply?view=graph-rest-1.0
- v1.0 replyAll documents MIME attachments and Mail.Send, although its JSON parameter table omits message. https://learn.microsoft.com/en-us/graph/api/message-replyall?view=graph-rest-1.0
- Microsoft's stable Go SDK exposes Message on BOTH reply and replyAll request bodies, including serializer/deserializer and getters/setters. Inspected from repository main on 2026-09-28: https://github.com/microsoftgraph/msgraph-sdk-go/blob/main/users/item_messages_item_reply_post_request_body.go and https://github.com/microsoftgraph/msgraph-sdk-go/blob/main/users/item_messages_item_reply_all_post_request_body.go
- fileAttachment JSON includes name, base64 contentBytes and @odata.type. https://learn.microsoft.com/en-us/graph/api/resources/fileattachment?view=graph-rest-1.0

## Inference and limits
The smallest delivery candidate is the existing rendered comment plus message.attachments, using the existing fileAttachments encoder; neither message.body nor recipient/header overrides are needed. This is supported by stable message parameters and attachment representation, but the exact combined JSON request and recipient experience require a controlled live verification. SDK support alone does not prove attachment delivery, size acceptance or quoted-thread rendering. Do not treat beta examples or a synthetic 202 as recipient proof.

Existing 10-MiB/file and 10-file local limits are validation ceilings, not a Graph delivery guarantee. Server rejections must propagate as failures; never retry by dropping attachments. MIME and draft upload workflows remain alternatives, not automatic fallback. Draft workflows require Mail.ReadWrite and are outside current authorization.

## Scope steering
The user requested actual email-reply attachment support. This supersedes rejection-only mail behavior. Focus on reply/replyAll; Teams work is deferred. Completing this work alone will not satisfy the Teams portion of SDO-564.
