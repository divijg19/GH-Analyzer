// Package engine owns search query execution, candidate filtering, matching,
// and result ordering.
//
// Consumes: indicators and evaluation.
//
// Produces: the Query/Condition model, the Execute entry point, distribution
// calibration, and ranked results.
//
// Never owns: observation acquisition, facts derivation, indicator
// computation, evidence, evaluation policy, or presentation.
package engine
