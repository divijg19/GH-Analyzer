# Atlas Intelligence Ontology

The Atlas Intelligence Ontology defines every concept Atlas understands,
following one guiding principle:

> **Atlas observes Observations. Atlas derives Facts. Atlas measures Indicators.
> Atlas performs Evaluation. Atlas projects Intelligence.**

Where [`ARCHITECTURE.md`](./ARCHITECTURE.md) defines layer ownership, this
document defines the concepts and their invariants.

The model is a strictly downward pipeline. Each stage transforms the previous
stage's output into a richer, still-deterministic representation. Nothing is
inferred probabilistically: given the same inputs, Atlas always produces the
same intelligence.

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

### Transport

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Communication with external providers. |
| **Inputs** | Request parameters. |
| **Outputs** | Raw provider payloads. |
| **Invariants** | Never parses domain types. Never interprets responses. |
| **Prohibited** | Domain logic, normalization, caching. |

### Acquisition

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Fetches external data and produces representations that mirror the provider schema. |
| **Inputs** | Candidate and repository identifiers, search queries. |
| **Outputs** | Provider-shaped representations. |
| **Invariants** | Representations are pure data — no intelligence, no computation. Acquisition mechanisms are not intelligence layers. |
| **Prohibited** | Domain model construction, scoring, evaluation, merge policy. |

### Normalization

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Maps provider representations to canonical observations. The boundary between what the provider returns and what Atlas knows. |
| **Inputs** | Provider-shaped representations. |
| **Outputs** | Repository observations, account metadata, contribution summaries. |
| **Invariants** | One representation maps to one observation. No aggregation, no derivation, no silent field loss. |
| **Prohibited** | Aggregation, scoring, evaluation, filtering, merge of multiple sources. |

### Observations

| Attribute | Definition |
|-----------|------------|
| **Purpose** | The canonical raw observation of a software development artifact. |
| **Inputs** | Normalized output from acquisition. |
| **Outputs** | Repository observations, activity observations, account metadata, contribution summaries. |
| **Invariants** | Observations are observables, not computations. |
| **Prohibited** | Derivation, aggregation, scoring, evaluation. |

A repository observation (`RepositoryVestige`) organizes its fields into
observation domains: Identity, Ownership, Timeline, Technology, Maintenance,
Structure. An activity observation (`ActivityObservation`) captures temporal
developer activity by kind. The complete catalogue, ownership, and acquisition
policy are normative in
[`OBSERVATION_SPECIFICATION.md`](./OBSERVATION_SPECIFICATION.md) and
[`ACTIVITY_SPECIFICATION.md`](./ACTIVITY_SPECIFICATION.md).

### Facts

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Deterministic aggregates computed from observations. Facts answer *"what do we know?"* without interpretation. |
| **Inputs** | Repository observations, activity observations. |
| **Outputs** | Repository facts, activity facts. |
| **Invariants** | Purely derivable from observations: given the same observations, identical facts. Never fetch, never observe, never evaluate. |
| **Prohibited** | Acquisition, scoring, evaluation, confidence. |

The complete fact catalogue and every derivation formula are normative in
[`REPOSITORY_FACTS.md`](./REPOSITORY_FACTS.md) and
[`ACTIVITY_INTELLIGENCE.md`](./ACTIVITY_INTELLIGENCE.md).

### Indicators

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Normalized measurements derived from facts. Indicators quantify observations. |
| **Inputs** | Repository facts. |
| **Outputs** | Four measurements in [0, 1]; three component scores in [0, 100]. |
| **Invariants** | Deterministic, reproducible, bounded, explainable, fact-derived. Indicators measure facts — they never read observations directly. |
| **Prohibited** | Acquisition, aggregation, scoring beyond measurement, confidence, ranking. |

The four measurements:

| Measurement | Range | Derives From |
|-------------|-------|--------------|
| Ownership | [0, 1] | share of valid repositories that are original |
| Consistency | [0, 1] | share of original repositories with recent activity |
| Depth | [0, 1] | share of original repositories that are substantial |
| Activity | {0.1, 0.4, 0.7, 1.0} | recency of latest activity |

An indicator is never a heuristic guess, an AI opinion, a reputation score, a
popularity metric, or an evaluation judgment.

### Profile

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Atlas's complete assembled understanding of a candidate. Profile is context, not intelligence. |
| **Inputs** | Repository facts, activity facts, indicators, account metadata, contribution summaries. |
| **Outputs** | The assembled candidate aggregate. |
| **Invariants** | Profile is assembled, not interpreted. It is the single source of truth for what Atlas knows about a candidate before interpretation. |
| **Prohibited** | Scoring, confidence, ranking, presentation, persistence. |

- **Profile** answers: *What do we know about this candidate?*
- **Intelligence** answers: *What does that knowledge mean?*

The interpretation of a Profile is Candidate Intelligence, normative in
[`CANDIDATE_INTELLIGENCE.md`](./CANDIDATE_INTELLIGENCE.md); the interpretation of
a single repository is Repository Intelligence, normative in
[`REPOSITORY_INTELLIGENCE.md`](./REPOSITORY_INTELLIGENCE.md).

### Evaluation

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Score interpretation, confidence classification, penalty application, and ranking policy. Evaluation owns the transition from "what we know" to "how to interpret it." |
| **Inputs** | Component scores, repository count, indicators. |
| **Outputs** | An overall score, a ranking score, a confidence classification. |
| **Invariants** | Never observes, never measures, never aggregates. Only interprets. |
| **Prohibited** | Acquisition, observation, fact derivation, measurement, presentation. |

### Engine

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Search query execution, candidate filtering, matching, and ordering. |
| **Inputs** | The candidate index, a query, a ranking policy. |
| **Outputs** | Filtered and ordered candidates. |
| **Invariants** | Pure computation: no acquisition, no observation. |
| **Prohibited** | Acquisition, scoring policy, presentation, evaluation. |

### Projection

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Consumer-shaped, read-only, deterministic views of the domain. Projection reshapes, never recomputes. |
| **Inputs** | The Profile and evaluation outputs. |
| **Outputs** | Consumer-facing views. |
| **Invariants** | No computation beyond reshaping and ordering. |
| **Prohibited** | Scoring, evaluation, confidence, acquisition, fact derivation, measurement. |

### Consumers

| Attribute | Definition |
|-----------|------------|
| **Purpose** | Presentation surfaces that render intelligence for consumption. |
| **Inputs** | Projections. |
| **Outputs** | Command-line output, API responses, web pages. |
| **Invariants** | Format only; never derive. |
| **Prohibited** | Any derivation of facts, indicators, scores, rankings, or projections. |

---

## Intelligence Domains

Atlas currently defines two intelligence domains: **Repository** and
**Activity**. Each owns a family of observations and facts.

| Domain | Observations | Facts | Measurements |
|--------|--------------|-------|--------------|
| Repository | repository observations | repository facts | Ownership, Consistency, Depth, Activity |
| Activity | activity observations | activity facts | — |

Atlas intentionally does not define Technology, Behaviour, Collaboration, or
Career as distinct domains: their knowledge is either owned by the Repository and
Activity domains or is not yet observable deterministically.

---

## Design Principles

Every stage of the intelligence model is governed by the same principles:

- **Deterministic** — identical inputs always yield identical output. Every
  time-dependent conclusion is deterministic given an explicit reference time.
- **Explainable** — every conclusion traces to the facts and weights that
  produced it.
- **Composable** — stages stack cleanly; each consumes only the output of the
  stage above it.
- **Inspectable** — there is no black box; the full pipeline is observable.
- **Evidence-backed** — every conclusion carries provenance naming the
  indicators, facts, and observations it consumed, so any conclusion can be
  traced end-to-end. See [`PROVENANCE.md`](./PROVENANCE.md).
- **Layer-owned** — each transformation has exactly one owner.
- **Transport-independent** — intelligence does not depend on any transport or
  provider.
- **Presentation-independent** — intelligence does not depend on any consumer
  surface.

All derived computations accept an explicit reference time. Every time-dependent
conclusion — notably the activity measurement — is deterministic given the same
reference time, removing any hidden dependency on the system clock.
