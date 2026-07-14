// Package evidence owns deterministic justification for indicators.
//
// Consumes: facts and indicators (signals).
//
// Produces: ExplainGroups that trace indicator values back to the facts and
// measurements that produced them.
//
// Never owns: indicator computation, facts derivation, evaluation, or
// presentation.
//
// Invariants:
//   - GenerateEvidence is a pure function of (facts, signals); the group order
//     (ownership, consistency, depth, activity) is fixed and never depends on
//     input ordering.
package evidence
