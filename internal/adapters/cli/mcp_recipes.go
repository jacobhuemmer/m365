package cli

const (
	recipeMailSearch = `mail-search

Search mail, then open a message or thread. No session required to read this recipe.

  mail list --folder all --search 'from:ajay'
  m365_run namespace=mail verb=list flags folder=all search=from:ajay

Also: search subject:… ; default --top applies. Then mail get and mail thread with the message id.

Do not dump the mailbox. Do not print secrets.
`

	recipeTeamsFind = `teams-find

Find a chat before notify. Use teams list and match members or topic. A dedicated name-lookup verb is not in this CLI yet.

  1:1: teams list, pick the one-on-one whose other member is the person (e.g. Ajay).
  Group: teams list, match topic (e.g. NOC) or members.

Several matches MUST NOT send. Zero matches: do not guess. Dry-run before notify (write_opt_in false / --dry-run).

Do not invent a Graph people-search tool.
`

	recipeCalendar = `calendar

  calendar list
  calendar free --when tomorrow
  calendar create --when 'tomorrow at 1:30 pm' --subject '…' --dry-run

MCP writes dry-run unless write_opt_in is true.
`

	recipeFiles = `files

  files list
  files download ITEM_ID --out /path/to/dest
  files upload --file /path/to/local --dry-run

Download writes only to --out. File bytes MUST NOT appear in MCP results.
`
)

var recipeBodies = map[string]string{
	"mail-search": recipeMailSearch,
	"teams-find":  recipeTeamsFind,
	"calendar":    recipeCalendar,
	"files":       recipeFiles,
}

var recipeNames = []string{"mail-search", "teams-find", "calendar", "files"}

func recipe(topic string) (string, bool) {
	s, ok := recipeBodies[topic]
	return s, ok
}
