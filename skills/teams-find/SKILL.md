---
name: teams-find
description: Find a Teams 1:1 or group chat by person name or topic before sending.
---

# teams-find

Find a chat before notify.

  1:1: teams find Ajay
  Group: teams find --group NOC
  Notify: teams send --to Ajay --text ping --dry-run

Several matches MUST NOT send. Zero matches: do not guess. Dry-run before notify (write_opt_in false).

Do not invent a Graph people-search tool.
