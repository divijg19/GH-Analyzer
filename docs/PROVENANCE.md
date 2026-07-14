# Provenance

*Status: Normative.*

---

## 1. Definition

**Provenance is the deterministic record of where every Atlas conclusion came
from.** Given any conclusion, Atlas names the repositories, indicators, facts,
and observations that produced it, down to the acquisition source.

Provenance is not observability, logging, or debugging. It is the reasoning
model made explicit. Evidence is the canonical explanation layer; provenance is
the vocabulary Evidence uses to bind the pipeline together.

> **Explanation philosophy.** Every conclusion emitted by Atlas is explainable
> without consulting hidden state, heuristics, or implementation details.
> Everything below follows from this single principle.

> **Explanation Invariant.** No conclusion exists unless Atlas can name exactly
> where it came from. Provenance never introduces information; it only names what
> was already consumed. Provenance leaves every conclusion unchanged — only its
> explainability becomes richer.

## 2. The Reference Model

Provenance is expressed with small, immutable, deterministic references. Every
layer names what it consumed using a shared vocabulary; no layer invents its own.

| Reference | Names | Consumed by |
|-----------|-------|-------------|
| Observation reference | one observation and, optionally, the observed field that mattered, plus its acquisition source | Facts, Repository Intelligence |
| Fact reference | a derived fact by canonical name | Indicators |
| Indicator reference | a normalized indicator by canonical name | Evidence |
| Repository reference | a repository that contributed, and the repository dimension consumed | Candidate Intelligence |
| Chain | the composition of the above for a single conclusion | Evidence |

An observation reference without an identity denotes the whole observation family
of its kind — used by portfolio facts that aggregate every repository
observation. An observation reference with an identity names a single observation.

A chain records only what a conclusion truly consumed; empty layers are absent.
Chains compose deterministically.

## 3. Observation Identity

Every observation carries a deterministic, Atlas-owned identity. Identity is
assigned during normalization and remains stable for identical observations
across executions. GitHub provides acquisition; Atlas owns identity.

- A repository observation is identified by its normalized repository name,
  unique within a candidate's portfolio.
- An activity observation is identified by its kind, its repository (if any), and
  its occurrence instant.

## 4. The Provenance Chain

Each layer records provenance for exactly the layer it consumes, never
duplicating another layer's knowledge:

```
Observation      identified, field-level, carries its acquisition source
        ↓
Fact             names the observation fields it derives from
        ↓
Indicator        names the facts it consumes; reaches observations through them
        ↓
Evidence         carries the chain for the conclusion
        ↓
Repository       names the observation fields each dimension interpreted
Intelligence
        ↓
Candidate        names the repositories and repository dimensions consumed
Intelligence
        ↓
Projection       presents the chain; the graph is navigable via repository
                 references into the repository intelligence views
```

Aggregation reorganizes knowledge; it never flattens it. A candidate dimension
names the repositories and repository dimensions it consumed; following a
repository reference into the repository intelligence view yields that
repository's full observation-level chain without duplicating it.

## 5. Ownership Boundaries

- **Provenance** owns the reference vocabulary only. It has no domain
  dependencies and may be consumed by any layer without a cycle.
- **Facts** own fact-to-observation provenance: which observed fields a fact
  derives from. Facts reference observations; they never duplicate them.
- **Indicators** own indicator-to-fact provenance. Indicators know only facts and
  reach observations transitively. They remain backend-agnostic.
- **Evidence** owns explanation. Every evidence group carries a chain. No layer
  above Evidence defines its own explanation format.
- **Repository Intelligence** consumes observations directly; its provenance is
  observation-level. It introduces no second repository fact layer.
- **Candidate Intelligence** consumes Repository Intelligence; its provenance is
  repository-level.
- **Projection** presents provenance; it never recomputes it.

## 6. Presentation

There is one canonical representation: the chain carried by every evidence group.

- JSON output always exposes the complete deterministic explanation graph.
- Human output surfaces the chains only on request; the default stays concise.

## 7. Invariants

1. **Determinism.** Provenance is a pure function of its inputs. Replay produces
   identical chains.
2. **No new information.** Provenance names only what was consumed.
3. **No duplication.** Each layer records provenance for the layer it consumes;
   deeper provenance is reached by navigation, not copying.
4. **Transport-agnostic.** The acquisition source names the backend without
   coupling any layer to its API surface.
5. **Completeness.** Every conclusion resolves to observations. No explanation
   relies on hidden state.
