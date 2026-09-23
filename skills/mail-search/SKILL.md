---
name: mail-search
description: Search Outlook mail by from, to, subject, or keywords, then get or thread the hit.
---

# mail-search

Search mail, then open a message or thread. No session required to read this recipe.

  mail list --folder all --search 'from:ajay'
  m365_run namespace=mail verb=list flags folder=all search=from:ajay

Also: search subject:… ; default --top applies. Then mail get and mail thread with the message id.

Do not dump the mailbox. Do not print secrets.
