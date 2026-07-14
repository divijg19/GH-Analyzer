# Atlas Repository Facts Specification

Every deterministic fact derived from repository observations — its derivation
formula, its observation inputs, its edge-case handling, and its invariants — is
defined by this specification.

`OBSERVATION_SPECIFICATION.md` defines observations (what Atlas acquires),
`INTELLIGENCE.md` defines the ontology (what intelligence is), `ARCHITECTURE.md`
defines layer ownership, and this specification defines the derivation contract
(how observations become facts).

## Normative Authority

Every repository fact is governed by this specification. The facts layer
implements it.

## Architectural Invariants

### No New Runtime Layer

Repository facts live **only** as a flat expansion of the repository fact
aggregate. There is no separate package, engine, or acquisition path. Derived
knowledge is a field on the existing repository fact aggregate.

### Facts, Not Signals

Every value defined here is a **Fact**: a deterministic aggregate derivable
purely from observations, given a reference time. The four named indicators
(Ownership, Consistency, Depth, Activity) are unchanged. Ratio aggregates
(ForkRatio, LicensedRatio, and similar) are descriptive portfolio facts, *not*
indicators: they are not part of the four-value measurement model and do not
feed the indicators layer. They are exposed for evidence.

### Determinism

Equivalent observation state plus an equivalent reference time must always
produce equivalent repository facts. No clock dependence beyond the explicit
reference time. No provider dependence beyond the normalized observations.

### Observation-Only Inputs

Every derived fact consumes only fields already present on the repository
observation. Repository facts add **no new observations**. They do not read user
profile, contribution, acquisition, or transport data. A fact requiring new
observations (contribution windows, language union across contributed
repositories, follower counts) is not a repository fact.

### Flat Expansion Only

The repository fact aggregate holds scalar fields. It does not hold nested
sub-structs, maps of arbitrary depth, or per-repository slices. Ranked languages
is the single allowed ordered list (a deterministic ranking, not a nested
object). This keeps facts flat, serializable, and inspectable.

---

## Observation Inputs

All derivations consume the following repository observation fields (owned by
Atlas per `OBSERVATION_SPECIFICATION.md`):

| Domain | Fields consumed |
| --- | --- |
| Identity | `Archived`, `Template` |
| Ownership | `Fork`, `CollaboratorCount` |
| Timeline | `CreatedAt`, `ReleaseCount`, `LatestReleaseAt` |
| Technology | `License`, `Topics`, `DefaultBranchProtected`, `LanguageDistribution` |
| Maintenance | `Stars`, `OpenIssues`, `PullRequestCount`, `Forks`, `Watchers` |
| Structure | `Size`, `DiscussionEnabled` |

None of these require new acquisition; repository facts are pure derivations.

**Provenance.** Facts reference observations; they never duplicate them. Each
fact exposes the repository observation fields it derives from, referencing the
repository observation family because portfolio facts aggregate every repository.
Indicators reach observations transitively through this mapping. See
[`PROVENANCE.md`](./PROVENANCE.md).

---

## Derived Fact Catalogue

Each fact lists: its field name on the repository fact aggregate, the
observation inputs, the deterministic formula, its type/range, and edge-case
handling. All formulas are evaluated against the repository set and a reference
time.

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
5. **Stable and reproducible.** Equivalent observation state always yields an
   identical `RankedLanguages` slice, on every toolchain.

Atlas keeps the deterministic language-ranking derivation and leaves behind the
presentation and gamification concerns that surrounded it in prior research.

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

These are pure summations/counts over fields already present on the repository
observation. They answer "how much release, PR, and collaboration infrastructure
does this portfolio exhibit?" without interpretation.

### 4. Ratio Facts (deterministic, in [0, 1])

| Field | Type | Formula | Edge cases |
| --- | --- | --- | --- |
| `ForkRatio` | `float64` | `ForkRepos / TotalRepos` | `0` if `TotalRepos == 0` |
| `LicensedRatio` | `float64` | `LicensedRepos / TotalRepos` | `0` if `TotalRepos == 0` |
| `ArchivedRatio` | `float64` | `ArchivedRepos / TotalRepos` | `0` if `TotalRepos == 0` |
| `MeanRepoSize` | `float64` | `sum(repo.Size) / TotalRepos` | `0` if `TotalRepos == 0` |
| `TopicBreadth` | `float64` | `TotalTopics / TotalRepos` | `0` if `TotalRepos == 0` |

These normalize counts against portfolio size. `ForkRatio` answers "what share
of this portfolio is original work versus forks?" `LicensedRatio` and
`ArchivedRatio` characterize portfolio hygiene. All are clamped to [0, 1] by
construction (`numerator ≤ denominator`).

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

## Excluded Concepts

The following concepts are deliberately **excluded** from the repository fact
layer, with rationale. Some belong to other layers; others are presentation or
gamification concerns that violate the Information Invariant.

| Concept | Reason |
| --- | --- |
| Six-stat athletic rating model | Gamified bounded rating; not a fact |
| Archetype naming (from activity shape) | Evaluation/Presentation; not a fact |
| Tiered finish labelling (bronze/silver/gold/…) | Evaluation/Presentation gamification |
| Legacy/founder forced overalls | Non-deterministic favouritism |
| Activity-shape style naming | Behaviour facts; needs contribution windows |
| Language-logo resolution | Presentation; not intelligence |
| Recent contributions / commits / active days | Owned by activity facts |
| Language union across contributed repositories | Collaboration facts; requires new acquisition |
| Followers / account age | Owned by user metadata, not repository facts |

The deterministic derivation technique behind log-scaling and language ranking
is adapted where it maps cleanly to a flat fact (star distribution, language
ranking). The gamified framing is never adopted.

---

## Out of Scope

Repository facts own only the repository fact aggregate (flat fields). They own
none of the following:

- Repository observation — facts add no observation fields
- Acquisition — facts own no acquisition, merge, or normalization
- Indicators — facts own no indicator formulas
- Evaluation, Projection — facts own no scoring or presentation
- CLI and server endpoints — facts own no presentation surface

Fact domains beyond the repository aggregate (contribution, technology,
behaviour, collaboration) are ownership boundaries, not repository facts.

---

## Conformance Invariants

For every repository fact, the following hold:

1. It is derivable solely from repository observation fields.
2. It is deterministic given the repository set and reference time.
3. Its edge cases (empty input, zero denominators) are handled explicitly.
4. It lives in the facts layer with no new package or layer.
5. It is exposed as a flat field on the repository fact aggregate (only ranked
   languages is an ordered list; no nested structs or maps).
6. It is not consumed by the indicators layer (which remains the four named
   signals).
7. This specification is its normative source; no implementation-only heuristics.
