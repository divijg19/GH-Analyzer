# Atlas Intelligence Ontology

This document is the canonical reference for the Atlas Intelligence Ontology.
It defines every concept Atlas understands, following one guiding principle:

> **Atlas observes Observations. Atlas derives Facts. Atlas measures Indicators. Atlas performs Evaluation. Atlas projects Intelligence.**

Where `ARCHITECTURE.md` answers *"which package owns this?"*, this document
answers *"what is the concept?"* and *"what are its invariants?"*

The model is a strictly downward pipeline. Each stage transforms the previous
stage's output into a richer, still-deterministic representation. Nothing is
ever inferred probabilistically: given the same inputs, Atlas always produces
the same intelligence.

---

## Canonical Vocabulary

Every concept has exactly one name, one owner, and one responsibility. No
synonyms, no overlaps, no ambiguous ownership.

```
Transport
    ↓
Acquisition
    ↓
Normalization
    ↓
    Observations
    ↓
    Facts
    ↓
    Indicators
    ↓
    Profile
    ↓
Evaluation
    ↓
Engine
    ↓
Projection
    ↓
Consumers
```

---

## Layer Definitions

### Transport — `internal/github`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Owns HTTP communication with external APIs (currently GitHub). |
| **Owner** | `internal/github` |
| **Inputs** | HTTP request parameters (URL, headers, auth). |
| **Outputs** | Raw JSON byte payloads. |
| **Invariants** | Never parses domain types. Never interprets responses. |
| **Prohibited** | Domain logic, normalization, caching. |

---

### Acquisition — `internal/acquisition`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Fetches external data via REST and GraphQL; produces DTOs that mirror the API schema. |
| **Owner** | `internal/acquisition` |
| **Inputs** | Username, owner+repo name, search query, context. |
| **Outputs** | `RepoDTO` (REST), `graphQLRepo` (GraphQL), `UserDTO`, `ContributionsDTO`. |
| **Invariants** | DTOs are pure data — no intelligence, no computation. REST and GraphQL are acquisition mechanisms only, not intelligence layers. |
| **Prohibited** | Domain model construction, scoring, evaluation, merge policy. |

---

### Normalization — `internal/acquisition/normalize.go`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Maps DTOs to domain vestiges. The boundary between "what GitHub returns" and "what Atlas knows." |
| **Owner** | `internal/acquisition` |
| **Inputs** | `RepoDTO` (REST), `graphQLRepo` (GraphQL), `UserDTO`, `ContributionsDTO`. |
| **Outputs** | `observations.RepositoryVestige` (full from REST, partial from GraphQL), `profile.UserMetadata`, `contributions.Summary`. |
| **Invariants** | One-to-one: one DTO → one vestige. No aggregation, no derivation, no omitting fields without documentation. Both REST and GraphQL DTOs pass through normalization. |
| **Prohibited** | Aggregation, scoring, evaluation, filtering, merge of multiple acquisition sources. |

---

### Observations

| Attribute | Definition |
|-----------|------------|
| **Purpose** | The canonical raw observation of a software development artifact. Every acquisition backend (REST, GraphQL, GitFut, local git) produces vestiges; Atlas never acquires from them directly. |
| **Owner** | `internal/observations` (repository and activity observations), `internal/profile` (metadata), `internal/contributions` (summary). |
| **Inputs** | Normalized output from acquisition. |
| **Outputs** | `observations.RepositoryVestige`, `observations.ActivityObservation`, `profile.UserMetadata`, `contributions.Summary`. |
| **Invariants** | Vestiges are observables, not computations. Fields are organized into observation domains: Identity, Ownership, Timeline, Technology, Maintenance, Structure. ActivityObservation captures temporal developer activity. |
| **Prohibited** | Derivation, aggregation, scoring, evaluation. |

**Observation domains of `RepositoryVestige`:**

- **Identity** — Name, Visibility, Archived, Template
- **Ownership** — Fork, ParentRepository, CollaboratorCount
- **Timeline** — CreatedAt, UpdatedAt, PushedAt, ReleaseCount, LatestReleaseAt
- **Technology** — License, Topics, DefaultBranch, DefaultBranchProtected, LanguageDistribution
- **Maintenance** — OpenIssues, Stars, Forks, Watchers, PullRequestCount
- **Structure** — Size, DiscussionEnabled

**`ActivityObservation` kinds** (v0.8.17):

- **Commit** — commit contributions within a time window
- **PullRequest** — PR contributions within a time window
- **Review** — code review contributions within a time window
- **Issue** — issue contributions within a time window
- **Discussion** — discussion contributions (deferred)
- **Release** — release contributions (deferred)
- **ActiveDay** — calendar activity metric
- **ContributionByRepo** — per-repository contribution breakdown
- **Aggregate** — lifetime/private contribution sums

**Future vestige families** (documented placeholders, no implementation):

- `CommitVestige` (individual commits)
- `PullRequestVestige` (individual PRs)
- `ReviewVestige` (individual reviews)
- `ReleaseVestige` (individual releases)
- `ContributorVestige`
- `OrganizationVestige`

---

### Facts — `facts.RepositoryFacts`, `facts.ActivityFacts`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Deterministic, evidence-backed aggregates computed from vestiges. Facts answer *"what do we know?"* without interpretation. |
| **Owner** | `internal/facts`. |
| **Inputs** | `[]observations.RepositoryVestige`. |
| **Outputs** | `facts.RepositoryFacts`. |
| **Invariants** | Purely derivable from vestiges: given the same vestiges, identical facts. Never fetch, never observe, never evaluate. |
| **Prohibited** | Observation (I/O), scoring, evaluation, confidence. |

**Fact families**:

| Fact Family | Status | Fields |
|-------------|--------|--------|
| `RepositoryFacts` | ✅ Implemented | TotalRepos, OriginalRepos, ForkRepos, RecentRepos, DeepRepos, ValidRepos, ValidOriginalRepos, LargestRepoSize, LatestActivity, ArchivedRepos, TemplateRepos, PublicRepos, PrivateRepos, LicensedRepos, TotalStars, TotalForks, TotalWatchers, TotalOpenIssues, TotalTopics, OldestCreated, NewestCreated, MaxRepoStars, MeanRepoStars, LanguageCount, RankedLanguages, TotalReleases, LatestReleaseAt, ReleasedRepos, TotalPullRequests, TotalCollaborators, ProtectedBranchRepos, DiscussionRepos, ForkRatio, LicensedRatio, ArchivedRatio, PortfolioAgeDays, NewestRepoAgeDays, DaysSinceLatestRelease, MeanRepoSize, TopicBreadth |
| `ActivityFacts` | ✅ Implemented (v0.8.17) | RecentCommits, RecentPullRequests, RecentReviews, RecentIssues, RecentPrivate, LifetimeCommits, LifetimePullRequests, LifetimeReviews, LifetimeIssues, LifetimePrivate, LifetimeTotal, ContributionBreadth, RepositoryBreadth, ActiveDays, ActivityCadence, ContributionFrequency, RepositoryDepth, YearCount, CommitCadence, ContributionRecency |

---

### Indicators — `indicators.Signals`, `indicators.RawScore`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Normalized measurements derived from facts. Signals quantify observations: ownership, consistency, depth, activity. |
| **Owner** | `internal/indicators`. |
| **Inputs** | `facts.RepositoryFacts`. |
| **Outputs** | `indicators.Signals` (four float64 values in [0, 1]), `indicators.RawScore` (three component integer scores in [0, 100]). |
| **Invariants** | Deterministic, reproducible, bounded [0,1], explainable, fact-derived. Signals measure facts — they never observe vestiges directly. |
| **Prohibited** | I/O, aggregation, scoring beyond measurement, confidence, ranking. |

**Current signals:**

| Signal | Range | Derives From |
|-----------|-------|--------------|
| `Ownership` | [0, 1] | `ValidOriginalRepos / ValidRepos` |
| `Consistency` | [0, 1] | `RecentRepos / max(10, OriginalRepos)` |
| `Depth` | [0, 1] | `DeepRepos / max(5, OriginalRepos)` |
| `Activity` | {0.1, 0.4, 0.7, 1.0} | Time since `LatestActivity` |

**Philosophy:**

Signals must satisfy:

- **Deterministic** — same facts → same indicators
- **Reproducible** — independent computation yields identical results
- **Bounded** — always in [0, 1]
- **Explainable** — every value traces to specific facts
- **Fact-derived** — indicators never read observations directly

Signals are **not**:

- Heuristic guesses
- AI or ML opinions
- Reputation scores
- Popularity metrics
- Evaluation judgments

---

### Profile — `index.Profile`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Atlas' complete assembled understanding of a candidate. Profile is context — not intelligence. It aggregates repository facts, activity facts, indicators, metadata, and contributions into a single canonical record. |
| **Owner** | `internal/index`. |
| **Inputs** | `facts.RepositoryFacts`, `facts.ActivityFacts`, `indicators.Signals`, `profile.UserMetadata`, `contributions.Summary`. |
| **Outputs** | `index.Profile` (the assembled candidate state before evaluation). |
| **Invariants** | Profile is assembled, not interpreted. It is the single source of truth for what Atlas knows about a candidate before evaluation. |
| **Prohibited** | Scoring, confidence, ranking, presentation, persistence concerns. |

**Key distinction:**

- **Profile** answers: *What do we know about this candidate?*
- **Intelligence** (future) answers: *What does that knowledge mean?*

Profile is not:
- A persistence format
- A presentation view
- An evaluation result
- An intelligence conclusion

---

### Evaluation — `internal/evaluation`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Score interpretation, confidence classification, penalty application, and ranking policy. Evaluation owns the transition from "what we know" to "how to interpret it." |
| **Owner** | `internal/evaluation`. |
| **Inputs** | `indicators.RawScore`, repository count, `index.Profile.Signals`. |
| **Outputs** | `int` (overall score 0–100), `evaluation.RankingPolicy.Score`, confidence classification. |
| **Invariants** | Never observes, never measures, never aggregates. Only interprets. |
| **Prohibited** | I/O, vestige observation, fact derivation, indicator measurement, presentation. |

**Owns:**

- `OverallScore` — weighted sum of component scores (ownership 0.3, consistency 0.4, depth 0.3)
- `ApplySmallSamplePenalty` — down-weights evaluations with <3 repositories
- `ClassifyConfidence` — repo-count → {high, moderate, low}
- `RankingPolicy` — type consumed by engine and projection for ordering

---

### Engine — `internal/engine`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Search query execution, candidate filtering, matching, ordering, and ranking. Engine is the orchestration layer between stored profiles and search consumers. |
| **Owner** | `internal/engine`. |
| **Inputs** | `index.Index`, `engine.Query`, `evaluation.RankingPolicy`. |
| **Outputs** | `[]engine.Result` (filtered and ordered candidates). |
| **Invariants** | Pure computation: no I/O, no acquisition, no vestige production. |
| **Prohibited** | I/O, scoring policy, presentation, evaluation. |

---

### Projection — `internal/projection`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Consumer-shaped, read-only, deterministic views of the domain. Projection is the terminal transformation before presentation — it reshapes, never recomputes. |
| **Owner** | `internal/projection`. |
| **Inputs** | `index.Profile`, `evaluation` outputs. |
| **Outputs** | `AnalyzeProjection`, `InspectProjection`, `SearchProjection`. |
| **Invariants** | No computation beyond reshaping and ordering. No business logic. |
| **Prohibited** | Scoring, evaluation, confidence, acquisition, fact derivation, indicator measurement. |

---

### Consumers — `cmd/atlas`, `cmd/server`, `web`

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Presentation surfaces that render intelligence for human consumption. |
| **Owner** | `cmd/atlas` (CLI), `cmd/server` (HTTP API), `web` (UI). |
| **Inputs** | Projections. |
| **Outputs** | CLI output, JSON responses, web pages. |
| **Invariants** | Format only; never derive. |
| **Prohibited** | Any derivation of facts, indicators, scores, rankings, or projections. |

---

## Intelligence Domains

Atlas organizes intelligence into six domains. Each domain owns a family of
vestiges, facts, and eventually indicators. No implementation exists beyond
`Repository` — the remaining five are documented ownership boundaries.

| Domain | Vestige Families | Fact Families | Signal Families | Status |
|--------|-----------------|---------------|-------------------|--------|
| **Repository** | `RepositoryVestige` | `RepositoryFacts` | Ownership, Consistency, Depth, Activity | ✅ Active |
| **Activity** | `ActivityObservation` | `ActivityFacts` | *(future)* | ✅ Active (v0.8.17) |
| **Technology** | *(future)* | *(future)* | *(future)* | 📝 Documented |
| **Behaviour** | *(future)* | *(future)* | *(future)* | 📝 Documented |
| **Collaboration** | *(future)* | *(future)* | *(future)* | 📝 Documented |
| **Career** | *(future)* | *(future)* | *(future)* | 📝 Documented |

---

## Design Principles

Every stage of the intelligence model is governed by the same principles:

- **Deterministic** — identical inputs always yield identical output. No
  randomness, no clock-dependent results in scored paths (the activity signal
  is deterministic *given a reference time*; see the determinism note below).
- **Explainable** — every score can be traced to the facts and weights that
  produced it.
- **Composable** — stages stack cleanly; each consumes only the output of the
  stage above it.
- **Observable** — intermediate stages are inspectable (`InspectProjection`
  exposes raw vestiges, facts, and signals).
- **Inspectable** — there is no "black box"; the full pipeline can be dumped.
- **Evidence-backed** — conclusions reference the observations that support
  them.
- **Layer-owned** — each transformation has exactly one owning package.
- **Transport-independent** — intelligence does not depend on HTTP, the GitHub
  API, or any transport.
- **Presentation-independent** — intelligence does not depend on CLI, JSON, or
  web rendering.

---

## Determinism Note

All derived computations accept an explicit `referenceTime` parameter. The
`activity` indicator is deterministic given the same reference time, removing
any hidden dependency on the system clock. See `facts.FromRepos`,
`indicators.ExtractSignals`, and `indicators.ExtractSignalsFromFacts`.

---

## Historical Evolution

Legacy terminology (e.g. `signals.*` as an umbrella package, "Signal Expansion"
as a roadmap milestone) may appear below for historical accuracy but does not
reflect the current frozen architecture.

### Roadmap (archival)

| Version | Theme | Scope |
|---------|-------|-------|
| **v0.8.14** | Intelligence Ontology | Vocabulary freeze, RepositoryVestige, RepositoryFacts, documentation, certification |
| **v0.8.15** | Deterministic Observation Acquisition | Observation specification, GitHub GraphQL acquisition, deterministic merge, expanded RepositoryVestige |
| **v0.8.16** | Repository Facts | Expanded RepositoryFacts with deterministic derivations |
| **v0.8.17** | Activity Intelligence | ActivityObservation model, ActivityFacts, Profile integration |
| **v0.9.0+** | Candidate Intelligence | All intelligence domains converge toward candidate assessment |

### Release history (archival)

- **v0.8.13** — Rebranded GH-Analyzer → Atlas; consolidated Evaluation.
- **v0.8.14** — Intelligence Ontology release. Canonical vocabulary,
  layer definitions, observation domains.
- **v0.8.15** — Deterministic Observation Acquisition release. GraphQL,
  deterministic merge.
- **v0.8.16** — Repository Intelligence release. Expanded Facts.
- **v0.8.17** — Activity Intelligence release. ActivityObservation,
  ActivityFacts, Profile integration.
