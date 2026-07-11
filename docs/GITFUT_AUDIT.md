# GitFut Intelligence Audit

**Purpose:** Classify every GitFut backend concept into the Atlas Intelligence
Ontology. This document is the blueprint for v0.8.15+ acquisition integration.

**Status:** Conceptual audit based on described GitFut architecture. Must be
validated against the actual GitFut codebase before v0.8.15 implementation.

**Audit date:** v0.8.14 (Intelligence Ontology release)

---

## Classification Rules

| Atlas Layer | Accepts From GitFut | Rejects From GitFut |
|-------------|---------------------|---------------------|
| **Vestige** | Raw observations of repositories, commits, PRs, reviews, contributors, languages | Aggregations, scores, computed metrics |
| **Fact** | Deterministic counts and sums derived from vestiges | Interpretations, rankings, comparisons |
| **Signal** | Nothing directly — signals derive from Facts, not acquisition | Raw scores, AI outputs, heuristic opinions |
| **Evaluation** | Nothing | Scores, confidence, penalties |
| **Projection** | Nothing | Presentation shapes, formatting |
| **Reject** | — | FIFA ratings, UI state, SVG templates, React components, styling, presentation logic |

---

## Concept Mappings

### GraphQL Queries (Acquisition source)

| GitFut Concept | Atlas Mapping | Rationale |
|----------------|---------------|-----------|
| Repository query fields | `RepositoryVestige` — Identity, Timeline, Technology, Maintenance | Raw observations; no computation |
| Language breakdown | `RepositoryVestige` — Technology group (new `Languages` field) | Observable property of a repo |
| Commit history | `RepositoryVestige` — Activity group (new `Commits` field), or future `CommitVestige` | Observable timeline |
| Release list | `RepositoryVestige` — Maintenance group (new `Releases` field) | Observable metadata |
| Discussions | `RepositoryVestige` — Collaboration group (new field) | Observable engagement metric |
| PR counts (total) | `RepositoryFacts` — new fact field (deterministic PR aggregate) | Count, not interpretation |
| Review counts | `RepositoryFacts` — new fact field (review count aggregate) | Count, not interpretation |
| Contributors list | Future `ContributorVestige` | Observable — each contributor is a vestige |
| Branch protection | `RepositoryVestige` — Structure group (new `Protected` field) | Observable boolean |

### Repository Observations

| GitFut Concept | Atlas Mapping | Rationale |
|----------------|---------------|-----------|
| Name, owner, description | RepositoryVestige Identity | Raw observations |
| Fork count, star count | RepositoryVestige Maintenance | Already exists |
| Language stats (byte counts) | RepositoryFacts — new `LanguageBreakdown` field | Deterministic aggregate of vestiges |
| Size, empty status | RepositoryVestige Structure | Already exists (Size) |
| Default branch | RepositoryVestige Technology | Already exists |
| Topics, license | RepositoryVestige Technology | Already exists |
| Created/pushed/updated dates | RepositoryVestige Timeline | Already exists |
| Visibility, archived, template | RepositoryVestige Identity | Already exists |

### Contribution Observations

| GitFut Concept | Atlas Mapping | Rationale |
|----------------|---------------|-----------|
| Total PRs | `RepositoryFacts` — new field | Deterministic count |
| Total issues | `RepositoryFacts` — new field | Already partially exists (OpenIssues is current, not total) |
| Merged PRs | `RepositoryFacts` — new `MergedPRs` field | Deterministic count |
| Review count | `RepositoryFacts` — new `TotalReviews` field | Deterministic count |
| Review comments | `RepositoryFacts` — new `TotalReviewComments` field | Deterministic count |
| PRs by author | Future `ContributionFacts` | Cross-cutting aggregation |
| Reviewers | Future `CollaborationFacts` | Cross-cutting aggregation |

### Commit Observations

| GitFut Concept | Atlas Mapping | Rationale |
|----------------|---------------|-----------|
| Commit count | `RepositoryFacts` — new `TotalCommits` field | Deterministic count |
| Commit frequency | Future `BehaviourFacts` — `CommitCadence` | Derived from commit timestamps |
| Recent commits | `RepositoryFacts` — new `RecentCommits` field | Windowed count |
| First/last commit dates | `RepositoryVestige` — Timeline group | Observable timestamps |
| Commit authors | Future `BehaviourFacts` | Cross-cutting pattern analysis |

### Review Observations

| GitFut Concept | Atlas Mapping | Rationale |
|----------------|---------------|-----------|
| PR review count | `RepositoryFacts` — new field | Deterministic count |
| Review state (approved/changes requested) | `RepositoryFacts` — new field | Deterministic count by state |
| Reviewers per PR | Future `CollaborationFacts` | Cross-cutting aggregation |
| Review comments per PR | Future `CollaborationFacts` | Cross-cutting aggregation |
| Review turnaround time | Future `BehaviourFacts` | Temporal derivation |

### Language Observations

| GitFut Concept | Atlas Mapping | Rationale |
|----------------|---------------|-----------|
| Language byte counts | `RepositoryFacts` — new `LanguageBytes` map field | Deterministic aggregate |
| Primary language | `RepositoryVestige` — Technology group | Observable property |
| Language percentage | `RepositoryFacts` — derived from byte counts | Deterministic derivation |
| Framework detection | Future `TechnologyFacts` | Requires taxonomy, not just observation |

### Concepts Rejected From Atlas

| GitFut Concept | Reason for Rejection |
|----------------|----------------------|
| FIFA rating | Subjective score; not observable, not deterministic |
| FIFA card layout | Presentation concern; not intelligence |
| SVG templates | Presentation concern; not intelligence |
| React components | UI concern; not intelligence |
| Styling/theming | Presentation concern; not intelligence |
| User profile page structure | Presentation concern; not intelligence |
| Search filters UI | Presentation concern; not intelligence |
| Comparison features | Belongs in Projection, not core intelligence |
| Ranking beyond Atlas metrics | Atlas is the ranking authority; GitFut should consume, not produce rankings |

---

## Implementation Priorities (v0.8.15 → v0.9.0)

### v0.8.15 — GraphQL Acquisition

Implement GraphQL client. No new intelligence.

- Full GraphQL query library
- Query fragments for repo, commit, PR, review data
- Batching and pagination support
- Rate-limit management
- DTOs mirroring GraphQL response schema
- Normalization into existing `RepositoryVestige` fields
- REST/GraphQL parity verification

### v0.8.16 — Repository & Technology Facts

Expand existing fact families with GitFut-enriched data.

- `RepositoryFacts` gains: `TotalCommits`, `RecentCommits`, `TotalPRs`, `MergedPRs`, `TotalReviews`, `TotalReviewComments`, `LanguageBytes`
- `RepositoryVestige` gains: `Languages` (byte breakdown), `Description`, `HasWiki`, `HasPages`
- `TechnologyFacts` — first implementation: language distribution, framework detection

### v0.8.17 — Behaviour & Collaboration Facts

New fact families from temporal and social observations.

- `BehaviourFacts` — commit cadence, review turnaround, activity windows, ownership evolution
- `CollaborationFacts` — reviewer distribution, team interaction, contribution spread
- New vestige families: `CommitVestige`, `PullRequestVestige`, `ReviewVestige`

### v0.8.18 — Signal Expansion

New signals from richer facts.

- New signals: Technology Breadth, Collaboration, Review Quality, Maintenance
- Signal philosophy maintained (deterministic, bounded, explainable)
- Existing signals improved via richer fact inputs (no heuristic changes)

### v0.9.0 — Candidate Intelligence

Full convergence.

- All vestige families populated
- All fact families implemented
- Signal set stabilized
- Projection consumes enriched intelligence
- Lattice integration ready
