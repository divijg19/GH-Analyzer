# Atlas Architecture

This document is the canonical reference for layer ownership in Atlas.
It is the source of truth for where new code belongs. The frozen layers and
their canonical packages are:

| Layer | Package |
|-------|---------|
| Observations | `internal/observations` |
| Facts | `internal/facts` |
| Indicators | `internal/indicators` |
| Evidence | `internal/evidence` |
| Evaluation | `internal/evaluation` |
| Profile | `internal/profile`, `internal/index` |
| Projection | `internal/projection` |

`internal/signals` exists only as a deprecated compatibility shim and owns
nothing.

The pipeline is strictly downward:

```
GitHub
    ↓
Transport              internal/github
    ↓
REST + GraphQL         internal/acquisition
    ↓
Normalization          internal/acquisition/normalize.go
    ↓
Merge                  internal/acquisition/merge.go
    ↓
Observations           observations.RepositoryVestige · observations.ActivityObservation · profile.UserMetadata · contributions.Summary
    ↓
Facts                  facts.RepositoryFacts · facts.ActivityFacts
    ↓
Indicators             indicators.Signals → indicators.RawScore
    ↓
Profile                index.Profile (assembles facts, indicators, metadata, contributions, activity)
    ↓
Evaluation             internal/evaluation (overall score, penalty, confidence, ranking policy)
    ↓
Projection             internal/projection (presentation shapes)
    ↓
Consumers              cmd/atlas · cmd/server · web
```

The conceptual model behind these layers — Observations, Facts, Indicators,
Profile, Evaluation, Intelligence, Projection, Consumers — is defined in
[`INTELLIGENCE.md`](./INTELLIGENCE.md).

---

## Layer Ownership

### Transport — `internal/github`

**Owns**

- HTTP client construction and lifecycle
- Authentication (`GITHUB_TOKEN` → `Authorization: Bearer`)
- Request headers (User-Agent, auth)
- Request execution (`Do`)

**Never owns**

- Facts, Indicators, Profiles, Contributions, Evidence
- Presentation
- GitHub endpoint semantics or DTOs

Transport is the only package permitted to perform raw `http` I/O on behalf of
authentication and execution. It has no knowledge of the domain.

### Acquisition — `internal/acquisition`

**Owns**

- GitHub REST endpoints and their URLs
- GitHub GraphQL queries and endpoint
- GitHub DTOs (mirroring GitHub's JSON schema exactly) for both REST and
  GraphQL (`repositories.go`, `graphql_types.go`)
- GitHub endpoint semantics, response handling, pagination
- GitHub API errors (`APIError`)
- The acquisition client (`Client`) used by consumers
- GraphQL query definitions (`graphql_query.go`, `graphql_activity.go`)
- Per-repo GraphQL execution (`graphql.go`: `queryRepo`, `fetchGraphQLVestiges`)
- Activity acquisition via `contributionsCollection` (`graphql_activity.go`)

**Never owns**

- Signals, Facts, Profiles, business logic
- Ranking, scoring, evidence
- Merge policy (merge is an Atlas concern, not acquisition — see Merge below)
- Presentation

Acquisition owns **GitHub's schema, not ours**. If GitHub changes its API,
only this layer changes; everything above remains untouched. DTOs mirror
GitHub field names and types verbatim — timestamps remain raw strings inside
DTOs and are parsed only during normalization.

Acquisition supports two mechanisms:
- **REST** via `repositories.go` (`FetchRepos`, `FetchReposNormalized`)
- **GraphQL** via `graphql.go` (`fetchGraphQLVestiges`)

Both produce DTOs. Normalization converts DTOs to domain vestiges. The
high-level entry point `FetchReposEnriched` orchestrates REST + GraphQL +
merge to produce fully observed vestiges. When GraphQL is unavailable,
REST-only vestiges are returned — GraphQL enrichment is always additive.

### Normalization — `internal/acquisition/normalize.go`

**Owns**

- Mapping GitHub DTOs to domain models:
  - `RepoDTO` → `observations.RepositoryVestige` (`normalizeRepo`)
  - `graphQLRepo` → `observations.RepositoryVestige` partial (`normalizeGraphQLRepo`)
  - `UserDTO` → `profile.UserMetadata` (`NormalizeUser`)
  - `ContributionsDTO` → `contributions.Summary` (`NormalizeContributions`)
  - Activity DTOs → `observations.ActivityObservation` (`normalizeActivityProfile`, `normalizeYearContrib`)
- Timestamp parsing (`created_at` / `updated_at` → `time.Time`, zero-fallback)

**Never performs**

- Network requests
- Ranking, indicator computation, or any business logic
- Merge of multiple acquisition sources (merge is separate — see Merge below)

Normalization is co-located in `internal/acquisition` (not a separate
package). It is the single boundary between GitHub's representation and the
domain's. Both REST and GraphQL DTOs pass through normalization before
reaching the domain.

### Merge — `internal/acquisition/merge.go`

**Owns**

- Combining REST-derived and GraphQL-derived vestiges into a single final
  vestige (`mergeVestiges`)
- Merge policy derived from `OBSERVATION_SPECIFICATION.md` — each field's
  owner and canonical acquisition mechanism determine which value wins
- Documentation of the specification-derived merge policy for each field

**Never owns**

- Network requests, DTOs, normalization, signal computation

Merge is Atlas logic, not acquisition logic. It is co-located in
`internal/acquisition` because it operates on vestiges produced by
normalization, but its policy is determined by the Observation
Specification, not by GitHub's API structure. Merge is unconditionally
additive for GraphQL-authoritative fields because they have no REST overlap;
no precedence heuristics are used.

---

### Observations — `observations.RepositoryVestige`, `observations.ActivityObservation`, `profile.UserMetadata`, `contributions.Summary`

**Owns**

- The raw, normalized observations recorded from external providers
- Repository observations (`observations.RepositoryVestige`), activity observations
  (`observations.ActivityObservation`), account metadata (`profile.UserMetadata`), and
  contribution totals (`contributions.Summary`)

**Never owns**

- Aggregation, signal extraction, scoring, or ranking

Observations are the lowest layer of the intelligence model: raw data Atlas
records but does not compute. They are the input to Facts. Each observation
type has a single owning package: `internal/observations` owns repository
and activity observations; `internal/profile` owns user metadata;
`internal/contributions` owns contribution summaries.

### Facts — `facts.RepositoryFacts`, `facts.ActivityFacts`

**Owns**

- Aggregated repository statistics (`facts.RepositoryFacts`)
- Aggregated activity statistics (`facts.ActivityFacts`)
- Deterministic derivation from observations
- Repository facts derive from `[]observations.RepositoryVestige`
- Activity facts derive from `[]observations.ActivityObservation`

**Never owns**

- Indicator computation, scoring, ranking, or presentation

Facts are structured observations. They answer: "What do we know about this
user's repository portfolio and activity?"

### Indicators — `indicators.Signals`, `indicators.RawScore`

**Owns**

- Indicator extraction (ownership, consistency, depth, activity)
- Indicator scoring (0-100 integer components)

**Never owns**

- Evidence generation (that is `internal/evidence`)
- Overall score computation (that is evaluation policy, not indicator extraction)
- Ranking weights or penalties (those are evaluation concerns)
- Network, persistence, or presentation

Indicators are normalized measurements. They answer: "How do we quantify what
we observe?"

**Key distinction:** `RawScore` contains only the three component scores
(ownership, consistency, depth). The overall score is computed by the
evaluation layer.

### Profile — `index.Profile`

**Owns**

- The canonical candidate aggregate
- Storage of facts, signals, metadata, contributions, and activity facts

**Never owns**

- Overall scores, rankings, or evaluations
- Presentation or interpretation

Profile is the single source of truth for what we know about a candidate.
It stores observations, not evaluations.

**Signals storage:** Profile stores signals as `map[string]float64` (0-1 float
values). This is the observation-level representation, not the scored
representation.

**Activity storage:** Profile stores ActivityFacts (not raw observations) as
`facts.ActivityFacts`. This keeps the profile compact and provider-independent.

### Evaluation — `internal/evaluation`

**Owns**

- Overall score assembly (`OverallScore`) from the three component scores and
  the canonical ranking weights (ownership `0.3`, consistency `0.4`, depth `0.3`)
- Small-sample penalty (`ApplySmallSamplePenalty`) for evaluations built on too
  few repositories
- Confidence classification (high, moderate, low)
- Ranking policy and weights (`RankingPolicy`)
- Future: percentile mapping, role-fit scoring, candidate scoring

**Never owns**

- Search execution, filtering, or ordering (those are engine concerns)
- Presentation or rendering
- Data acquisition or normalization

Evaluation is the single source of truth for how scores are interpreted.
It transforms raw scores into meaningful evaluations that projection can
present directly.

**Key distinction:** Evaluation owns score interpretation. Engine owns search
execution. Projection owns presentation. Presentation owns rendering.

### Engine — `internal/engine`

**Owns**

- Search query execution
- Candidate filtering and matching
- Result ordering and ranking
- Query parsing and condition matching

**Never owns**

- Score interpretation or confidence classification (that is evaluation)
- Presentation or rendering
- Data acquisition or normalization

Engine is the orchestration layer for search. It filters, matches, and orders
candidates based on query conditions.

### Projection — `internal/projection`

**Owns**

- Consumer-facing read models
- Repository ordering (top repos by size, relevance, etc.)
- Deterministic formatting and ordering

**Never owns**

- Overall score or penalty computation — those are Evaluation's
  (`evaluation.OverallScore`, `evaluation.ApplySmallSamplePenalty`)
- Ranking weights or policy — those are Evaluation's (`evaluation.RankingPolicy`)
- Data acquisition, normalization, or storage
- Intelligence computation (facts, signals, evidence)

Projection is a read-only, deterministic view of the domain for presentation.
It transforms domain data into shapes optimized for specific consumers. Scored
values (overall score, penalties, confidence) are supplied by Evaluation and
re-shaped here; they are never recomputed.

**Projection types:**

- `AnalyzeProjection` — Deep-dive analysis (overall score, top repos, component scores)
- `InspectProjection` — Raw data inspection (everything + evidence)
- `SearchProjection` — Search results (username, score, confidence, signals, reasons)

### Presentation — `cmd/atlas`, `cmd/server`, `web`

**Owns**

- CLI command handling and formatting
- HTTP API surface and JSON shaping
- Web UI

**Never computes**

- Intelligence. Presentation formats; it does not derive facts, signals, or
  rankings. All intelligence is produced by the domain and reached through
  projection.

---

## Decision Rule

Before implementing any change, ask:

> **Which layer owns this?**

- Network call to GitHub? → **Acquisition**
- GitHub JSON → internal model? → **Normalization**
- A fact or observation? → **Facts**
- A measurement or indicator? → **Indicators**
- Evidence or justification? → **Evidence**
- The candidate aggregate? → **Profile**
- Score interpretation or confidence? → **Evaluation**
- Search execution or filtering? → **Engine**
- A view for a consumer? → **Projection**
- Output text or HTTP response? → **Presentation**

If a change spans multiple layers, it is likely too large — split it. If no
single layer clearly owns it, the architecture must be revisited before code
is written.

---

## Historical Evolution

This section documents the evolution of the Atlas architecture. Legacy
terminology (e.g. `signals.*` as an umbrella package) may appear here for
historical accuracy but does not reflect the current frozen architecture.

**v0.8.11** consolidated all GitHub REST access into `internal/acquisition`,
made the domain packages pure, and established the normalization boundary.

**v0.8.12** completed the presentation architecture:
- Removed `Report` (deleted `signals.RepositoryVestigert`, `signals.Scores`, `signals.BuildReport`)
- Established projection layer with `CandidateProjection`, `AnalyzeProjection`, `InspectProjection`, `SearchProjection`
- All CLI/server presentation paths now consume projections
- Introduced `internal/evaluation` as the single owner of score interpretation
- Removed projection dependency on engine; projection depends only on domain + evaluation
- Documented the future Insights layer (not yet implemented)
- Profile remains the canonical aggregate; projections are presentation shapes

**v0.8.13** rebranded the project and consolidated evaluation ownership:

- Renamed the module `github.com/divijg19/GH-Analyzer` →
  `github.com/divijg19/Atlas`; removed the legacy `cmd/gha` and
  `internal/ghanalyzer` packages (Phase 1).
- Consolidated scoring into `internal/evaluation` (`scoring.go`):
  `RankingPolicy`, `OverallScore`, and `ApplySmallSamplePenalty`.
  (Phase 6).
- Enriched repository observations and facts with metadata — `RepoDTO`,
  `RepositoryVestige`, `NormalizeRepos`, `RepositoryFacts` — without introducing new
  indicators (Phase 9, Repository Intelligence Foundation).
- Documented the canonical intelligence model in
  [`INTELLIGENCE.md`](./INTELLIGENCE.md): Observations, Facts, Indicators, Profile,
  Evaluation, Intelligence, Projections, Consumers.

**v0.8.13 (final conformance & release-freeze pass)** completed simplification
and canonicalization with no new intelligence:

- Canonicalized signal-name constants (`SignalOwnership`,
  `SignalConsistency`, `SignalDepth`, `SignalActivity`) and applied them across
  `indicators`, `evaluation`, `search`, `engine`, and `presets`, removing
  duplicated string literals.
- Extracted the single canonical profile builder `index.BuildProfile` (with the
  `index.ProfileFetcher` interface) and rewired onto it.
- Removed the convenience `search → evaluation` dependency: `search.Search` now
  passes `nil` to `engine.Execute`, which defaults to `evaluation.RankingPolicy{}`.
- Deleted the unused `CandidateProjection` / `BuildCandidateProjection`
  (zero production callers) and its tests.
- Removed the speculative future "Insights" stage from the architecture and
  pipeline diagram; `Profile → Evaluation → … → Projection` is now the
  documented canonical flow.

**v0.8.15** completed the observation acquisition architecture:

- Established `docs/OBSERVATION_SPECIFICATION.md` as the normative source for
  observation ownership and merge policy.
- Added GitHub GraphQL acquisition: `graphql_types.go` (DTOs),
  `graphql_query.go` (query), `graphql.go` (executor).
- Extended normalization with `normalizeGraphQLRepo` for GraphQL DTOs.
- Added `merge.go` with specification-driven `mergeVestiges`.
- Enriched `RepositoryVestige` with 8 new GraphQL-authoritative fields.
- REST remains the baseline; GraphQL enrichment is additive with graceful
  degradation.
- No downstream layers (Facts, Indicators, Profile, Evaluation, Engine,
   Projection, Consumers) were modified.

**v0.8.17** introduced Activity Intelligence:
- `observations.ActivityObservation` type with strongly-typed metadata
- GraphQL activity acquisition via `contributionsCollection`
- `facts.ActivityFacts` with deterministic derivations
- Integration into `index.Profile` and `projection.InspectProjection`
- `OBSERVATION_SPECIFICATION.md` and `ACTIVITY_INTELLIGENCE.md` as normative specs

**v0.8.17 (ontology consolidation)** completed the architectural freeze:
- Migrated all canonical logic from `internal/signals` to `internal/observations`,
  `internal/facts`, `internal/indicators`, `internal/evidence`, `internal/evaluation`,
  `internal/profile`, `internal/projection`.
- `internal/signals` became a pure compatibility shim with zero ownership.
- Reduced public API surfaces to only what is consumed externally.
- Aligned all normative documentation to frozen architecture.

