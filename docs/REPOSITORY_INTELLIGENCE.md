# Repository Intelligence

*Layer: Repository Intelligence (v0.9.0)*
*Status: Normative. Frozen.*
*Owner: `internal/repositoryintelligence`. Canonical consumer: `atlas intelligence` (via Candidate Intelligence aggregation).*

---

## 1. Definition

**Repository Intelligence is the deterministic semantic interpretation of a single
repository (an `observations.RepositoryVestige`).** It answers *what this repository
is and how it behaves* — distinct from *who built it* (Candidate Intelligence) and
*how it changed over time* (Activity Intelligence).

Repository Intelligence is the second interpretive layer in the Atlas ontology. It
sits directly above the repository facts family and below Candidate Intelligence,
which aggregates it across a portfolio. This is the structure GitFut lacked: GitFut
conflated repository traits with candidate traits. Atlas separates them so that the
same repository can be interpreted once and aggregated many times (per owner, per
organization, per project family).

> **Information Invariant.** Repository Intelligence NEVER introduces information.
> Every value it reports is a deterministic function of the `RepositoryVestige`
> (and the reference time). It reorganizes observable repository facts into
> interpretable dimensions; it does not infer facts that were not observed.

## 2. Scope

In scope:
- The intrinsic character of one repository: identity, health, maintenance,
  delivery, architecture, technology, community, documentation, quality, complexity,
  lifecycle, governance, risk.
- Deterministic, replayable derivation at any reference time.

Out of scope (owned by other layers):
- *Who* contributed and in what role → Candidate / Contribution Intelligence.
- *How* the repository evolved over time → Activity Intelligence.
- Aggregation across a portfolio → Candidate Intelligence (see
  `docs/CANDIDATE_INTELLIGENCE.md`).
- Relationship inference between repositories → Project Families (v0.10.x).

## 3. Pipeline Position

```
GitHub → Acquisition → Observations (RepositoryVestige)
                              ↓
                   Facts (facts.RepositoryFacts, per-portfolio aggregate)
                              ↓
        Repository Indicators  →  Repository Evidence  →  Repository Intelligence
                              ↓
                   Candidate Intelligence (aggregation)
                              ↓
                   atlas intelligence → Analyze / Inspect / Search / API
```

Repository Intelligence is computed **per `RepositoryVestige`**. The portfolio
aggregate `facts.RepositoryFacts` remains the canonical fact family for the
repository domain; Repository Intelligence derives its per-repo indicators directly
from the vestige, mirroring how the aggregate is derived, so the two never disagree.

## 4. The Dimension Contract

Every dimension satisfies the same contract as Candidate Intelligence
(`docs/CANDIDATE_INTELLIGENCE.md` §8):

```go
type RepositoryDimension interface {
    Name() string
    Evidence() []evidence.EvidenceGroup
    Summary() string
}
```

- `Name()` returns the stable, lowercase dimension identifier.
- `Evidence()` returns the observable facts that support the conclusion, grouped by
  signal. Evidence traces to the repository observation; no evidence is invented.
- `Summary()` is a deterministic, human-readable sentence derived purely from
  `Evidence()`. Dimension → Evidence → Summary: the summary is a rendering of the
  evidence, never an independent computation.

Every dimension also exposes a `Level` (deterministic qualitative classification)
and a `Confidence` (how much evidence supported the conclusion).

## 5. The Thirteen Dimensions

Repository Intelligence interprets each repository across thirteen dimensions. The
order is canonical.

| # | Dimension | Question answered |
|---|-----------|-------------------|
| 1 | `identity` | Is this repository clearly identified and distinguishable? |
| 2 | `health` | Is this repository alive and receiving attention? |
| 3 | `maintenance` | Is this repository actively maintained? |
| 4 | `delivery` | Does this repository ship releases with discipline? |
| 5 | `architecture` | How is this repository structured? |
| 6 | `technology` | What stack does this repository use? |
| 7 | `community` | How much external engagement does this repository attract? |
| 8 | `documentation` | Is this repository documented? |
| 9 | `quality` | What quality signals does this repository exhibit? |
| 10 | `complexity` | How large and complex is this repository? |
| 11 | `lifecycle` | What maturity stage is this repository in? |
| 12 | `governance` | How is this repository owned and protected? |
| 13 | `risk` | How fragile or at-risk is this repository? |

Each dimension's derivation is fully deterministic from the vestige fields:

- **identity**: presence of `Description`, `Topics` (count), `License`,
  `DefaultBranch`, and non-`template` status.
- **health**: `days since last push` (≤90 active, ≤365 dormant, else inactive) and
  `Stars`, `OpenIssues` as attention signals.
- **maintenance**: `days since last push`, `Archived`, and release recency. A
  repository pushed within 90 days and not archived is *maintained*.
- **delivery**: `ReleaseCount` and `days since latest release`. No releases ⇒
  *undisciplined*; recent releases ⇒ *disciplined*.
- **architecture**: `Size` and `language count`. Single-language small repos are
  *narrow*; multi-language or large repos are *broad*. Empty (`Size == 0`) ⇒ *empty*.
- **technology**: primary language (ranked by bytes) and language breadth. Breadth
  of stack is reported; "modernity" is intentionally not inferred (no evidence).
- **community**: `Forks`, `Watchers`, `CollaboratorCount`, `PullRequestCount`,
  `DiscussionEnabled`. Engagement is reported, not judgmentally scored.
- **documentation**: presence of `Description`, `Topics`, `License`, and
  `DiscussionEnabled`. (README is not part of the vestige and is not inferred.)
- **quality**: `License` present, `DefaultBranchProtected`, `CollaboratorCount > 0`,
  `DiscussionEnabled`. These are observable quality signals only.
- **complexity**: `Size` and language count as scale proxies. LOC is not observed
  and is not approximated.
- **lifecycle**: `age` (days since `CreatedAt`), `Archived`, `ReleaseCount`,
  `Template`. Mature repositories are old, released, and not archived.
- **governance**: `CollaboratorCount`, `DefaultBranchProtected`, `Visibility`,
  `Fork`. Protected branches with collaborators indicate shared governance; forks
  and solo repos indicate individual ownership.
- **risk**: composite of `Archived`, stale push (>365d), `CollaboratorCount <= 1`,
  missing `License`, and `Fork`. Each is a discrete, evidence-backed flag.

## 6. Determinism and Replay

`BuildRepositoryIntelligence(ctx, vestige, referenceTime)` is pure. Identical inputs
yield a byte-identical `RepositoryIntelligence`. All time-relative thresholds use the
supplied `referenceTime`, never `time.Now`, so intelligence is replayable at any
historical point.

## 7. Aggregation Contract (consumed by Candidate Intelligence)

Candidate Intelligence aggregates a slice of `RepositoryIntelligence` (one per
repository) into its own seven portfolio dimensions. The aggregation is defined in
`docs/CANDIDATE_INTELLIGENCE.md` and is the ONLY way portfolio intelligence is
produced: Candidate Intelligence does not re-read `RepositoryVestige` directly for
these dimensions. This keeps a single owner per ontology layer.

Reserved for future layers (not implemented in v0.9.0): Contribution Intelligence,
User Intelligence, Behavior Intelligence. When added, Candidate Intelligence will
aggregate those as well.

## 8. Forbidden Patterns

- Repository Intelligence MUST NOT import `internal/intelligence` (the aggregation
  direction is one-way; this prevents an import cycle).
- Repository Intelligence MUST NOT call the GitHub API or any acquisition backend.
- Repository Intelligence MUST NOT emit a dimension conclusion not backed by
  `Evidence()`.
- Repository Intelligence MUST NOT approximate unobserved properties (e.g. LOC,
  commit counts, CI presence) with weak proxies. Absent data is reported as
  *unknown*, not guessed.
