// Package evidence owns deterministic justification for indicators.
//
// It generates ExplainGroups that trace indicator values back to the
// facts and measurements that produced them.
//
// Evidence never computes indicators, evaluates candidates, derives facts,
// or performs presentation.
//
// Invariants:
//   - GenerateEvidence is a pure function of (facts, signals); the group order
//     (ownership, consistency, depth, activity) is fixed and never depends on
//     input ordering.
//
// Consumed by: projection, engine.
package evidence
