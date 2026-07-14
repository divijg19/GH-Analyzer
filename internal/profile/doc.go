// Package profile owns canonical candidate identity and profile information.
//
// Consumes: identity observations produced by acquisition and normalization.
//
// Produces: UserMetadata — normalized identity (name, bio, location, company,
// followers).
//
// Never owns: acquisition, facts derivation, indicators, evaluation, or
// presentation.
package profile
