# `orven validate` — Requirements

This is the specification for the plugin validator. It is written down
so the checks are implemented from an agreed spec, and so contributors
know exactly what will be checked before they submit. The validator is
also the future CI gate for the community plugin repository.

Usage:

```bash
go run ./cmd/orven validate ./plugins/my-plugin
```

(or `orven validate <dir>...` with a built binary; exit code 0 = no
errors, 1 = errors, and warnings are listed either way)

The validator runs the plugin against its own fixtures via the real
engine runner — executing it on the alphabetically first file in
`fixtures/`, and scanning the rest only for credential-shaped
content — then inspects the manifest and output.

## Severity model

- **Error** — the plugin violates the contract or cannot run; validation
  fails.
- **Warning** — the plugin works but departs from house style or best
  practice; validation passes with the issues listed. CI may choose to
  require zero warnings for the default repository.

Every finding must name the file/field it refers to and say how to fix
it. The validator helps contributors succeed; it never merely rejects.

## Errors (fail validation)

1. `plugin.yaml` missing or unparsable; missing `id`, `name`,
   `version`, or `entrypoint`; `schema_version` unknown.
2. `engine.min_contract` newer than the validating engine.
3. Config schema fields with unknown `type`, duplicate keys, or
   defaults that don't match their declared type.
4. Entrypoint fails to start, exits non-zero on fixture input, or
   produces output that is not a single JSON object.
5. Output `contract_version` missing; `status` not one of the six
   contract statuses.
6. Observations missing `title`; unknown `scope` value (only `event`,
   `state`, or absent are legal).
7. Declared `timeout`, or collection intervals that don't parse, or
   `min_interval` > `max_interval`.
8. Secret leakage: any configured secret value (from the test input)
   appearing verbatim in summary, observations, or error text.
9. Forbidden voice in output: recommendation or remediation language
   ("you should", "we recommend", "try restarting", "to fix",
   "consider", "please"). Facts only. The list is maintained in the
   validator (`internal/validate/validate.go`) and favors precision
   over recall: factual past-tense wording ("the container restarted
   4 times") must never be flagged, so only clearly advisory or
   imperative forms are matched.

10. Credential-shaped content in output regardless of value:
    Authorization/API-key header fragments or credential-bearing query
    parameters (`api_key=`, `token=`, `password=`, …). The runtime
    sanitizer would redact these; a plugin should never emit them in
    the first place. Shares its patterns with the engine's sanitizer
    so the two can't drift apart.

## Warnings

11. Missing `README.md`, `fixtures/`, or `tests/`.
12. Fixture files containing credential-shaped content — never commit
    real credentials to fixtures; invent obviously fake values that
    don't use header or query-parameter shapes.
13. Missing `collection.freshness` declaration (the engine will guess
    2× the recommended interval).
14. No permissions declared (every plugin touches *something*; say so).
15. `summary` longer than one sentence, or absent on a non-`ok` status
    (failure results should explain themselves).
16. Observation `kind` outside `fact | count | change | notice` —
    kind is advisory metadata with no engine behavior, and the
    vocabulary is bounded so it cannot drift before it ever gains
    semantics.

### Observation-title house style (warnings)

Titles are the headlines of short news briefs: sentence case, concise,
factual, no trailing period, detail in the body. The validator checks
what can be checked *safely*:

17. **Trailing period or exclamation mark.** Show the corrected title
    with the punctuation removed (punctuation-only fix — safe to
    suggest verbatim).
18. **Obvious title casing.** Heuristic: three or more words, and most
    non-leading words of four or more letters begin with a capital
    ("2 New Requests Awaiting Approval"). Show a suggested sentence-case
    version produced by *capitalization changes only*: lowercase
    non-leading words, but leave fully-uppercase tokens (RAID, GB,
    S02E04) untouched. Mark the suggestion "verify proper nouns" —
    the validator cannot distinguish "The Bear" from "the queue".
19. **All-caps titles.** Warn. Do not auto-suggest here: acronyms make
    a mechanical lowercase unsafe.
20. **Excessive length** (over ~60 characters): suggest moving the
    specifics into `body`. No rewrite offered.

### Hard boundaries for the title checks

- The validator **never rewrites** plugin output — suggestions are
  display-only examples in the report.
- It **never attempts headlinese** (dropping "is", "the", …); removing
  words can change meaning, and meaning belongs to the plugin author.
- Where a suggestion is shown, it must differ from the original by
  **capitalization or trailing punctuation only**.
- Style checks apply to `title` only; `body` and `summary` are ordinary
  sentences and are not style-checked beyond the voice rules (9).

## Non-goals

- Judging `event` vs `state` correctness — a domain decision
  (see PLUGIN_SDK.md, "The deciding question").
- Linting the plugin's source code. Any language is welcome; only the
  manifest and the observable contract behavior are validated.
