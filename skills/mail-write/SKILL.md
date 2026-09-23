---
name: mail-write
description: Write Outlook mail or replies as HTML paragraphs, lists, and links with m365 mail send/reply --html.
---

# mail-write

Write Outlook mail as real HTML, not a markdown dump. No session required to read this recipe.

Use <p> for paragraphs, <ul>/<ol> for lists, and <a href> for links. Do not send a markdown heading as the mail body.

Plain text is fine for a short reply: blank lines become paragraphs and newlines stay line breaks. Replies keep the quoted thread either way.

  mail send --to you@example.com --subject 'Status' --html --body '<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul><p>Track it in <a href="https://example.com/ticket">the ticket</a>.</p>' --dry-run
  mail reply MESSAGE_ID --html --body '<p>Agreed.</p><p>I will update the ticket.</p>' --dry-run

MCP (flags.body string, not body-file=-):

  m365_run namespace=mail verb=send flags to=you@example.com subject=Status html=true body='<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>'
  m365_run namespace=mail verb=reply args=[MESSAGE_ID] flags html=true body='<p>Agreed.</p><p>I will update the ticket.</p>'

Happy-path examples stay dry-run (write_opt_in false). MCP writes dry-run unless write_opt_in is true.
