// Package intelligence owns the deterministic semantic interpretation of a
// Profile: Candidate Intelligence.
//
// Consumes: a Profile, which carries repository intelligence, facts,
// indicators, evidence, and evaluation.
//
// Produces: CandidateIntelligence dimensions — a pure synthesis that
// reorganizes deterministic knowledge already present in the Profile.
//
// Never owns: acquisition, normalization, facts derivation, indicator
// computation, evaluation, ranking, or presentation.
//
// Invariants (normative, see docs/CANDIDATE_INTELLIGENCE.md):
//   - Deterministic: identical Profile + referenceTime => byte-identical output.
//   - Explainable: every output traces to Evidence.
//   - Compositional: dimensions are independent.
//   - Backend-agnostic: no GitHub, HTTP, GraphQL, or REST.
//   - Information: never introduces information not already in the Profile.
package intelligence
