# Orven

**A read-only daily briefing for your self-hosted world.**

Orven exists to answer one question, on one calm page:

> *What changed while I wasn't looking?*

Plugins observe the systems you already run — media servers, backups,
containers, certificates — and the engine compiles their observations
into a morning briefing that reads like a newspaper front page: facts,
in sentences, finishable before your coffee is.

## What Orven is

- **A briefing, not a feed.** One page, prepared on the schedule you
  choose, from observations collected on each plugin's own schedule.
  If nothing happened, it says *"All quiet"* — and means it, because
  it only makes that claim when every source was checked.
- **Strictly read-only.** Orven never changes anything on your
  systems, and it never tells you what to do. *"The backup failed at
  3:12 AM"* — never *"you should re-run the backup."* What matters is
  your judgment, not the software's.
- **Honest about what it doesn't know.** An unreachable source is
  reported as unreachable, stale information says it may be out of
  date, and missing coverage is never presented as good news. When
  coverage is incomplete, the briefing says so before anything else.
- **Small and self-hosted.** One binary or container, plain-file
  storage, no external services. A backup is a copy of one folder.

## What Orven is not

- **Not a monitoring dashboard.** No gauges, graphs, or live tiles —
  Grafana and Uptime Kuma do that well already.
- **Not an alerting system.** Orven will never page you at 3 AM. It
  assumes the news can wait until you sit down with it.
- **Not an automation platform.** There are no fix-it buttons and
  there never will be.
- **Not AI-powered.** Briefings are compiled deterministically from
  structured observations. (AI assistants are great at *writing
  plugins* for Orven — see below — but the product itself doesn't
  need or use one.)

## Maturity

Early and honest about it: **v0.1.x**. What works today: the full
briefing pipeline (collect → observe → compile → read → history →
print/PDF), the plugin contract (v1, frozen, with a backwards-
compatibility promise), the validator, a demo plugin and an HTTP
reference plugin, Docker deployment. What doesn't exist yet: sign-in
(keep it on a trusted network or behind an authenticating reverse
proxy — the app warns about this), in-app plugin installation
(installing is a folder-drop today), and update checks. See
[CHANGELOG.md](CHANGELOG.md).

## Quick start

```bash
docker run -d -p 8420:8420 -v ./orven-data:/app/data ghcr.io/mmagtech/orven:latest
```

Open <http://localhost:8420>, enable the bundled **Demo Activity**
plugin under *Plugins*, press **Run now**, then **Prepare the first
briefing**. You'll be reading a briefing inside two minutes.

Compose, Unraid, and persistent-storage details:
[docs/DEPLOY.md](docs/DEPLOY.md). From source: `go run ./cmd/orven`
(Go 1.26+; Python 3 for the demo plugin).

## How the plugin model works

A plugin is one folder: a `plugin.yaml` manifest and an entrypoint in
any language. The engine runs it as a subprocess on a schedule —
config in on stdin, one JSON object of **observations** out on stdout.
Observations are facts with a scope: **events** (it happened — a
backup completed) accumulate into the next briefing; **states** (it's
true right now — a download is stuck) appear once per briefing and
return each morning until they clear.

Everything else is the engine's job: scheduling, timeouts, credential
handling, storage, and how the briefing reads. Plugins never schedule
themselves, never persist their own data, and never render UI.

```bash
go run ./cmd/orven validate ./path/to/plugin   # the same gate CI uses
```

### Where plugins live

- **The catalog —
  [orven-plugins](https://github.com/MMagTech/orven-plugins)** — is
  where real, installable plugins are published (curated and
  community). Browsing and installing from the catalog inside the app
  is planned; today, installing means copying a plugin folder into
  your plugins directory and restarting.
- **Your plugins directory** (`/app/plugins` in Docker) holds what you
  have installed. The bundled **Demo Activity** plugin is pre-installed
  sample content: it observes nothing real, and exists so your first
  run has a briefing to show.
- **`examples/` in this repository** is teaching material, never
  installed: `examples/jobs-example` is the generic HTTP-source
  reference the SDK walks through. Real integrations belong in the
  catalog, not here.

To write your own plugin, start from
[docs/PLUGIN_SDK.md](docs/PLUGIN_SDK.md) and copy a reference — or
point an AI coding assistant at this repository (it will find
[CLAUDE.md](CLAUDE.md)) and describe what you want observed in plain
language.

## Trust and boundaries

The read-only covenant is enforced, not promised: plugins state facts
through a narrow contract, the validator rejects advisory language,
and the engine owns every side effect. Each plugin declares the access
it needs, and you see that before enabling it. Credentials are scoped
per plugin, stored write-only, and scrubbed from everything a plugin
returns before it can reach a briefing, a log, or a page.

Equally important is what we don't claim: **valid is not trusted.**
Installing a plugin is a trust decision — curated catalog plugins are
reviewed; third-party plugins mean *you* are the reviewer (they're
deliberately small enough to read). The current enforcement limits are
documented plainly in
[docs/CONSTRAINTS.md](docs/CONSTRAINTS.md#known-enforcement-gaps-recorded-planned)
rather than papered over.

## Documentation

| Document | What it covers |
|---|---|
| [docs/DEPLOY.md](docs/DEPLOY.md) | Docker, compose, Unraid, persistent storage |
| [docs/CONSTRAINTS.md](docs/CONSTRAINTS.md) | The product constitution: architecture boundaries, language rules, security model, plugin identity |
| [docs/PLUGIN_SDK.md](docs/PLUGIN_SDK.md) | Writing plugins: the contract, scopes, HTTP pattern, house style |
| [docs/VALIDATOR.md](docs/VALIDATOR.md) | Every check `orven validate` performs, and its hard boundaries |
| [CLAUDE.md](CLAUDE.md) | Entry point for AI coding assistants working in this repo |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). The short version: read the
constraints first (they win every argument), keep changes small and
tested, and send plugins to the
[catalog repository](https://github.com/MMagTech/orven-plugins).
Issues and ideas are welcome — especially reports of any place where
Orven's output is less than calm, factual, and honest.

## Project direction and license

Orven Community — this repository — is the product, licensed under
[Apache 2.0](LICENSE). It is and will remain complete and first-class:
never a trial edition, never intentionally diminished. If commercial
offerings appear later, they will expand the platform (enterprise
capabilities, managed services, professional tooling) around a
Community Edition that stands on its own.
