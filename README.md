# m365

Microsoft 365 CLI for one signed-in user. JSON on stdout by default; `--human` for text.

## Install

From source (Go 1.25+):

```sh
make install
```

Homebrew (HEAD):

```sh
brew tap jacobhuemmer/m365 https://github.com/jacobhuemmer/m365
brew install --HEAD jacobhuemmer/m365/m365
```

## Auth

Copy `.env.example` to `~/.config/m365/.env` (or export the same names) with an Entra app `M365_CLIENT_ID` and `M365_TENANT_ID`. Then:

```sh
m365 auth login
m365 auth status
```

Login is a browser/device flow in a terminal. Tokens live in the macOS keychain (file fallback if needed).

## Usage

Agent CLI contract (JSON, exits, verbs): [docs/m365.md](docs/m365.md).

```
m365 <namespace> <verb> [flags]
```

| Namespace | Verbs |
| --- | --- |
| `auth` | `login`, `status`, `logout` |
| `mail` | `list`, `get`, `thread`, `watch`, `send`, `reply`, `attachments`, `save-attachment` |
| `teams` (`chat`) | `list`, `find`, `get`, `send` |
| `calendar` | `calendars`, `list`, `get`, `create`, `update`, `delete`, `free` |
| `files` | `root`, `list`, `get`, `download`, `upload`, `create-folder`, `delete`, `move` |
| `mcp` | `serve` |

Exit classes: `0` success, `3` usage/config, `4` auth, `5` service, `6` not-found.

## Experimental mail response triage

`mail watch` performs one finite, folder-scoped Outlook delta poll and exits. The first poll records a quiet baseline unless `--include-existing` is set; later polls resume from the protected local checkpoint. JSON output is one body-free `mail.changed` event per line.

Response-responsibility classification is experimental and disabled by default. Enable it explicitly in `$XDG_CONFIG_HOME/m365/config.json` (or `~/.config/m365/config.json`):

```json
{
  "experimental": {
    "mail_response_classification": {
      "enabled": true,
      "provider": "jev",
      "model": "jev-latest",
      "actionable_threshold": 0.8
    }
  }
}
```

Set `TYPESAFE_API_KEY` in the environment, never in `config.json` or a command flag. A present key does not enable the feature, and an enabled feature with no key fails before polling mail.

```sh
m365 mail watch --classify --target-address mason@example.com --target-name "Mason Huemmer"
```

Classification sends TypeSafe Jev only the latest 10 chronological text messages within a 64 KiB body budget, plus participant, timestamp, subject, target, and truncation metadata. It never sends attachment bytes, Graph cursors, revisions, or credentials. Output uses `mail.response_classified` with one of `waiting_on_target`, `waiting_on_other`, `no_response_expected`, or `unclear`; `actionable` is only a routing hint.

An agent must keep classification and sending as separate trust decisions:

```sh
# 1. Consume an actionable event, then inspect its cited conversation.
m365 mail thread MESSAGE_ID --bodies

# 2. Compose text externally and preview the reply.
m365 mail reply MESSAGE_ID --body-file reply.txt --dry-run

# 3. Send only after review with a separate explicit command.
m365 mail reply MESSAGE_ID --body-file reply.txt
```

The classifier does not draft, create Outlook Draft items, or send. Through MCP, the final reply still requires `write_opt_in=true`; without it, `m365_run` forces a dry run even after an actionable event.

## MCP

```sh
m365 mcp serve
```

Stdio JSON-RPC for agents. Do not pass `--human`. Login stays `m365 auth login` in a terminal.

Three tools (`m365_status`, `m365_help`, `m365_run`) and six recipe prompts: `mail-search`, `teams-find`, `calendar`, `files`, `mail-write`, `teams-write`. Writes through `m365_run` stay dry-run unless `write_opt_in` is true.

## Agent skills (copy)

This repo has no `m365 skill` CLI. Copy `skills/<topic>/SKILL.md` into an agent skill root:

```text
Cursor    .cursor/skills/<topic>/SKILL.md
Claude    .claude/skills/<topic>/SKILL.md
Codex     .codex/skills/<topic>/SKILL.md
Grok      .grok/skills/<topic>/SKILL.md
OpenCode  .opencode/skills/<topic>/SKILL.md
```

Topics: `mail-search`, `teams-find`, `calendar`, `files`, `mail-write`, `teams-write`. MCP-only agents already get the same text from `m365_help` / `prompts/get`.

## Develop

```sh
sh scripts/install-tools.sh
make verify
```

CI runs `make verify` on `macos-latest` for pushes and PRs to `main`.

## License

MIT
