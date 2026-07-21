# CLAUDE.md — Contributor Guide

Orven is a **read-only daily briefing platform** for self-hosted
infrastructure. Plugins collect facts; the engine compiles them into a
calm morning briefing that reads like a newspaper page. Orven observes.
It never changes anything, and it never tells the reader what to do.

## Before any change

Read **`docs/CONSTRAINTS.md`**. It is the product's constitution: if a
change conflicts with it, the change is wrong, not the file. The three
rules most often at stake:

1. **Facts only, everywhere.** No recommendations, remediation steps,
   or advice in any briefing, observation, or summary — "The backup
   failed at 3:12 AM," never "You should re-run the backup."
2. **The briefing is the product.** Calm, concise, readable in
   minutes. Never a dashboard, never an alert stream. Missing coverage
   is stated as missing — absence of data is never an all-clear.
3. **Importance is the reader's judgment.** Orven reports what
   changed; it never ranks or classifies what "needs attention."

## Architecture — do not blur these lines

| Where            | What                                                        | May import        |
|------------------|-------------------------------------------------------------|-------------------|
| `contract/`      | The versioned plugin contract (v1). **Additive-only, forever.** | (nothing)     |
| `internal/engine`| Scheduling, plugin subprocess execution, observations, brief compilation, storage | `contract` |
| `internal/core`  | Web UI, settings, history, export — the shell               | `engine`, `contract` |
| `internal/validate` | `orven validate` (spec: `docs/VALIDATOR.md`)             | `engine`, `contract` |
| `plugins/`       | Installed plugins (one folder each)                         | —                 |
| `examples/`      | Reference plugins that are documentation, not installed     | —                 |

- A plugin written against contract version N must work on **every**
  engine ≥ N. Contract fields are never removed or repurposed.
- The engine owns all scheduling. Plugins are stateless subprocesses:
  JSON in on stdin, one JSON object out on stdout, then exit. No
  daemons, timers, servers, or plugin-owned UI — ever.
- Dark mode may change **colors only** (`CONSTRAINTS.md` §11); the
  palette is `light-dark()` pairs in `internal/core/static/style.css`.
- Secrets are write-only after entry: never in briefs, logs, output,
  errors, or exports.

## Building a plugin

1. Copy a reference: `plugins/demo-activity/` (fixture-driven basics)
   or `examples/radarr-queue/` (observing an HTTP API — the common
   case).
2. Read `docs/PLUGIN_SDK.md`. The two decisions that matter most:
   - **`event` vs `state` scope** — "If the condition resolves before
     the next briefing, should the reader still be told it happened?"
     Yes → event. No → state.
   - **Title house style** — sentence case, concise, factual, no
     trailing period; detail goes in the body.
3. Ship the full folder: `plugin.yaml`, entrypoint, `README.md`,
   `fixtures/`, `tests/`. Tests must run without the real external
   system (the engine passes `fixture` in test input).
4. Validate and test:

```bash
go run ./cmd/orven validate ./path/to/plugin
```

```bash
python -m unittest discover -s tests
```

A plugin is done when `orven validate` reports zero errors and zero
warnings and its tests pass.

## Developing the app

```bash
go run ./cmd/orven        # serves on :8420 (ORVEN_DATA, ORVEN_PLUGINS, ORVEN_ADDR)
```

```bash
go test ./...             # engine + validator suites
```

Docker: `docker compose up --build`. Runtime state lives in `data/`
(gitignored); deleting it resets the app.

When changing the engine, run the full Go suite — the tests encode the
product's semantic guarantees (briefing states, scope lifecycle, quiet
rules, concurrency ceiling), not just code correctness. When changing
templates or CSS, verify in the browser; the three briefing states are
easiest to produce with the demo plugin's scenario setting.

## Never

- Remediation or advisory language in anything a reader sees.
- Plugin-owned schedulers, persistence, network servers, or UI.
- Removing or repurposing a contract field.
- A dark-mode style that changes more than color.
- New runtime dependencies without strong justification — the Go app
  currently depends on `yaml.v3` alone, and plugins use stdlib only.
