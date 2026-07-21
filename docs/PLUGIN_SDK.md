# Writing an Orven Plugin

A plugin is a folder containing a `plugin.yaml` and an executable
entrypoint. The engine runs your entrypoint as a subprocess, writes one
JSON object to your **stdin**, and expects one JSON object on your
**stdout**. That is the whole interface. Any language works.

```text
my-plugin/
  plugin.yaml      required — identity, entrypoint, intervals, config schema
  main.py          your entrypoint (any language)
  README.md        what it observes, what it needs
  fixtures/        sample source data for tests
  tests/           tests that run without the real external system
```

Copy `plugins/demo-activity/` as a starting point.

For Python plugins, declare `entrypoint: ["python", "main.py"]` and
stop there: if a system only ships `python3` (or only `python`), the
engine resolves between those two standard names automatically. That
is the only substitution it ever makes — every other entrypoint
command is executed exactly as declared.

## The contract (version 1)

**Input (stdin):**

```json
{
  "contract_version": 1,
  "plugin_id": "my-plugin",
  "now": "2026-07-21T07:30:00Z",
  "window_start": "2026-07-20T07:30:00Z",
  "config": { "url": "http://sonarr:8989", "max_items": 5 },
  "secrets": { "api_key": "..." },
  "fixture": "only set during tests — path to fixture data"
}
```

`window_start` is your last successful run (zero value on the first
run). Report what changed since then.

**Output (stdout):**

```json
{
  "contract_version": 1,
  "status": "ok",
  "summary": "3 movies finished downloading overnight.",
  "observations": [
    { "title": "3 movies finished downloading", "body": "…", "kind": "count", "scope": "event" },
    { "title": "1 import failed", "body": "…", "kind": "notice", "scope": "state" }
  ]
}
```

### Observation scope: did it happen, or is it true right now?

Every observation has a `scope`, and getting it right is what keeps
briefings calm:

- **`event`** (the default) — something that *happened*: a backup
  finished, a certificate was renewed, three movies downloaded. Events
  from every collection since the last briefing are accumulated, so
  nothing that happened between briefings is lost. Report only events
  that are new since `window_start`.
- **`state`** — a condition that is *currently true*: an import is
  failed, an episode is stuck in the queue, an update is available.
  Report the condition on every collection for as long as it remains
  true, and simply stop reporting it once it clears. The engine shows
  each condition once per briefing (from your most recent collection),
  and it will reappear in every later briefing until you stop
  observing it — you never need to remember what you reported before.

If you mark a state as an event, readers will see it duplicated once
per collection run. If you mark an event as a state, it may vanish
from the briefing before anyone reads it.

#### The deciding question

> **If the condition resolves before the next briefing, should the
> reader still be told it happened?**
>
> - Yes → `event`
> - No → `state`

This is the test that settles the ambiguous cases. A backup failure is
a condition that may clear on its own (the next backup succeeds) — but
the reader absolutely should still hear about it, so it's an event. A
stuck download that un-sticks by morning was never worth reading
about, so it's a state.

#### Examples across domains

| Observation                                   | Scope   | Why |
|-----------------------------------------------|---------|-----|
| 3 movies finished downloading                 | `event` | Happened; worth knowing even after they're watched. |
| 1 episode is stuck in the download queue      | `state` | If it clears before the briefing, nobody needed to know. |
| 2 requests are awaiting approval              | `state` | Only the current queue matters; approved ones vanish. |
| Last night's backup failed                    | `event` | Even if tonight's succeeds, the reader must hear a backup failed. |
| Backup completed, 412 GB                      | `event` | A fact about last night, true forever. |
| No successful backup in 5 days                | `state` | A condition; disappears the moment a backup succeeds. |
| Container update available                    | `state` | Once updated, it's not news. |
| Container restarted 4 times overnight         | `event` | Even if it's stable now, the reader should know it happened. |
| Container is unhealthy                        | `state` | Report while true; stop when healthy. |
| Disk is 87% full                              | `state` | A reading of now. |
| Disk usage grew 200 GB this week              | `event` | A change that occurred over the window. |
| SMART reported 3 new reallocated sectors      | `event` | The increase happened; silence later doesn't unhappen it. |
| Certificate renewed                           | `event` | A completed occurrence. |
| Certificate expires in 9 days                 | `state` | Resolves when renewed; then it's no longer worth saying. |

#### Conditions worth remembering: report both

Some conditions are serious enough that the reader should learn about
them even if they resolve quickly — a RAID array that degraded and
rebuilt overnight, for example. Model these as a pair: an `event` for
the transition ("Array became degraded at 2:14 AM") plus a `state` for
the ongoing condition ("Array is degraded, rebuild 40% complete"). The
event guarantees the occurrence is never lost; the state keeps the
condition in every briefing until it clears.

The right scope is a domain judgment only you can make — automated
validation can check your output's shape, but not this choice. When
unsure, apply the deciding question above.

`status` must be one of:

| status        | meaning                                      |
|---------------|----------------------------------------------|
| `ok`          | relevant activity was found                  |
| `nothing`     | checked successfully, nothing relevant       |
| `partial`     | only part of the information was available   |
| `unavailable` | the source system could not be reached       |
| `auth_failed` | credentials were rejected                    |
| `error`       | the check could not complete (put detail in `error`) |

Never report `nothing` when you actually failed — missing coverage must
never look like good news.

### Title house style

An observation's `title` is the headline of a short news brief, set in
the calm school of headline writing:

- **sentence case**, not Title Case — "1 episode is stuck in the
  queue", never "1 Episode Stuck in Queue";
- **concise and factual** — state what is, without drama;
- **no trailing period** (or exclamation mark — the voice is calm);
- **detail belongs in the body** — if a title runs long, move the
  specifics after it into `body`.

`orven validate` checks this style and warns on violations (see
`docs/VALIDATOR.md`), but it never rewrites your words.

### What `summary` is for

`summary` is one factual sentence describing **this single collection
run** — not a headline for the finished briefing. The engine composes
the briefing; when a section has observations, they are shown by
themselves and your summary is not displayed there. Your summary
appears in two places:

- the plugin's **run history** ("last run: 3 new items found");
- the **briefing, only when there are no observations to show**, where
  it explains the result — "No new activity on Sonarr", "Sonarr could
  not be reached", "The API key was rejected".

So write it as a calm statement about the check itself, and make it
most informative for the empty and failure cases — that's when a
reader will actually see it.

## Observing an HTTP API (the common case)

Most real plugins read a service's HTTP API — Sonarr, Radarr,
Overseerr, Traefik. The complete reference implementation is
**`examples/radarr-queue/`**; copy it and change what it observes.
The pattern it demonstrates:

- **Standard library HTTP only** (Python `urllib`, Go `net/http`, …).
  No package installs — a plugin folder must run as dropped in.
- **GET requests only.** A plugin never changes the system it observes.
- **Secrets go in headers, never URLs.** The API key arrives in the
  engine input's `secrets` and is sent as a header; query strings end
  up in proxy and server logs.
- **Map transport failures to statuses honestly:**

  | what happened                        | status        |
  |--------------------------------------|---------------|
  | connection refused / DNS / timeout   | `unavailable` |
  | HTTP 401 or 403                      | `auth_failed` |
  | other HTTP error, unparsable body    | `error`       |
  | reached it, queue/list is empty      | `nothing`     |

- **Fixtures replace the network.** When the engine passes `fixture`
  in the input (tests and `orven validate` do), read that file as the
  canned API response instead of calling anything. Your plugin is
  fully testable without the real service. **Never commit a real
  credential to a fixture** — invent obviously fake values, and don't
  give them header or query-parameter shapes (the validator warns on
  those).
- **Credentials must never appear in your output.** Not in summaries,
  observations, or error text — even when an upstream API echoes your
  key back in an error page. The engine redacts assigned secret values
  and credential-shaped fragments from everything you return before it
  is stored or shown, but that scrubber is a backstop against
  accidents, not permission: `orven validate` treats credential-shaped
  output as an error.
- **Declare the access** in `permissions:` so the user sees it before
  enabling.

### Choosing a data source

This is not a choice between "HTTP plugins" and something else — every
plugin is a subprocess the engine runs on its schedule. The choice is
what your subprocess reads:

- **A service's HTTP API** — prefer this when the service has one
  (most self-hosted apps do): clean authentication, no filesystem
  coupling, works across hosts and containers.
- **Local files or directories** (log files, exported reports) — needs
  a declared path permission, and Docker users must mount it.
- **Local sockets or read-only CLIs** (the Docker socket, `smartctl`)
  — the most privileged option; declare it plainly and read only.

Whatever the source: no daemons, no servers, no background threads
that outlive the run. The engine starts you, you observe, you print
one JSON object, you exit.

## Rules

1. **Facts only.** Observations state what is, never what to do.
   "The backup failed at 3:12 AM" — yes. "You should re-run the
   backup" — never.
2. **No scheduling.** The engine decides when you run. No timers, cron
   jobs, daemons, or background threads that outlive your run.
3. **No state.** Don't write files outside your own folder. If you
   genuinely need persistent state, request plugin storage (future
   contract addition) rather than inventing your own.
4. **Secrets stay out of output.** Never echo config or secrets into
   summaries, observations, or error text.
5. **Be quick and bounded.** Declare a realistic `timeout`; the engine
   kills you at 5 minutes regardless.
6. **Declare everything you touch** in `permissions:` — hosts, files,
   sockets. Users see this list before enabling you.
7. **Backwards compatibility.** Written against contract v1? You must
   keep working on every future engine. The engine guarantees it will
   never remove or repurpose contract fields.

## plugin.yaml

See `plugins/demo-activity/plugin.yaml` for a complete annotated
example.

Declare `collection.freshness` honestly: it is how long your results
stay trustworthy. If a briefing has to fall back on data older than
this window (because your source was unreachable, for instance), the
briefing tells the reader that the information may be out of date —
so a download queue might declare minutes-to-hours, while certificate
expiry could declare a day or more. Config field types the settings form can render: `text`,
`number`, `boolean`, `url`, `duration`, `select`, `secret`. `secret`
values are stored separately, shown as "configured", and never
displayed again.

## Testing

Ship fixtures and tests that run without the real system: the engine
passes `fixture` in test input so your plugin can read canned source
data instead of calling the network. Cover at minimum: normal activity,
nothing new, source unreachable, bad credentials, malformed response.
