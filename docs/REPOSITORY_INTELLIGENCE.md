# Atlas Repository Intelligence Specification

This document is the **canonical specification for derived repository
intelligence** in Atlas. It defines every deterministic fact derived from
`RepositoryVestige` observations, its derivation formula, its observation
inputs, its edge-case handling, and its invariants.

It is the specification for **v0.8.16: Repository Intelligence**.

Where `OBSERVATION_SPECIFICATION.md` defines *observations* (what Atlas
acquires), `INTELLIGENCE.md` defines the *ontology* (what intelligence is),
and `ARCHITECTURE.md` defines *layer ownership*, this document defines the
*derivation contract* (how observations become facts). The three are
deliberately separate: acquisition evolves, ontology is stable, derivation is
now specified.

## Normative Authority

This document is the **normative source** for every `RepositoryFacts` field
added in v0.8.16. The code in `internal/facts/repository.go` is an
**implementation** of this specification.

If a derived fact's formula, input set, or edge-case handling changes, this
document **must be updated first**. The code is then updated to match. No code
change that alters a derivation is valid without a corresponding change to this
specification.

## Architectural Invariants

### No New Runtime Layer

Repository intelligence in v0.8.16 lives **only** as a flat expansion of
`facts.RepositoryFacts`. There is no new package, no new engine, no GraphQL
change, no acquisition change. Derived knowledge is a field on the existing
`RepositoryFacts` struct, computed in `facts.FromRepos`.

### Facts, Not Signals

Every value defined here is a **Fact**: a deterministic aggregate derivable
purely from observations, given a `referenceTime`. The four named indicators
(Ownership, Consistency, Depth, Activity) are unchanged. New ratio aggregates
(ForkRatio, LicensedRatio, etc.) are descriptive portfolio facts, *not*
indicators: they are not part of the four-value measurement model and do not
feed `indicators.ExtractSignals`. They are exposed for evidence.

### Determinism

Equivalent vestige state plus equivalent `referenceTime` must always produce
equivalent `RepositoryFacts` values. No clock dependence beyond the explicit
`referenceTime` parameter. No provider dependence beyond the normalized
vestiges.

### Vestige-Only Inputs

Every derived fact consumes only fields already present on `RepositoryVestige`
(after v0.8.15). v0.8.16 adds **no new observations**. It does not read
`UserProfile`, `Contributions.Summary`, acquisition DTOs, or HTTP. Any
intelligence requiring new observations (contribution windows, language union
across contributed repos, follower counts) is deferred.

### Flat Expansion Only

`RepositoryFacts` gains scalar fields. It does not gain nested sub-structs,
maps of arbitrary depth, or per-repo slices. `RankedLanguages` is the single
allowed ordered list (a deterministic ranking, not a nested object). This keeps
the fact struct flat, serializable, and inspectable.

---

## Observation Inputs

All derivations consume the following `RepositoryVestige` fields (already owned
by Atlas per `OBSERVATION_SPECIFICATION.md`):

| Domain | Fields consumed |
| --- | --- |
| Identity | `Archived`, `Template` |
| Ownership | `Fork`, `CollaboratorCount` |
| Timeline | `CreatedAt`, `ReleaseCount`, `LatestReleaseAt` |
| Technology | `License`, `Topics`, `DefaultBranchProtected`, `LanguageDistribution` |
| Maintenance | `Stars`, `OpenIssues`, `PullRequestCount`, `Forks`, `Watchers` |
| Structure | `Size`, `DiscussionEnabled` |

None of these require new acquisition. v0.8.16 is a pure derivation release.

---

## Derived Fact Catalogue

Each fact lists: its field name on `RepositoryFacts`, the vestige inputs, the
deterministic formula, its type/range, and edge-case handling. All formulas are
evaluated inside `FromRepos(repos, referenceTime)`.

### 1. Star Distribution

| Field | Type | Formula | Edge cases |
| --- | --- | --- | --- |
| `MaxRepoStars` | `int` | `max(repo.Stars)` over all repos | `0` if no repos |
| `MeanRepoStars` | `float64` | `TotalStars / TotalRepos` | `0` if `TotalRepos == 0` |

`MaxRepoStars` captures the candidate's single highest-profile repository
independent of portfolio size. `MeanRepoStars` normalizes total stardom by
volume, distinguishing a few popular repos from broadly popular ones.

### 2. Technology Breadth

| Field | Type | Formula | Edge cases |
| --- | --- | --- | --- |
| `LanguageCount` | `int` | count of distinct keys across all `LanguageDistribution` maps | `0` if no language data |
| `RankedLanguages` | `[]string` | union of language keys, each summed by bytes; sort by total bytes descending; tiebreak by language name ascending | empty slice if no language data |

`LanguageCount` is the **distinct** language universe of the portfolio (the
union across repos, not a per-repo count). `RankedLanguages` is the same union
ordered by aggregate byte share — a deterministic primary-to-secondary
language ranking.

**Ordering contract (must hold across Go versions):**

1. **Sum before order.** Byte counts are summed across all repos into a single
   per-language total *before* any comparison. No per-repo ordering is used.
2. **Primary key: byte total descending.** Higher aggregate bytes rank first.
3. **Tiebreak: language name ascending.** When two languages have identical
   byte totals, the lexicographically smaller name ranks first.
4. **Total order.** Because the tiebreak is a strict total order over language
   names (names are unique), the final ordering is fully determined and does
   **not** depend on `map` iteration order or `sort` implementation details.
5. **Stable and reproducible.** Equivalent vestige state always yields an
   identical `RankedLanguages` slice, on every Go toolchain.

This is the GitFut `languages` / `rankedLanguages` idea (signals.ts:18,22),
adapted: the *deterministic derivation* is kept, the Devicon-logo presentation
and FIFA `skillMoves` scale are rejected.

### 3. Release & Maintenance Aggregates

| Field | Type | Formula | Edge cases |
| --- | --- | --- | --- |
| `TotalReleases` | `int` | `sum(repo.ReleaseCount)` | `0` if none |
| `LatestReleaseAt` | `time.Time` | `max(repo.LatestReleaseAt)` over repos with releases | zero time if none |
| `ReleasedRepos` | `int` | count of repos with `ReleaseCount > 0` | `0` |
| `TotalPullRequests` | `int` | `sum(repo.PullRequestCount)` | `0` |
| `TotalCollaborators` | `int` | `sum(repo.CollaboratorCount)` | `0` |
| `ProtectedBranchRepos` | `int` | count of repos with `DefaultBranchProtected == true` | `0` |
| `DiscussionRepos` | `int` | count of repos with `DiscussionEnabled == true` | `0` |

These are pure summations/counts over fields already present after v0.8.15. They
answer "how much release, PR, and collaboration infrastructure does this
portfolio exhibit?" without interpretation.

### 4. Ratio Facts (deterministic, in [0, 1])

| Field | Type | Formula | Edge cases |
| --- | --- | --- | --- |
| `ForkRatio` | `float64` | `ForkRepos / TotalRepos` | `0` if `TotalRepos == 0` |
| `LicensedRatio` | `float64` | `LicensedRepos / TotalRepos` | `0` if `TotalRepos == 0` |
| `ArchivedRatio` | `float64` | `ArchivedRepos / TotalRepos` | `0` if `TotalRepos == 0` |
| `MeanRepoSize` | `float64` | `sum(repo.Size) / TotalRepos` | `0` if `TotalRepos == 0` |
| `TopicBreadth` | `float64` | `TotalTopics / TotalRepos` | `0` if `TotalRepos == 0` |

These normalize counts against portfolio size. `ForkRatio` answers "what share
of this portfolio is original work versus forks?" — the GitFut `fork` signal,
adapted to a ratio fact. `LicensedRatio` and `ArchivedRatio` characterize
portfolio hygiene. All are clamped to [0, 1] by construction (`numerator ≤
denominator`).

### 5. Freshness & Age (deterministic given `referenceTime`)

| Field | Type | Formula | Edge cases |
| --- | --- | --- | --- |
| `PortfolioAgeDays` | `int` | `referenceTime - OldestCreated`, in days | `0` if `OldestCreated` is zero |
| `NewestRepoAgeDays` | `int` | `referenceTime - NewestCreated`, in days | `0` if `NewestCreated` is zero |
| `DaysSinceLatestRelease` | `int` | `referenceTime - LatestReleaseAt`, in days | `0` if no release |

`PortfolioAgeDays` is the **PortfolioAge** concept; `NewestRepoAgeDays` is
**RepositoryFreshness** (how recently the candidate started something new).
`DaysSinceLatestRelease` is **ReleaseCadence** (recency of shipping). All age
values accept `referenceTime` explicitly — no system-clock dependence.

(`MeanRepoSize` and `TopicBreadth` are averages, grouped under Ratio Facts
above; they are ratio-class, not age-class.)

---

## Rejected Derivations (from GitFut research)

Prior GitFut research classified many GitFut concepts.
The following were evaluated and **rejected** for v0.8.16 with rationale:

| Concept | Reason |
| --- | --- |
| FIFA six-stat model (PAC/SHO/PAS/DRI/DEF/PHY) | Gamified 0–99 rating; not a fact |
| `archetypeFromShape` (Poacher/Regista/…) | Evaluation/Presentation; not a fact |
| `pickFinish` (bronze/silver/gold/toty/icon) | Evaluation/Presentation gamification |
| `legacyScore` / founder forced-overalls | Non-deterministic favouritism |
| `deriveStyle` activity-shape naming | Future behaviour facts; needs contribution windows |
| Language-logo resolution (Devicon CDN) | Presentation; not intelligence |
| `recent_contributions` / `recent_commits` / `active_days` | Realized via `ActivityFacts` (v0.8.17) |
| Language union across contributed repos | Future collaboration facts; requires new acquisition |
| `followers` / `account_age_years` | Lives on `UserMetadata`, not repository facts |

The *deterministic derivation technique* behind GitFut's log-scaling and
language ranking is acknowledged and adapted where it maps cleanly to a flat
fact (star distribution, language ranking). The *FIFA framing* is never ported.

---

## Out of Scope (Frozen Layers)

v0.8.16 changed **only** `RepositoryFacts` (flat fields) and `FromRepos`. The
following remain untouched in the v0.8.17 architectural baseline:

- `observations.RepositoryVestige` — no new observation fields
- `internal/acquisition` (REST/GraphQL/merge/normalize) — frozen
- `indicators.Signals` / `indicators.RawScore` — frozen; formulas stable
- `internal/evaluation`, `internal/engine`, `internal/projection` — frozen
- CLI version, server endpoints — frozen
- Speculative `ContributionFacts` / `TechnologyFacts` / `BehaviourFacts` /
  `CollaborationFacts` placeholders — removed in v0.8.17 verification pass;
  future domains are documented ownership boundaries, not exported types

---

## Fact Certification Checklist

For every fact added in v0.8.16, the following must hold before certification:

1. Derivable solely from `RepositoryVestige` fields present after v0.8.15.
2. Deterministic given `repos` and `referenceTime`.
3. Edge cases (empty input, zero denominators) handled explicitly.
4. Implemented in `FromRepos` with no new package or layer.
5. Exposed as a flat field on `RepositoryFacts` (only `RankedLanguages` is an
   ordered list; no nested structs or maps).
6. Not consumed by `ExtractSignals` (it remains the four named signals).
7. Covered by unit tests in `internal/signals/facts_test.go` (legacy) or
   `internal/facts/repository_test.go` (canonical).
8. Documented here as the normative source; no implementation-only heuristics.

---

## Historical Note

- **v0.8.16** — Repository Intelligence release. Established this specification
  as the normative source for derived `RepositoryFacts`. Expanded
  `RepositoryFacts` with flat deterministic aggregates: star distribution
  (`MaxRepoStars`, `MeanRepoStars`), technology breadth (`LanguageCount`,
  `RankedLanguages`), release/maintenance aggregates (`TotalReleases`,
  `LatestReleaseAt`, `ReleasedRepos`, `TotalPullRequests`, `TotalCollaborators`,
  `ProtectedBranchRepos`, `DiscussionRepos`), ratio facts (`ForkRatio`,
  `LicensedRatio`, `ArchivedRatio`), and freshness/age facts
  (`PortfolioAgeDays`, `NewestRepoAgeDays`, `DaysSinceLatestRelease`,
  `MeanRepoSize`, `TopicBreadth`). No new observations, no new signals, no new
  layers.
