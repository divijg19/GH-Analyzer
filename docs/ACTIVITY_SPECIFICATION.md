# Atlas Activity Observation Specification

This document is the **canonical activity observation specification** for
Atlas. It defines the `ActivityObservation` model, its kinds, its acquisition
mechanisms, and its observation catalogue.

It is the specification for **v0.8.17: Activity Intelligence**.

Where `ACTIVITY_INTELLIGENCE.md` defines the derived facts, this document
defines the raw observations from which those facts derive.

---

## Normative Authority

This document is the **normative source** for all activity observation
acquisition decisions. The code in `internal/acquisition/` is an
**implementation** of this specification.

If an observation's kind, acquisition path, or metadata changes, this document
**must be updated first**. No code change that alters activity observation
semantics is valid without a corresponding change to this specification.

---

## Architectural Invariants

### Atlas Owns the Ontology

Atlas defines `ActivityObservation`. GitHub REST and GraphQL are acquisition
mechanisms that produce Atlas' canonical type. GitHub API shapes never
influence Atlas' observation model.

### Activity Is Not a Vestige

Repositories are long-lived stateful entities. Activity is temporal.
`ActivityObservation` is a fundamentally different concept from
`RepositoryVestige`.

### Observation → Fact Pipeline

```
ActivityObservation[]
        ↓
   ActivityFacts   (derivation, deterministic)
        ↓
   Profile         (fact aggregation only)
```

Raw observations are acquisition data. Only `ActivityFacts` enter `Profile`.

### Determinism

Equivalent activity state must always produce equivalent
`ActivityObservation` values regardless of acquisition mechanism.

### Provider Independence

`ActivityObservation` has no provider-specific fields. GitHub-specific types
never leave `internal/acquisition`. The model is designed for future providers.

### Current Package Placement

The v0.8.17 implementation colocates activity observations and activity facts
with `internal/signals` for historical continuity. That is an implementation
detail, not
an architectural statement about long-term ownership.

---

## Canonical Model

### ActivityKind

`ActivityKind` is a typed string enumerating the kinds of activity Atlas
observes.

```go
type ActivityKind string

const (
    ActivityKindCommit              ActivityKind = "commit"
    ActivityKindPullRequest         ActivityKind = "pull_request"
    ActivityKindReview              ActivityKind = "review"
    ActivityKindIssue               ActivityKind = "issue"
    ActivityKindDiscussion          ActivityKind = "discussion"
    ActivityKindRelease             ActivityKind = "release"
    ActivityKindActiveDay           ActivityKind = "active_day"
    ActivityKindContributionByRepo  ActivityKind = "contribution_by_repo"
    ActivityKindAggregate           ActivityKind = "aggregate"
)
```

Extensible by design. New kinds are added as the ontology grows.

### ActivityObservation

```go
type ActivityObservation struct {
    Kind       ActivityKind
    Count      int
    Repository string       // "owner/name" for per-repo observations; empty for global
    Actor      string       // Always the observation subject (username)
    OccurredAt time.Time    // Observation timestamp or window boundary
    Metadata   ActivityMetadata
}
```

### ActivityMetadata

Strongly typed fields. Every field has a documented zero-value default.
Fields are populated based on `Kind`.

```go
type ActivityMetadata struct {
    // Calendar window
    WindowStart time.Time
    WindowEnd   time.Time
    ActiveDays  int         // Days with ≥1 contribution in window
    TotalDays   int         // Total days in window

    // Year-specific (Kind: aggregate / per-year buckets)
    Year int

    // Per-repo (Kind: contribution_by_repo)
    RepoCommitCount int

    // Private activity (Kind: aggregate)
    RestrictedCount int
}
```

### Zero-Value Rules

Every `ActivityMetadata` field has a deterministic zero value:

| Field | Zero Value | Meaning |
|-------|-----------|---------|
| `WindowStart` | `time.Time{}` | No windowing applied |
| `WindowEnd` | `time.Time{}` | No windowing applied |
| `ActiveDays` | `0` | No active day data |
| `TotalDays` | `0` | No window defined |
| `Year` | `0` | Not year-specific |
| `RepoCommitCount` | `0` | No per-repo breakdown |
| `RestrictedCount` | `0` | No private activity data |

---

## Observation Catalogue

### Existing Observations (from v0.8.15 contributions)

| # | Observation | Kind | Semantic Purpose | Status |
|---|-------------|------|------------------|--------|
| 1 | PullRequestCount (total) | `pull_request` | Total user PRs (REST search API) | **Existing** — migrate to GraphQL |
| 2 | IssueCount (total) | `issue` | Total user issues (REST search API) | **Existing** — migrate to GraphQL |

These are acquired via `internal/acquisition/contributions.go` using the
GitHub search API. They will be **migrated** to GraphQL-acquired
`ActivityObservation` in v0.8.17. The REST-based `ContributionsDTO` is
deprecated.

### Tier 1 Observations

| # | Observation | Kind | Semantic Purpose | Owner | Canonical Acquisition | Merge Policy | Fallback | Nullable | Degradation |
|---|-------------|------|------------------|-------|----------------------|--------------|----------|----------|-------------|
| 3 | CommitCount (1-year) | `commit` | Total commits in last 365 days | Atlas | GraphQL contributionsCollection | GraphQL authoritative | 0 | No | Defaults to 0 |
| 4 | PullRequestCount (1-year) | `pull_request` | Total PRs in last 365 days | Atlas | GraphQL contributionsCollection | GraphQL authoritative | 0 | No | Defaults to 0 |
| 5 | ReviewCount (1-year) | `review` | Total reviews in last 365 days | Atlas | GraphQL contributionsCollection | GraphQL authoritative | 0 | No | Defaults to 0 |
| 6 | IssueCount (1-year) | `issue` | Total issues in last 365 days | Atlas | GraphQL contributionsCollection | GraphQL authoritative | 0 | No | Defaults to 0 |
| 7 | PrivateContributionCount (1-year) | `aggregate` | Private activity in last 365 days | Atlas | GraphQL contributionsCollection | GraphQL authoritative | 0 | No | Defaults to 0 |
| 8 | ActiveDays (1-year) | `active_day` | Days with ≥1 contribution in 365-day window | Atlas | GraphQL contributionCalendar | GraphQL authoritative | 0 | No | Defaults to 0 |
| 9 | ContributionByRepo (1-year) | `contribution_by_repo` | Commit contributions per repository | Atlas | GraphQL commitContributionsByRepository | GraphQL authoritative | Empty | Yes | Empty if absent |
| 10 | LifetimeContributions (all years) | `aggregate` | Sum of all contributions across all years | Atlas | GraphQL multi-year batches | GraphQL authoritative | 0 | No | Defaults to 0 |
| 11 | YearlyCommitCount | `commit` | Commits in a specific year | Atlas | GraphQL per-year batch | GraphQL authoritative | 0 | No | Defaults to 0 |
| 12 | YearlyPullRequestCount | `pull_request` | PRs in a specific year | Atlas | GraphQL per-year batch | GraphQL authoritative | 0 | No | Defaults to 0 |
| 13 | YearlyReviewCount | `review` | Reviews in a specific year | Atlas | GraphQL per-year batch | GraphQL authoritative | 0 | No | Defaults to 0 |
| 14 | YearlyIssueCount | `issue` | Issues in a specific year | Atlas | GraphQL per-year batch | GraphQL authoritative | 0 | No | Defaults to 0 |
| 15 | YearlyPrivateCount | `aggregate` | Private activity in a specific year | Atlas | GraphQL per-year batch | GraphQL authoritative | 0 | No | Defaults to 0 |

### Tier 2 Observations (Deferred)

| # | Observation | Kind | Notes | Target |
|---|-------------|------|-------|--------|
| 16 | CommitDetails (per-commit) | `commit` | Individual commit SHA, message, timestamp | Post-v0.8.17 |
| 17 | PRDetails (per-PR) | `pull_request` | Individual PR title, body, timeline | Post-v0.8.17 |
| 18 | ReviewDetails (per-review) | `review` | Individual review state, comments | Post-v0.8.17 |
| 19 | DiscussionDetails (per-discussion) | `discussion` | Discussion title, category, comments | Post-v0.8.17 |
| 20 | ReleaseDetails (per-release) | `release` | Release name, tag, prerelease flag | Post-v0.8.17 |

### Rejected Observations

| Observation | Reason |
|-------------|--------|
| Starred repositories | Not activity — relates to repository state, not developer activity |
| Followers/following | Not activity — social graph, not temporal |
| Organization membership | Not activity — identity, not temporal |
| Contribution heatmap positions | Presentation concern, not intelligence |
| Contribution spike boolean (binary) | Too coarse — Atlas derives multi-window recency |

---

## Acquisition Mapping

### GraphQL Queries

#### Profile Activity Query (1-year window)

Acquired via `contributionsCollection` on the base `user` query. This is the
primary acquisition path for Tier 1 observations 3–9.

```graphql
query ActivityProfile($login: String!) {
  user(login: $login) {
    recent: contributionsCollection {
      totalCommitContributions
      totalPullRequestContributions
      totalPullRequestReviewContributions
      totalIssueContributions
      restrictedContributionsCount
      commitContributionsByRepository(maxRepositories: 100) {
        contributions { totalCount }
        repository { nameWithOwner }
      }
      contributionCalendar {
        weeks {
          contributionDays { contributionCount }
        }
      }
    }
  }
}
```

**Normalization path:**

| GraphQL Field | ActivityObservation |
|---------------|-------------------|
| `totalCommitContributions` | `{Kind: commit, Count: N, OccurredAt: now}` |
| `totalPullRequestContributions` | `{Kind: pull_request, Count: N, OccurredAt: now}` |
| `totalPullRequestReviewContributions` | `{Kind: review, Count: N, OccurredAt: now}` |
| `totalIssueContributions` | `{Kind: issue, Count: N, OccurredAt: now}` |
| `restrictedContributionsCount` | `{Kind: aggregate, Count: N, OccurredAt: now, Metadata.RestrictedCount: N}` |
| `commitContributionsByRepository[]` | `{Kind: contribution_by_repo, Count: N, Repository: owner/name, OccurredAt: now, Metadata.RepoCommitCount: N}` |
| `contributionCalendar.weeks[].days[]` | `{Kind: active_day, Count: N, Metadata.ActiveDays: days_with_count>0, Metadata.TotalDays: 365}` |

#### Lifetime Activity Query (multi-year, batched)

Acquired via aliased `contributionsCollection` calls, one per year, batched
4 years per request.

```graphql
query LifetimeActivity($login: String!) {
  user(login: $login) {
    y2022: contributionsCollection(from: "2022-01-01T00:00:00Z", to: "2022-12-31T23:59:59Z") {
      totalCommitContributions
      totalPullRequestContributions
      totalPullRequestReviewContributions
      totalIssueContributions
      restrictedContributionsCount
    }
    y2023: contributionsCollection(from: "2023-01-01T00:00:00Z", to: "2023-12-31T23:59:59Z") { ... }
    y2024: contributionsCollection(from: "2024-01-01T00:00:00Z", to: "2024-12-31T23:59:59Z") { ... }
    y2025: contributionsCollection(from: "2025-01-01T00:00:00Z", to: "2025-12-31T23:59:59Z") { ... }
  }
}
```

**Normalization path:**

| GraphQL Field | ActivityObservation |
|---------------|-------------------|
| `y2022.totalCommitContributions` | `{Kind: commit, Count: N, Metadata.Year: 2022}` |
| `y2022.totalPullRequestContributions` | `{Kind: pull_request, Count: N, Metadata.Year: 2022}` |

### Merge Policy

All Tier 1 activity observations are **GraphQL-authoritative**. REST cannot
produce these observations (the search API returns totals, not windowed
counts). Merge is trivially the GraphQL value.

### Graceful Degradation

If the GraphQL activity query fails:

| Observation | Degradation | Impact |
|-------------|-------------|--------|
| 1-year counts (3–7) | Default to 0 | ActivityFacts show no recent activity |
| Active days (8) | Default to 0 | Cadence metrics are zero |
| Contribution by repo (9) | Empty slice | Breadth/depth metrics are zero |
| Lifetime/yearly (10–15) | Default to 0 | Lifetime cadence unavailable |

Observations 1–2 (existing REST-based) remain as fallbacks for total counts
but are deprecated in favour of GraphQL activity observations.

---

## Acquisition Implementation

### New Files

```
internal/acquisition/
    activity_types.go     ActivityKind, ActivityObservation, ActivityMetadata
    graphql_activity.go   FetchActivityObservations, ActivityGraphQL queries
    normalize_activity.go ActivityObservation normalization
```

### Implementation Constraints

Every acquisition query exists because this specification requires it. No
speculative additions. No GraphQL query is added without a corresponding row
in the Tier 1 table above.

### Pagination

Cursor-based pagination is not needed for Tier 1. The 1-year
`contributionsCollection` is a single object (not a connection). Lifetime
queries are batched 4 years per request with retry.

### Incremental Acquisition

Designed for future incremental sync. The `OccurredAt` field supports future
`since` parameters. Do not implement synchronization in v0.8.17.

---

## Type Definitions (Normative)

```go
package acquisition

import "time"

// ActivityKind is the canonical kind of an activity observation.
// Every kind corresponds to a type of observable developer activity.
// Kinds are extensible: new values must be added to this type and
// documented in ACTIVITY_SPECIFICATION.md before implementation.
type ActivityKind string

const (
    ActivityKindCommit             ActivityKind = "commit"
    ActivityKindPullRequest        ActivityKind = "pull_request"
    ActivityKindReview             ActivityKind = "review"
    ActivityKindIssue              ActivityKind = "issue"
    ActivityKindDiscussion         ActivityKind = "discussion"
    ActivityKindRelease            ActivityKind = "release"
    ActivityKindActiveDay          ActivityKind = "active_day"
    ActivityKindContributionByRepo ActivityKind = "contribution_by_repo"
    ActivityKindAggregate          ActivityKind = "aggregate"
)

// ActivityObservation is the canonical observation unit for developer activity.
//
// It represents a single observable fact about a developer's activity on a
// software collaboration platform. Activity is temporal: each observation
// has an OccurredAt timestamp or window.
//
// ActivityObservation is NOT a RepositoryVestige. Repositories are
// long-lived stateful entities. Activity is temporal. These are
// fundamentally different ontological concepts.
type ActivityObservation struct {
    Kind       ActivityKind
    Count      int              // Quantity observed (0 for boolean observations)
    Repository string           // "owner/name" — empty for global observations
    Actor      string           // The observation subject (username)
    OccurredAt time.Time        // Observation timestamp or window end
    Metadata   ActivityMetadata // Strongly typed, zero-value defaults
}

// ActivityMetadata holds all optional typed fields for ActivityObservation.
//
// Fields are populated based on ActivityKind. Every field has a deterministic
// zero value. No field is mandatory for all observation kinds.
type ActivityMetadata struct {
    WindowStart     time.Time // Calendar window start
    WindowEnd       time.Time // Calendar window end
    ActiveDays      int       // Days with ≥1 contribution (Kind: active_day)
    TotalDays       int       // Total days in window (Kind: active_day)

    Year            int       // Specific calendar year (Kind: aggregate, per-year)

    RepoCommitCount int       // Commits in this repo (Kind: contribution_by_repo)

    RestrictedCount int       // Private activity count (Kind: aggregate)

    ReleaseName     string    // (Kind: release)
    ReleaseTag      string    // (Kind: release)
    IsPrerelease    bool      // (Kind: release)

    PRNumber        int       // (Kind: pull_request)
    PRState         string    // "OPEN" | "MERGED" | "CLOSED" (Kind: pull_request)
    PRAdditions     int       // (Kind: pull_request)
    PRDeletions     int       // (Kind: pull_request)

    ReviewState     string    // "APPROVED" | "CHANGES_REQUESTED" | "COMMENTED" (Kind: review)
}
```

---

## Observation Certification Checklist

Every Tier 1 observation must satisfy these gates before v0.8.17 certification:

1. Kind is defined in `ActivityKind`.
2. Canonical acquisition mechanism is documented (GraphQL query + field).
3. Deterministic normalization path exists in `internal/acquisition`.
4. Graceful degradation is documented and implemented.
5. Unit tests cover normalization.
6. `ActivityObservation` values are serializable.
7. This specification is the normative source for the observation.
8. No downstream package references GraphQL DTOs or provider mechanisms.

---

## Out of Scope (Frozen Layers)

The following are untouched in v0.8.17:

- `RepositoryVestige` — frozen (no new fields)
- `RepositoryFacts` — frozen (no new fields)
- `Signals` / `Indicators` — frozen (no changes)
- `Evaluation` — frozen
- `Engine` / `Search` / `Ranking` — frozen
- `Projection` — frozen
- CLI behaviour — frozen
- Individual event storage/indexing — deferred
- Multi-provider acquisition — deferred
- Live synchronization — deferred

v0.8.17 exists solely to define, acquire, and derive Activity Intelligence.
It changes nothing in the downstream layers.
