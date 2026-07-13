// Package profile owns canonical candidate identity and profile information.
//
// It defines UserMetadata: normalized identity observations (name, bio,
// location, company, followers) produced by acquisition/normalization.
//
// Profile never acquires data, derives facts, computes indicators, evaluates
// candidates, or performs presentation.
//
// Consumed by: index, projection.
package profile
