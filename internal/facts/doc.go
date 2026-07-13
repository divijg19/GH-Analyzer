// Package facts owns deterministic derivations from observations.
//
// It defines RepositoryFacts and ActivityFacts: aggregates computed
// exclusively from observations.RepositoryVestige and
// observations.ActivityObservation respectively.
//
// Facts never acquire data, compute indicators, evaluate policy, or
// explain results.
//
// Invariants:
//   - FromRepos and ActivityFactsFromObservations are pure: identical
//     vestiges plus an identical referenceTime always yield byte-identical
//     facts. No clock, randomness, or hidden state participates.
//   - Facts never store scores, confidence, or presentation data.
//
// Consumed by: indicators, evidence, index, projection.
package facts
