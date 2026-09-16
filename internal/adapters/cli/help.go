package cli

const rootHelp = `m365 — Microsoft 365 CLI for one signed-in user

Usage: m365 <namespace> <verb> [flags]

Namespaces:
  auth    sign in, status, logout (core, not a workload)
  mail    Outlook mail
  teams   Teams chats (alias: chat)

Output: JSON on stdout by default. --human for text. Diagnostics on stderr.
Exit classes: 0 success; 3 usage/config; 4 auth; 5 service; 6 not-found.

List limits: --top default 10 (mail list), 20 (teams list/messages/watch); max 50.
Attachments: max 10 MiB per file, max 10 files per send/reply.
`

const mailListHelp = `m365 mail list — list messages

Flags: --folder (default inbox) --unread --search --top (default 10, max 50) --page-token
Well-known folders: inbox, sentitems, drafts, all
Output: JSON (default) or --human. Exit 0 empty list.
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

const teamsSendHelp = `m365 teams send — send a chat message

Required: CHAT_ID --text or --text-file
Optional: --attach (repeatable) --html --format md --dry-run
Attachment caps: 10 MiB per file, 10 files.
Output modes: --json (default) --human
`
