// Package observations owns Atlas' canonical observation model.
//
// Consumes: nothing. Observations are the lowest domain layer; they are
// acquired externally and normalized into canonical form before entry.
//
// Produces: RepositoryVestige (repository state), ActivityObservation
// (temporal activity), and ActivityKind (observation type).
//
// Never owns: facts derivation, indicators, evaluation, or presentation.
package observations
