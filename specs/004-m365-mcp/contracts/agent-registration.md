# Agent registration

Dedicated MCP server (like `kata mcp serve`). MUST NOT be a lazy-mcp backend. MUST NOT use Microsoft Enterprise MCP or `npx`.

Command is the stowed CLI on PATH. Args are exactly `mcp` then `serve`. Same Keychain session as the terminal. No Client ID, Tenant ID, or tokens in this file or in the MCP config.

```json
{
  "mcpServers": {
    "m365": {
      "command": "m365",
      "args": ["mcp", "serve"]
    }
  }
}
```

Harnesses (Grok, Claude, Cursor, Codex, OpenCode) register this as its own server. If `m365_status` reports `signed_in` false or `session_usable` false, tell the operator to run `m365 auth login` in a terminal. Agents MUST NOT start a browser from MCP.
