# The Orven Design System

**Status: Draft 2 — Design System Office, Phase 2.** Draft 2 closes the
conformance passes Draft 1 deferred (the Archive, a beat's record, the
action rooms, and the loading state — §4.2–4.4, §5.4) and performs the
two audits Draft 1 called for but had not run: the CG-9 warn-colour
scope audit (§1.3, §9.4) and the CG-11 accessibility contrast
measurement (§8). Every reader-facing room has now been assessed against
the Charter, and every audit the document promised has been executed and
recorded. One law-level ambiguity surfaced during the passes — the
evaluative "Healthy" health-state word — and, with the owner's judgment
delegated to this office, was ruled at the level of the standard: the
label should read as the factual "Reporting". Enacting it is a
one-string implementation change, recorded as conformance debt CG-14 and
left to implementation, not made here (§9.5, SR-1).

This is the normative implementation specification for everything a
reader of Orven sees. It translates the ratified **Design Charter**
([`docs/DESIGN_CHARTER.md`](DESIGN_CHARTER.md)) into engineering
standards that an engineer — or an AI coding agent — can build against
without reinterpreting the Charter.

The Charter defines **what** Orven should feel like. This document
defines **how** that feeling is implemented.

## 0. Preface

### 0.1 What this document is — and is not

This Design System **implements** the Charter; it never modifies,
reinterprets, or extends it. It introduces no new design philosophy. It
specifies the target state derived from the Charter's laws and grammar,
and — where today's frontend differs from that target — it records the
difference as implementation debt, never as an alternative standard.

The current implementation is **informative only**. A value or pattern
does not become normative merely because it ships today. Where this
document and the running code disagree, this document (as a faithful
reading of the Charter) describes the destination, and §9 records the
gap.

### 0.2 Authority hierarchy

This document obeys, and may never contradict, everything above it:

1. Product Vision
2. Architecture (`CLAUDE.md`, `docs/CONSTRAINTS.md`)
3. Design Charter (`docs/DESIGN_CHARTER.md`)
4. **Design System (this document)**
5. Components
6. Application

When implementation exposes an ambiguity or a conflict with anything
above, this document **surfaces it for owner review** (§9) rather than
inventing a resolution.

### 0.3 How to read a specification

Every major section is written in three clearly separated registers.
An engineer building a feature reads all three; a reviewer checking
conformance reads Normative and Current Conformance.

- **Normative** — mandatory. Requirements derived directly from the
  Charter or the constitution, each cited to its source. Building
  something that violates a Normative clause produces a
  non-conformant, and in the strong cases a non-Orven, result.
- **Implementation Notes** — optional engineering guidance: how to
  satisfy the Normative clauses in this codebase, worked values,
  pitfalls. Advisory. An engineer may achieve the Normative outcome
  another way.
- **Current Conformance** — whether today's code already satisfies the
  section, and precisely where it does not. Every "no" or "partial"
  here has a matching entry in the §9 conformance register.

Each Normative clause is also labelled by **Charter layer**, because
the layer decides how freely it may change:

- **(law)** — never changes; a violation is a different product.
- **(grammar)** — Orven's chosen device; changes only by owner
  amendment.
- **(furniture)** — per-room, per-medium arrangement; free to vary
  within the laws.

Citations use the short forms **[Law N]**, **[grammar: *device*]**,
**[Charter §]** for Charter clauses, and **[C§n]** for
`docs/CONSTRAINTS.md`.

### 0.4 Owner-defined parameters

The Charter deliberately leaves some values to the product owner, to be
set from experience with real data rather than from speculation. This
document specifies the **mechanism** around each such value and marks
the value itself **⟨OWNER-DEFINED⟩**. It never invents one. Every
⟨OWNER-DEFINED⟩ marker is collected in the register at §9.2. An
implementation encountering an unfilled parameter stops and requests a
product decision; it does not choose a default that would silently
become the standard.

### 0.5 The implementation constraint that shapes everything

Orven's architecture forbids new runtime and build dependencies
(`CLAUDE.md`, "Never"): the Go app depends on `yaml.v3` alone and
plugins use the standard library only. Therefore:

- **Design tokens are CSS custom properties** declared in `:root`
  ([`internal/core/static/style.css`](../internal/core/static/style.css)).
  There is no token-build pipeline, no preprocessor, no CSS framework,
  and this document must not introduce one.
- The two appearance modes are expressed with the CSS `light-dark()`
  function and `color-scheme`, never with duplicated rule sets
  ([C§11]).
- Any interactivity is progressive enhancement over server-rendered
  HTML templates ([`internal/core/templates/`](../internal/core/templates/)).
  The Brief must render and read correctly with no JavaScript.

---

## 1. Foundations — the tokens

The tokens are the vocabulary every later section spends. They are
furniture in the Charter's sense — the specific palette and faces are
"today's choices, not the identity" [grammar: two inks] — but within
this Design System their **names and roles are normative**: components
reference tokens, never raw values, so that a future owner-approved
palette change is a single edit.

### 1.1 Normative

**Token discipline (furniture role, normative here).** Every color,
type face, size, and spacing value a component uses **must** be a named
token from this section. No component may hard-code a hex color, a
pixel size, or a font stack. New tokens are added here, in the open, so
the vocabulary stays auditable.

**Colour — two inks, and colour is the only thing night changes.**

- There are exactly **two inks**: a **report ink** for the reporting
  voice and an **apparatus ink** for the labelling voice [Law III],
  [grammar: two inks]. "Ink" here is role, not a single token — each
  voice draws its foreground, and shares the surface and rule tokens
  below.
- Every colour token is a `light-dark()` pair. Day and night are **the
  same page in different colours, and colours only** [Law III],
  [C§11]. No token may introduce a glow, gradient, neon, or a
  night-only accent.
- **No third voice.** There is **no status, alarm, success, or
  attention colour** anywhere a reader reads the report [Law III]. Red,
  green, amber, and any semantic colour channel are forbidden as
  meaning-bearers. Prominence is never carried by colour (see §2.5).
- **The single warning exception (settled by owner ruling, 23 July
  2026).** A soft warning colour (the `--warn-*` tokens) is permitted
  **only in action-oriented areas outside the Brief**, and **only** when
  the reader must understand a **genuine risk or consequence before
  proceeding** — a destructive action, a security consequence, a trust
  acknowledgement, or an irreversible decision. It must **never** appear
  in the Brief, never mark routine status, never manufacture urgency for
  an ordinary condition, and never become a general third visual voice.
  Its use stays rare and confined to those cases. This is a narrow,
  fenced action-room exception — the reader's report is still governed
  by "no third voice" above [Law III].

The canonical token set (roles are normative; the paired values are the
current working default and are owner-revisable furniture):

| Token | Role | Voice |
|---|---|---|
| `--paper` | Outermost background (the desk) | surface |
| `--sheet` | The document surface (the page) | surface |
| `--ink` | Report foreground — the reporting voice | report ink |
| `--mid` | Secondary apparatus foreground | apparatus ink |
| `--faint` | Quiet apparatus foreground (labels, captions) | apparatus ink |
| `--rule` | Hairlines, borders, the bounds | structure |
| `--accent` | The single restrained editorial accent | report ink |
| `--field` | Form input surface (action rooms) | surface |

**The accent is not a status colour.** `--accent` is one restrained
editorial colour (links, the curated mark). It never signals
importance, health, or urgency, and it never appears as a fill behind
text meant to draw the eye [Law III]. Its use is enumerated in §3; any
new use is an amendment-level decision.

**Typography — the two inks are two faces.**

- The **report** is set in a reading **serif** (Georgia-class); the
  **apparatus** is set in a quiet **sans** (Segoe-class) [grammar: two
  inks]. Two families, one each per voice. A third family is permitted
  only for genuinely monospaced technical text (e.g. raw logs), which
  is neither report nor apparatus.
- In **rooms where the reader acts, the voices invert**: the sans
  apparatus leads and the serif explains [grammar: two inks], (§4.4).
- Tokens: `--face-report` (serif stack), `--face-apparatus` (sans
  stack), `--face-mono` (monospace stack, technical only).

**Typographic scale — prominence is a small set of measured steps.**

- Prominence is expressed as **size**, drawn from a **small, fixed
  ordered scale of steps** [Law II], [grammar: size as measurement].
  Bigger type means *more happened*, never *this matters more*.
- The **number of steps** and the **count thresholds** that map an
  observation count to a step are **⟨OWNER-DEFINED⟩** (§9.2). This
  document defines the scale as an ordered token series
  `--tier-1 … --tier-T` where step 1 is the most prominent; it does not
  fix `T` or the thresholds.
- Non-prominence sizes (body, apparatus label, caption) are fixed
  furniture tokens and carry no prominence meaning.

**Spacing & measure.**

- Spacing is drawn from a **single named scale** (`--space-*`); no
  component invents an off-scale margin.
- The report is **one readable column, ragged right — never justified**
  [Law II], [grammar: size as measurement]. The column has a fixed
  maximum **measure** token (`--measure`). Parallel stories may sit side
  by side **only where the medium is wide** [Law VI] (§6).

### 1.2 Implementation Notes

- The current working palette (light / dark pairs), suitable to carry
  forward verbatim as the default values behind the token roles above:

  ```
  --paper:  light-dark(#f7f3ea, #211f1b);
  --sheet:  light-dark(#fffdf7, #282521);
  --ink:    light-dark(#2b2a26, #d8d2c3);
  --faint:  light-dark(#8a8578, #96907f);
  --mid:    light-dark(#6f6a5e, #aaa393);
  --rule:   light-dark(#d8d2c2, #3d392f);
  --accent: light-dark(#6b4f2a, #c2a172);
  --field:  light-dark(#ffffff, #1d1b17);
  ```

- Report face today: `Georgia, "Times New Roman", serif`. A suitable
  apparatus sans stack for `--face-apparatus` would be a system-sans
  chain (e.g. `"Segoe UI", system-ui, sans-serif`) so that no font is
  shipped or fetched (dependency rule, §0.5).
- Prefer defining the tier scale as CSS custom properties even before
  `T` is fixed, so that filling the owner decision is one edit and no
  component hard-codes a size:
  `--tier-1: …; --tier-2: …;` (values pending §9.2).
- `--measure` is `46rem` today. Keep the measure a token so §6 and §7
  can reference one value.

### 1.3 Current Conformance

- **Token vocabulary — partial.** The colour tokens exist and are
  already `light-dark()` pairs (good). Type, size, and spacing are
  **not** tokenised: font stacks and sizes are written inline across
  [`style.css`](../internal/core/static/style.css) (e.g. the body font
  literal, per-component `font-size` values). → **CG-10**.
- **Two inks — not met.** Only the serif is present; the apparatus is
  currently rendered as *smaller serif*, not a sans voice
  (`.pill`, `.coverage h3`, nav, and captions all inherit Georgia). The
  second ink face does not exist yet. → **CG-7**.
- **Warn palette — permitted; audit performed (Draft 2).** A warn
  palette (`--warn-ink`, `--warn-bg`, `--warn-border`) exists and renders
  a muted red. The CG-9 audit (§9.4) has now been run. **The
  load-bearing check passes:** neither the Brief nor its print rendering
  ([`front.html`](../internal/core/templates/front.html),
  [`printbrief.html`](../internal/core/templates/printbrief.html))
  contains any `--warn-*` usage — the reader's report is clean of the
  soft-alarm channel [Law III]. The residual is action-room debt: the
  warn colour is also spent on **error/failure states** (a repository
  fetch error, a plugin load error, an engine-incompatibility notice, an
  install error) and on **one list-status marking** (a failed plugin's
  load error in the Installed list), which are error reporting governed
  by "no alarm" (§5.6), not "genuine risk or consequence before
  proceeding" (§9.4). → **CG-9** (audited; residual itemised).
- **Prominence scale — absent.** No tier tokens exist; stories render
  at one uniform size. → **CG-3**.
- **Measure & ragged-right — met.** `max-width: 46rem`; body text is
  not justified.

---

## 2. The grammar as components

Six devices are Orven's identity — another product could obey the same
laws with different grammar; this grammar is what makes Orven look like
Orven [Charter, "The grammar"]. Each device below is specified as a
component: the law it serves, its anatomy, the tokens it spends, its
behaviour across media, and its revisable (furniture) parts.

### 2.1 The bounds

**Normative.**

- A **double rule opens every document and a double rule closes it**
  [grammar: the bounds], [Law I], [Law V]. The edition lives between
  the bounds; nothing outside them is part of the edition.
- The opening and closing bounds are **both required** and appear on
  **every** edition — heaviest and emptiest day alike [Law V].
- The bounds are drawn with `--rule`. That the mark is specifically a
  *double* rule is the revisable signature; that a boundary opens and
  closes the document is law.

**Implementation Notes.**

- The closing bound pairs with the end mark (§2.6) at the foot of the
  document; together they are the reader's signal that the edition is
  complete.
- A double rule is cheaply expressed as a `3px double var(--rule)`
  border, matching the existing opening rule, or as a dedicated
  bounding element so print (§7) can keep it.

**Current Conformance.**

- **Partial.** An opening rule exists beneath the dateline
  ([`style.css:51`](../internal/core/static/style.css), `3px double`).
  There is **no closing double rule** at the foot of the document.
  → **CG-1**.

### 2.2 The certification line

**Normative.**

- One quiet line at the head of every document carries **identity ·
  date · edition number · accounting**, identical in structure every
  day [grammar: the certification line], [Law V].
- The **accounting** states the edition's own arithmetic and its
  **coverage window** — e.g. "47 observations from 15 of 17 plugins ·
  observations through 6:45" [Law V, as amended]. Both the edition's
  completeness and its staleness are **declared, not discovered**.
- The line is set in the **apparatus** voice (§1.1). Its composition
  and position are revisable; its **presence and constancy are not**
  [grammar: the certification line].
- The **edition number** and the scheme that assigns it are
  **⟨OWNER-DEFINED⟩** (§9.2): whether Orven maintains a monotonic
  edition sequence, and how it is displayed, is a product decision.
  This document requires a stable, non-ranking identifier in this slot
  and invents none.

**Implementation Notes.**

- The certification line draws its accounting from the manifest's
  figures (§2.3, single source of truth), so the head of the page and
  its body always agree.
- "n of m plugins" is a count relationship, never a percentage or grade
  [Law VII]; do not render it as "88% coverage."

**Current Conformance.**

- **Partial.** The dateline carries identity (masthead) and date, and a
  prepared-at time and an edition *phrase* ("covers activity since the
  previous Brief") ([`front.html:6`](../internal/core/templates/front.html)).
  It carries **no edition number** and **no accounting/coverage-window**
  on the line. → **CG-5**.

### 2.3 The manifest

**Normative.**

- The **whole before the parts**: every source and its count appears
  **up front**, before the stories [grammar: the manifest], [Law II],
  [Law V].
- On **quiet days the manifest may name the sources instead of
  counting** them, because presence is the day's only news [grammar:
  the manifest], (§5.2).
- The manifest's **form and its truncation rules on small media** are
  revisable furniture and, for truncation, **⟨OWNER-DEFINED⟩** (§9.2,
  §6).

**Implementation Notes.**

- **Single source of truth for counts.** The certification-line
  accounting (totals, §2.2) and the per-story printed counts (§2.5) are
  the manifest's per-source figures at coarser and finer grains. Compute
  the per-source counts once and derive all three surfaces from them, so
  the head, the manifest, and the stories can never disagree.

**Current Conformance.**

- **Not met (placement).** The equivalent content — the "Coverage"
  section listing contributors — renders at the **foot** of the Brief
  ([`front.html:48`](../internal/core/templates/front.html)), not up
  front, and is framed as closing coverage rather than an opening
  manifest. → **CG-2**.

### 2.4 Two inks

**Normative.**

- The **report speaks in full declarative sentences** in the serif; the
  **apparatus labels in a smaller, plainer sans** [Law III], [grammar:
  two inks].
- **No third voice** — no icons, badges, or alarm colour — ever speaks
  [Law III] (§3.6 enumerates the only permitted glyphs, all of them
  furniture in the chrome, none in the report).
- In **action rooms the voices invert**: sans leads, serif explains
  [grammar: two inks], (§4.4). Day and night still differ in colour only
  (§1.1) [Law III], [C§11].

**Implementation Notes.**

- Inversion is a room-level rule, not a component rule: the same button
  or field inherits its leading voice from the room it sits in.

**Current Conformance.**

- **Not met.** The apparatus sans is not implemented; all voices are
  serif (§1.3). Inversion in action rooms therefore cannot be
  expressed yet. → **CG-7**.

### 2.5 Size as measurement

**Normative.**

- Prominence is **size**, and size is **assigned by observation count
  alone** — bigger means more happened, never more important [Law II],
  [grammar: size as measurement].
- **The count is printed beside the claim** — e.g. "Containers · 9
  observations" — so the reader can **audit the layout every morning**
  [Law II]. The printed count is not decoration; it is the defense
  against verbosity masquerading as importance [Charter, commentary on
  Law II]. **A prominence step without its printed count is
  non-conformant.**
- A **perfectly flat page is the floor**, not a second mode: it is what
  the tiers become when counts are equal [Law II]. There is one visual
  language; flatness is never offered as a toggle.
- **Position carries no verdict** [Law VI]. Order is not ranking.
- The **number of steps** and the **thresholds** are **⟨OWNER-DEFINED⟩**
  (§9.2). Specified mechanism: a pure, total function
  `tier(count) → step` that is (a) identical every day, (b) monotonic
  (never-decreasing prominence for higher counts), and (c) defined for
  every possible count including zero and the overflow beyond the
  largest threshold. This document fixes the function's *properties*
  and leaves its *values* to the owner.

**Implementation Notes.**

- The honest ceiling observed in discovery was ~60–70 observations
  before bands repeat and the page simply grows taller [Sprint 03,
  informative]; growing taller is acceptable — the edition ends
  regardless (§2.6).
- Render the printed count in the apparatus voice adjacent to the story
  title, drawn from the same per-source figure as the manifest (§2.3).
- Map `tier(count)` to the `--tier-n` type token; never to a colour or
  weight-only change (colour is forbidden as prominence, §1.1).

**Current Conformance.**

- **Not met.** Stories render at a single uniform size
  (`.story h2`, [`style.css:58`](../internal/core/static/style.css)); no
  tier function exists and no per-source count is printed beside a
  story. → **CG-3**, **CG-4**.

### 2.6 The reserved seat & the end

**Normative.**

- **Silence sits in the design's only enclosure** — a box — wherever it
  occurs, structurally, never as a footnote and never as an implied
  all-clear [Law IV], [grammar: the reserved seat & the end], [C§9].
  The enclosure is the *only* boxed element in the reading surface;
  nothing else earns a box, so the box unmistakably means silence.
- The document **closes with an explicit end mark and its own
  accounting** [Law I], [Law V], pairing with the closing bound (§2.1).
- The **end mark's wording** and the **box's per-room placement** are
  revisable furniture; the end mark's wording is **⟨OWNER-DEFINED⟩**
  (§9.2).

**Implementation Notes.**

- "When you reach the end, you are caught up" [C§27] is exactly what the
  end mark certifies; its wording should say completeness, not farewell.
- The reserved seat is the same structure across rooms (the Brief's
  quiet edition, a source's silence in its record) — one enclosure
  component, reused (§5.2, §4.3).

**Current Conformance.**

- **Partial / not met.** The quiet state is centred italic text
  (`.quiet`, [`style.css:65`](../internal/core/static/style.css)), **not
  an enclosure/box**. There is **no explicit end mark** closing the
  document. → **CG-6**, **CG-8**.

---

## 3. Shared furniture — app-wide components

These are reusable parts that are **not** Charter grammar but must stay
consistent across every room so the app reads as one identity [C§10].
They are furniture: free to rearrange within the laws. Each is specified
with its states.

### 3.1 Normative (applies to all furniture in this section)

- Every component draws only from §1 tokens and never introduces a
  colour channel of meaning [Law III].
- Every interactive control is reachable and operable without
  JavaScript where it performs a server action (§0.5), and has a
  visible focus state (§8).
- Trust distinctions are shown wherever a plugin is seen or acted on,
  not only at install [C§17] (§3.4).

### 3.2 Masthead & navigation

- **Normative.** A quiet wordmark identifies the app; primary
  navigation is a small apparatus-voice row. The masthead is chrome,
  not part of the edition (it sits outside the bounds, §2.1) and is
  omitted in print (§7). The current-room link is marked as current
  (furniture).
- **Implementation Notes.** Today: `.masthead` / `.wordmark` /
  `.masthead nav` in [`style.css`](../internal/core/static/style.css),
  rendered by [`_chrome.html`](../internal/core/templates/_chrome.html).
  The wordmark uses tracked uppercase — furniture, retainable.
- **Current Conformance.** Met as furniture. The nav renders in serif;
  once the apparatus sans exists (§2.4) the nav should adopt it as
  apparatus voice. → tracked under **CG-7**.

### 3.3 Buttons, quiet actions & the service line

- **Normative.** Actions in reading rooms speak in the **margin's
  voice** — quiet, secondary, never competing with the news [Law III
  spirit; C§2 "never an alert stream"]. A reader action is offered, not
  urged. **Destructive or irreversible actions** must be clearly
  labelled and confirmed (product/UX safety); these are one of the
  sanctioned cases where the soft warning colour is permitted (§1.1) —
  in the action room, never in the Brief.
- **Implementation Notes.** The service line
  (`.service-line`, `.quiet-action`) and standard `button` styles exist
  and match this intent. Keep primary-action buttons for **action
  rooms** (§4.4) and quiet actions for **reading rooms**.
- **Current Conformance.** Met. Service line renders quiet, borderless,
  centered actions ([`front.html:62`](../internal/core/templates/front.html)).

### 3.4 Trust labelling — curated vs. third-party

- **Normative.** The distinction between curated and third-party
  plugins is shown at **every** surface where a plugin is seen or acted
  on — install flow, plugin list, plugin page, permission prompts — and
  a third-party plugin stays visibly third-party forever [C§17]. **Valid
  is never trusted** [C§17]: successful validation or install is never
  presented as endorsement. The distinction must be legible **without
  relying on colour alone** [Law III; §8].
- **Implementation Notes.** Today the pills encode the distinction with
  **border style and the accent ink** (`.pill-curated` uses `--accent`;
  `.pill-third` uses a dashed border) — a non-colour-channel encoding,
  which is Law III-compatible. Preserve the non-colour encoding if the
  pill visuals change.
- **Current Conformance.** Met, with a caveat: verify the third-party
  cue survives for colour-blind and monochrome readers (§8).

### 3.5 Forms & fields (action-room furniture)

- **Normative.** Where the reader acts, **every choice is explained in a
  sentence of consequence**, and conventional controls are permitted —
  friction is a cost paid by a person [Charter, Settings room], [Law
  VIII]. Fields draw `--field`; help text is apparatus voice. A change
  is recorded as an attributed, dated fact, effective at the next
  edition [Law VIII].
- **Implementation Notes.** `.form`, `.help`, `.form-inline`,
  `.inline-input` exist and match. In action rooms the leading voice is
  the sans (§2.4 inversion) once available.
- **Current Conformance.** Met as furniture; voice inversion pending
  **CG-7**.

### 3.6 Iconography — the deliberate near-absence

- **Normative.** There is **no iconography in the report and no icon
  used to carry status or meaning** [Law III]. Icons are a third voice
  and are forbidden as meaning-bearers. The **only** permitted glyphs
  are non-semantic **control affordances in the chrome/apparatus**: the
  lighting control's mode glyphs and a disclosure caret for
  progressive-disclosure controls. No new glyph may encode a fact,
  status, or importance.
- **Implementation Notes.** Today: `◐ ☾ ☀` on the appearance control
  and `▾` on the download disclosure
  ([`_chrome.html`](../internal/core/templates/_chrome.html),
  [`front.html`](../internal/core/templates/front.html)). These are
  affordances, not status — permitted.
- **Current Conformance.** Met.

### 3.7 Notices, tabs, pills, tables (secondary furniture)

- **Normative.** Secondary furniture (flash notices, tab strips,
  metadata pills, list tables) is apparatus-voiced, token-drawn, and
  colour-neutral. A notice that must draw the eye still may not use an
  alarm colour in a reading room [Law III].
- **Implementation Notes.** `.flash`, `.tabs`, `.pill`, `.count`,
  `table.list` exist. The `.warn` family is governed by the settled rule
  (§1.1, §9.4): confined to action rooms, genuine-consequence cases only,
  and never reaching the Brief or marking routine status.
- **Current Conformance.** Met; the `.warn` scope audit is tracked as
  **CG-9**.

---

## 4. The rooms — page templates

The Charter defines **five rooms** by what they fundamentally *are*;
everything else about a room is furniture [Charter, "The rooms"]. Each
template below states the room's essence (law), which grammar devices it
must show, and its furniture. The app's remaining pages are mapped to
the room-identity that governs them.

### 4.1 The Brief — `front.html`

- **Normative.** One **finite, dated, numbered** edition of what
  happened, what changed, and what could not be seen [Law I], composed
  by the six devices (§2), readable in minutes. It carries, in order:
  the opening **bounds** → the **certification line** → the **manifest**
  → the stories sized by **measurement** with printed counts → the
  **reserved seat** for any silence → the **end mark** and closing
  **bounds**. Its length **follows the news, never the number of
  installed plugins** [C§27]. A source that could not be checked, or is
  stale or partial, is itself news and is **never folded into quiet**
  [C§27], [Law IV].
- **Implementation Notes.** This is the room with the largest gap to the
  Charter; §9.1 lists the specific device work. Compose the page from
  the §2 components so the ordering above is structural, not incidental.
- **Current Conformance.** Partial: identity, dateline, stories, quiet
  and coverage content, freshness, and the service line exist; the
  bounds (close), manifest placement, measurement + printed counts,
  edition number/accounting, reserved-seat enclosure, and end mark do
  not. → **CG-1…CG-8**.

### 4.2 The Archive — `history.html`

- **Normative.** The **unbroken chronological record** of editions,
  **every one present with equal standing** — a quiet day is a full
  citizen [Charter, the Archive], [Law IV]. It **states its own extent**.
  **No aggregation that ranks days; no streaks; no scores** [Law VII].
  Each edition appears as its certification line (§2.2).
- **Implementation Notes.** Ensure quiet and incomplete editions are
  listed identically to eventful ones — equal standing is the law here.
- **Current Conformance (assessed, Draft 2).** Against
  [`history.html`](../internal/core/templates/history.html):
  - **Equal standing — met.** Editions are grouped by day and each is
    listed as a row (a time-link with a muted "all quiet" or "*n*
    section(s)" note). A quiet day is listed identically to an eventful
    one — a full citizen [Law IV]. That a single-edition day renders
    open is furniture, not ranking.
  - **No ranking / no aggregation — met.** There are no streaks, scores,
    or day-vs-day comparisons; "*n* sections" is a count relationship,
    not a grade [Law VII].
  - **States its own extent — partial.** The room states its retention
    window ("kept for *N* days") but not the actual span it holds (the
    earliest edition still present). Retention is a policy, extent is a
    fact; the Charter asks the Archive to state its extent.
  - **Editions as certification lines — not met.** Editions are listed
    as bare time-links plus a muted section count, not as the
    certification line (§2.2). When the certification line gains its
    edition number and accounting (CG-5), the Archive's rows should adopt
    that same structure so every edition is cited identically wherever it
    appears. → tracked under **CG-5**.
  - **Empty state — not the enclosure.** "No briefings yet" renders as
    `.quiet` centred italic, not the reserved-seat enclosure (§2.6).
    → **CG-8**.

### 4.3 A beat's record — `plugin.html`

- **Normative.** The **public record of a reporting source**: its
  **credentials as counts and dates**, its **record as dated
  sentences**, and its **silences seated chronologically** within that
  record [Charter, a beat's record], [Law IV]. About the source's own
  operation, Orven speaks **only in counts and dates — never a score,
  streak, or grade** [Law VII]. Trust labelling (§3.4) appears here
  [C§17].
- **Implementation Notes.** "filed in 213 of 217 editions," never "98.2
  percent" [Law VII]. The silence enclosure (§2.6) is reused for a
  source's gaps.
- **Current Conformance (assessed, Draft 2).** Against
  [`plugin.html`](../internal/core/templates/plugin.html):
  - **Trust labelling — met.** The curated / third-party / manual
    distinction is shown via the pills (§3.4) [C§17].
  - **Credentials as counts and dates — partial.** The record shows the
    source's collection interval and relative "last ran / last success"
    times, and a "Recent runs" table (when · result · note). It does
    **not** present the Charter's signature credential — the count of
    editions the source has filed in ("filed in 213 of 217 editions").
    The record reads as a controls-and-runs panel, not the "record as
    dated sentences" the room is defined by. → **CG-13**.
  - **Silences seated chronologically — not met.** A source's gaps are
    not seated as the reserved-seat enclosure within its record (§2.6);
    the runs table lists runs but does not structurally seat silence.
    → tracked under **CG-8**.
  - **Health label — evaluative "Healthy" (specified fix, not yet
    implemented).** The `{{.Health}}` pill renders a derived current-state
    word. All but one of its values are neutral facts about the last run
    or configuration ("Disabled", "Ready", "Waiting for next scheduled
    run", "Running", "Partial data", "Source unavailable",
    "Authentication failed", "Timed out", "Failed" —
    [`engine.go`](../internal/engine/engine.go)), which are Law
    VII-compatible state facts. The one evaluative value, **"Healthy"**,
    is Orven characterising its own witness's condition, where Law VII
    holds that about its own operation Orven speaks "only in counts and
    dates … never a score, streak, or grade" and "does not rate its
    witnesses." **The Design System's standard: this label must read as a
    factual state, not a rating — "Reporting"** (the last collection ran
    and completed; what the source *is doing*). Applying it is a
    one-string change in `engine.go` that belongs to implementation, not
    to this specification — tracked as **CG-14**. Rationale and the
    considered alternative ("Filed") are recorded at **§9.5, SR-1**.

### 4.4 Settings — the room where the reader acts

- **Normative.** The **identity is the writing, not the widgetry**:
  every choice explained in a **sentence of consequence**; every change
  an **attributed, dated fact, effective at the next edition** — "Paused
  by you, 12 July" [Charter, Settings], [Law VIII]. **The two voices
  invert** here (§2.4). Conventional controls are permitted [Charter,
  Settings]. Nothing in this room advises or recommends [C§1], [C§8].
- **Implementation Notes.** Applies to `settings.html` and, by
  room-identity, to every acting surface: `discover.html`,
  `install.html`, `uninstall.html`, `restore.html`, backups, and secret
  entry. Secrets are write-only after entry and never echoed [C§13].
- **Current Conformance (assessed, Draft 2).** Against `settings.html`,
  `discover.html`, `install.html`, `uninstall.html`, `backups.html`,
  `restore.html`, `restored.html`:
  - **Sentence of consequence — met.** The rooms explain choices in
    consequence sentences rather than leaning on widgetry; the backups,
    restore, and uninstall flows are exemplary ("Restoring overwrites the
    present with the past. A safety backup … is written first") [Law
    VIII]. Conventional controls are used freely, as permitted.
  - **Nothing advises — met, with two borderline lines.** The rooms
    state facts and consequences, not recommendations. Two lines lean
    toward guidance and are worth an editorial eye without being clear
    violations: install's "treat that silence as a reason for more
    caution, not less" and backups' "point this at a mapped path your
    existing backup tooling already covers." Both concern the reader's
    own action (a trust decision, an operational setup), the room where
    consequence-guidance is permitted [Charter, Settings] — distinct from
    the forbidden advice *about an observed system* [C§1, C§8].
  - **Attributed, dated fact effective at next edition — not met.** This
    is the room's defining law and its clearest gap. A change is applied
    and confirmed with a transient `.flash` line, but it is not recorded
    as an attributed, dated fact effective at the next edition — there is
    no "Disabled by you, 23 July, effective next Brief" [Law VIII]. The
    fact of the change, its attribution, and its effective edition are
    not surfaced. → **CG-12**.
  - **Voice inversion — not met.** The sans apparatus does not exist, so
    the action rooms cannot invert their voices (§2.4). → **CG-7**.
  - **Secrets — met.** Secret fields are write-only after entry and
    render only "configured — leave blank to keep", never the value
    ([`plugin.html`](../internal/core/templates/plugin.html),
    [`backups.html`](../internal/core/templates/backups.html)) [C§13].

### 4.5 The quiet edition

- **Normative.** The **full frame around a stated nothing**: the
  apparatus (bounds, certification line, manifest, end) certifies the
  quiet; the **whitespace is the news** [Charter, the quiet edition],
  [Law IV]. **"If it says quiet, it looked"** — silence is earned by
  verified coverage; anything unchecked, stale, or partial is excluded
  from quiet [C§27]. On a quiet day the manifest may **name** the
  sources rather than count them (§2.3).
- **Implementation Notes.** The quiet edition is the Brief room (§4.1)
  with its stories replaced by the reserved seat (§2.6) — same frame,
  same devices, different centre. It is not a separate template.
- **Current Conformance.** Partial: a quiet message exists
  ([`front.html:12`](../internal/core/templates/front.html)) but without
  the enclosure, the up-front manifest, or the certifying frame.
  → **CG-2**, **CG-6**, **CG-8**.

### 4.6 Room map for the remaining pages

The following pages have no distinct Charter room; each inherits the
identity of the room named. All obey §3 furniture and the laws.

| Template | Governing room-identity |
|---|---|
| `discover.html`, `install.html`, `uninstall.html` | Settings / action room (§4.4) |
| `plugins.html` | Between the Archive's listing discipline and the action room; a list of sources, acting furniture permitted |
| `backups.html`, `restore.html`, `restored.html` | Action room (§4.4) |
| `logs.html` | Technical surface; monospace face permitted (§1.1); still token-drawn and colour-neutral |
| `printbrief.html` | The Brief (§4.1) in the print medium (§7) |

---

## 5. States

Every reader-facing state is specified once here and reused by the rooms
(§4). States are where the promise most often breaks, so each is anchored
to a law.

### 5.1 The three briefing states

**Normative.**

1. **News** — one or more stories. Composed per §4.1, sized by
   measurement (§2.5).
2. **Quiet** — verified full coverage and nothing changed. Rendered as
   the quiet edition (§4.5). Quiet is a **positive, certified**
   statement, never an absence of rendering [Law IV], [C§27].
3. **Incomplete coverage** — one or more sources could not be checked,
   or returned stale/partial data. This is **news, stated
   structurally**, and is **never folded into quiet** [Law IV], [C§27].
   The uncheckable source is named; the reason is a fact, with **no
   remediation** [C§8].

**Normative — how the states combine.** These are not three mutually
exclusive layouts. **Quiet is exclusive**: it requires verified full
coverage *and* no changes, and **any incomplete coverage precludes it**
("if it says quiet, it looked" [C§27]). **Incomplete coverage is an
overlay** that can accompany news — a Brief may carry stories *and* a
named source it could not check. So a Brief is exactly one of {news,
quiet}, optionally carrying the incomplete-coverage overlay (which
forces the news form whenever it is present).

**Implementation Notes.** Coverage completeness is the
certification-line accounting (§2.2), so the top-of-page state and the
head of the page always agree; per-story freshness (§5.5) composes
independently.

**Current Conformance.** All three states exist in
[`front.html`](../internal/core/templates/front.html) (`.confidence`,
`.quiet`, stories). Their *framing* is non-conformant per §2 (manifest,
bounds, enclosure, end). → carried by CG-1…CG-8.

### 5.2 Silence / empty (the reserved seat)

- **Normative.** Absence of data is **stated wherever it occurs**, in
  the design's enclosure, never as a blank and never as an implied
  all-clear [Law IV], [C§9]. This covers the quiet edition, an empty
  Archive, and a source with no record.
- **Implementation Notes.** One enclosure component (§2.6), reused.
- **Current Conformance.** → **CG-8**.

### 5.3 First run / onboarding

- **Normative.** A genuinely fresh install may seed-enable exactly one
  bundled, zero-permission demonstration plugin so the first Brief has
  something to show [C§28]; nothing from a catalog is auto-enabled
  [C§15]. Wherever the demo contributes, coverage states its **events
  are fiction** [C§28]. Onboarding copy states facts and invites; it
  does not advise [C§1].
- **Implementation Notes.** Present in the current welcome/first-Brief
  flow; keep the fiction disclosure attached to any demo contribution.
- **Current Conformance.** Met (onboarding + fiction note exist in
  [`front.html`](../internal/core/templates/front.html)); framing
  inherits the §2 gaps.

### 5.4 Loading / preparing

- **Normative.** Preparing a Brief is an occasional, explicit action
  with a finite end. Any progress indication must **not read as urgency
  and must not invite the reader to keep watching** [Law I "not asking
  the reader back early"], [C§2]. There is **no perpetual live state**:
  the edition either exists or is being prepared.
- **Implementation Notes.** A plain "Preparing…" apparatus line is the
  simplest way to satisfy the rule; a spinner or a live-updating counter
  tends to read as an alert and is best avoided. Any motion used must
  respect reduced-motion (§8).
- **Current Conformance (assessed, Draft 2).** **Met by synchronous
  design.** Preparing a Brief is a discrete `POST /generate` from a
  quiet-action button that returns the finished edition; there is no
  spinner, no live counter, and **no perpetual or self-refreshing
  state** anywhere in the templates or the chrome script — the edition
  either exists or is being prepared, exactly as the law requires [Law
  I], [C§2]. No progress indication exists to read as urgency, so the
  rule is satisfied. Note for the future: should generation ever become
  slow or asynchronous, the sanctioned pattern is a plain "Preparing…"
  apparatus line (never a spinner or live tally), honouring reduced
  motion (§8).

### 5.5 Stale / partial (freshness)

- **Normative.** **Freshness speaks only when noteworthy** — fresh data
  earns no caption; stale or delayed data is stated in **one plain
  sentence** [C§12], [grammar spirit]. **Partial** availability from a
  source is stated as fact. Neither is decorated, colour-coded, or
  repeated as routine bookkeeping [C§12], [Law III].
- **Implementation Notes.** `.freshness`, `.freshness.stale`,
  `.status-note` exist and match the one-sentence rule.
- **Current Conformance.** Met.

### 5.6 Error presentation

- **Normative.** A reader-facing failure is **reported as a fact,
  attributed and dated, with no remediation and no alarm** — "Could not
  be checked: *name* — *reason*" [C§8], [Law IV], [Law III]. Failures
  and degraded systems **belong in the publication**; the response
  belongs to the reader [C§27]. Secrets never appear in an error [C§13].
  Operator-facing raw logs (`logs.html`) are a technical surface, not
  the reader's voice, and are exempt from the reporting-voice rules but
  not from the secret rule.
- **Implementation Notes.** The Coverage "Could not be checked" line is
  the model; reuse its phrasing for any reader-facing error.
- **Current Conformance (audited, Draft 2).** Met for the reader-facing
  surfaces. The Brief reports failures as attributed facts with no
  remediation ("Could not be checked: *name* — *reason*",
  [`front.html`](../internal/core/templates/front.html)); no
  reader-facing error surface carries remediation phrasing, and
  `logs.html` is a technical operator surface (monospace, exempt from the
  reporting voice but not the secret rule) [C§13]. The one cross-cutting
  finding is presentational, tracked under CG-9: several **action-room**
  error notices (repository fetch error, plugin load error, engine
  incompatibility, install error) render in the warn colour, which is a
  soft-alarm channel — errors are to be stated "with no alarm." The
  wording is conformant; the colouring is the residual CG-9 debt.

---

## 6. Responsive behaviour

**Normative.**

- **Geometry flexes; grammar does not** [Law VI]. The same grammar holds
  on desktop, phone, and print — **order and apparatus never change with
  width**; only the arrangement does.
- The report stays **one readable column** at the `--measure` cap;
  **parallel stories sit side by side only where the medium is wide**
  [Law II], [Law VI]. Narrow media are always single-column.
- The **manifest's truncation rules on small media** are revisable
  furniture and **⟨OWNER-DEFINED⟩** (§9.2): what a long manifest does on
  a phone (wrap, summarise to a count, collapse) is a product decision,
  not an invented default.
- **Position carries no verdict** at any width [Law VI]; a reflow must
  never read as re-ranking.

**Implementation Notes.**

- Desktop, mobile, and print are **three equal mediums of the same
  Brief** [Sprint 05], not a primary plus degraded fallbacks.
- Use relative units and the `--measure` token; let the single column be
  the default and opt **into** side-by-side only at wide breakpoints.

**Current Conformance.**

- Partial/met: a single fluid column at `max-width: 46rem` works across
  widths; there is no side-by-side treatment (acceptable — it is a
  wide-only option, not a requirement) and no manifest yet to truncate.
  Revisit once the manifest (§2.3) lands. → tracked with **CG-2**.

---

## 7. Print / PDF

**Normative.**

- Print/PDF is an **equal medium of the same Brief** [Sprint 05], not an
  afterthought. **Grammar and order are identical to screen**
  [Law VI]; only furniture changes.
- What changes in print: **lighting resolves to daylight** (the page
  prints on white in the day ink) [C§11 spirit], and **chrome is
  removed** (masthead, nav, service line, other `no-print` affordances).
- What must not change: the **bounds, certification line, manifest,
  measurement, reserved seat, and end mark** all print [§2]. The edition
  must be **recognisably the same document** on paper.

**Implementation Notes.**

- Today's `@media print` block resolves `color-scheme: light`, hides
  `.no-print`, and drops the sheet border/shadow — a good base. When the
  bounds and end mark (§2.1, §2.6) are added, ensure they are **not**
  marked `no-print`.
- `printbrief.html` is the print rendering of the Brief room (§4.1) and
  must track its device set.

**Current Conformance.**

- Partial: the print stylesheet exists and behaves correctly for
  today's page; it will need the new devices included as they land.
  → tracked with **CG-1**, **CG-6**.

---

## 8. Accessibility

Except for the first clause, this section's specific targets are
**engineering standards introduced by the Design System**, not
derivations from the Charter (which is silent on accessibility metrics).
They are chosen as sound, conventional baselines; the Charter's "no
third voice" already does much of the work by forbidding colour-coded
meaning.

**Normative.**

- **No meaning is carried by colour alone** [Law III already guarantees
  this by forbidding a third voice; restated as an accessibility
  requirement]. Every distinction a reader must perceive — including
  trust labelling (§3.4) and prominence (§2.5, carried by size + printed
  count) — is legible in monochrome.
- Text meets at least **WCAG 2.1 AA contrast** (4.5:1 body, 3:1 large
  text) in **both** lightings. AA is the Design System's baseline;
  every `light-dark()` pair must pass against its surface.
- Markup is **semantic and in reading order**: the DOM order matches the
  edition order (§4.1) so a screen reader and a sighted reader receive
  the same sequence, including the manifest **before** the stories
  (§2.3).
- Every interactive control has a **visible, non-colour-only focus
  state** and is keyboard-operable.
- Motion respects **`prefers-reduced-motion`** (§5.4); Orven has no
  essential motion.
- The **printed count** beside each story (§2.5) is real text, available
  to assistive tech, not an aria-hidden decoration.

**Implementation Notes.**

- Contrast has now been **measured** (Draft 2) with the sRGB WCAG 2.1
  formula; results are recorded below. When a token value changes,
  re-measure — the pairs are owner-revisable furniture (§1.1) and a new
  value re-opens the question.
- Ensure the third-party trust cue (§3.4) does not depend on the accent
  colour alone (the dashed border satisfies this today — keep it).

**Current Conformance (measured, Draft 2).**

Contrast of every text token against both reading surfaces (`--sheet`,
`--paper`), each lighting, WCAG 2.1 (4.5:1 body / 3:1 large):

| Token | Light on sheet / paper | Dark on sheet / paper | Verdict |
|---|---|---|---|
| `--ink` | 14.1 / 13.0 | 10.1 / 10.9 | AA-body, both lightings |
| `--mid` | 5.3 / 4.9 | 6.1 / 6.6 | AA-body, both lightings |
| `--accent` | 7.4 / 6.8 | 6.3 / 6.8 | AA-body, both lightings |
| `--warn-ink` | 8.3 (sheet), 7.4 (own bg) | 6.8 | AA-body, both lightings |
| `--faint` | **3.6 / 3.3** | 4.8 / 5.2 | **light: AA-large only** |

- **One measured failure: `--faint` in light mode.** At 3.6 / 3.3 it
  clears AA-large (3:1) but **fails AA-body (4.5:1)** — and `--faint`
  carries small apparatus copy well under the large-text threshold (help
  text at `.82rem`, captions, freshness, coverage, table headers). In
  dark mode `--faint` passes (4.8 / 5.2). This is a genuine AA gap for
  light-mode apparatus text. → **CG-11**.
  - *The tension the fix exposes.* Darkening `--faint` (light) to reach
    4.5:1 on both surfaces lands it at roughly `#726d5d`, essentially the
    current `--mid` (`#6f6a5e`) — collapsing the two-level apparatus
    quietness in light mode. So the resolution is a furniture tradeoff
    with an accessibility floor, not a free edit: either accept a
    quieter-but-distinct faint that meets only AA-large for genuinely
    secondary text, or darken faint toward mid and lose one level of
    apparatus quiet. Both are owner-revisable furniture (§1.1, OD-7); the
    Design System records the measured floor and the tradeoff and leaves
    the palette choice to the owner.
- **Non-text: `--rule`.** Hairlines measure ~1.4–1.5:1 against the
  surface — appropriate for decorative structural rules, but note that
  form-field and tab boundaries drawn in `--rule` fall below the 3:1 of
  WCAG 1.4.11 for meaningful UI boundaries. Recorded as an observation
  under CG-11; the trust cues that *must* be legible do not rely on
  `--rule` (§3.4 uses border-style and accent).
- **Focus states — still unspecified.** No focus styling is declared in
  [`style.css`](../internal/core/static/style.css); controls fall back to
  browser defaults. A visible, non-colour-only focus ring remains an
  open CG-11 item.

---

## 9. Conformance & change control

### 9.1 How to classify a new decision

Before building any reader-facing change, label it by layer using the
Charter's **placement test**: *could a different designer, given only
the layers above the one you are working in, produce a different-looking
version that is still unmistakably Orven?* If yes, the decision belongs
below that line [Charter, "The three layers"].

- Touching a **law** → not a design decision; it requires a Charter
  amendment ratified by the owner. Stop and surface.
- Touching **grammar** → an owner amendment; propose, do not
  self-approve.
- Touching **furniture** → free within the laws; proceed, keeping to the
  tokens (§1) and this document's furniture specs (§3).

A change that would introduce a colour of meaning, a third voice, an
importance ranking, remediation language, or a live/infinite surface is
**never** furniture — it contradicts a law and is out of scope by
definition [Law I, II, III; C§1, C§8].

### 9.2 Owner-defined parameters (surfaced, not invented)

Each item below is a value the Charter leaves to the owner. This
document specifies the mechanism around it and **must not** fill it in.
Implementation blocks on these where noted.

| ID | Parameter | Mechanism specified | Blocks |
|---|---|---|---|
| OD-1 | Number of prominence steps (`T`) | Ordered `--tier-1…T` scale, §1.1/§2.5 | Measurement (§2.5) |
| OD-2 | Count thresholds for `tier(count)` | Pure, daily-identical, monotonic, total function, §2.5 | Measurement (§2.5) |
| OD-3 | Edition-number scheme | Stable non-ranking identifier in the certification line, §2.2 | Certification line (§2.2) |
| OD-4 | End-mark wording | Explicit end mark closing the document, §2.6 | End mark (§2.6) |
| OD-5 | Manifest truncation on small media | Revisable furniture rule, §2.3/§6 | Manifest on mobile |
| OD-6 | Whether per-source counts are retrofitted into today's Brief now | The Law II defense (§2.5); a timing decision | Brief reconciliation |
| OD-7 | Any future change to faces/palette | Token roles fixed (§1.1); values are revisable furniture | — (default in place) |

### 9.3 Conformance register (implementation debt)

Divergences between today's frontend and this specification. These are
**implementation debt, not design decisions** — the standard is the
Charter-faithful target above; the code is expected to move toward it.
Sequencing and prioritisation are product decisions for the owner.

| ID | Gap | Spec | Severity (law weight) |
|---|---|---|---|
| CG-1 | Bounds do not close (no closing double rule) | §2.1 | Law I/V grammar |
| CG-2 | Manifest placed at foot, not up front | §2.3 | Law II/V grammar |
| CG-3 | Size-as-measurement absent (uniform stories) | §2.5 | Law II — load-bearing |
| CG-4 | Per-source counts not printed beside stories | §2.5 | Law II defense |
| CG-5 | Certification line lacks edition number + accounting/coverage window | §2.2 | Law V |
| CG-6 | No explicit end mark | §2.6 | Law I/V grammar |
| CG-7 | Two inks: apparatus sans not implemented; no voice inversion | §2.4 | Law III grammar |
| CG-8 | Silence not in an enclosure (centred italic, not a box) | §2.6 | Law IV grammar |
| CG-9 | **Audited (Draft 2).** Brief clean of `--warn-*` (load-bearing check passes); residual: warn colour spent on action-room error/failure states and one list-status marking, outside the "before-proceeding consequence" fence | §1.1/§3.7/§9.4 | Law III — rule settled; residual is presentational debt |
| CG-10 | Type/size/spacing not tokenised (inline values) | §1.1 | Token discipline |
| CG-11 | **Measured (Draft 2).** `--faint` fails AA-body in light mode (3.6/3.3; AA-large only); all other text tokens pass AA in both lightings; focus states still unspecified | §8 | Accessibility |
| CG-12 | Reader's changes applied but not recorded as an attributed, dated fact effective at the next edition ("Paused by you, 12 July") | §4.4 | Law VIII |
| CG-13 | A beat's record shows credentials as relative times + a runs table, not as the counts-and-dates record ("filed in *n* of *m* editions") and dated sentences | §4.3 | Law VII/IV |
| CG-14 | The evaluative health-state word "Healthy" should read as a factual state, "Reporting" (`engine.go`) — a one-string implementation change | §4.3/§9.5 | Law VII |

All five reader-facing rooms have now received a conformance pass (Draft
2): the Brief (§4.1), the Archive (§4.2), a beat's record (§4.3), the
action rooms (§4.4), and the quiet edition (§4.5), plus the loading
state (§5.4). No room remains unassessed. Detailed room-by-room
*template specs* beyond the essence-and-devices treatment in §4 are a
furniture-level elaboration that can follow the conformance-debt work;
the conformance status of every room is now recorded, not deferred.

### 9.4 The warn-colour ruling (CG-9), settled 23 July 2026

The question of whether a muted warn colour is a Law III third voice or
permitted action-room furniture was settled by owner ruling. The rule
(specified in §1.1): a soft warning colour is permitted **only in
action-oriented areas outside the Brief**, and **only** to convey a
**genuine risk or consequence before proceeding** — a destructive
action, a security consequence, a trust acknowledgement, or an
irreversible decision. It must never appear in the Brief, never mark
routine status, never manufacture urgency for an ordinary condition, and
never become a general third visual voice; its use stays rare and
confined to those cases. The reader's report remains governed by "no
third voice" [Law III].

**Audit outcome (Draft 2).** The CG-9 audit has been run. The
load-bearing check **passes**: the Brief and its print rendering carry no
`--warn-*` usage — the reader's report is clean of the soft-alarm
channel. Inside the fence, correctly: the install "no permissions"
caution, the Settings "no sign-in protection" security notice, the
backups/restored "credentials will not be included" consequence, the
manual-files deletion warning, and the restore/uninstall acknowledgement
labels — each a genuine before-proceeding risk, consequence, or trust
acknowledgement in an action room. Outside the fence (residual CG-9
debt): the warn colour is also spent on **error/failure states** (a
repository fetch error, a plugin load error, an engine-incompatibility
notice, an install error) and on **one list-status marking** (a failed
plugin's load error in the Installed list). Errors are governed by "no
alarm" (§5.6) and are not "genuine risk or consequence before
proceeding"; the fix is to state them in the neutral apparatus voice and
reserve `--warn-*` for the before-proceeding cases. This is presentation
debt, not a design question — the wording is already conformant.

### 9.5 Law-level items surfaced from the conformance passes

Per §0.2 and §9.1, an implementation ambiguity that touches a **law** is
surfaced for owner review rather than resolved inside this office. The
Draft 2 conformance passes surfaced one such item; the owner delegated
the terminology judgment to this office, which sets the standard here
and leaves enactment to implementation (CG-14).

- **SR-1 — Is a derived "Healthy" state a Law VII grade? — ruled at the
  standard.** The beat's record and the plugin lists render a source's
  `Health` as a single word (§4.3). Every value but one is a neutral
  state fact about the last run or configuration ("Disabled", "Ready",
  "Partial data", "Source unavailable", "Authentication failed", "Timed
  out", "Failed") — clearly Law VII-compatible. The value **"Healthy"**
  is evaluative: Orven characterising its own witness's condition, where
  Law VII holds that about its own operation Orven speaks "only in counts
  and dates … never a score, streak, or grade" and "does not rate its
  witnesses." Because the question touched a law it was surfaced; the
  owner delegated the wording judgment to this office. **The standard:**
  the last-run-succeeded state must read as a factual state, not a
  rating — **"Reporting"** (what the source *is doing*), parallel to the
  present-tense "Running"/"Waiting" and to the factual outcome states,
  with no evaluative charge. "Filed" was the considered newspaper-native
  alternative; "Reporting" was chosen for unambiguous legibility to a
  self-hosting operator. **Scope note:** this office specifies the
  wording; it does not change production code. The edit itself — a
  one-string change in [`engine.go`](../internal/engine/engine.go) — is
  implementation work, recorded as **CG-14** and applied when
  implementation reaches this area, alongside the frontend/backend work.

---

**End of the Design System, Draft 2.** It translates 8 laws · 6 devices
· 5 rooms into token, component, template, and state specifications;
assesses **every** reader-facing room for conformance; marks 7
owner-defined parameters; records 14 conformance gaps (CG-1…CG-14) as the
roadmap to full Charter conformance; carries the owner's settled
warn-colour ruling and its executed audit (§1.1, §9.4); records the
measured contrast results (§8); and rules one surfaced law-level item at
the level of the standard, leaving its enactment to implementation (§9.5,
SR-1 → CG-14). It adds no design philosophy of its own, invents no
owner-reserved value, resolves nothing that belongs to the owner, and
changes no production code — it specifies the standard, and the code is
expected to move toward it.
