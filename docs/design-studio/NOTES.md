# Orven Product Design Studio — The Historical Record

Design-language discovery for Orven, conducted as a dedicated design
conversation (no implementation). Session date: 23 July 2026.
Eight sprints, concluded by the ratification of the Orven Design Charter.

**Outcome: the Orven Design Charter — ratified by the product owner on
23 July 2026.** The charter (`orven-design-charter.html` in this package)
is the governing document; these notes are the record of how it was
discovered.

---

## The arc in one paragraph

Eight divergent concepts (01) produced a gut winner, the Broadsheet, and a
runner-up, the Observatory. Load tests at 21 and then 47 observations
(02, 03) proved the Broadsheet could scale on a mechanical rule —
prominence by volume, never importance. Portability proofs (04) extracted
six principles and showed them working in three non-newspaper languages,
proving the principles, not the costume, were the identity. A tribunal
(05) tried every newspaper element individually — keeping grammar,
dropping paper geometry — and proved the distilled result across desktop,
mobile, and print as three equal mediums. The grammar was extended to the
secondary rooms (06); the owner caught the drift from language into
application, prompting the separation of law, grammar, and furniture (07).
An adversarial review (08) attacked the laws with four steelman heresies —
two failed and strengthened their targets, one was subordinated by ruling,
and one succeeded, forcing Law VII's rewriting. The corrected charter was
then consolidated, given a preamble, and ratified.

## Sprint-by-sprint record

### Sprint 01 — Eight Briefs (`sprint-01-eight-briefs.html`)
Eight fundamentally different concepts, same morning (edition No. 214).

| # | Concept | Verdict |
|---|---------|---------|
| 1 | The Broadsheet — morning newspaper | **Kept — first choice.** "The one I'd want to see 10 years from now." |
| 2 | The Reading Room — pure prose | Scratched — no improvement over the current minimal UI |
| 3 | The Ledger — ruled daybook | Scratched — too close to current |
| 4 | The Wire — amber teletype | Scratched — "ancient computer" look |
| 5 | The Almanac — daily register | Scratched |
| 6 | The Observatory — night log, time rail | Kept as second (calming colors, structure); scratched in Sprint 02 review |
| 7 | The Minutes — memorandum | Not selected |
| 8 | The Forecast — shipping-forecast litany | Not selected |

Owner's concern raised here — how does the UI cope with many plugins? —
drove Sprints 02–03.

### Sprint 02 — The Heavy Morning (`sprint-02-scaling-study.html`)
Introduced the governing rule: **prominence follows volume, never
importance** (space assigned by observation count; arithmetic the reader
can audit). Studies: Broadsheet at 21 observations (liked; owner asked to
double the load); Broadsheet quiet day; Observatory at load (scratched);
Broadsheet/Observatory hybrid (scratched). Owner on the volume rule:
"not sure about honest, but I like it."

### Sprint 03 — The Frantic Edition (`sprint-03-frantic-edition.html`)
The Broadsheet at 17 plugins / 47 observations on a genuinely bad night
(failed offsite backup, container restart loop ending stopped, hot disk,
two silent plugins) — no red, no badges. Mechanism: **headline size is a
unit of measurement**, three tiers assigned purely by count. Honest
ceiling: ~60–70 observations before bands repeat and the page grows
taller. Owner then pivoted to principles extraction before answering this
sprint's review questions.

### Sprint 04 — The Principles, Proven (`sprint-04-principles-proof.html`)
Six principles extracted and proven portable by rendering the same
frantic night as an engineering Drawing, a concert Programme (silent
plugins marked *tacet*), and a railway Timetable — zero newspaper DNA,
all six principles obeyed. Conclusion: the principles are the identity;
"newspaper" was one garment.

### Sprint 05 — The Distillation (`sprint-05-distillation.html`)
Owner's directive: strip newspaper imitation, keep timeless editorial
principle; desktop, mobile, and print/PDF are **three equally important
consumptions of the same Brief**; the resemblance must be "earned, not
imitated." The tribunal tried each element (pattern: *what survives is
grammar, what falls is geometry*): kept edition number/dateline, tiered
headlines, boxed silence, the end mark; transformed masthead (quiet
wordmark), "Inside" (→ the manifest), rules (only the bounding double
rules), serif (→ two inks: Georgia-class report, sans apparatus), kickers,
paper tint; dropped snaking columns, justified text, italic decks.
Result: the six-device grammar, proven in all three mediums.

### Sprint 06 — The Other Rooms (`sprint-06-other-rooms.html`)
The grammar extended to five rooms: Archive (editions as certification
lines), a beat's record (credentials + record + gaps in place), Settings
(two-inks inverted; "Paused by you, 12 July"), the quiet edition (manifest
names beats — presence is the news), the night ink (Observatory-derived
blue-black; colors only). Owner: successful, but flagged drift from
discovering the language into designing the application.

### Sprint 07 — The Charter, Draft 1 (`sprint-07-charter.html`)
The three-layer model: **Laws** (never change) / **Grammar** (rarely,
deliberately) / **Furniture** (free), with the placement test ("could a
different designer, given only the layers above, build a different version
still unmistakably Orven?"). Eight laws (six from Sprint 04 + VII "counts,
never grades" + VIII "the reader's actions are facts too"). The owner's
three room questions answered: Archive = chronological record, not
one-row-per-edition; Beat page = public record of a source, not a
newspaper section; Settings = the room where the reader acts — the
writing carries the identity, conventional controls permitted.

### Sprint 08 — The Heresies (`sprint-08-heresies.html`)
Adversarial review: assume every law is wrong; four steelman
counter-languages, mocked up at full strength.

| Heresy | Target | Verdict |
|---|---|---|
| The Current Page (living document, no editions) | Law I | **Failed** — a page that can be stale demands refreshing; "the edition is the license to not look." Scar kept: the edition declares its coverage window (folded into Law V) |
| The Even Page (perfect flatness, no tiers) | Law II | **Partial** — legitimate rival; exposed verbosity-gaming; the printed count promoted to load-bearing defense. Owner's ruling: subordinated — one visual language, flatness is the floor when counts are equal, not a mode |
| The Red Ink (bookkeeper's red for completion-state facts) | Law III | **Failed instructively** — choosing which facts get the channel is design-time classification; within a week the reader scans instead of reading. Kept as charter commentary |
| The Honest Fraction (grades are just arithmetic) | Law VII | **Succeeded** — caught our own mockups ("13 percent blocked" vs "98.2% coverage"); the line is who is being measured. Law VII rewritten: "Facts pass through; the record stays long. Orven does not rate its witnesses." |

Owner's assessment: "The exercise succeeded because it found a real
weakness instead of simply defending the existing design."

## Ratification — 23 July 2026

The consolidated charter (preamble → layers → laws → grammar → rooms →
process record) was reviewed in full and **ratified**, with two
observations entered into the record:

1. **Editorial (accepted):** the preamble's promise changed from "Orven
   watches" to "Orven **observes** so that you may stop watching" — the
   platform's own verb, and the sharper one: observing without watching is
   precisely Orven's distinction.
2. **Expectation (recorded):** Law II (prominence is arithmetic) is the
   law most likely to be revisited after real-world use. Judged no reason
   to delay — "constitutions are amended through experience, not
   speculation."

## What happens next

The Design Studio's objective is met. Implementation belongs to a separate
conversation, governed by the ratified charter. Future design work should
label every decision by layer (law / grammar / furniture); amendments to
laws follow experience with real data, with Law II the watched candidate.

## Package contents

- `sprint-01-eight-briefs.html` … `sprint-08-heresies.html` — the eight sprints, self-contained
- `orven-design-charter.html` — **the ratified Design Charter** (opens in any browser; day/night ink follows the system; prints cleanly to PDF)
- `NOTES.md` — this record

## Artifact links (live, private)

- Sprint 01: https://claude.ai/code/artifact/184196b4-914b-422a-9205-5c66009f28bf
- Sprint 02: https://claude.ai/code/artifact/bea96cc5-f161-4d05-b905-7de3ac5c3f0e
- Sprint 03: https://claude.ai/code/artifact/864c74ad-fde9-4951-bb9c-6ec522664ae6
- Sprint 04: https://claude.ai/code/artifact/0ac7bc58-64de-4b0f-a29c-70f18c8cfcc9
- Sprint 05: https://claude.ai/code/artifact/2e965b6f-afe9-4057-8c36-56d6f22de175
- Sprint 06: https://claude.ai/code/artifact/4c644387-b466-4ee7-aa7e-8fc1482b50b5
- Sprint 07: https://claude.ai/code/artifact/3461722b-a63d-4afe-b3c5-850e105a0ef6
- Sprint 08: https://claude.ai/code/artifact/f250ddb7-5264-452a-9b5c-7dcdb7658c4f
- **The Design Charter (ratified):** https://claude.ai/code/artifact/1271be7c-ec41-4dac-ad0e-e0d5585fa9b7
