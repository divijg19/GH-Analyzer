// Package indicators owns deterministic measurements derived from facts.
//
// Consumes: facts (RepositoryFacts, ActivityFacts).
//
// Produces: Signals and RawScore — normalized measurements (ownership,
// consistency, depth, activity).
//
// Never owns: observation acquisition, facts derivation, evidence,
// evaluation, or presentation.
//
// Invariants:
//   - ExtractSignals and ExtractSignalsFromFacts are pure functions of
//     (facts, referenceTime). Identical inputs yield identical signals.
//   - Signal values are always clamped to [0, 1]; RawScore values are always
//     in [0, 100].
package indicators
