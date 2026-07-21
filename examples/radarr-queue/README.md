# Radarr Queue

The reference **HTTP-source plugin** for Orven. It watches the Radarr
download queue and reports, as facts:

- how many downloads are stuck (a `state` — re-reported until it clears);
- how many are transferring normally (a `state`);
- an empty queue as `nothing`.

It is strictly read-only: only GET requests, and it never suggests
what to do about anything it sees.

## What it needs

- Radarr URL (settings form)
- API key (stored write-only as a secret; sent as the `X-Api-Key`
  header, never in the URL)

## Test

```bash
python -m unittest discover -s tests
```

Tests and `orven validate` run entirely from `fixtures/` — no Radarr
required. Copy this plugin as the starting point for any plugin that
observes a service's HTTP API; the pattern is documented in
`docs/PLUGIN_SDK.md`.
