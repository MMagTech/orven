# Contributing to Orven

Thanks for wanting to help. Orven is small on purpose, and its rules
are written down — most review friction disappears if you read two
documents first:

1. **[docs/CONSTRAINTS.md](docs/CONSTRAINTS.md)** — the product
   constitution. If a change conflicts with it, the change is wrong,
   not the file. The rules most often at stake: Orven is strictly
   read-only, output is facts and never advice, and the briefing is
   calm — never a dashboard.
2. **[docs/PLUGIN_SDK.md](docs/PLUGIN_SDK.md)** — if you're writing a
   plugin.

## Where contributions go

- **Plugins** → the [catalog repository](https://github.com/MMagTech/orven-plugins).
  New observers for real systems are the most valuable contribution
  there is. The catalog gate is `orven validate` at zero errors and
  zero warnings, plus tests that run without the real external system.
- **Platform changes** (engine, UI, contract, validator, docs) → this
  repository, as a pull request.

## Working on the platform

```bash
go run ./cmd/orven          # run the app on :8420
go test ./...               # engine, contract, and validator suites
go run ./cmd/orven validate plugins/demo-activity examples/jobs-example
```

Ground rules:

- **Tests are part of the change.** The Go suites encode product
  semantics (briefing states, scope lifecycle, quiet rules, the
  credential boundary), not just code correctness. A behavior change
  without a test pinning it is incomplete.
- **The contract is additive-only, forever.** A plugin written against
  contract v1 must work on every future engine. Never remove or
  repurpose a contract field.
- **Respect the architecture seams.** `internal/core` (shell) and
  `internal/engine` (scheduling, execution, compilation) talk only
  through the `Engine` type; `contract/` imports nothing.
- **Keep dependencies near zero.** The app depends on `yaml.v3` alone;
  plugins use stdlib only. New runtime dependencies need a strong
  case.
- **Match the voice.** User-visible text is calm, factual, and plain.
  When in doubt, read a briefing first.

## AI-assisted contributions

Welcome — this repo is built for them. Point your assistant at the
repository root ([CLAUDE.md](CLAUDE.md) is its entry point) and review
what it produces before submitting; you are the author of record.

## Pull requests

Small and focused beats large and sweeping. CI must pass (tests,
validator at zero findings on the reference plugins, Docker smoke
test). Describe *what changed and why* — reviewers here care about the
product reasoning as much as the diff.

By contributing you agree your contribution is licensed under
[Apache 2.0](LICENSE).
