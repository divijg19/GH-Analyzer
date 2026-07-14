# Atlas Activity Intelligence Specification

This document is the **normative specification** for all Activity Intelligence
derivations in Atlas. It defines every field in `facts.ActivityFacts`, its
purpose, its input observations, its derivation, its edge cases, and its
invariants.

It is the specification for **v0.8.17: Activity Intelligence**.

Where `ACTIVITY_SPECIFICATION.md` defines the raw observations, this document
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

This document is the **normative source** for every `ActivityFacts` field.
The code in `internal/facts/activity.go` is an **implementation** of this
specification.

If a fact's derivation, inputs, or semantics change, this document **must be
updated first**. The code is then updated to match.

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
normalizeActivityProfile → ActivityObservation[] (Tier 1, kinds: commit/PR/review/issue/aggregate/active_day/contribution_by_repo)
        ↓
queryLifetimeBatch (GraphQL, per-year batches)
        ↓
normalizeYearContrib → ActivityObservation[] (Tier 1, kinds: commit/PR/review/issue/aggregate, with Metadata.Year)
        ↓
ActivityFactsFromObservations([]ActivityObservation, referenceTime)
        ↓
ActivityFacts
```

### Storage Resolution

```
ActivityFacts
        ↓
Profile.ActivityFacts (internal/index/index.go)
```

Raw `ActivityObservation[]` values are NOT stored in `Profile`. They are
ephemeral acquisition data.

---

## Certification Checklist

Every `ActivityFacts` field must satisfy these gates:

1. Exists in this specification with documented purpose and derivation.
2. Derives exclusively from `ActivityObservation[]` + `referenceTime`.
3. Derivation is deterministic (same inputs → same outputs).
4. Formula is documented in this specification.
5. Edge cases are documented in this specification.
6. Dedicated unit tests exist in `internal/facts` (canonical activity-fact tests).
7. No duplicate derivation exists in `RepositoryFacts` or elsewhere.
8. Downstream consumers (`Profile`, `InspectProjection`) remain unchanged.

---

## Explicitly Out of Scope (v0.8.17)

The following concepts are deliberately excluded from `ActivityFacts`:

| Concept | Reason | Target |
|---------|--------|--------|
| Individual event facts | No per-event observations in Tier 1 | Post-v0.8.17 |
| Discussion participation | No discussion activity observation | Post-v0.8.17 |
| Release participation | No release activity observation | Post-v0.8.17 |
| Review quality metrics | Requires per-review state data | Post-v0.8.17 |
| PR merge rate | Requires per-PR state data | Post-v0.8.17 |
| Time-of-day patterns | Requires per-event timestamps | Post-v0.8.17 |
| Weekday/weekend patterns | Requires per-event timestamps | Post-v0.8.17 |
| Trend analysis (increasing/decreasing) | Requires 3+ years of data with trend | Post-v0.8.17 |
| Comparison to peers | Requires multi-profile evaluation | v0.9.0+ |
| Scoring/normalization (0–99) | Belongs in future signal redesign | v0.9.0+ |
