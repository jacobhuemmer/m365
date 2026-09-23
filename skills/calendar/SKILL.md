---
name: calendar
description: List own calendar, find free slots, and dry-run create events with English times like tomorrow at 1:30 pm.
---

# calendar

  calendar list
  calendar free --when tomorrow
  calendar create --when 'tomorrow at 1:30 pm' --subject '…' --dry-run

MCP writes dry-run unless write_opt_in is true.
