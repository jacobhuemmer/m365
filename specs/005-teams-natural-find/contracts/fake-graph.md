# Fake Graph (teams find)

Synthetic only. No live member lists.

Existing seed chats (`chat-1`, `chat-2` NOC) remain so 001 tests stay valid.

**Add** (find tokens / ids):

| Chat id | Type | Topic / other member | Notes |
| --- | --- | --- | --- |
| `chat-ajay` | oneOnOne | Ajay Kumar `<ajay@example.com>` | **Not** among the first 20 list items |
| `chat-ajay-b` | oneOnOne | Ajay Singh `<ajay.singh@example.com>` | Second Ajay for several-match |
| `chat-group-ajay` | group | members include Ajay Kumar | Must not appear in default person find Ajay |
| `chat-noc-dev` | group | topic `NOC-Dev` | `--group NOC` |
| fillers | oneOnOne | synthetic-1 … | Pad list so `chat-ajay` is off page one (`teams list` default 20) |

Find without Teams consent still 403 → exit 4 (`fake-mail` / no teams). `fake-both` / `fake-all` include these chats.

List paging: `$top` honored; `chat-ajay` only on a later page.
