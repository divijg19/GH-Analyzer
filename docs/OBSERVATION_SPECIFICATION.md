# Atlas Observation Specification

Every repository observation Atlas understands has a single canonical owner
(Atlas), an acquisition mechanism, a merge policy, and a graceful degradation
behaviour. This specification defines all four for every observation.

`INTELLIGENCE.md` defines the ontology (what intelligence is), `ARCHITECTURE.md`
defines layer ownership, and this specification defines the acquisition contract
(how observations enter Atlas).

## Normative Authority

Observation ownership and merge policy are governed by this specification. The
acquisition layer implements it.

The merge policy for every field is derived from the ownership documented
here, never from implementation heuristics (protocol detection or runtime
branching). Each field's table row defines its single canonical behaviour.

---

## Architectural Invariants

### Observation First

Atlas acquires observations. It does **not** acquire REST responses. It does
**not** acquire GraphQL responses. Protocols exist only to acquire
observations. `observations.RepositoryVestige` is the architectural boundary.

### Atlas Owns Observations

Atlas owns every observation. Providers never own observations. Protocols
never own observations. A new provider requires only new acquisition mappings,
never ontology changes.

### RepositoryVestige = Current Understanding

`observations.RepositoryVestige` represents **every repository observation
Atlas currently understands**. It does not mirror provider schemas. Growth is
deliberate: every new observation must justify its existence.

### Determinism

Equivalent repository state must always produce equivalent `RepositoryVestige`
values regardless of acquisition mechanism.

### Observation Identity

Every observation carries a deterministic, Atlas-owned identity. A repository
observation is identified by its normalized repository name; an activity
observation by its kind, its repository, and its occurrence instant. Identity is
assigned during normalization and remains stable for identical observations
across executions. GitHub provides acquisition; Atlas owns identity. Provenance
references observations by this identity plus a transport-agnostic acquisition
source. See [`PROVENANCE.md`](./PROVENANCE.md).

### Backend Isolation

Everything below the acquisition layer remains unaware of HTTP, REST,
GraphQL, pagination, cursors, or DTOs. Everything above
`RepositoryVestige` remains unaware of acquisition mechanisms.

### Minimal Abstraction

Atlas defines no planners, routing frameworks, provider registries, generic
executors, or speculative provider interfaces. Atlas supports one provider
(GitHub) with two acquisition mechanisms (REST and GraphQL).

---

## Current Observation Inventory

The following observations already exist in the repository observation. Each is
acquired via the GitHub REST backend and normalized within the acquisition
layer.

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

**Canonical acquisition:** REST, owned by the acquisition layer.
**Merge:** All observations are REST-authoritative. Where no GraphQL overlap
exists, merge is trivially the REST value. Merge policy is derived from
documented observation ownership, never from implementation heuristics.

---

## GraphQL-Authoritative Observations

The following observations are acquired via GraphQL, which REST either cannot
supply or cannot supply practically at profile scale. Each satisfies the
acceptance criteria for a canonical observation:

* deterministic
* publicly observable
* stable
* broadly useful
* provider-independent after normalization

| # | Observation            | Domain       | Semantic Purpose                                 | Owner | Canonical Acquisition          | Merge Policy                 | Fallback         | Nullable | Degradation                |
| - | ---------------------- | ------------ | ------------------------------------------------ | ----- | ------------------------------ | ---------------------------- | ---------------- | -------- | -------------------------- |
| 17| LanguageDistribution   | Technology   | Map of language → byte percentage               | Atlas | GraphQL                        | GraphQL authoritative        | Empty map       | Yes      | Empty map if absent        |
| 18a| ReleaseCount           | Timeline     | Total release count                              | Atlas | GraphQL                        | GraphQL authoritative        | 0               | No       | Defaults to 0               |
| 18b| LatestReleaseAt        | Timeline     | Latest release timestamp                         | Atlas | GraphQL                        | GraphQL authoritative        | Zero time       | Yes      | Zero time if no releases    |
| 19| PullRequestCount       | Maintenance  | Total pull requests (open + closed)              | Atlas | GraphQL                        | GraphQL authoritative        | Zero             | No       | Defaults to 0              |
| 20| DiscussionEnabled      | Structure    | Whether the repo has discussions enabled         | Atlas | GraphQL                        | GraphQL authoritative        | false            | No       | Defaults to false          |
| 21| ParentRepository       | Ownership    | Parent repo for forks (owner/login)              | Atlas | GraphQL                        | GraphQL authoritative        | Empty string    | Yes      | Empty string if not a fork |
| 22| CollaboratorCount      | Ownership    | Approximate collaborator count                  | Atlas | GraphQL                        | GraphQL authoritative        | Zero             | No       | Defaults to 0              |
| 23| DefaultBranchProtected | Technology   | Whether the default branch has protection        | Atlas | GraphQL                        | GraphQL authoritative        | false            | No       | Defaults to false          |

### Acquisition rationale

These observations are assigned to GraphQL as the canonical acquisition
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

**Merge:** Every observation in this group is GraphQL-authoritative with no REST
fallback because REST cannot produce it. Merge policy is derived from documented
ownership, never from runtime protocol-detection heuristics.

---

## Observations Atlas Does Not Own

The following observations lie outside Atlas's observation model. They are
recorded here to make the boundary of that model explicit.

### Metadata

| Observation        | Notes                                             |
| ------------------ | ------------------------------------------------- |
| MergeStrategy      | mergeCommitAllowed, rebaseMergeAllowed, etc.     |
| BranchProtection   | Full rule set (not just default-branch boolean)  |
| WikiEnabled        | hasWikiEnabled                                    |
| PagesEnabled       | hasPagesEnabled                                   |
| Packages           | Package count by ecosystem                       |
| ProjectsV2         | Project count                                    |

### Rare / Enterprise / Low-Value

| Observation        | Notes                                             |
| ------------------ | ------------------------------------------------- |
| SecuritySettings   | Only partially public                             |
| Environments       | Deploy environments (enterprise)                  |
| ActionsPresence    | Whether Actions are configured                    |
| Mirror             | Whether the repo is a mirror                      |
| IssuesEnabled      | hasIssuesEnabled (rarely false)                  |

---

## Observation Conformance Invariants

For every repository observation, the following hold:

1. Owner is Atlas.
2. One documented canonical acquisition mechanism exists.
3. One deterministic normalization path exists.
4. One explicit merge policy exists (derived from observation ownership).
5. Graceful degradation is documented and implemented.
6. No downstream layer references REST, GraphQL, HTTP, DTOs, pagination, or
   provider concepts.
7. This specification is the normative source for the observation's ownership
   and merge policy. No implementation-level heuristics or precedence rules
   exist outside this document.

---

## Out of Scope

Acquisition owns none of the layers above it. Their ownership is defined in
`INTELLIGENCE.md` and `ARCHITECTURE.md`; respecting that ownership is part of the
acquisition contract.

- Facts — acquisition introduces no fact families
- Indicators — acquisition introduces no indicators
- Evaluation — acquisition owns no scoring policy
- Profile — acquisition owns no profile assembly
- Projection — acquisition owns no presentation shape
- CLI and server endpoints — acquisition owns no presentation surface
