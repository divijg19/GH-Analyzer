// Package projection owns consumer-shaped, read-only presentation
// transformations.
//
// It defines presentation models (SearchProjection, AnalyzeProjection,
// InspectProjection) that reshape domain data for specific consumers.
// Projections never recompute intelligence - they only reorder and
// reformat.
//
// Projection never acquires observations, derives facts, computes
// indicators, generates evidence, or evaluates candidates.
//
// Consumed by: cmd/atlas, cmd/server.
package projection
