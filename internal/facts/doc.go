// Package facts owns deterministic derivations from observations.
//
// Consumes: observations (RepositoryVestige, ActivityObservation).
//
// Produces: RepositoryFacts and ActivityFacts — aggregates computed
// exclusively from observations.
//
// Never owns: acquisition, indicators, evaluation, or explanation.
//
// Invariants:
//   - FromRepos and ActivityFactsFromObservations are pure: identical
//     observations plus an identical referenceTime always yield byte-identical
//     facts. No clock, randomness, or hidden state participates.
//   - Facts never store scores, confidence, or presentation data.
package facts
