# Changelog

All notable changes to Orven are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions
follow [Semantic Versioning](https://semver.org/). The plugin
contract's compatibility promise is stronger than semver: contract v1
plugins work on every future engine, period.

## [0.3.0] — 2026-07-23

The Brief Experience: the reading contract, kept on every page.

### Added

- **The edition projection.** Sections with observations — or that
  could not be checked, or whose information is stale or partial —
  are the Brief's stories; failure-to-observe is itself news. Sources
  that checked successfully and observed nothing collapse into one
  line ("Also checked: … No new observations."), so the page's length
  follows the news, never the number of installed plugins. A story
  leads with at most eight items; the rest wait behind a quiet
  expansion on the reading page only — the stored Brief, print, and
  exports remain complete.
- **The dateline** states the window it can prove: "covers activity
  since the previous Brief," or "the first Brief." Briefs stored
  before the field existed stay silent rather than claim a window.
- **The first-run experience.** A fresh installation seed-enables the
  bundled demonstration plugin (zero permissions, fixture-only — the
  constitution's sole exception) and collects once immediately; the
  Today empty state becomes a welcome card with one action, "Prepare
  your first Brief." Wherever the demo contributes, Coverage states
  that its events are fiction. Onboarding is permanent and
  backup-aware; it ends when a real plugin is enabled or the demo is
  disabled or uninstalled, and never returns. A restored demo comes
  back installed but disabled.
- **The constitution grew three sections**: the ecosystem rules
  (§22–26), the reading contract (§27), and seeded content (§28).

### Changed

- `/print` before any briefing exists returns the reader to Today
  instead of a bare 404.
- Destructive and occasional actions on secondary pages — delete a
  backup, uninstall a plugin, restore the demo — now speak in the
  quiet service-line voice instead of competing with content.
- Plugin SDK documentation: which fixture the validator executes,
  config-default injection, the `publisher` field, the multi-endpoint
  fixture pattern, timezone-aware timestamp parsing, and a PowerShell
  stdin note.

### Fixed

- The engine now sends `now` and `window_start` in canonical UTC. On
  hosts in a non-UTC timezone it previously marshaled local offsets
  while every SDK example showed UTC — an offset-naive plugin parser
  would get window comparisons silently wrong by hours.

## [0.2.0] — 2026-07-22

Plugin management and Daily Care: plugins now install from catalogs
after an explicit trust decision, and Orven backs up, restores, and
exports what it holds.

### Added

- Daily Care. **Backups**: download on demand, automatic daily backups
  with retention, a backup browser, and restore. Credentials are only
  ever included encrypted (AES-256-GCM, passphrase-derived key) —
  never plain, in any archive. **Restore means "put me back exactly
  where I was"**: the backed-up domains are reproduced strictly
  (anything created since the backup is preserved only in the
  automatic pre-restore safety backup); credentials are reproduced
  only when the archive carries them. The restore flow answers the
  questions that matter — a plain-English confirmation, a completion
  page stating what was restored and what still needs attention, an
  "Awaiting reinstall" list for plugins recorded in the backup, and a
  visible warning when automatic backups are set to include
  credentials but the (never-backed-up) passphrase is missing.
  **Collection schedule** overview in Settings: every plugin's
  interval, last run, and next expected run in one place. **Print
  preview** (`/print`): the briefing exactly as it prints, with a
  running header and footer and paper typography; PDF remains the
  browser's reliable print-to-PDF. **Exports**: any briefing downloads
  as Markdown or JSON. The Brief page's actions became a quiet
  service line — Prepare Brief · Print · Download — because the page
  is a publication first.

- Plugin management, the complete three-tier experience. Settings
  manages trusted repositories (default catalog vs third-party,
  labeled). **Plugins → Discover** browses what those repositories
  publish — name, publisher, version, curated/community standing, and
  requested permissions — and installs after an explicit trust-decision
  page; every install passes the validator regardless of source, and
  provenance (catalog, publisher, version) is recorded. **Plugins →
  Installed** gains uninstall: configuration, credentials, and raw
  observations are removed; run history is preserved unless
  deliberately deleted; historical briefings are never altered.
  Folders added manually are never deleted without an explicit
  acknowledgment.
- Seed-once demo lifecycle: a fresh installation is seeded with the
  demo plugin exactly once. Uninstalling it is permanent across
  restarts and container updates, and Settings offers a deliberate
  "Restore the demo plugin".
- `orven index` subcommand: generates a catalog repository's
  index.json (the file Discover reads).

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
