# Atlas Activity Intelligence Specification

Every field in `ActivityFacts` — its purpose, its input observations, its
derivation, its edge cases, and its invariants — is defined by this
specification.

`ACTIVITY_SPECIFICATION.md` defines the raw observations; this specification
defines the derived facts. The pipeline is:

```
observations.ActivityObservation[]
        ↓
   facts.ActivityFacts   (this document)
        ↓
   Profile         (fact aggregation)
```

---

## Normative Authority

Every `ActivityFacts` field is governed by this specification. The facts layer
implements it.

---

## Architectural Invariants

### Derive Only from ActivityObservation

Every `ActivityFacts` field derives exclusively from
`observations.ActivityObservation[]` and `referenceTime`. No `RepositoryFacts`,
`Indicators`, `Evaluation`, or other intelligence domains are consulted.

### Determinism

Equivalent `ActivityObservation[]` + `referenceTime` must always produce
equivalent `ActivityFacts`.

### No Downstream Changes

`ActivityFacts` is consumed by `Profile` only. No changes to `Indicators`,
`Evaluation`, `Engine`, `Projection`, or `CLI`.

### Zero-Value Meaning

| Field | Zero Value | Meaning |
|-------|-----------|---------|
| All int fields | `0` | No activity observed |
| All float64 fields | `0.0` | No activity observed or division by zero avoided |
| `ActivityCadence` | `0.0` | No calendar data available |
| `ContributionFrequency` | `0.0` | No active days |
| `RepositoryDepth` | `0.0` | No per-repo contributions |
| `CommitCadence` | `0.0` | No yearly data |
| `ContributionRecency` | `0.0` | No lifetime data |

---

## Fact Catalogue

### Field Group 1: Aggregate Volume (1-year window)

These fields count raw contributions in the last 365 days. They derive from
`ActivityObservation{Kind: commit|pull_request|review|issue|aggregate, Metadata.Year == 0}`.

| # | Field | Type | Input Kind | Derivation | Edge Cases |
|---|-------|------|------------|------------|------------|
| 1 | `RecentCommits` | int | commit | Sum of all `Count` where `Kind == commit` and `Metadata.Year == 0` | Defaults to 0 |
| 2 | `RecentPullRequests` | int | pull_request | Sum of all `Count` where `Kind == pull_request` and `Metadata.Year == 0` | Defaults to 0 |
| 3 | `RecentReviews` | int | review | Sum of all `Count` where `Kind == review` and `Metadata.Year == 0` | Defaults to 0 |
| 4 | `RecentIssues` | int | issue | Sum of all `Count` where `Kind == issue` and `Metadata.Year == 0` | Defaults to 0 |
| 5 | `RecentPrivate` | int | aggregate | Sum of all `Count` where `Kind == aggregate` and `Metadata.Year == 0` | Defaults to 0 |

**Purpose:** These fields answer "How active was the developer in the last year?"

### Field Group 2: Lifetime Aggregates (all years)

These fields count raw contributions across all observed years. They derive
from `ActivityObservation{Kind: commit|pull_request|review|issue|aggregate}`
regardless of `Metadata.Year`.

| # | Field | Type | Input Kind | Derivation | Edge Cases |
|---|-------|------|------------|------------|------------|
| 6 | `LifetimeCommits` | int | commit | Sum of all `Count` where `Kind == commit` | Defaults to 0 |
| 7 | `LifetimePullRequests` | int | pull_request | Sum of all `Count` where `Kind == pull_request` | Defaults to 0 |
| 8 | `LifetimeReviews` | int | review | Sum of all `Count` where `Kind == review` | Defaults to 0 |
| 9 | `LifetimeIssues` | int | issue | Sum of all `Count` where `Kind == issue` | Defaults to 0 |
| 10 | `LifetimePrivate` | int | aggregate | Sum of all `Count` where `Kind == aggregate` | Defaults to 0 |
| 11 | `LifetimeTotal` | int | all types | `LifetimeCommits + LifetimePullRequests + LifetimeReviews + LifetimeIssues + LifetimePrivate` | Defaults to 0 |

**Purpose:** These fields answer "What is the developer's lifetime output?"

### Field Group 3: Contribution Breadth

These fields measure the diversity of the developer's activity.

| # | Field | Type | Inputs | Derivation | Edge Cases |
|---|-------|------|--------|------------|------------|
| 12 | `ContributionBreadth` | int | Recent counts | Count of `Recent*` fields with value > 0. Range: 0–5 | 0 if no recent activity |
| 13 | `RepositoryBreadth` | int | contribution_by_repo | Count of distinct `Repository` values where `Kind == contribution_by_repo` and `Metadata.Year == 0` | 0 if no per-repo data |

**Derivation detail (ContributionBreadth):**

| Condition | Breadth |
|-----------|---------|
| All Recent* fields are 0 | 0 |
| Exactly one Recent* field > 0 | 1 |
| Two Recent* fields > 0 | 2 |
| Three Recent* fields > 0 | 3 |
| Four Recent* fields > 0 | 4 |
| All five Recent* fields > 0 | 5 |

### Field Group 4: Contribution Cadence

These fields measure the developer's activity rhythm.

| # | Field | Type | Inputs | Derivation | Edge Cases |
|---|-------|------|--------|------------|------------|
| 14 | `ActiveDays` | int | active_day | `ActiveDays` from `Metadata` where `Kind == active_day` and `Metadata.Year == 0` | 0 if no calendar data |
| 15 | `ActivityCadence` | float64 | active_day | `ActiveDays / TotalDays` (from active_day metadata). Range: [0, 1] | 0.0 if totalDays == 0 |
| 16 | `ContributionFrequency` | float64 | all recent types | `(RecentCommits + RecentPullRequests + RecentReviews + RecentIssues + RecentPrivate) / ActiveDays` | 0.0 if activeDays == 0 |
| 17 | `CommitCadence` | float64 | commit (yearly) | `LifetimeCommits / YearCount`. Average commits per year | 0.0 if yearCount == 0 |

**Example:**

| Scenario | ActiveDays | TotalDays | ActivityCadence | ContributionFrequency |
|----------|-----------|-----------|-----------------|---------------------|
| Daily active | 365 | 365 | 1.0 | ~3.0 (1100/365) |
| Weekender | 104 | 365 | ~0.2849 | ~10.0 (1040/104) |
| Rare | 12 | 365 | ~0.0329 | ~1.0 (12/12) |

### Field Group 5: Contribution Depth

These fields measure engagement depth per repository.

| # | Field | Type | Inputs | Derivation | Edge Cases |
|---|-------|------|--------|------------|------------|
| 18 | `RepositoryDepth` | float64 | contribution_by_repo | Sum of all `Count` where `Kind == contribution_by_repo` divided by `RepositoryBreadth` | 0.0 if RepositoryBreadth == 0 |
| 19 | `YearCount` | int | all yearly types | Count of distinct `Metadata.Year` values across all observations | 0 if no yearly data |

### Field Group 6: Contribution Recency

These fields measure how recent the developer's activity is relative to their
lifetime baseline.

| # | Field | Type | Inputs | Derivation | Edge Cases |
|---|-------|------|--------|------------|------------|
| 20 | `ContributionRecency` | float64 | all types | `totalRecent / (LifetimeTotal / YearCount)`. Ratio of recent activity to expected yearly average | 0.0 if yearCount == 0 or expected == 0 |

**Interpretation:**

| Value | Meaning |
|-------|---------|
| < 0.5 | Significantly less active recently than historically |
| ~ 1.0 | Activity level consistent with historical average |
| > 2.0 | Significantly more active recently than historically |

---

## Derivation Lifecycle

### Input Resolution

```
queryActivityProfile (GraphQL, 1-year window)
        ↓
normalize → ActivityObservation[] (kinds: commit/PR/review/issue/aggregate/active_day/contribution_by_repo)
        ↓
lifetime acquisition (GraphQL, per-year batches)
        ↓
normalize → ActivityObservation[] (kinds: commit/PR/review/issue/aggregate, with Metadata.Year)
        ↓
ActivityFactsFromObservations([]ActivityObservation, referenceTime)
        ↓
ActivityFacts
```

### Storage Resolution

```
ActivityFacts
        ↓
Profile.ActivityFacts
```

Raw `ActivityObservation[]` values are NOT stored in `Profile`. They are
ephemeral acquisition data.

---

## Conformance Invariants

For every `ActivityFacts` field, the following hold:

1. It exists in this specification with a documented purpose and derivation.
2. It derives exclusively from `ActivityObservation[]` + `referenceTime`.
3. Its derivation is deterministic (same inputs → same outputs).
4. Its formula is documented in this specification.
5. Its edge cases are documented in this specification.
6. No duplicate derivation exists in `RepositoryFacts` or elsewhere.

---

## Out of Scope

The following concepts are deliberately excluded from `ActivityFacts`:

| Concept | Reason |
|---------|--------|
| Individual event facts | No per-event observations acquired |
| Discussion participation | No discussion activity observation |
| Release participation | No release activity observation |
| Review quality metrics | Requires per-review state data |
| PR merge rate | Requires per-PR state data |
| Time-of-day patterns | Requires per-event timestamps |
| Weekday/weekend patterns | Requires per-event timestamps |
| Trend analysis (increasing/decreasing) | Requires multi-year trend data |
| Comparison to peers | Requires multi-profile evaluation |
| Scoring/normalization (0–99) | Belongs to interpretation layers |
