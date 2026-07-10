# GH-Analyzer Architecture

This document is the canonical reference for layer ownership in GH-Analyzer.
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
Facts              signals.Facts
   ↓
Signals            signals.Signals → RawScore
   ↓
Profile            index.Profile (aggregates facts, signals, metadata, contributions)
   ↓
Insights           (future: summaries, highlights, narratives)
   ↓
Evaluation         internal/evaluation (score calibration, confidence, ranking policy)
   ↓
Engine             internal/engine (filtering, matching, searching, ordering)
   ↓
Projection         internal/projection (presentation shapes)
   ↓
Presentation       cmd/gh-analyzer · cmd/server · web
```

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
  - `RepoDTO` → `signals.Repo`
  - `UserDTO` → `profile.UserMetadata`
  - `ContributionsDTO` → `contributions.Summary`
- Timestamp parsing (`created_at` / `updated_at` → `time.Time`, zero-fallback)

**Never performs**

- Network requests
- Ranking, signal computation, or any business logic

Normalization is co-located in `internal/acquisition` (not a separate
package). It is the single boundary between GitHub's representation and the
domain's.

### Facts — `signals.Facts`

**Owns**

- Aggregated repository statistics (counts, sizes, timestamps)
- Deterministic derivation from `[]signals.Repo`

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

### Insights — (future layer)

**Will own**

- Natural-language summaries
- Highlights and strengths/weaknesses
- Evidence aggregation for human consumption
- Narrative observations

**Does not exist yet.** Insights will sit between Profile and Projection.
It will interpret observations into human-readable intelligence while
remaining deterministic (no AI/ML).

### Evaluation — `internal/evaluation`

**Owns**

- Raw weighted score handling
- Score calibration
- Confidence classification (high, moderate, low)
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
- Presentation scoring (overall computation, penalties)
- Repository ordering (top repos by size, relevance, etc.)
- Deterministic formatting and ordering

**Never owns**

- Data acquisition, normalization, or storage
- Intelligence computation (facts, signals, evidence)
- Natural-language generation (deferred to Insights)

Projection is a read-only, deterministic view of the domain for presentation.
It transforms domain data into shapes optimized for specific consumers.

**Projection types:**

- `CandidateProjection` — Minimal listing view (username, display name, signals, facts)
- `AnalyzeProjection` — Deep-dive analysis (overall score, top repos, component scores)
- `InspectProjection` — Raw data inspection (everything + evidence)
- `SearchProjection` — Search results (username, score, confidence, signals, reasons)

### Presentation — `cmd/gh-analyzer`, `cmd/server`, `web`

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
- Human-readable interpretation? → **Insights** (future)
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
- Removed `Report` (deleted `signals.Report`, `signals.Scores`, `signals.BuildReport`)
- Established projection layer with `CandidateProjection`, `AnalyzeProjection`, `InspectProjection`, `SearchProjection`
- All CLI/server presentation paths now consume projections
- Introduced `internal/evaluation` as the single owner of score interpretation
- Removed projection dependency on engine; projection depends only on domain + evaluation
- Documented the future Insights layer (not yet implemented)
- Profile remains the canonical aggregate; projections are presentation shapes
