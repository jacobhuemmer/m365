# Data model
No new public types. mail.ReplyInput carries ID, All, Files and Rendered. Existing domain.OutboundFile carries Path/Name/Size; existing encoder reads and base64-encodes bytes. JSON adds message.attachments only for files; comment retains rendered reply content. No message.body or recipient overrides. Failure returns no successful ID.
