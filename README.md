# Orven

A read-only daily briefing for your self-hosted world. Orven sits
quietly, lets plugins collect facts from the systems you care about,
and prepares a calm front page you can read with your coffee — like a
morning newspaper, not a monitoring dashboard.

Orven observes. **It never changes anything, and it never tells you
what to do.**

## Quick start (Docker)

```bash
docker compose up --build
```

Open <http://localhost:8420>. Enable the bundled **Demo Activity**
plugin under *Plugins*, press **Run now**, then **Prepare the first
briefing** on the front page.

## Quick start (from source)

Requires Go 1.26+ (and Python 3 for the demo plugin).

```bash
go run ./cmd/orven
```

## How it fits together

- **Core** (`internal/core`) — the shell: front page, history,
  settings, print/PDF export. Knows nothing about running plugins.
- **Engine** (`internal/engine`) — owns all scheduling, runs plugins as
  isolated subprocesses with timeouts, stores their observations, and
  compiles briefings.
- **Contract** (`contract/`) — the versioned boundary plugins are
  written against. Plugins built for contract v1 work on every future
  engine, guaranteed.
- **Plugins** (`plugins/`) — one folder each. Any language; JSON on
  stdin/stdout. See [docs/PLUGIN_SDK.md](docs/PLUGIN_SDK.md).

Collection and briefing are separate: plugins gather observations on
their own intervals; one briefing is assembled at the time you choose
in *Settings*.

## Building a plugin (with or without AI)

Point your AI assistant at this repository and say what you want
observed — [docs/PLUGIN_SDK.md](docs/PLUGIN_SDK.md) and
[docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) contain everything it needs
to produce a conforming plugin. `plugins/demo-activity/` is the
reference implementation.

## Configuration

| env variable    | default   | meaning                  |
|-----------------|-----------|--------------------------|
| `ORVEN_ADDR`    | `:8420`   | listen address           |
| `ORVEN_DATA`    | `data`    | data directory (back this up) |
| `ORVEN_PLUGINS` | `plugins` | installed plugins        |

All application state lives in the data directory as plain files — a
backup is a copy of that folder. The `secrets/` subfolder holds plugin
credentials; encrypt backups that include it.

## Tests

```bash
go test ./...
cd plugins/demo-activity && python -m unittest discover -s tests
```

## Status

Phase One vertical slice: schedule a briefing, configure and run a
plugin from the UI, collect observations, generate and read briefings,
browse history, print/export. Not yet built: sign-in, installing
plugins from repositories in-app, update checks, scheduled backups.
