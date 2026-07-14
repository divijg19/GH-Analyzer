// Package intelligence synthesizes a Profile into a CandidateIntelligence.
//
// Candidate Intelligence is the deterministic semantic interpretation of a
// Profile. It is a pure synthesis layer: it never acquires, normalizes, derives
// facts, calculates indicators, evaluates, ranks, or presents. It only
// reorganizes existing deterministic knowledge already present in the Profile.
//
// Invariants (normative, see docs/CANDIDATE_INTELLIGENCE.md):
//   - Deterministic: identical Profile + referenceTime => byte-identical output.
//   - Explainable: every output traces to Evidence.
//   - Compositional: dimensions are independent.
//   - Backend-agnostic: no GitHub, HTTP, GraphQL, or REST.
//   - Information: never introduces information not already in the Profile.
//
// Consumed by: projection.
package intelligence
