// Package engine owns search query execution, candidate filtering, matching,
// and result ordering.
//
// It defines the Query/Condition model, the Execute entry point, distribution
// calibration, and the rankingStrategy interface. Engine orchestrates the
// search path from query to ranked results.
//
// Engine never acquires observations, derives facts, computes indicators,
// generates evidence, evaluates candidates, or performs presentation.
//
// Consumed by: search, cmd/atlas.
package engine
