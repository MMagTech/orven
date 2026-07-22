# Jobs Example

The reference **HTTP-source plugin** for Orven. It observes a made-up
jobs API — deliberately generic, because any service with an HTTP
endpoint and an API token is observed the same way. It reports, as
facts:

- each job completed since the last run (`event`s, filtered by
  `window_start` so nothing is double-reported);
- jobs in a failed state (a `state` — re-reported until it clears);
- how many jobs are running (a `state`);
- an empty queue as `nothing`.

It is strictly read-only: GET requests only, and it never suggests
what to do about anything it sees.

## Test

```bash
python -m unittest discover -s tests
```

Tests and `orven validate` run entirely from `fixtures/` — no real
service required. Copy this folder as the starting point for any
plugin that observes an HTTP API; the pattern is documented in
`docs/PLUGIN_SDK.md`, and real-world plugins built on it live in the
[plugin catalog](https://github.com/MMagTech/orven-plugins).
