// Package projection owns consumer-shaped, read-only presentation
// transformations.
//
// Consumes: domain outputs (profile, intelligence, evaluation, evidence).
//
// Produces: presentation models (SearchProjection, AnalyzeProjection,
// InspectProjection) that reshape domain data for specific consumers.
//
// Never owns: acquisition, facts derivation, indicator computation, evidence,
// evaluation, or intelligence computation. Projections reorder and reformat;
// they never recompute intelligence.
package projection
