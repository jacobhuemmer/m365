# m365

A command-line tool for **your** Microsoft 365 account: Outlook, Teams, calendar, and OneDrive.

JSON on stdout by default. Add `--human` if you want plain text. Writes stay dry-run until you drop `--dry-run` (or, for agents, set `write_opt_in`).

Current release: **0.1.1**.

## Install

### macOS (Homebrew)

```sh
brew tap jacobhuemmer/tap
brew install m365
m365 --version    # 0.1.1
```

Upgrade later with `brew update && brew upgrade m365`.

To build the latest `main` instead of a numbered release:

```sh
brew install --HEAD jacobhuemmer/tap/m365
```

### Windows (Scoop)

```powershell
scoop bucket add jacobhuemmer https://github.com/jacobhuemmer/scoop-bucket
scoop install m365
m365 --version
```

### Windows (WinGet)

The first community package is in review: [microsoft/winget-pkgs#439501](https://github.com/microsoft/winget-pkgs/pull/439501). After that merges:

```powershell
winget install JacobHuemmer.m365
```

Until then, use Scoop or a zip from [Releases](https://github.com/jacobhuemmer/m365/releases).

### From source

Go 1.25+:

```sh
git clone https://github.com/jacobhuemmer/m365.git
cd m365
make install
```

## Sign in

You need an Entra app registration that you are allowed to use (delegated, as yourself). Put the IDs in `~/.config/m365/.env`:

```sh
mkdir -p ~/.config/m365
cp .env.example ~/.config/m365/.env
```

```
M365_CLIENT_ID=...
M365_TENANT_ID=...
```

You can export the same names instead of a file. Then, in a real terminal (browser login):

```sh
m365 auth login
m365 auth status
```

`signed_in` and `session_usable` should be true, with the namespaces you granted (`mail`, `teams`, `calendar`, `files`). Tokens go in the macOS keychain, with a file fallback if needed.

Do not run `auth login` through MCP. If status is not usable, stop and log in from a terminal.

## Everyday use

```
m365 <area> <command> [flags]
```

`--help` on any command does not need a session. `chat` is an alias for `teams`.

| Area | What it does |
| --- | --- |
| `auth` | `login`, `status`, `logout` |
| `mail` | list, read, send, reply |
| `teams` | find a chat, read, send |
| `calendar` | list, free slots, create |
| `files` | OneDrive list, download, upload |
| `mcp` | `serve` for agents |

Exit codes: `0` ok, `3` usage/config, `4` sign-in, `5` Microsoft Graph, `6` not found.

Preview any send with `--dry-run` first.

### Mail

```sh
m365 mail list --unread --top 10 --human
m365 mail get MESSAGE_ID
m365 mail send --to you@example.com --subject 'Status' --body 'In UAT.' --dry-run
m365 mail send --note-to-self --body 'Remember this.' --dry-run
```

HTML mail (paragraphs, lists, links): `--html` and a real HTML `--body`. Replies with `--html` go out as HTML, not a jammed comment.

### Teams

```sh
m365 teams find Ajay
m365 teams send --to Ajay --text 'Looking into this.' --dry-run
m365 teams send --note-to-self --text 'Scratch note.' --dry-run
```

`--note-to-self` is **Chat with yourself** (`48:notes`), not a hidden 1:1. Do not combine it with `--to` or a chat id.

For a list or a link, use `--format md` (a small markdown subset becomes HTML) or `--html` with HTML already written.

After a unique `teams find` or a listed 1:1, later `--to Ajay` remembers the chat id so it does not scan the whole list again.

### Calendar and files

```sh
m365 calendar list --human
m365 calendar free --when tomorrow
m365 calendar create --when 'tomorrow at 1:30 pm' --subject 'Sync' --dry-run
m365 files list --human
```

## Agents (MCP)

```sh
m365 mcp serve
```

Stdio JSON-RPC. Three tools: `m365_status`, `m365_help`, `m365_run`. Six recipe prompts: `mail-search`, `teams-find`, `calendar`, `files`, `mail-write`, `teams-write`. Do not pass `--human`. Writes through `m365_run` stay dry-run unless `write_opt_in` is true.

This repo has no `m365 skill` command. Copy `skills/<topic>/SKILL.md` into an agent skill root if you want files on disk:

```text
Cursor    .cursor/skills/<topic>/SKILL.md
Claude    .claude/skills/<topic>/SKILL.md
Codex     .codex/skills/<topic>/SKILL.md
Grok      .grok/skills/<topic>/SKILL.md
OpenCode  .opencode/skills/<topic>/SKILL.md
```

Full CLI contract for agents: [docs/m365.md](docs/m365.md).

## Experimental: mail watch

`mail watch` is one folder-scoped poll, then exit. Classification is off unless you enable it in `~/.config/m365/config.json`. Keep `TYPESAFE_API_KEY` in the environment, never in that file. An `actionable` event is not permission to send; still `--dry-run`, then an explicit send.

## Develop

```sh
sh scripts/install-tools.sh
make verify
```

## License

MIT
