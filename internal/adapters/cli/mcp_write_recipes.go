package cli

const (
	recipeMailWrite = `mail-write

Write Outlook mail as real HTML, not a markdown dump. No session required to read this recipe.

Use <p> for paragraphs, <ul>/<ol> for lists, and <a href> for links. Do not send a markdown heading as the mail body.

  mail send --to you@example.com --subject 'Status' --html --body '<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul><p>Track it in <a href="https://example.com/ticket">the ticket</a>.</p>' --dry-run
  mail reply MESSAGE_ID --html --body '<p>Agreed.</p><p>I will update the ticket.</p>' --dry-run

MCP (flags.body string, not body-file=-):

  m365_run namespace=mail verb=send flags to=you@example.com subject=Status html=true body='<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>'
  m365_run namespace=mail verb=reply args=[MESSAGE_ID] flags html=true body='<p>Agreed.</p><p>I will update the ticket.</p>'

Happy-path examples stay dry-run (write_opt_in false). MCP writes dry-run unless write_opt_in is true.
`

	recipeTeamsWrite = `teams-write

Write a Teams chat message that renders: paragraphs, lists, links. No session required to read this recipe.

Use --html with real HTML, or --format md with the documented subset (# / ## / ###, **bold**, - / * / 1. lists, [label](url), fenced code, blank-line paragraphs). --format md converts that subset to HTML. --html posts the body as HTML already.

  teams send --to Ajay --format md --text 'The change is in UAT.

- Rollback is the previous chart.

See [the ticket](https://example.com/ticket).' --dry-run
  teams send --to Ajay --html --text '<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>' --dry-run

MCP (flags.text string, not text-file=-):

  m365_run namespace=teams verb=send flags to=Ajay format=md text='The change is in UAT.

- Rollback is the previous chart.'
  m365_run namespace=teams verb=send flags to=Ajay html=true text='<p>The change is in UAT.</p><ul><li>Rollback is the previous chart.</li></ul>'

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
