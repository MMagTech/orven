# The Orven Design Charter

**Ratified by the product owner on 23 July 2026.** This is the
governing document for Orven's design language: every change to
anything a reader sees is made inside it. `docs/CONSTRAINTS.md` says
what Orven *is*; this charter says how Orven *looks and reads*, and
why.

**About this file.** The charter was ratified as a designed artifact —
[`docs/design-studio/orven-design-charter.html`](design-studio/orven-design-charter.html),
which is set in the charter's own language and remains the ratified
original. This file is its verbatim transcription, kept in Markdown so
the charter is diffable, reviewable, and citable like every other
governing document. The two carry the same text; an amendment changes
both. The full record of how the charter was discovered — eight design
sprints, concluded by an adversarial review — is preserved in
[`docs/design-studio/`](design-studio/NOTES.md).

**How to use it.** Label every design decision by layer: **law** /
**grammar** / **furniture** (defined below). Laws and grammar change
only by amendment — ratified by the product owner, arising from real
implementation experience, never resolved silently inside
implementation work. Ambiguities and conflicts are surfaced for owner
review. The owner's recorded expectation: Law II is the law most
likely to be revisited after real-world use.

---

*The laws, grammar, and commentary of Orven's design language — set in
that language, in its day ink or its night ink, as your system
prefers.*

## Preamble — why Orven exists

A self-hosted machine is a quiet responsibility. It works through the
night — backing up, serving, renewing, recording — and the person
responsible for it carries a low, constant question: *is everything
all right?*

Nearly every tool built to answer that question does so by asking for
vigilance. Dashboards want watching. Alerts want reacting. Graphs want
interpreting. Each of them, however well made, converts a quiet
responsibility into a standing anxiety — and none of them will ever
tell you that you may stop looking.

Orven is built on the opposite promise. Once a day it compiles what
its plugins observed into a single finite brief: what happened, what
changed, and what it could not see. It reports in plain sentences,
ranks nothing, recommends nothing, and ends. The reader — the
operator, the only person with the judgment to decide what matters on
their own machine — reads it in minutes and puts it down.

The promise, in one sentence: **Orven observes so that you may stop
watching — and earns that trust by never once telling you what to
think.**

Everything below exists to protect that promise. The laws are not
aesthetic preferences; each one guards a place where the promise could
quietly break.

## The three layers

Every design decision in Orven sits at one of three altitudes.
Confusing them is how design languages rot.

- **Laws** — What makes something Orven at all. Violating one produces
  a different product. Never change.
- **Grammar** — Orven's chosen expressions of the laws — the
  recognizable devices. Change rarely and deliberately.
- **Furniture** — Per-room, per-medium arrangement. Free to vary
  whenever a better arrangement is found.

*The placement test: could a different designer, given only the layers
above, produce a different-looking version that is still unmistakably
Orven? If yes, it belongs below the line they worked from. If a page
must look exactly one way to feel right, furniture has been mistaken
for law.*

## The laws — as amended after adversarial review

### I. The edition, not the feed

Every Orven document is finite, dated, and numbered, with a beginning
and a declared end. Nothing is infinite, live, or asking the reader
back early.

> **Commentary — from the heresy of the Current Page.** *A page that
> can be stale makes the reader responsible for refreshing it; a page
> you must keep checking is an alert stream with good typography. The
> edition is the license to not look. That license is the product.*

### II. Prominence is arithmetic

Whatever form prominence takes, it derives from countable fact — never
from judged importance — and the count is printed beside the claim, so
the reader can audit the layout every morning.

> **Commentary — from the heresy of the Even Page.** *The printed
> count is not decoration; it is the defense. Verbosity becomes
> visible as verbosity ("Containers · 9 observations" names who is
> talking a lot). A perfectly flat page is what the tiers become when
> counts are equal — it is the language's floor, not a second grammar,
> and it is not offered as a mode. Orven has one visual language.*

### III. Two voices, one of them quiet

The report speaks in full declarative sentences; the apparatus labels
in a smaller, plainer voice. No third voice — no icons, badges, or
alarm color — ever speaks.

> **Commentary — from the heresy of the Red Ink.** *Even one
> restrained accent, assigned by the fairest mechanical rule, is a
> needs-attention classification performed at design time — and within
> a week the reader scans for it instead of reading. The moment there
> is a third voice, it becomes the only voice anyone hears.*

### IV. Silence has a reserved seat

Absence of data is stated, structurally, wherever it occurs — in the
brief, in history, in a source's record. It is never a footnote and
never an implied all-clear.

### V. The frame certifies

A fixed apparatus — identity, date, number, and accounting — appears
identically on the heaviest and emptiest days. The accounting states
the edition's own arithmetic and its coverage window ("47 observations
from 15 of 17 plugins · observations through 6:45"), so the document's
completeness and its staleness are both declared, never discovered.

> **Amended at ratification.** *The coverage-window clause is the scar
> left by the Current Page heresy: the edition answers its strongest
> rival by being honest about what it does not yet know.*

### VI. Content breathes; structure doesn't

Density expands and contracts with the day; order and apparatus never
do. Position carries no verdict. The same grammar holds on a desktop,
a phone, and a printed page — geometry flexes, grammar does not.

### VII. Facts pass through; the record stays long

Orven reports what sources say in whatever form they say it, ratios
and percentages included. But about its own operation — coverage,
filings, gaps — Orven speaks only in counts and dates: "filed in 213
of 217 editions," never "98.2 percent," never a score, streak, or
grade. **Orven does not rate its witnesses.**

> **Rewritten at ratification — the successful heresy.** *The original
> law ("counts, never grades") could not distinguish a fact a source
> reported about the world from Orven evaluating its own sources. The
> line is who is being measured. A self-grade's only purpose is to
> invite "is that good?" — and the road from 98.2 percent to a letter
> grade took one design meeting.*

### VIII. The reader's actions are facts too

When the reader changes something, Orven records it as it records
everything: attributed, dated, stated plainly — "Paused by you,
12 July" — taking effect at the next edition, with the page saying so.
Even the instrument's room is written, and nothing in it advises.

## The grammar — Orven's chosen devices

Six devices express the laws. They are the identity — another product
could obey the same laws with different grammar; this grammar is what
makes Orven look like Orven. Each names its revisable part, so future
change is deliberate rather than drift.

### The bounds

A double rule opens every document and a double rule closes it. The
edition lives between them; nothing outside them exists. (Laws I, V.)

*Revisable — that the mark is a double rule. The boundary is law; the
rule style is signature.*

### The certification line

Identity · date · edition number · accounting, one quiet line at the
head of every document, identical every day. (Law V.)

*Revisable — composition and position; never presence or constancy.*

### The manifest

The whole before the parts: every source and its count, up front. On
quiet days it may name the sources instead, because presence is the
day's only news. (Laws II, V.)

*Revisable — its form, and its truncation rules on small media.*

### Two inks

A reading serif for the report; a quiet sans for the apparatus. In
rooms where the reader acts, the voices invert — the sans leads, the
serif explains. Day and night are the same page in different colors,
and colors only. (Law III.)

*Revisable — the specific faces and palettes. Georgia-class,
Segoe-class, the warm day paper, and the blue-black night are today's
choices, not the identity.*

### Size as measurement

A small fixed scale of prominence steps assigned by count alone —
bigger type means more happened, never that it matters more. Body text
is one readable column, ragged right; parallel stories sit side by
side only where the medium is wide. (Laws II, VI.)

*Revisable — the number of steps and the tier thresholds.*

### The reserved seat & the end

Silence sits in the design's only enclosure. The document closes with
an explicit end mark and its own accounting. (Laws I, IV, V.)

*Revisable — the box's placement per room; the end mark's wording.*

## The rooms — law extracted from furniture

What each room fundamentally is. Everything else about a room —
layouts, groupings, controls — is furniture, free to vary within the
laws.

- **The Brief** — The product itself: one finite edition of what
  happened, what changed, and what could not be seen, composed by the
  six devices, readable in minutes.
- **The Archive** — The unbroken chronological record of editions,
  every one present with equal standing — a quiet day is a full
  citizen. The archive states its own extent. No aggregation that
  ranks days; no streaks.
- **A beat's record** — The public record of a reporting source: its
  credentials as counts and dates, its record as dated sentences, and
  its silences seated chronologically within that record.
- **Settings** — The room where the reader acts. The identity is the
  writing, not the widgetry: every choice explained in a sentence of
  consequence, every change an attributed dated fact, effective at the
  next edition. Conventional controls are permitted — friction is a
  cost paid by a person.
- **The quiet edition** — The full frame around a stated nothing. The
  apparatus certifies the quiet; the whitespace is the news.

## Process record

This charter was discovered, not decreed, across eight design sprints:
eight divergent concepts (01); load tests at 21 and 47 observations
(02, 03); portability proofs in three non-newspaper languages (04); an
element-by-element distillation against three equal mediums — desktop,
mobile, print (05); extension to the secondary rooms (06); the
separation of law, grammar, and furniture (07); and an adversarial
review in which four steelman heresies were built against the laws
(08). Two heresies failed and strengthened the laws they attacked; one
was subordinated by ruling; one succeeded and forced Law VII's
rewriting. A charter that has corrected itself under attack can be
trusted further than one that has only been admired.

## Ratification record

Ratified by the product owner on 23 July 2026, with two observations
entered into the record. First, an editorial change accepted at
ratification: the preamble's promise now reads "Orven *observes*"
rather than "Orven watches" — the platform's own verb, and the sharper
one, since observing without watching is precisely the distinction
Orven exists to make. Second, the owner's recorded expectation:
Law II, prominence by arithmetic, is the law most likely to be
revisited once the product has lived with real data. That uncertainty
was judged no reason to delay — constitutions are amended through
experience, not speculation.

---

**End of the charter** — 8 laws · 6 devices · 5 rooms · discovered
over 8 sprints · ratified by the product owner, 23 July 2026
