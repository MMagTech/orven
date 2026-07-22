# Changelog

All notable changes to Orven are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions
follow [Semantic Versioning](https://semver.org/). The plugin
contract's compatibility promise is stronger than semver: contract v1
plugins work on every future engine, period.

## [Unreleased]

### Changed

- The in-tree HTTP reference plugin is now `examples/jobs-example`, a
  deliberately generic observer of a made-up jobs API (it also
  demonstrates window-filtered events). The Radarr plugin it replaces
  lives where real plugins belong: the
  [plugin catalog](https://github.com/MMagTech/orven-plugins).

## [0.1.0] — 2026-07-21

The first release: a complete, small, working briefing platform.

### Added

- The briefing pipeline: plugins collect observations on their own
  schedules; the engine compiles one briefing at the time you choose.
- Three honest briefing states: *all quiet* (only when every source
  was checked and nothing changed), *changes* (facts without
  importance-ranking), and *unable to verify all sources* (partial
  coverage leads with reduced confidence). A Coverage section defines
  each briefing's scope.
- Plugin contract v1 (frozen): subprocess plugins in any language,
  JSON over stdin/stdout; observations with event/state scope; engine-
  enforced timeouts, overlap prevention, interval clamping, and a
  concurrency ceiling.
- Credential boundary: per-plugin write-only secrets delivered via
  stdin, and a sanitizer that redacts assigned secret values and
  credential-shaped content from everything a plugin returns.
- Newspaper-style UI with light/dark/system appearance (palette-only
  dark mode), history, print/PDF export, schema-generated plugin
  settings, freshness/staleness wording, and a log viewer.
- `orven validate`: the plugin validator and CI gate (contract
  conformance, forbidden advisory language, credential shapes, title
  house style).
- Reference plugins: `demo-activity` (fixture-driven) and
  `radarr-queue` (HTTP-source pattern).
- Docker image with health check; deployment docs for compose and
  Unraid.

### Known limitations

- No authentication yet — run on a trusted network or behind an
  authenticating reverse proxy (the app warns about this).
- Plugin installation is manual (folder drop); in-app catalog install
  is planned.
- Plugin network egress and filesystem isolation are documented as
  not yet enforced (`docs/CONSTRAINTS.md`, "Known enforcement gaps").
