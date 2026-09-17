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

```
m365 <namespace> <verb> [flags]
```

| Namespace | Verbs |
| --- | --- |
| `auth` | `login`, `status`, `logout` |
| `mail` | `list`, `get`, `send`, `reply`, `save-attachment` |
| `teams` (`chat`) | `list`, `find`, `get`, `send` |
| `calendar` | `calendars`, `list`, `get`, `create`, `update`, `delete`, `free` |
| `files` | `root`, `list`, `get`, `download`, `upload`, `create-folder`, `delete`, `move` |
| `mcp` | `serve` |

Exit classes: `0` success, `3` usage/config, `4` auth, `5` service, `6` not-found.

## MCP

```sh
m365 mcp serve
```

Stdio JSON-RPC for agents. Do not pass `--human`. Login stays `m365 auth login` in a terminal.

## Develop

```sh
sh scripts/install-tools.sh
make verify
```

CI runs `make verify` on `macos-latest` for pushes and PRs to `main`.

## License

MIT
