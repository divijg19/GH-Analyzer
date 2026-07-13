// Package observations owns Atlas' canonical observation model.
//
// It defines immutable observations acquired from external providers:
// RepositoryVestige (repository state), ActivityObservation (temporal
// activity), ActivityKind (observation type enumeration).
//
// Observations never derive facts, compute indicators, evaluate candidates,
// or perform presentation.
//
// Consumed by: facts, indicators, acquisition, projection, index, evidence.
package observations
