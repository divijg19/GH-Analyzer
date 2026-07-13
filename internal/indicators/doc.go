// Package indicators owns deterministic measurements derived from facts.
//
// It defines Signals and RawScore: normalized measurements (ownership,
// consistency, depth, activity) computed exclusively from facts.
//
// Indicators never acquire observations, derive facts, generate evidence,
// evaluate candidates, or perform presentation.
//
// Invariants:
//   - ExtractSignals and ExtractSignalsFromFacts are pure functions of
//     (facts, referenceTime). Identical inputs yield identical signals.
//   - Signal values are always clamped to [0, 1]; RawScore values are always
//     in [0, 100].
//
// Consumed by: evidence, evaluation, projection, engine, index, search, presets.
package indicators
