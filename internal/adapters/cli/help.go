package cli

const rootHelp = `m365 — Microsoft 365 CLI for one signed-in user

Usage: m365 <namespace> <verb> [flags]

Namespaces:
  auth      sign in, status, logout (core, not a workload)
  mail      Outlook mail
  teams     Teams chats (alias: chat)
  calendar  own calendars and events
  files     own OneDrive
  mcp       stdio MCP for agents (serve)

Output: JSON on stdout by default. --human for text. Diagnostics on stderr.
Exit classes: 0 success; 3 usage/config; 4 auth; 5 service; 6 not-found.

List limits: --top default 10 (mail list, calendar list), 20 (teams list, calendar calendars, files list); max 50.
Attachments: max 10 MiB per file, max 10 files per send/reply.
Upload: max 100 MiB per file.
`

const calendarHelp = `m365 calendar — own calendars and events

Verbs: calendars, list, get, create, update, delete, free
Create: --subject --when 'tomorrow at 1:30 pm' --until --duration (default 30 minutes) --dry-run
Free: --when (default tomorrow) --duration (default 30 minutes) --hours 9:00-17:00 --top (default 5, max 20)
Working hours 09:00-17:00 local; 30-minute grid. --dry-run does not apply to free.
Flags: --calendar --start --end --top (events default 10, calendars default 20, max 50) --page-token
Window for list: default now through +7 days.
Output: JSON (default) or --human.
`

const filesHelp = `m365 files — own OneDrive

Verbs: root, list, get, download, upload, create-folder, delete, move
Flags: --folder --top (default 20, max 50) --page-token --out --overwrite --file --name --dry-run
Upload: max 100 MiB per file. Download requires --out; existing file refused without --overwrite.
Output: JSON (default) or --human. No file bytes on stdout.
`

const mailListHelp = `m365 mail list — list messages

Flags: --folder (default inbox) --unread --search --top (default 10, max 50) --page-token
Well-known folders: inbox, sentitems, drafts, all
Output: JSON (default) or --human. Exit 0 empty list.
`

const mailWatchHelp = `m365 mail watch — poll one folder for incremental message changes

Flags: --folder (default inbox) --include-existing
The first poll records a quiet baseline unless --include-existing is set.
Folder-scoped only: the mailbox-wide all value is not supported.
JSON output is one body-free mail.changed object per line; empty polls write no lines.
The cursor is committed only after the complete delta round is emitted.
`

const mailSendHelp = `m365 mail send — send mail

Required: --to --subject --body or --body-file
Optional: --cc --attach (repeatable) --html --dry-run
Attachment caps: 10 MiB per file, 10 files. --top N/A.
Output modes: --json (default) --human
`

const teamsListHelp = `m365 teams list — list chats

Flags: --top (default 20, max 50) --page-token
Output: JSON (default) or --human.
`

const teamsFindHelp = `m365 teams find — find a chat by person or group

Examples: teams find Ajay
          teams find --group NOC
Flags: --group --top (default 10, max 20)
Person query prefers 1:1. --group matches group topic/members.
Scan ceiling 10 pages of 50 chats; incomplete=true if not finished.
Empty list exit 0. Never sends.
Output: JSON (default) or --human.
`

const teamsSendHelp = `m365 teams send — send a chat message

Required: CHAT_ID or --to Ajay, plus --text or --text-file
Examples: teams send --to Ajay --text ping --dry-run
Optional: --attach (repeatable) --html --format md --dry-run
--to and chat id together exit 3. Attachment caps: 10 MiB per file, 10 files.
Output modes: --json (default) --human
`
