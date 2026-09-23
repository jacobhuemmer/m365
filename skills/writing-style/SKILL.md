---
name: writing-style
description: >
  Write Outlook email and Teams chat in a plain, direct, human voice.
  Use whenever drafting, replying, or sending mail or Teams with m365,
  including a message that will go out under the user's name, or a
  status note that will be mailed or pasted into chat.
---

# Writing voice

Load this before any draft that goes out under the user's name.
Structure and CLI flags stay in `mail-write` and `teams-write`. This file
is the voice.

Ground every sentence in the thread you already read. If a fact is not
there, say so or ask. Do not invent owners, dates, or ticket keys.

## Register

Plain, first person, present tense. The point is in the first sentence.
One idea per sentence. Short words. Contractions (`I'll`, `I'd`, `don't`).

Write like you are talking to a coworker who already has context:
direct, specific, a little informal. It should read like a person, not
a status page or a vendor newsletter.

Length matches the thread. A two-line chat gets a two-line reply. A
request for a fact gets one or two sentences. A request for a decision
gets the decision, then one reason.

## Words to reach for

Prefer: `FYI`, `Just an idea`, `Let me`, `I'll`, `I assume`, `I'm thinking`,
`I'm not entirely sure`, `no worries`, `got it`, `sounds good`, `See
attachment.`

Ops updates name the thing, the number, and the next step: `The API
gateway dropped from about 5.8 million requests an hour to about 2.9
million. Now that I know it works, we may increase the interval to 60
seconds.`

Uncertainty is named, then a next step: `I'm not entirely sure how the
proxied records work, but I assume once we switch it off that will close
the connections. Let me check.`

Ticket keys sit in the sentence (`OPS-1042`, `CHG-117`), not in a
preamble. Times carry a timezone (`3:30 PM CT`, `15:00 UTC`).

## Email

1. Greeting on a new mail or a cold thread: `Hi Sam,` / `Hi team,` /
   `Hi all,` / `Hi there,` (unknown vendor). Skip it on a fast
   back-and-forth that already dropped greetings.
2. The answer or the ask in the first sentence.
3. One or two sentences of reason, a link, or a ticket. Bullets only
   when there are real items (bucket names, owner+action). An owner
   bullet looks like `- Priya: reply with approval of the attached form.`
4. Close with `Thanks,` and the signed-in user's first name on the next
   line (from `m365 auth status`; the examples show `<first name>`):
   ```
   Thanks,
   <first name>
   ```
   That is the close. Not `Best,` `Kind regards,` `Thanks so much,` or a
   signature block.

Keep it under ~80 words unless they asked for a write-up. A whole mail
can be `See attachment.` plus the close. A correction can be `Ignore my
last comment about the background image. I see Jordan's email down in
the thread.`

New subject: a noun phrase, ticket key first when there is one. Replies
keep the thread subject.

## Teams

No greeting and no `Thanks,` sign-off in an active chat. The first line is
the point; that is what shows up on a phone.

Fragments are fine: `Found the root cause. Doing a write up.` `Gimme 5.`
`Call you in 30 min.` `Wait maybe a day and then switch over.`

With close peers the chat can be colloquial. With a manager or a
customer-facing group, stay plain and specific, still short. Flag
something upward with `FYI,`.

One message. Line breaks are ok. Emoji only if that chat already uses
them. Long content belongs in mail or a wiki link, not in chat.

## What this voice is not

Corporate filler reads as someone else. Write the fact instead of
`please don't hesitate`, `hope this helps`, `I'd be happy to`, `feel
free to`, `I wanted to reach out`, `please find the details`, `to ensure
the stability and security of our systems`, `at your earliest
convenience`, `let me know if you have any questions`.

Do not stack even, 16-word sentences with an em dash and a hedge-offer
(`If the bucket does not show, reply and I will check`). Do not add
`#MadeWithAI`. Do not open a routine reply with `I hope this email finds
you well.`

## Examples

Email, ask:

```
Hi Priya,

Please approve the attached access form. It requests my access to the
staging cluster through the VPN. The vendor has this as OPS-1042.

- Priya: reply with approval of the attached form.

Thanks,
<first name>
```

Email, status:

```
Hi team,

Helm chart 1.4.2 and image version 2.8.0-rc3 have been pushed to
staging.

Thanks,
<first name>
```

Email, idea:

```
Hi Sam,

Let me follow up with the vendor and request another set of traces.

If we must enable DEBUG, my preference would be a feature flag on that
code path so we can lift the payloads without turning on global DEBUG.
I don't really want DEBUG unless we have to. It produces so many traces.

Thanks,
<first name>
```

Teams, ops:

```
Found the traffic was real. Devices were sending 5s HTTP requests asking
for new commands. Updating the config to 30s. Told Chris we will
include this change in our release.
```

Teams, peer:

```
Gimme 5
```

Not this (AI-smooth):

```
Hi Alex,

The firmware drop is live. There is no public URL; it is a private
object storage bucket. If the bucket does not show, reply and I will
check your group membership. Please send that list when you have it.

Thanks,
<first name>
```

Rewrite in this voice:

```
Hi Alex,

The firmware drop is live in the private bucket device-firmware
(us-east). No public URL. Sign in with your work account; the
engineering groups should already see it.

Send me the publisher names and emails when you have them for OPS-1051.

Thanks,
<first name>
```
