package cli

const (
	recipeMailWrite = `mail-write

Write Outlook mail as real HTML, not a markdown dump. No session required to read this recipe.

Use <p> for paragraphs, <ul>/<ol> for lists, and <a href> for links. Do not send a markdown heading as the mail body.

Plain text is fine for a short reply: blank lines become paragraphs and newlines stay line breaks. Replies keep the quoted thread either way.

--format md works for mail too, with the same subset as teams-write. Dry-run JSON shows rendered (the exact body that is sent) and format_problems. A send with any format problem fails with a usage error; fix the body and dry-run again.

  mail send --to you@example.com --subject 'Status' --html --body '<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul><p>Track it in <a href="https://example.com/ticket">the ticket</a>.</p>' --dry-run
  mail reply MESSAGE_ID --html --body '<p>Agreed.</p><p>I will update the ticket.</p>' --dry-run

MCP (flags.body string, not body-file=-):

  m365_run namespace=mail verb=send flags to=you@example.com subject=Status html=true body='<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>'
  m365_run namespace=mail verb=reply args=[MESSAGE_ID] flags html=true body='<p>Agreed.</p><p>I will update the ticket.</p>'

Attach local files with --attach PATH, repeatable, on mail send and mail reply (10 MiB per file, 10 files; missing, empty or oversize files exit 3 and send nothing). Dry-run lists each file's name and size, so check it before the real send.

  mail send --to you@example.com --subject 'Report' --body 'Report attached.' --attach ~/report.pdf --dry-run

MCP: flags attach=[/path/report.pdf]. Paths are local to the machine running m365.

Happy-path examples stay dry-run (write_opt_in false). MCP writes dry-run unless write_opt_in is true.
`

	recipeTeamsWrite = `teams-write

Write a Teams chat message that renders: paragraphs, lists, links. No session required to read this recipe.

Use --html with real HTML, or --format md with the documented subset (# / ## / ###, **bold**, - / * / 1. lists, [label](url), inline code in backticks, fenced code, blank-line paragraphs). --format md converts that subset to HTML. --html posts the body as HTML already. Plain text keeps its paragraphs and line breaks.

  teams send --to Ajay --format md --text 'The change is in UAT.

- Rollback is the previous chart.

See [the ticket](https://example.com/ticket).' --dry-run
  teams send --to Ajay --html --text '<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>' --dry-run

MCP (flags.text string, not text-file=-):

  m365_run namespace=teams verb=send flags to=Ajay format=md text='The change is in UAT.

- Rollback is the previous chart.'
  m365_run namespace=teams verb=send flags to=Ajay html=true text='<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>'

Dry-run JSON shows rendered (the exact body that is sent) and format_problems; a send with any format problem fails with a usage error.
Teams cannot send file attachments yet. A live teams send with --attach exits 3 and sends nothing; --dry-run still previews the file, which does not mean a live send would carry it. Do not retry with a text-only message that claims a file is attached. Tell the user, and offer to send the file by mail (mail-write) or share a link they already have.
Do not post one run-on --text string with markdown left unconverted. Mentions, Adaptive Cards, and Graph beta markdown are out of scope.
Happy-path examples stay dry-run (write_opt_in false).
`
)

func recipePromptDescription(name string) string {
	switch name {
	case "mail-write", "teams-write":
		return "Write recipe " + name
	default:
		return "Lookup recipe " + name
	}
}
