# Atlas Architecture

This document is the canonical reference for layer ownership in Atlas.
It is the source of truth for where new code belongs. When proposing a change,
first answer: **Which layer owns this?** If the answer is not obvious, do not
implement it yet — revisit the architecture.

The pipeline is strictly downward:

```
GitHub
   ↓
Transport          internal/github
   ↓
Acquisition        internal/acquisition
   ↓
Normalization      internal/acquisition/normalize.go
   ↓
Vestiges           signals.RepositoryVestige · profile.UserMetadata · contributions.Summary
   ↓
Facts              signals.RepositoryFacts
    ↓
Signals            signals.Signals → RawScore
   ↓
Profile            index.Profile (aggregates facts, signals, metadata, contributions)
   ↓
Evaluation         internal/evaluation (overall score, penalty, confidence, ranking policy)
   ↓
Intelligence       Profile + Evaluation output
   ↓
Projection         internal/projection (presentation shapes)
   ↓
Consumers          cmd/atlas · cmd/server · web
```

The conceptual model behind these layers — Vestiges, Facts, Signals,
Profile, Evaluation, Intelligence, Projections, Consumers — is defined in
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

- Facts, Signals, Profiles, Contributions
- Presentation
- GitHub endpoint semantics or DTOs

Transport is the only package permitted to perform raw `http` I/O on behalf of
authentication and execution. It has no knowledge of the domain.

### Acquisition — `internal/acquisition`

**Owns**

- GitHub REST endpoints and their URLs
- GitHub DTOs (mirroring GitHub's JSON schema exactly)
- GitHub endpoint semantics, response handling, pagination
- GitHub API errors (`APIError`)
- The acquisition client (`Client`) used by consumers

**Never owns**

- Signals, Facts, Profiles, business logic
- Ranking, scoring, evidence
- Presentation

Acquisition owns **GitHub's schema, not ours**. If GitHub changes its API,
only this layer changes; everything above remains untouched. DTOs mirror
GitHub field names and types verbatim — timestamps remain raw strings inside
DTOs and are parsed only during normalization.

### Normalization — `internal/acquisition/normalize.go`

**Owns**

- Mapping GitHub DTOs to domain models:
  - `RepoDTO` → `signals.RepositoryVestige`
  - `UserDTO` → `profile.UserMetadata`
  - `ContributionsDTO` → `contributions.Summary`
- Timestamp parsing (`created_at` / `updated_at` → `time.Time`, zero-fallback)

**Never performs**

- Network requests
- Ranking, signal computation, or any business logic

Normalization is co-located in `internal/acquisition` (not a separate
package). It is the single boundary between GitHub's representation and the
domain's.

### Vestiges — `signals.RepositoryVestige`, `profile.UserMetadata`, `contributions.Summary`

**Owns**

- The raw, normalized observations recorded from GitHub
- Repository observations (`signals.RepositoryVestige`), account metadata
  (`profile.UserMetadata`), and contribution totals (`contributions.Summary`)

**Never owns**

- Aggregation, signal extraction, scoring, or ranking

Vestiges are the lowest layer of the intelligence model: observations Atlas
records but does not compute. They are the input to Facts. Phase 9 enriched
`signals.RepositoryVestige` with repository metadata (visibility, archived, template,
license, topics, stars, forks, watchers, open issues, created/pushed dates,
default branch) without changing its role as a raw observation.

### Facts — `signals.RepositoryFacts`

**Owns**

- Aggregated repository statistics (counts, sizes, timestamps)
- Deterministic derivation from `[]signals.RepositoryVestige`

**Never owns**

- Signal computation, scoring, ranking, or presentation

Facts are structured observations. They answer: "What do we know about this
user's repository portfolio?"

### Signals — `signals.Signals`, `signals.RawScore`

**Owns**

- Signal extraction (ownership, consistency, depth, activity)
- Signal scoring (0-100 integer components)
- Evidence generation

**Never owns**

- Overall score computation (that is evaluation policy, not signal extraction)
- Ranking weights or penalties (those are presentation concerns)
- Network, persistence, or presentation

Signals are measurements. They answer: "How do we quantify what we observe?"

**Key distinction:** `RawScore` contains only the three component scores
(ownership, consistency, depth). The overall score is computed by the
projection layer when needed for presentation.

### Profile — `index.Profile`

**Owns**

- The canonical candidate aggregate
- Storage of facts, signals, metadata, contributions

**Never owns**

- Overall scores, rankings, or evaluations
- Presentation or interpretation

Profile is the single source of truth for what we know about a candidate.
It stores observations, not evaluations.

**Signals storage:** Profile stores signals as `map[string]float64` (0-1 float
values). This is the observation-level representation, not the scored
representation.

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

**Post-Phase-6 consolidation (v0.8.13):** overall-score computation, ranking
weights, the `RankingPolicy` type, and the small-sample penalty were moved here
from `internal/projection` and `internal/engine`. `internal/engine/ranking.go`
now defines only the `RankingStrategy` interface; all policy and scoring live
in `internal/evaluation`.

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
- A measurement or signal? → **Signals**
- The candidate aggregate? → **Profile**
- Score interpretation or confidence? → **Evaluation**
- Search execution or filtering? → **Engine**
- A view for a consumer? → **Projection**
- Output text or HTTP response? → **Presentation**

If a change spans multiple layers, it is likely too large — split it. If no
single layer clearly owns it, the architecture must be revisited before code
is written.

---

## Historical Note

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
  `internal/engine/ranking.go` now defines only the `RankingStrategy` interface
  (Phase 6).
- Enriched repository observations and facts with metadata — `RepoDTO`,
  `signals.RepositoryVestige`, `NormalizeRepos`, `signals.RepositoryFacts` — without introducing new
  indicators (Phase 9, Repository Intelligence Foundation).
- Documented the canonical intelligence model in
  [`INTELLIGENCE.md`](./INTELLIGENCE.md): Vestiges, Facts, Indicators, Profile,
  Evaluation, Intelligence, Projections, Consumers.

**v0.8.13 (final conformance & release-freeze pass)** completed simplification
and canonicalization with no new intelligence:

- Canonicalized signal-name constants (`signals.SignalOwnership`,
  `SignalConsistency`, `SignalDepth`, `SignalActivity`) and applied them across
  `signals`, `evaluation`, `search`, `engine`, and `presets`, removing
  duplicated string literals.
- Extracted the single canonical profile builder `index.BuildProfile` (with the
  `index.ProfileFetcher` interface) and rewired both `index.Build` (batch) and
  `live.BuildLiveIndex` (live) onto it; removed the duplicated
  `live.buildLiveProfile`.
- Removed the convenience `search → evaluation` dependency: `search.Search` now
  passes `nil` to `engine.Execute`, which defaults to `evaluation.RankingPolicy{}`.
- Deleted the unused `CandidateProjection` / `BuildCandidateProjection`
  (zero production callers) and its tests.
- Removed the speculative future "Insights" stage from the architecture and
  pipeline diagram; `Profile → Evaluation → … → Projection` is now the
  documented canonical flow.
