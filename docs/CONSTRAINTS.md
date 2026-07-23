# Orven Constraints

These rules define the product. Contributions — human or AI-assisted —
must fit inside them. If a change conflicts with this file, the change
is wrong, not the file.

## The product

1. **Orven is strictly read-only.** It observes and reports. It never
   changes, restarts, fixes, or manages anything, and it never suggests
   doing so. No feature may add remediation actions or advice.
2. **The briefing is the product.** Calm, concise, factual, readable in
   minutes. If nothing happened, say so confidently. Never a monitoring
   dashboard, never an alert stream.
3. **Core and engine are separate.** `internal/core` (UI, settings,
   export) may change freely without touching `internal/engine`
   (scheduling, plugin execution, observations, briefs), and vice
   versa. They interact only through the `Engine` type. App version and
   engine contract version are independent.

## The contract

4. **Backwards compatibility is absolute.** A plugin written for
   contract version N works on every engine ≥ N. Contract fields in
   `contract/` may be added, never removed or repurposed. New engines
   may add capabilities; old plugins must not need them.
5. **Plugins are workers.** They collect observations on the engine's
   schedule, stateless, least-privilege, isolated from each other, with
   declared permissions and no access to other plugins' credentials or
   the application's storage.
6. **The engine owns everything:** schedules, observations, briefs,
   config, secrets, run history, retention. Plugins own nothing.
7. **Presentation belongs to the brief.** Plugins emit structured
   observations, never layouts, HTML, or UI.

## Language

8. **Forbidden in any briefing or observation:** recommendations
   ("you should", "consider", "we recommend"), remediation steps
   ("restart", "reinstall", "run this command to fix"), speculation
   presented as fact, and alarmism. State what was observed; stop.
9. **Missing data is stated as missing.** "No information was
   collected" — never silence, never an implied all-clear.

## Presentation

10. **One visual identity.** The briefing is a newspaper page: serif
    typography, calm spacing, restrained color, generous margins. Every
    surface of the app shares that identity.
11. **Dark mode is lighting, not a theme.** Reading the briefing in
    dark mode must feel like reading the exact same page under
    different light — the way an e-reader switches between day and
    night. Typography, spacing, hierarchy, and personality stay
    identical; **only the color palette changes.** No glowing accents,
    gradients, neon status colors, or a separate "dashboard" identity
    after dark. If a dark-mode change would alter anything other than
    colors, the change is wrong.
12. **Freshness speaks only when noteworthy.** Fresh data earns no
    caption. Stale or delayed data is stated in one plain sentence.
    Never decorate every section with routine bookkeeping.

## Security

13. Secrets are write-only after entry: never in briefs, logs, plugin
    output, error messages, exports, or unencrypted backups.
14. **The credential publication boundary is structural.** A plugin
    receives only its own assigned credentials, via stdin, in a
    stripped environment. Everything a plugin returns — stdout fields
    and captured stderr alike — passes through the engine's sanitizer
    before storage, briefing, logging, or display: assigned secret
    values and credential-shaped content (authorization headers,
    credential query parameters) are redacted; text with no credential
    in it passes through unchanged. **The boundary's honest limit:**
    this protects against accidental disclosure; it cannot stop a
    malicious plugin from deliberately transforming (encoding,
    splitting) or exfiltrating a secret it was given. That residual
    risk is governed by the install-time trust decision, declared
    permissions, and plugin review — which is why installing a plugin
    is, and must remain presented as, a trust decision.
15. Installing a plugin is a trust decision: show source, publisher,
    version, and requested permissions before enablement. Third-party
    repositories are labeled as such.
16. Updates never silently expand permissions.
17. **Valid is never trusted.** Passing validation, matching the
    schema, or installing successfully must never be presented as an
    endorsement. The distinction between curated plugins (from the
    default repository, reviewed before acceptance) and third-party
    plugins (the user is the reviewer) must be extremely clear at
    every surface where a plugin is seen or acted on — install flow,
    plugin list, plugin page, and permission prompts — not only at
    install time. A third-party plugin stays visibly third-party
    forever.

## Plugin identity and catalogs

18. **A plugin's identity is (catalog, plugin ID).** Version is
    release metadata used for update tracking; it is not part of the
    identity. Within an installation, the bare plugin ID is the
    storage key for config, secrets, observations, and history —
    flat, stable across the plugin's life, and never
    catalog-qualified.
19. **An installation may contain only one plugin with a given plugin
    ID.** This is a product rule; the loader and duplicate detection
    enforce it, and a future installer refuses an install whose ID is
    already taken, naming the incumbent and its catalog.
20. **Orven never resolves a plugin ID across catalogs.** There is no
    search order and no catalog priority; every install is an explicit
    user pick of a specific plugin from a specific catalog. A same-ID
    plugin in another catalog is a different plugin — never a match,
    never an update.
21. **Provenance is install metadata, not identity.** The installer
    records catalog, publisher, and version alongside the plugin.
    That record powers the curated-vs-third-party labeling (§17) and
    pins updates to the recording catalog: updates only ever come from
    the catalog a plugin was installed from.

## The ecosystem

22. **The Brief is why Orven is installed; the ecosystem is why it
    stays installed.** The ecosystem succeeds when people build the
    plugins they wished existed — not when the maintainers build
    plugins for them.
23. **Difficulty must come from the observed system, never from
    Orven.** The intended plugin author runs a homelab, understands
    the system they want observed, and may never have published
    software. Understanding that system — its API, its
    authentication, its data model — is their domain knowledge, and
    may be genuinely hard; Orven adds as little as possible on top.
    When an author who understands the system they are observing is
    blocked by Orven itself — the contract, the documentation, a
    validator message, an error — the defect belongs to Orven, not
    the author: it is treated as a product defect and never closed as
    author error.
24. **The minimum path never grows.** A plugin is, and remains: one
    folder, one manifest, one entrypoint in any language, JSON in on
    stdin, JSON out on stdout, testable against a fixture without the
    real system. No change may add a required build step, framework,
    toolchain, registry, account, or ceremony to this path.
    Conveniences may be added, but each one is optional forever, and
    the plain path remains the **reference path**: documentation,
    examples, and onboarding lead with the simplest complete
    workflow. Advanced workflows may exist, but the plain path stays
    first-class in capability and presentation alike, and is never
    described as legacy.
25. **Barriers fall by subtraction, not construction.** The order is
    fixed: first, refuse to add requirements; second, improve what
    exists — documentation, validator messages, reference plugins,
    error output; only last, build something new — and a new
    authoring tool is core surface that must justify itself under
    ordinary roadmap discipline. "It lowers the barrier" is never, by
    itself, a reason to build.
26. **Maintainer plugins are ordinary plugins.** First-party plugins
    are written against the same public contract, the same
    documentation, the same validator, and the same install flow as
    anyone's. An engine capability reachable only by first-party
    plugins is a contract defect. When the maintainers meet friction
    writing a plugin, they fix the shared path; they never route
    around it privately. (The bundled demo's seed-once lifecycle is
    an installation convenience, not a capability, and remains the
    sole documented exception.)

## The reading contract

27. **The Brief prepares the reader; the reader prepares the
    response.** The plugin reports observations; Orven prepares the
    Brief; the operator decides what, if anything, to do. The
    editorial standard is the traditional nightly news: what
    happened, without speculation, prediction, recommendation, or
    importance ranking. Kept on every page, the contract is three
    promises. **When you reach the end, you are caught up** — the
    edition is complete for its window, and its length follows the
    news, never the number of installed plugins. **If it says quiet,
    it looked** — silence is earned by verified coverage; a source
    that could not be checked, or whose information is stale or
    partial, is itself news and is never folded into quiet. **The
    page tells you what happened; it never tells you what to do** —
    failures and degraded systems belong in the publication, and the
    response belongs to the reader. A Brief is an edition: it ends.

## Seeded content

28. **Seed-enablement is a single, narrow exception.** Only a
    bundled, first-party plugin that declares zero permissions and
    observes nothing but its own fixtures may be enabled by the
    application itself — exactly once, on a genuinely fresh
    installation, so the first Brief has something to show. Nothing
    installed from a catalog, curated included, is ever enabled
    without the user's explicit action (§15). Once the user disables
    or uninstalls seeded content, Orven never re-enables or
    re-installs it on its own; the deliberate "Restore the demo
    plugin" action returns it installed but disabled, because
    restoring code must not silently restore permission to run.
    Wherever the demonstration plugin contributes to a briefing,
    Coverage states that its events are fiction.

## Known enforcement gaps (recorded, planned)

These are honest limits of today's containment. Contributors must not
write documentation or UI copy that implies otherwise, and fixes are
planned as incremental milestones (OS-level plugin isolation):

- **Network egress is not enforced.** A plugin's declared network
  access is informed consent shown to the user, not a technical
  restriction; a running plugin can reach any host.
- **Plugins share the application's OS user.** A malicious plugin
  could read the application's data directory, including other
  plugins' stored secrets, directly from disk. The per-plugin
  credential boundary is structural at the input channel and against
  accidental disclosure, not against determined malice on the same
  filesystem.

Until these close, the compensating controls are the curated default
repository, small auditable plugins, declared permissions, and the
user's trust decision (§15, §17).
