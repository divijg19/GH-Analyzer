// Package provenance owns the canonical, transport-agnostic reference types
// that bind every deterministic transformation in the Atlas pipeline.
//
// Consumes: nothing. Provenance is the lowest layer of the explanation model
// and has no domain dependencies, so any package may import it without creating
// a cycle.
//
// Produces: small immutable reference value types and Chain, which name where a
// conclusion came from — which observations, facts, indicators, and
// repositories contributed — without duplicating the underlying data.
//
// Never owns: the underlying domain data. Provenance references it, never
// copies it. See docs/PROVENANCE.md.
package provenance
