# Atlas Observation Specification

This document is the **canonical observation acquisition specification** for
Atlas. It defines every repository observation Atlas currently understands,
its single canonical owner (Atlas), its acquisition mechanism, its merge
policy, its normalization path, and its graceful degradation behaviour.

It is the specification for **v0.8.15: Deterministic Observation Acquisition**.

Where `INTELLIGENCE.md` defines the ontology (what intelligence is), and
`ARCHITECTURE.md` defines layer ownership, this document defines the
acquisition contract (how observations enter Atlas). The three are deliberately
separate: ontology is stable, architecture is stable, acquisition evolves.

## Normative Authority

This document is the **normative source** for all observation ownership and
merge policy decisions. The code in `internal/acquisition/` is an
**implementation** of this specification.

If a field's owner, acquisition mechanism, or merge policy changes, this
document **must be updated first**. The code is then updated to match. No
code change that alters observation ownership or merge behaviour is valid
without a corresponding change to this specification.

The merge policy for every field is derived from the ownership documented
here, never from implementation heuristics (`if graphql != nil`, protocol
detection, or runtime branching). Each field's table row defines its
single canonical behaviour. The code implements that behaviour.

---

## Architectural Invariants

### Observation First

Atlas acquires observations. It does **not** acquire REST responses. It does
**not** acquire GraphQL responses. Protocols exist only to acquire
observations. `RepositoryVestige` is the architectural boundary.

### Atlas Owns Observations

Atlas owns every observation. Providers never own observations. Protocols
never own observations. This distinction ensures future providers require only
new acquisition mappings rather than ontology changes.

### RepositoryVestige = Current Understanding

`RepositoryVestige` represents **every observation Atlas currently
understands**. It does not mirror provider schemas. Growth is deliberate:
every new observation must justify its existence.

### Determinism

Equivalent repository state must always produce equivalent `RepositoryVestige`
values regardless of acquisition mechanism.

### Backend Isolation

Everything below `internal/acquisition` remains unaware of HTTP, REST,
GraphQL, pagination, cursors, or DTOs. Everything above
`RepositoryVestige` remains unaware of acquisition mechanisms.

### Minimal Abstraction

No planners, routing frameworks, provider registries, generic executors, or
future-provider interfaces. Atlas currently supports one provider (GitHub)
with two acquisition mechanisms (REST and GraphQL). Design for exactly that.

---

## Current Observation Inventory

The following observations already exist in `RepositoryVestige`. Each is
currently acquired via the GitHub REST executor and normalized in
`internal/acquisition/normalize.go`.

| # | Observation     | Domain       | Semantic Purpose                          | Owner | Canonical Acquisition | Merge Policy          | Fallback           | Nullable | Degradation            |
| - | --------------- | ------------ | ----------------------------------------- | ----- | --------------------- | --------------------- | ------------------ | -------- | ---------------------- |
| 1 | Name            | Identity     | Repository identifier                     | Atlas | REST                  | REST authoritative    | None               | No       | Required               |
| 2 | Visibility      | Identity     | public / private / internal               | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to "public"  |
| 3 | Archived        | Identity     | Archived state flag                      | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to false      |
| 4 | Template        | Identity     | Template repository flag                  | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to false      |
| 5 | Fork            | Ownership    | Whether the repo is a fork               | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to false      |
| 6 | CreatedAt       | Timeline     | Repository creation time                  | Atlas | REST                  | REST authoritative    | None               | Yes      | Zero time if absent    |
| 7 | UpdatedAt       | Timeline     | Last update time                         | Atlas | REST                  | REST authoritative    | None               | Yes      | Zero time if absent    |
| 8 | PushedAt        | Timeline     | Last push time                           | Atlas | REST                  | REST authoritative    | None               | Yes      | Zero time if absent    |
| 9 | License         | Technology   | SPDX license key                         | Atlas | REST                  | REST authoritative    | None               | Yes      | Empty string if absent |
| 10| Topics          | Technology   | Topic tags                               | Atlas | REST                  | REST authoritative    | None               | Yes      | Empty slice if absent  |
| 11| DefaultBranch   | Technology   | Default branch name                      | Atlas | REST                  | REST authoritative    | None               | No       | Empty string if absent |
| 12| OpenIssues      | Maintenance  | Open issue count                         | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to 0          |
| 13| Stars           | Maintenance  | Stargazer count                          | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to 0          |
| 14| Forks           | Maintenance  | Fork count                               | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to 0          |
| 15| Watchers        | Maintenance  | Watcher count                            | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to 0          |
| 16| Size            | Structure    | Repository size in KB                    | Atlas | REST                  | REST authoritative    | None               | No       | Defaults to 0          |

**Canonical acquisition:** REST via `internal/acquisition`.
**Normalization:** `normalizeRepo` in `internal/acquisition/normalize.go`.
**Merge:** All observations are REST-authoritative. No GraphQL overlap exists
yet, so merge is trivially the REST value. Merge policy is derived from
documented observation ownership, never from implementation heuristics.

---

## Tier 1 Expansion Candidates (Under Research)

The following observations are candidates for v0.8.15. Each is being evaluated
in Phase 2 (Observation Research) against the acceptance criteria:

* deterministic
* publicly observable
* stable
* broadly useful
* provider-independent after normalization

They are listed here as **candidates only**. Final acceptance is determined
by the Phase 2 observation audit (GitFut + REST + GraphQL).

| # | Observation            | Domain       | Semantic Purpose                                 | Owner | Proposed Canonical Acquisition | Merge Policy                 | Fallback         | Nullable | Degradation                |
| - | ---------------------- | ------------ | ------------------------------------------------ | ----- | ------------------------------ | ---------------------------- | ---------------- | -------- | -------------------------- |
| 17| LanguageDistribution   | Technology   | Map of language → byte percentage               | Atlas | GraphQL                        | GraphQL authoritative        | Empty map       | Yes      | Empty map if absent        |
| 18a| ReleaseCount           | Timeline     | Total release count                              | Atlas | GraphQL                        | GraphQL authoritative        | 0               | No       | Defaults to 0               |
| 18b| LatestReleaseAt        | Timeline     | Latest release timestamp                         | Atlas | GraphQL                        | GraphQL authoritative        | Zero time       | Yes      | Zero time if no releases    |
| 19| PullRequestCount       | Maintenance  | Total pull requests (open + closed)              | Atlas | GraphQL                        | GraphQL authoritative        | Zero             | No       | Defaults to 0              |
| 20| DiscussionEnabled      | Structure    | Whether the repo has discussions enabled         | Atlas | GraphQL                        | GraphQL authoritative        | false            | No       | Defaults to false          |
| 21| ParentRepository       | Ownership    | Parent repo for forks (owner/login)              | Atlas | GraphQL                        | GraphQL authoritative        | Empty string    | Yes      | Empty string if not a fork |
| 22| CollaboratorCount      | Ownership    | Approximate collaborator count                  | Atlas | GraphQL                        | GraphQL authoritative        | Zero             | No       | Defaults to 0              |
| 23| DefaultBranchProtected | Technology   | Whether the default branch has protection        | Atlas | GraphQL                        | GraphQL authoritative        | false            | No       | Defaults to false          |

### Acquisition rationale (preliminary)

These are provisionally assigned to GraphQL as the canonical acquisition
mechanism because REST either cannot supply the observation or makes it
impractical at profile scale:

- **LanguageDistribution** — REST exposes only a `language` scalar (single
  dominant language) via the repositories endpoint. GraphQL's `languages`
  connection returns byte percentages per language.
- **Releases** — REST requires a separate paginated `/releases` call per
  repository. GraphQL supplies count via a connection field.
- **PullRequestCount** — REST has no direct count without pagination. GraphQL
  supplies totalCount via a connection field.
- **DiscussionEnabled** — Not available in the REST repository object.
  GraphQL provides `hasDiscussionsEnabled`.
- **ParentRepository** — REST repository object does not include parent
  metadata. GraphQL provides a `parent` object.
- **CollaboratorCount** — REST requires a separate paginated call. GraphQL
  supplies totalCount via a connection.
- **DefaultBranchProtected** — Not in the REST repository object. GraphQL
  supplies `branchProtectionRules`.

**Merge:** Every Tier 1 candidate is GraphQL-authoritative with no REST
fallback because REST cannot produce the observation. Merge policy is derived
from documented ownership, never from `if graphql != nil` heuristics.

**The Phase 2 audit (GitFut + REST + GraphQL) will confirm or revise these
assignments.**

---

## Tier 2 / Tier 3 (Deferred, Documented)

The following observations are recognised as potentially valuable but
explicitly deferred. They are documented here so the observation model has a
known frontier; they are **out of scope for v0.8.15**.

### Tier 2 (metadata)

| Observation        | Notes                                             |
| ------------------ | ------------------------------------------------- |
| MergeStrategy      | mergeCommitAllowed, rebaseMergeAllowed, etc.     |
| BranchProtection   | Full rule set (not just default-branch boolean)  |
| WikiEnabled        | hasWikiEnabled                                    |
| PagesEnabled       | hasPagesEnabled                                   |
| Packages           | Package count by ecosystem                       |
| ProjectsV2         | Project count                                    |

### Tier 3 (rare / enterprise / low-value)

| Observation        | Notes                                             |
| ------------------ | ------------------------------------------------- |
| SecuritySettings   | Only partially public                             |
| Environments       | Deploy environments (enterprise)                  |
| ActionsPresence    | Whether Actions are configured                    |
| Mirror             | Whether the repo is a mirror                      |
| IssuesEnabled      | hasIssuesEnabled (rarely false)                  |

---

## Observation Certification Checklist

For every field in `RepositoryVestige` (current + Tier 1), the following must
hold before v0.8.15 is certified:

1. Owner is Atlas.
2. One documented canonical acquisition mechanism exists.
3. One deterministic normalization path exists.
4. One explicit merge policy exists (derived from observation ownership).
5. Graceful degradation is documented and implemented.
6. No downstream package references REST, GraphQL, HTTP, DTOs, pagination, or
   provider concepts.
7. Unit tests cover acquisition and normalization behaviour.
8. This specification is the normative source for the field's ownership and
   merge policy. No implementation-level heuristics or precedence rules exist
   outside this document.

---

## Out of Scope (Frozen Layers)

The following layers are deliberately untouched in v0.8.15. Their owners are
documented in `INTELLIGENCE.md` and `ARCHITECTURE.md`; respecting them is part
of the release discipline.

- `RepositoryFacts` — no new fact families
- `Signals` / `Indicators` — frozen
- `Evaluation` — frozen
- `Engine` / `Search` / `Ranking` — frozen
- `Profile` — frozen
- `Projection` — frozen
- CLI and server endpoints — frozen

v0.8.15 exists solely to complete the observation boundary. It changes nothing
below `RepositoryVestige`.
