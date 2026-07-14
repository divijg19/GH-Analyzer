// Package repositoryintelligence provides the deterministic semantic
// interpretation of a single repository (an observations.RepositoryVestige).
//
// It is the second interpretive layer in the Atlas ontology, sitting above the
// repository facts family and below Candidate Intelligence, which aggregates it
// across a portfolio. See docs/REPOSITORY_INTELLIGENCE.md.
//
// The package MUST NOT import internal/intelligence (the aggregation direction
// is one-way) and MUST NOT call any acquisition backend.
package repositoryintelligence
