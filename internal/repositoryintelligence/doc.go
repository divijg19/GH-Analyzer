// Package repositoryintelligence owns the deterministic semantic
// interpretation of a single repository.
//
// Consumes: observations (RepositoryVestige) and the repository facts family.
//
// Produces: RepositoryIntelligence dimensions — the second interpretive layer
// in the Atlas ontology, below Candidate Intelligence, which aggregates it
// across a portfolio. See docs/REPOSITORY_INTELLIGENCE.md.
//
// Never owns: acquisition, or portfolio aggregation. It MUST NOT import the
// Candidate Intelligence layer (the aggregation direction is one-way) and MUST
// NOT call any acquisition backend.
package repositoryintelligence
