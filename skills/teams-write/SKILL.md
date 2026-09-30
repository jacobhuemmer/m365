---
name: teams-write
description: Write Teams chat messages as HTML or converted markdown so they render as paragraphs, lists, and links.
---

# teams-write

Write a Teams chat message that renders: paragraphs, lists, links. No session required to read this recipe.

Use --html with real HTML, or --format md with the documented subset (# / ## / ###, **bold**, - / * / 1. lists, [label](url), fenced code, blank-line paragraphs). --format md converts that subset to HTML. --html posts the body as HTML already. Plain text keeps its paragraphs and line breaks.

  teams send --to Ajay --format md --text 'The change is in UAT.

- Rollback is the previous chart.

See [the ticket](https://example.com/ticket).' --dry-run
  teams send --to Ajay --html --text '<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>' --dry-run

MCP (flags.text string, not text-file=-):

  m365_run namespace=teams verb=send flags to=Ajay format=md text='The change is in UAT.

- Rollback is the previous chart.'
  m365_run namespace=teams verb=send flags to=Ajay html=true text='<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>'

Teams: --attach PATH (repeatable) uploads each file to your OneDrive (Microsoft Teams Chat Files), gives every other chat member read access without emailing them, and posts the message with a file card. Dry-run lists the files and share_with (who gets access): check it first. A chat member with no email address makes the send fail before anything is uploaded. Notes (--note-to-self) keeps the file private to you. If sharing or posting fails, nothing is reported as sent; the error names any file already uploaded.

  teams send --to Ajay --text 'Report attached.' --attach ~/report.pdf --dry-run

MCP: flags attach=[/path/report.pdf]; requires write_opt_in for the real send.

Do not post one run-on --text string with markdown left unconverted. Mentions, Adaptive Cards, and Graph beta markdown are out of scope.
Happy-path examples stay dry-run (write_opt_in false).
