# Roadmap

How Orven decides what happens when. `docs/CONSTRAINTS.md` says what
Orven *is*; this file says what is being worked on. One milestone at a
time. Ideas that arrive while a milestone is under review are captured
here — in Backlog or Ideas — rather than expanding work that is
already being reviewed.

## Now — Deployment & SDK validation

Prove the shipped product against reality:

- cut and publish v0.2.0, so the image being deployed is the product
  being validated;
- deploy Orven on a real server (Unraid) with persistent storage, as
  a daily driver, treating every gap between the deployment docs and
  reality as a finding;
- then the ecosystem acid test: a fresh AI context builds a useful
  plugin using only the public repositories, SDK documentation, and
  validator — installed into the running deployment and verified in a
  real morning briefing.

## Next

Chosen deliberately at milestone close, never automatically.

## Backlog

Captured and roughly ordered. Scheduled deliberately, never
automatically.

- **First-run experience** — the Today empty state becomes the
  onboarding surface (no wizard): a welcome card in the paper's own
  voice with one action, "Prepare your first Brief." On a genuinely
  fresh installation the demo is seed-enabled and collects once
  immediately, so the first Prepare has material to compile. Only a
  bundled, first-party, zero-permission, fixture-only plugin may
  ever be seed-enabled — document that narrow exception in
  `docs/CONSTRAINTS.md` when implementation begins; catalog plugins,
  curated included, never auto-enable. The demo is never
  automatically re-enabled after the user disables or uninstalls it.
  Coverage discloses, inside the briefing, that Demo Activity
  produces fictional demonstration events. While onboarding is
  active, a quiet shell line after the first Brief: "This Brief is a
  demonstration. Tomorrow's prepares itself at 07:00. Install a
  plugin from Discover when you are ready to prepare Briefs from
  your own systems." Onboarding state is permanent and backup-aware
  (a restored installation is never a first run); onboarding ends
  when a real plugin is installed and enabled, or when the demo is
  disabled or uninstalled. Sub-decision for implementation:
  "Restore the demo plugin" returns it installed but disabled
  (current preference — restoring code should not silently restore
  permission to run).
- **Restore, finished — informed-consent reinstall** — availability
  probe per recorded plugin (available / removed / unreachable /
  manually added, with reasons); backed-up version, currently
  available version, and current permissions shown before anything is
  fetched; per-plugin opt-out; bulk reinstall through the existing
  validated install path; per-plugin failure reporting that never
  stops the rest; foundations: install records gain the granted
  permissions, backup manifests gain application and storage-format
  versions. Orven never silently installs newer plugin versions or
  accepts changed permissions on the user's behalf.
- **Backups list polish** — present backup kind in user terms
  (safety / automatic / manual) instead of filenames; Restore primary
  and Delete quiet; rows lead with kind and date. Needs an additive
  `kind` field in the backup manifest (manual and automatic backups
  are indistinguishable by filename).
- **Briefing section ordering** — currently plugin-folder
  alphabetical; an editorial order would read better.
- **Engine-owned per-section item cap** — very long sections fold
  into "N more items," so calm doesn't depend on every plugin
  author's restraint.
- **Quiet-action pass on secondary pages** — apply the service-line
  demotion where actions compete with content outside the Brief page.
- **Print empty-state handling** — direct access to `/print` before
  any briefing exists returns a bare Go 404; provide an Orven-native
  empty state or redirect.
- **Catalog release versioning and checksums** (trust roadmap tier
  2) — installs verify code matches the published release; recorded
  versions become reinstallable.
- **OS-level plugin isolation** (trust roadmap tier 3) — separate
  plugin user, filesystem confinement, network egress limits where
  technically practical; closes the documented enforcement gaps in
  `docs/CONSTRAINTS.md`.
- **Authentication and read-only roles** — required before
  internet-exposed installations are reasonable.
- **Update checks with permission-diff pause** — updates never
  silently expand permissions (constitution §16); builds on the
  granted-permissions records from "Restore, finished."
- **Acid test as a standing gate** — any milestone that changes the
  contract surface, SDK documentation, or validator closes by
  re-running the fresh-context acid test (a fresh AI context builds
  a useful plugin from the public repositories alone), keeping
  constitution §23 and §26 measured rather than assumed.

## Ideas

Captured, not commitments.

- **The delivery concept** — exports, email, notifications, Discord,
  webhooks may be one product concept ("the briefing reaching you")
  rather than five features. Awaiting a design discussion held
  against the finished Daily Care implementation.
- **Routine success vs quiet mornings** — nightly "backup completed"
  events make "All quiet" rare; current stance is a plugin-level
  "report successful runs" setting pattern. Revisit with ecosystem
  evidence.
- **Missed-briefing catch-up** — if the host slept through briefing
  time, should one be prepared on wake?
- **AI-assisted briefing enhancement** (original vision) — would
  consume the structured observation model, never replace it, and
  stay traceable to observations.

## Done

- 2026-07-21 — **Phase One through contract freeze**: briefing
  pipeline, three briefing states, credential publication boundary,
  validator, plugin contract v1 frozen — `1b1a848`…`7efec41`
- 2026-07-21 — **GitHub publication**: both repositories public, CI,
  branch protection, v0.1.0, multi-arch image on GHCR — `09a62d7`,
  tag `v0.1.0`
- 2026-07-22 — **Generic HTTP example** replaces the product-named
  example in core — `3db1ce3`
- 2026-07-22 — **Plugin management**: repositories in Settings,
  Discover with the install trust flow, Installed with uninstall,
  seed-once demo lifecycle; the catalog publishes `index.json` —
  `57936bd`, `2a3a420`; catalog `1dd5876`
- 2026-07-22 — **Daily Care**: backups with encrypted credentials,
  strict restore ("put me back exactly where I was"), schedule
  overview, print preview, exports, the service line — `b42b675`…`b808d4d`
