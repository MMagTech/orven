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
