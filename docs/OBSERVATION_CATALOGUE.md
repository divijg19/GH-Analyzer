# Atlas Observation Catalogue v0.8.15

This document is the **frozen canonical observation catalogue** produced by
Phase 2 (Observation Research) of v0.8.15: Deterministic Observation
Acquisition.

It audits three sources:
1. **GitHub REST** — the current acquisition mechanism
2. **GitHub GraphQL** — the proposed secondary acquisition mechanism
3. **GitFut backend concepts** — observations validated by an existing
   developer-profile platform

Every observation candidate is classified as one of:
- **Already Observed** — exists in `RepositoryVestige` (no change needed)
- **Candidate (REST)** — REST can supply; proposed for REST expansion
- **Candidate (GraphQL)** — GraphQL-native; proposed for GraphQL acquisition
- **Future** — valuable but deferred to a later release
- **Reject** — fails acceptance criteria

All candidates accepted for v0.8.15 satisfy:

* deterministic
* publicly observable
* stable
* broadly useful
* provider-independent after normalisation

This catalogue is **frozen**. No additional observations enter v0.8.15 after
this document.

---

## 1. GitHub REST Audit

### Observations currently acquired

| # | Observation     | Status           | Notes                                          |
| - | --------------- | ---------------- | ---------------------------------------------- |
| 1 | Name            | Already Observed | `RepoDTO.Name`                                 |
| 2 | Visibility      | Already Observed | `RepoDTO.Visibility`                           |
| 3 | Archived        | Already Observed | `RepoDTO.Archived`                             |
| 4 | Template        | Already Observed | `RepoDTO.IsTemplate`                           |
| 5 | Fork            | Already Observed | `RepoDTO.Fork`                                 |
| 6 | CreatedAt       | Already Observed | `RepoDTO.CreatedAt` (parsed)                   |
| 7 | UpdatedAt       | Already Observed | `RepoDTO.UpdatedAt` (parsed)                   |
| 8 | PushedAt        | Already Observed | `RepoDTO.PushedAt` (parsed)                    |
| 9 | License         | Already Observed | `RepoDTO.License.Key`                          |
| 10| Topics          | Already Observed | `RepoDTO.Topics`                               |
| 11| DefaultBranch   | Already Observed | `RepoDTO.DefaultBranch`                         |
| 12| OpenIssues      | Already Observed | `RepoDTO.OpenIssues`                           |
| 13| Stars           | Already Observed | `RepoDTO.Stars`                                |
| 14| Forks           | Already Observed | `RepoDTO.Forks`                                |
| 15| Watchers        | Already Observed | `RepoDTO.Watchers`                             |
| 16| Size            | Already Observed | `RepoDTO.Size`                                 |

Observations 1–16 constitute the complete REST-acquired `RepositoryVestige`.

### Observations REST cannot supply (practical limitations)

| Observation          | Limitation                                                   |
| -------------------- | ------------------------------------------------------------ |
| LanguageDistribution | REST returns only `language` scalar (single dominant lang)   |
| Releases             | Separate paginated `/repos/{owner}/{name}/releases` endpoint |
| PullRequestCount     | No direct count; separate paginated endpoint                 |
| DiscussionEnabled    | Not present in REST repository response                      |
| ParentRepository     | Not present in REST repository response                      |
| CollaboratorCount    | Separate paginated `/collaborators` endpoint                 |
| DefaultBranchProtected| Not present in REST repository response                     |

All seven require GraphQL to acquire efficiently at profile scale.

---

## 2. GitHub GraphQL Audit

### GraphQL-native observation candidates

The following observations are available via the GitHub GraphQL API
(`Repository` type on `node` query).

| Observation          | GraphQL Field(s)                                       | Type          | Deterministic | Nullable |
| -------------------- | ------------------------------------------------------ | ------------- | ------------- | -------- |
| LanguageDistribution | `languages(first:10) { edges { size node { name } }`  | `[Edge!]`     | Yes           | Yes      |
| Releases             | `releases { totalCount }`                              | `Int`         | Yes           | Yes      |
| PullRequestCount     | `pullRequests { totalCount }`                          | `Int`         | Yes           | Yes      |
| DiscussionEnabled    | `hasDiscussionsEnabled`                                | `Boolean`     | Yes           | Yes      |
| ParentRepository     | `parent { name owner { login } }`                      | `Repository`  | Yes           | Yes      |
| CollaboratorCount    | `collaborators { totalCount }`                         | `Int`         | Yes           | Yes      |
| DefaultBranchProtected| `branchProtectionRules { totalCount }`                 | `Int`         | Yes           | Yes      |

All seven satisfy the acceptance criteria:

* **Deterministic:** All return stable values derived from repository state.
  No heuristic computation involved.
* **Publicly observable:** All available through the public GitHub GraphQL
  schema without special authentication beyond the standard token.
* **Stable:** All are documented GitHub GraphQL fields with long-term
  stability guarantees.
* **Broadly useful:** All directly support future Facts and Indicators
  (e.g., language diversity → TechnologyFacts, release freshness → Activity).
* **Provider-independent after normalisation:** All normalise to intrinsic
  repository properties (byte map, count, boolean, string).

---

## 3. GitFut Observation Audit

This audit evaluates observations surfaced by the GitFut backend (a
developer-profile platform) against Atlas' ontology. The purpose is not to
port GitFut code but to learn which observations are valuable enough that a
product was built around them.

### Observations already in Atlas

| GitFut Observation       | Atlas Equivalent            | Status           |
| ------------------------ | --------------------------- | ---------------- |
| Repository list          | `RepositoryVestige` fields  | Already Observed |
| Star count               | Stars                       | Already Observed |
| Fork count               | Forks                       | Already Observed |
| Primary language         | License, Topics             | Partial          |
| Contribution summary     | `contributions.Summary`     | Already Observed |

### GitFut observations Atlas should acquire (accepted for v0.8.15)

| GitFut Observation       | Atlas Candidate             | Mechanism | Rationale                                    |
| ------------------------ | --------------------------- | --------- | -------------------------------------------- |
| Language breakdown       | LanguageDistribution        | GraphQL   | Valuable for TechnologyFacts (v0.8.16)       |
| Pull request activity    | PullRequestCount            | GraphQL   | Valuable for Maintenance/Activity signals    |
| Release frequency        | Releases (count + date)     | GraphQL   | Valuable for Activity/Depth signals          |
| Issue activity           | DiscussionEnabled           | GraphQL   | Broadens Maintenance domain                  |
| Collaborator network     | CollaboratorCount           | GraphQL   | Valuable for Ownership evaluation            |
| Fork network             | ParentRepository            | GraphQL   | Valuable for Ownership evaluation            |

### GitFut observations Atlas should not acquire (rejected)

| GitFut Observation       | Rejection Rationale                                           |
| ------------------------ | ------------------------------------------------------------- |
| Social graph / followers  | Not repository-scoped; belongs in profile, not vestige        |
| "Activity score"         | Heuristic computation; Atlas derives signals, not scores      |
| "Impact score"           | Heuristic computation; same reason                            |
| UI/FIFA card layout      | Presentation concern; belongs in projection layer             |
| Badge/achievement system | Gamification; outside Atlas' observation model                |

### GitFut observations deferred to future releases

| GitFut Observation       | Future Release | Notes                                       |
| ------------------------ | -------------- | ------------------------------------------- |
| Commit frequency         | v0.8.17        | ContributionFacts scope                     |
| Issue velocity           | v0.8.17        | ContributionFacts scope                     |
| Review activity          | v0.8.17        | BehaviourFacts scope                        |
| Community health         | v0.8.18        | Signals expansion                           |

---

## 4. Frozen Observation Catalogue — v0.8.15

The following observations are **accepted** for v0.8.15. This catalogue is
frozen. Phase 3 (RepositoryVestige Expansion) and Phase 4 (Acquisition
Completion) implement strictly against this catalogue.

### Existing (Already Observed) — No Change

Observations 1–16 from the REST audit (see section 1). Already in
`RepositoryVestige`. REST-authoritative. No behavioural change.

### Expansion (New for v0.8.15)

| # | Observation            | Domain       | Owner | Canonical Acquisition | Merge Policy          | Fallback           |
| - | ---------------------- | ------------ | ----- | --------------------- | --------------------- | ------------------ |
| 17| LanguageDistribution   | Technology   | Atlas | GraphQL               | GraphQL authoritative | Empty map          |
| 18| Releases               | Timeline     | Atlas | GraphQL               | GraphQL authoritative | Zero count         |
| 19| PullRequestCount       | Maintenance  | Atlas | GraphQL               | GraphQL authoritative | 0                  |
| 20| DiscussionEnabled      | Structure    | Atlas | GraphQL               | GraphQL authoritative | false              |
| 21| ParentRepository       | Ownership    | Atlas | GraphQL               | GraphQL authoritative | Empty string       |
| 22| CollaboratorCount      | Ownership    | Atlas | GraphQL               | GraphQL authoritative | 0                  |
| 23| DefaultBranchProtected | Technology   | Atlas | GraphQL               | GraphQL authoritative | false              |

**Merge justification:** Every observation is GraphQL-authoritative with no
REST fallback because no overlapping REST-supplied field exists. Merge policy
is derived from documented observation ownership, not implementation
heuristics.

**Fallback behaviour:** When GraphQL is unavailable (REST-only mode), each
observation degrades to its documented default. The resulting
`RepositoryVestige` remains valid and downstream layers (`RepositoryFacts`,
`Signals`, `Evaluation`) continue producing correct results from the smaller
observation set.

### Rejected (for v0.8.15)

Observations rejected during audit:

- Social graph / followers — not repository-scoped
- Activity score — heuristic; Atlas derives signals
- Impact score — heuristic
- Badge/achievement system — gamification
- UI/FIFA card layout — presentation concern

### Deferred (to v0.8.16+)

Observations accepted as valuable but deferred:

- Commit frequency → v0.8.17 (ContributionFacts)
- Issue velocity → v0.8.17 (ContributionFacts)
- Review activity → v0.8.17 (BehaviourFacts)
- Community health signals → v0.8.18 (Signals expansion)
- Full branch protection rules → v0.8.16 (extended facts)
- Merge strategy settings → v0.8.16 (extended facts)
- Wiki/pages presence → v0.8.16 (extended facts)
- Package ecosystem → v0.8.16 (extended facts)
- Security settings → v0.8.17+ (low priority)

---

## Certification Gates (Phase 5)

Every observation in the frozen catalogue must satisfy:

1. Owner is Atlas.
2. One documented canonical acquisition mechanism exists.
3. One deterministic normalisation path exists.
4. One explicit merge policy exists (derived from observation ownership).
5. Graceful degradation is documented and implemented.
6. No downstream package references REST, GraphQL, HTTP, DTOs, pagination, or
   provider concepts.
7. Unit tests cover acquisition and normalisation behaviour.

---

## Summary

| Category           | Count | Mechanism      |
| ------------------ | ----- | -------------- |
| Already Observed   | 16    | REST           |
| Expansion (accepted) | 7    | GraphQL        |
| Rejected           | 5     | N/A            |
| Deferred           | 9     | Future release |
| **Total v0.8.15**  | **23**| REST + GraphQL |
