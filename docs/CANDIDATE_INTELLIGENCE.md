# Atlas Candidate Intelligence Specification

**Status**: Normative
**Authority**: This is the single source of truth for Candidate Intelligence.
Nothing outside it defines Candidate Intelligence.

---

## 1. Philosophy

Atlas builds a deterministic engineering pipeline: Observations flow into
Facts, Indicators, Evidence, and Evaluation, aggregated into a `Profile`.
Candidate Intelligence graduates that pipeline into a **domain model for
understanding software engineers**.

Candidate Intelligence is the layer where Atlas stops describing *what was
observed* and begins stating *what the observations mean*.

## 2. Definition

> **Candidate Intelligence is the deterministic semantic interpretation of a
> Profile.**

This definition is deliberate. The layer exists to answer *what the Profile
means*, not merely to transform one struct into another. Every output is a
conclusion drawn from knowledge already present in the Profile — never a new
measurement.

## 3. Pipeline

```
Observations → Facts → Indicators → Evidence → Evaluation → Profile
                             ↓
                    Candidate Intelligence   (this layer)
                             ↓
                        Projection            (presentation only)
```

Candidate Intelligence consumes a `Profile` and produces a
`CandidateIntelligence`. It never reaches downward (acquisition, normalization,
fact derivation, indicator calculation, evaluation) or sideways (ranking,
presentation).

## 4. Inputs

The only input is `index.Profile`:

| Field | Type | Role in intelligence |
|-------|------|----------------------|
| `Username` | `string` | Identity |
| `Signals` | `map[string]float64` | Indicators (ownership, consistency, depth, activity, …) |
| `Facts` | `*facts.RepositoryFacts` | Repository domain aggregates |
| `ActivityFacts` | `*facts.ActivityFacts` | Activity domain aggregates |
| `Metadata` | `*profile.UserMetadata` | Account-level observable metadata |
| `Contributions` | `*contributions.Summary` | Contribution summary |

No dimension may read `RepositoryVestige`, `ActivityObservation`, DTOs,
GraphQL, REST, HTTP, or any upstream source. Dimensions receive a `Profile`
already resolved by earlier layers.

## 5. Invariants

Candidate Intelligence is:

- **Deterministic** — identical `Profile` + `referenceTime` ⇒ byte-identical output.
- **Explainable** — every output traces to Evidence; every sentence to a fact/signal.
- **Compositional** — dimensions are independent; no dimension depends on another's output.
- **Backend-agnostic** — no knowledge of GitHub, GraphQL, REST, or transport.
- **Specification-driven** — conforms to this document, not to convenience.
- **Evidence-backed** — every score references `[]evidence.EvidenceGroup`.
- **Reproducible** — replaying at a past `referenceTime` yields the past interpretation.
- **Presentation-independent** — produces no UI, formatting, or rendering.

### 5.1 The Information Invariant (foundational)

> **Candidate Intelligence never introduces information.**
> It only reorganizes existing deterministic knowledge.

Every value a dimension emits must originate in one of:

- Observation
- Fact
- Indicator
- Evidence
- Evaluation

If a proposed builder cannot answer *"Where did this knowledge originate?"*
with one of those five, the implementation is wrong — by construction, not by
review. This invariant is the primary gate for all new dimensions.

## 6. Prohibited Responsibilities

Candidate Intelligence must **never**:

- fetch data (HTTP, GraphQL, REST, filesystem acquisition)
- normalize or transform raw input
- derive Facts
- calculate Indicators / Signals
- perform Evaluation (scoring, weighting, confidence)
- rank candidates or produce comparisons
- generate presentation (UI, templates, SVG, formatting)
- call AI / LLM / embedding services
- depend on GitHub or any provider
- emit heuristic, probabilistic, or non-deterministic output

These responsibilities belong to their respective pipeline layers or to consumers
(search, projection). The intelligence layer is pure synthesis.

## 7. Intelligence Dimensions

### 7.1 Implemented

| # | Dimension | Core question |
|---|-----------|---------------|
| 1 | **Ownership Intelligence** | How much of this portfolio is original rather than forked? |
| 2 | **Delivery Intelligence** | Do the candidate's repositories ship releases with discipline? |
| 3 | **Breadth Intelligence** | How diverse is the candidate's technology footprint? |
| 4 | **Focus Intelligence** | How specialized/concentrated is the candidate around a dominant technology? |
| 5 | **Maintenance Intelligence** | Are the candidate's repositories maintained, not abandoned? |
| 6 | **Project Intelligence** | Does the candidate build coherent, mature projects versus scattered code? |
| 7 | **Collaboration Intelligence** | How does the candidate work with others (collaborators, forks, engagement)? |
| 8 | **Documentation Intelligence** | How well is the candidate's portfolio documented? |
| 9 | **Portfolio Intelligence** | What is the shape and risk profile of the candidate's body of work? |

> **Breadth vs. Focus.** These are deliberately distinct nouns, not two ends of
> one score. Breadth measures *diversity* (how many distinct technologies);
> Focus measures *concentration* (how dominant a single technology is). A
> candidate can be both broad and focused, or neither. Preserving both is the
> kind of semantic distinction this layer exists to capture.

> **Naming note — Portfolio, not Experience.** Atlas observes repositories,
> engineering history, maintenance, continuity, and project evolution. It does
> **not** observe years employed, promotions, company history, or seniority.
> "Experience" would falsely imply career knowledge Atlas lacks. The canonical
> name is **Portfolio Intelligence**.

Each dimension aggregates `RepositoryIntelligence` across `Profile.Repositories`
(see `docs/REPOSITORY_INTELLIGENCE.md`). Candidate Intelligence is a pure
aggregation layer: it does not read repository facts directly. Repository
Intelligence is the single owner of repository interpretation; Candidate
Intelligence combines it into a portfolio view.

### 7.2 Growth Intelligence

Atlas defines Growth Intelligence — how a candidate's portfolio changes over
time — as a dimension of the ontology, but **intentionally does not compute it**.
Growth requires historical observations of the same candidate at different
instants, and Atlas observes a candidate only at a single instant. Atlas does not
substitute a weak point-in-time proxy for a measurement it cannot make from a
single observation. The ontology accommodates the dimension; the evidence for it
does not exist within a single observation.

## 8. Dimension Contract

Every dimension satisfies the same normative contract:

```
Name()     → stable dimension identifier
Evidence() → the deterministic basis for the conclusion (a set of evidence groups)
Summary()  → one deterministic rendering of the evidence
```

The contract is mandatory: every dimension exposes exactly these three
capabilities, and `Summary` is derived solely from `Evidence`.

## 9. Evidence and Summary are distinct products

A dimension is not its summary. The product chain is:

```
Dimension
   ↓
Evidence      (the deterministic facts + indicators the conclusion rests on)
   ↓
Summary       (one deterministic rendering of the Evidence)
```

`Summary` is merely a rendering of `Evidence`, not the intelligence itself.
Projection consumes `Summary` for display; downstream logic consumes `Evidence`
for reasoning. The two must never be conflated.

## 10. Per-Dimension Specification

Each implemented dimension defines, at minimum:

- **Purpose** — one sentence.
- **Inputs** — the `RepositoryIntelligence` dimensions it aggregates (see
  `docs/REPOSITORY_INTELLIGENCE.md`), plus `Metadata` where relevant. Candidate
  Intelligence NEVER reads `RepositoryVestige` facts directly; Repository
  Intelligence is the single owner of repository interpretation.
- **Outputs** — structured fields (score/level where meaningful), `Evidence`, `Summary`, confidence.
- **Derivation** — deterministic aggregation rule; explicit `referenceTime` where
  time-sensitive; no branching on presentation.
- **Invariants** — bounds, determinism, information-origin.
- **Edge cases** — empty portfolio, zero denominators, missing sub-fields.
- **Evidence mapping** — which `EvidenceGroup`s support it.
- **Explainability contract** — template for `Summary`, assembled from Evidence only.
- **Prohibited inputs** — anything outside `Profile` (and, transitively, outside
  the `RepositoryIntelligence` already derived from `Profile.Repositories`).

### 10.1 Ownership Intelligence

- **Purpose**: Quantify how much of the portfolio is original work versus forks.
- **Inputs**: aggregated `RepositoryIntelligence.Governance.Fork` and
  `RepositoryIntelligence.Governance.Collaborators` across the portfolio.
- **Outputs**: originality level, fork ratio, total collaborators, evidence, summary, confidence.
- **Derivation**: originality = original repos / total repos (guard zero); original =
  `!Governance.Fork`. Collaborator breadth = Σ `Governance.Collaborators`.
- **Invariants**: ratio ∈ [0,1]; deterministic; originates in Repository Intelligence.
- **Edge cases**: no repos ⇒ level "unknown", confidence low, summary states absence.
- **Evidence mapping**: ownership evidence group (original, repositories, forks, collaborators).
- **Summary form**: "Ownership is {level} ({pct}% of {n} repositories are original)."

### 10.2 Delivery Intelligence

- **Purpose**: Determine whether the candidate's repositories ship releases with discipline.
- **Inputs**: aggregated `RepositoryIntelligence.Delivery.Level` and
  `RepositoryIntelligence.Delivery.ReleaseCount` across the portfolio.
- **Outputs**: delivery level, disciplined ratio, released repos, total releases, evidence, summary, confidence.
- **Derivation**: released = repos with `ReleaseCount > 0`; disciplined = repos with
  `Delivery.Level == disciplined`. Level from disciplined ratio.
- **Invariants**: deterministic; originates in Repository Intelligence.
- **Edge cases**: no repos ⇒ level "unknown".
- **Evidence mapping**: delivery evidence group (released, repositories, disciplined, total releases).
- **Summary form**: "Delivery is {level} ({n} of {m} repositories ship releases; {k} total)."

### 10.3 Breadth Intelligence

- **Purpose**: Measure diversity of the technology footprint.
- **Inputs**: aggregated `RepositoryIntelligence.Technology.Stack`,
  `RepositoryIntelligence.Technology.LanguageCount` across the portfolio.
- **Outputs**: breadth level, distinct languages, multi-language repos, max languages in a repo, evidence, summary, confidence.
- **Derivation**: distinct languages = union of `Technology.Stack`; multi-language repos =
  count with `LanguageCount >= 2`. Level broad if distinct ≥ 5 or multi-lang ≥ half.
- **Invariants**: deterministic; originates in Repository Intelligence.
- **Edge cases**: no repos ⇒ level "unknown".
- **Evidence mapping**: breadth evidence group (distinct languages, 2+ language repos, max languages).
- **Summary form**: "Breadth is {level} ({k} distinct languages across {n} repositories)."

### 10.4 Focus Intelligence

- **Purpose**: Measure specialization — how concentrated the portfolio is around a
  dominant technology. The semantic complement of Breadth.
- **Inputs**: aggregated `RepositoryIntelligence.Technology.PrimaryLang` across the portfolio.
- **Outputs**: focus level, dominant language, dominant share, distinct languages, evidence, summary, confidence.
- **Derivation**: dominant = most frequent primary language (ties broken lexically);
  share = dominant count / repos with a primary language. Level high if share ≥ 0.6,
  moderate if ≥ 0.35, else low.
- **Invariants**: share ∈ [0,1]; deterministic; originates in Repository Intelligence.
- **Edge cases**: no repos, or none with a primary language ⇒ level "unknown".
- **Evidence mapping**: focus evidence group (dominant language, dominant share, distinct languages, repos with a primary language).
- **Summary form**: "Focus is {level} ({lang} is the primary language in {pct}% of repositories)."

### 10.5 Maintenance Intelligence

- **Purpose**: Assess whether repositories are maintained rather than abandoned.
- **Inputs**: aggregated `RepositoryIntelligence.Maintenance.Level` and
  `RepositoryIntelligence.Maintenance.Archived` across the portfolio.
- **Outputs**: maintenance level, maintained ratio, stale repos, archived repos, evidence, summary, confidence.
- **Derivation**: maintained = repos with `Maintenance.Level == maintained`; stale =
  `maintained == inactive`; archived = `Maintenance.Archived`. Level from maintained ratio.
- **Invariants**: deterministic; originates in Repository Intelligence.
- **Edge cases**: no repos ⇒ level "unknown".
- **Evidence mapping**: maintenance evidence group (maintained, repositories, stale, archived).
- **Summary form**: "Maintenance is {level} ({n} of {m} repositories actively maintained)."

### 10.6 Project Intelligence

- **Purpose**: Distinguish coherent, mature projects from scattered, low-substance code.
- **Inputs**: aggregated `RepositoryIntelligence.Architecture.Size`,
  `RepositoryIntelligence.Lifecycle.Level`, `RepositoryIntelligence.Technology.PrimaryLang`.
- **Outputs**: project level, total size, mature repos, early repos, distinct stacks, evidence, summary, confidence.
- **Derivation**: mature = `Lifecycle.Level == mature`; early = `Lifecycle.Level == early`.
  Level broad if mature ≥ half; narrow if early ≥ half; else moderate.
- **Invariants**: deterministic; originates in Repository Intelligence.
- **Edge cases**: no repos ⇒ level "unknown".
- **Evidence mapping**: project evidence group (total size, mature, early, distinct primary languages).
- **Summary form**: "Project is {level} ({n} mature of {m} repositories; {size}KB total)."

### 10.7 Collaboration Intelligence

- **Purpose**: Characterize how the candidate works with others.
- **Inputs**: aggregated `RepositoryIntelligence.Community.Collaborators`,
  `RepositoryIntelligence.Community.Forks`, `RepositoryIntelligence.Community.Watchers`.
- **Outputs**: collaboration level, total collaborators, repos with collaborators, total forks, total watchers, evidence, summary, confidence.
- **Derivation**: level high if collaborators ≥ 5 or watchers ≥ 50 or forks ≥ 10;
  moderate if any engagement; else low.
- **Invariants**: deterministic; originates in Repository Intelligence.
- **Edge cases**: no repos ⇒ level "unknown"; solo-only account ⇒ level "individual" (not penalized).
- **Evidence mapping**: collaboration evidence group (collaborators, repos with collaborators, forks, watchers).
- **Summary form**: "Collaboration is {level} ({n} collaborators across {k} repositories)."

### 10.8 Documentation Intelligence

- **Purpose**: Assess how well the portfolio is documented.
- **Inputs**: aggregated `RepositoryIntelligence.Documentation.Level` and
  `RepositoryIntelligence.Documentation.HasLicense` across the portfolio.
- **Outputs**: documentation level, documented ratio, documented repos, licensed repos, evidence, summary, confidence.
- **Derivation**: documented = repos with `Documentation.Level` high or moderate.
  Level from documented ratio (high ≥ 0.6, moderate ≥ 0.3, else low).
- **Invariants**: ratio ∈ [0,1]; deterministic; originates in Repository Intelligence.
- **Edge cases**: no repos ⇒ level "unknown". README is not observed and is never inferred.
- **Evidence mapping**: documentation evidence group (documented, repositories, licensed).
- **Summary form**: "Documentation is {level} ({n} of {m} repositories are documented)."

### 10.9 Portfolio Intelligence

- **Purpose**: Describe the candidate's body of work on GitHub.
- **Inputs**: `RepositoryIntelligence.Risk.Level` aggregated across the portfolio, plus `Metadata.CreatedAt`.
- **Outputs**: portfolio level, repository count, at-risk repos, account age, evidence, summary, confidence.
- **Derivation**: at-risk = repos with `Risk.Level == at-risk`. Level broad if ≥ 10
  repos, moderate if ≥ 3, narrow if 1–2, unknown if 0. Account age from `Metadata.CreatedAt`.
- **Invariants**: deterministic; originates in Repository Intelligence + Metadata.
- **Edge cases**: no repos ⇒ level "unknown".
- **Evidence mapping**: portfolio evidence group (repositories, at-risk, account age).
- **Summary form**: "Portfolio is {level} ({n} repositories; {k} at risk)."

> Renamed from "Experience" deliberately — see §7.1 naming note.

#### 10.9.1 Portfolio Intelligence Facts (deterministic aggregates)

The repository facts layer additionally exposes deterministic Portfolio
Intelligence aggregates directly from the observations already collected
(no new acquisition, no scoring). Candidate Intelligence surfaces these as
explanation only; it does not interpret them. See
[`REPOSITORY_FACTS.md` §6](../REPOSITORY_FACTS.md) for the authoritative
contract.

- **Inputs**: `RepositoryVestige.Topics`, `Fork`, `ParentRepository`,
  `CreatedAt`, `LanguageDistribution`, `PushedAt`.
- **Aggregates**: ranked topics (`RankedTopics`, `TopicUniverse`), fork
  lineage concentration (`ForkLineage`), all languages per creation year as a
  lossless observational timeline (`TechnologyTimeline`), and a maintenance
  recency histogram (`MaintenanceBuckets`).
- **Invariants**: deterministic given `referenceTime`; observational aggregates
  only — they answer "what was observed?", not "what does it mean?".
- **Evidence mapping**: a dedicated **Portfolio evidence group** (not an
  indicator signal — signals are reserved for the scored families ownership,
  consistency, depth, activity), with provenance resolving each fact to its
  observation fields (see [`PROVENANCE.md`](./PROVENANCE.md)).
- **Deferred**: Project Families. Real family detection requires combining
  multiple orthogonal observations (naming affinity, shared topics, primary
  language, repository continuity, ownership/origin). The earlier topic-leading
  heuristic was a placeholder and has been removed.

#### 10.10 Topology Intelligence (synthesis)

Topology Intelligence is genuine **synthesis**: it interprets the portfolio's
structural shape from deterministic `RepositoryFacts` rather than re-listing
observations. It answers "is the work concentrated, distributed, or fragmented?"

- **Inputs**: `RepositoryFacts.OriginalRepos`, `TotalRepos`, `MaintenanceBuckets`,
  `ForkLineage`.
- **Synthesis**: `OriginalityRatio` and `MaintenanceShare` (deterministic getters
  on `RepositoryFacts`); `TopUpstreamShare` = share of forks from the single most
  forked upstream; `Concentration` classified from those three values
  (concentrated / distributed / fork-heavy / upstream-concentrated /
  abandoned-leaning).
- **Invariants**: deterministic; derived only from the canonical facts aggregate
  (never raw `RepositoryVestige`); provenance references the contributing
  repositories and their governance/maintenance/health dimensions.
- **Evidence mapping**: `topology` evidence group.

#### 10.11 Technology Intelligence (synthesis)

Technology Intelligence synthesizes the candidate's **technology trajectory**
from the lossless `TechnologyTimeline` — diversification, specialization, and
transitions — without collecting anything new.

- **Inputs**: `RepositoryFacts.TechnologyTimeline` (`year → []language`).
- **Synthesis**: `Diversification` = distinct languages across all years;
  `Specialization` = share of creation years in which the most persistent
  language appears; `Transitions` = deterministic adoptions (first appearance in
  a later year) plus drops (language absent from the latest-year footprint).
- **Invariants**: deterministic; lossless timeline in, deterministic summary out;
  provenance references the contributing repositories and their technology
  dimension.
- **Evidence mapping**: `technology` evidence group.

These two dimensions are the v0.9.5 answer to "is Atlas an intelligence engine?":
they derive new deterministic knowledge from facts Atlas already owned.

## 10.12 Confidence Model

Confidence classifies **how much evidence supported a dimension's conclusion**,
not whether the conclusion is favourable. It is derived deterministically from
sample size (the volume of evidence), not from opinion:

- `low` — empty or tiny sample (≤ 2 repositories / creation years).
- `moderate` — small sample (3–9).
- `high` — substantial sample (≥ 10).

The thresholds are documented calibration constants (`confidenceLowSample`,
`confidenceModerateSample` in `internal/intelligence/explain.go`), not magic
numbers. Confidence is surfaced on every `DimensionView` so consumers can weigh
conclusions by evidence volume. It never alters a Level; it explains the strength
of the supporting observation set.

## 11. Explainability Model

- Every `Summary` is assembled **deterministically** from `Evidence`.
- No LLM, no generated prose, no probabilistic language ("likely", "probably").
- Every sentence is traceable to a specific `Evidence` item.
- Summary templates are fixed strings with value interpolation.
- Projection may reformat `Summary` for display but must not alter its meaning.

**Provenance.** Every candidate dimension's evidence records which repositories
contributed and which repository intelligence dimensions were consumed.
Aggregation reorganizes knowledge; it never flattens it. Deeper provenance —
each repository's supporting evidence, indicators, facts, and observations — is
reached by navigating from a repository reference into the repository
intelligence view, so no provenance is duplicated at the portfolio level. See
[`PROVENANCE.md`](./PROVENANCE.md).

## 12. Out of Scope

Candidate Intelligence does not own:

- New Observations, Facts, Indicators, or Evaluation policy.
- Acquisition, or any provider or transport concern.
- Multi-provider support.
- AI, embeddings, or probabilistic ranking.
- Presentation systems.

Atlas intentionally does not compute Growth Intelligence: it requires historical
observations Atlas does not retain (see §7.2).
