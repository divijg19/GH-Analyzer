# Atlas Architecture

Atlas is a strictly layered system. Each layer owns a distinct kind of
knowledge and exposes it through a stable contract. Knowledge flows in one
direction: a layer depends only on the layers beneath it, never on the layers
above.

> **Invariant: every layer owns knowledge, not implementation.** A layer is
> defined by the knowledge it owns and the contract it exposes, never by the
> transport, provider, or presentation mechanism it happens to use. Mechanisms
> may change beneath a layer without changing the layer; the knowledge a layer
> owns cannot move without moving the layer.

The conceptual model behind these layers is defined in
[`INTELLIGENCE.md`](./INTELLIGENCE.md).

## The Pipeline

```
GitHub
    ↓
Transport                request execution and authentication
    ↓
Acquisition              provider schema, endpoints, and responses
    ↓
Normalization            provider representation → canonical observations
    ↓
Merge                    one canonical observation per artifact
    ↓
Observations             raw, canonical records of what was observed
    ↓
Facts                    deterministic aggregates of observations
    ↓
Indicators               normalized measurements derived from facts
    ↓
Evidence                 the explanation carried by every conclusion
    ↓
Profile                  the canonical candidate aggregate
    ├─▶ Evaluation                interpretation of scores
    └─▶ Repository Intelligence   interpretation of one repository
            ↓
        Candidate Intelligence    interpretation of a portfolio
    ↓
Projection               read-only views for consumers
    ↓
Presentation             consumer surfaces
```

Provenance is a cross-cutting vocabulary consumed by every layer to name what it
took from the layer beneath it. It is the lowest layer of the explanation model
and depends on nothing else. See [`PROVENANCE.md`](./PROVENANCE.md).

---

## Layer Ownership

### Provenance

**Owns**

- The reference vocabulary that binds every deterministic transformation:
  references to observations, facts, indicators, and repositories, and the chain
  that composes them
- The record of which acquisition source produced an observation, naming the
  provider without depending on it

**Never owns**

- Domain data, derivation, interpretation, or presentation
- Any dependency on another layer

Provenance is a pure vocabulary of small, immutable values. It answers: "How
does one layer name what it consumed from the layer below?" See
[`PROVENANCE.md`](./PROVENANCE.md).

### Transport

**Owns**

- Request execution and lifecycle
- Authentication and request headers

**Never owns**

- Facts, indicators, profiles, contributions, or evidence
- Provider endpoint semantics or schemas
- Presentation

Transport is the only layer permitted to perform raw network I/O. It has no
knowledge of the domain.

### Acquisition

**Owns**

- Provider endpoints, queries, and their schemas
- Provider response handling, pagination, and errors
- Repository and activity acquisition

**Never owns**

- Facts, profiles, or business logic
- Ranking, scoring, or evidence
- Merge policy — merge is an Atlas concern, not a provider concern
- Presentation

Acquisition owns the **provider's schema, not Atlas's**. When the provider
changes its API, only this layer changes; everything above remains untouched.
Acquisition mirrors provider field names and types verbatim; canonical meaning
is assigned only during normalization. Provider values are additive: enrichment
sources augment the baseline and never override it with precedence heuristics.

### Normalization

**Owns**

- Mapping provider representations to canonical observations
- Parsing provider values into canonical types

**Never owns**

- Network requests
- Ranking, measurement, or any business logic
- Merge of multiple acquisition sources

Normalization is the single boundary between the provider's representation and
Atlas's. Every acquired value passes through normalization before reaching the
domain.

### Merge

**Owns**

- Combining values acquired from multiple mechanisms into one canonical
  observation per artifact
- Merge policy, derived from the Observation Specification: each field's owner
  and canonical acquisition mechanism determine which value is authoritative

**Never owns**

- Network requests, provider schemas, or normalization

Merge is Atlas logic, not provider logic. Its policy is determined by the
[Observation Specification](./OBSERVATION_SPECIFICATION.md), not by the
provider's API structure. Merge is additive: an authoritative field has a single
source, so no precedence heuristics are used.

---

### Observations

**Owns**

- The raw, canonical observations recorded from external providers: repository
  observations, activity observations, account metadata, and contribution totals

**Never owns**

- Aggregation, measurement, scoring, or ranking

Observations are the lowest layer of the intelligence model: raw records Atlas
keeps but does not compute. They are the input to Facts. See
[`OBSERVATION_SPECIFICATION.md`](./OBSERVATION_SPECIFICATION.md).

### Facts

**Owns**

- Deterministic aggregates of observations: repository facts and activity facts
- Derivation of those aggregates from observations alone

**Never owns**

- Measurement, scoring, ranking, or presentation

Facts are structured observations. They answer: "What do we know about this
candidate's repositories and activity?" See
[`REPOSITORY_FACTS.md`](./REPOSITORY_FACTS.md) and
[`ACTIVITY_INTELLIGENCE.md`](./ACTIVITY_INTELLIGENCE.md).

### Indicators

**Owns**

- Extraction of the four measurements — ownership, consistency, depth, activity
- Component scoring of those measurements

**Never owns**

- Evidence
- Overall score computation, ranking weights, or penalties
- Network, persistence, or presentation

Indicators are normalized measurements. They answer: "How do we quantify what we
observe?" Indicators know only facts and reach observations transitively; they
remain provider-agnostic.

### Evidence

**Owns**

- The canonical explanation unit, its human-readable items, and the provenance
  chain binding each conclusion to the indicators, facts, and observations that
  produced it

**Never owns**

- Interpretation, scoring, or presentation
- The provenance vocabulary itself

Evidence is Atlas's explanation layer. Every interpretive layer expresses its
reasoning as evidence carrying provenance, so no conclusion exists without a
complete, deterministic explanation. See [`PROVENANCE.md`](./PROVENANCE.md).

### Profile

**Owns**

- The canonical candidate aggregate: facts, indicators, metadata, contributions,
  and activity facts

**Never owns**

- Overall scores, rankings, or evaluations
- Presentation or interpretation

Profile is the single source of truth for what Atlas knows about a candidate
before interpretation. It stores observations, not evaluations.

### Evaluation

**Owns**

- Overall score assembly from component scores and canonical ranking weights
- The small-sample penalty for evaluations built on too few repositories
- Confidence classification
- Ranking policy and weights

**Never owns**

- Search execution, filtering, or ordering
- Presentation or rendering
- Acquisition or normalization

Evaluation is the single source of truth for how scores are interpreted. It
transforms raw scores into evaluations that projection can present directly.

### Repository Intelligence

**Owns**

- The deterministic semantic interpretation of a single repository across its
  dimensions
- Deterministic evidence and summary per dimension, via the dimension contract

**Never owns**

- Acquisition, normalization, fact derivation, measurement, or evaluation
- Aggregation across a portfolio — that is Candidate Intelligence's role
- Any dependency on Candidate Intelligence; the dependency direction is one-way
- Any information not already present in the repository observation

Repository Intelligence is a pure synthesis layer computed per repository. It is
the building block Candidate Intelligence aggregates. See
[`REPOSITORY_INTELLIGENCE.md`](./REPOSITORY_INTELLIGENCE.md).

### Candidate Intelligence

**Owns**

- The deterministic semantic interpretation of a Profile across its dimensions
- Deterministic evidence and summary per dimension, via the dimension contract

**Never owns**

- Acquisition, normalization, fact derivation, measurement, or evaluation
- Ranking, comparison, or presentation
- Any dependency on a provider or transport
- Any information not already present in the Profile

Candidate Intelligence is a pure synthesis layer. It aggregates Repository
Intelligence across a portfolio and reorganizes existing deterministic knowledge
into interpretable dimensions. See
[`CANDIDATE_INTELLIGENCE.md`](./CANDIDATE_INTELLIGENCE.md).

### Engine

**Owns**

- Search query execution, candidate filtering and matching, result ordering

**Never owns**

- Score interpretation or confidence classification
- Presentation or rendering
- Acquisition or normalization

Engine is the orchestration layer for search. It filters, matches, and orders
candidates by query conditions.

### Projection

**Owns**

- Consumer-facing read models and their deterministic ordering and formatting

**Never owns**

- Overall score, penalty, ranking weights, or confidence — those are Evaluation's
- Acquisition, normalization, or storage
- Intelligence computation

Projection is a read-only, deterministic view of the domain for presentation. It
reshapes domain data into forms optimized for specific consumers. Scored values
are supplied by Evaluation and reshaped here; they are never recomputed.

### Presentation

**Owns**

- Consumer surfaces: command-line output, the HTTP API, and the web UI

**Never computes**

- Intelligence. Presentation formats; it does not derive facts, measurements, or
  rankings. All intelligence is produced by the domain and reached through
  projection.

## Documentation Invariant

Atlas's documentation describes contracts, ownership, and invariants. It does
not describe implementations, history, releases, or development process. A
specification names what a layer owns and the contract it exposes; it never
names the packages, files, or mechanisms that satisfy that contract, and it
never narrates how the system reached its present form.
