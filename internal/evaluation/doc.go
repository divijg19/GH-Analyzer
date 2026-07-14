// Package evaluation owns interpretation of indicators into candidate
// assessment.
//
// Consumes: indicators (signals).
//
// Produces: ranking policy (RankingPolicy), confidence classification
// (ClassifyConfidence), overall score computation (OverallScore), and
// small-sample penalty (ApplySmallSamplePenalty).
//
// Never owns: observation acquisition, facts derivation, indicator
// computation, evidence, or presentation.
//
// Invariants:
//   - RankingPolicy, OverallScore, and ApplySmallSamplePenalty are pure; the
//     ranking weights are the single source of ranking policy.
//   - RankingPolicy.Score takes a signal map, never a profile, so evaluation
//     stays independent of the acquisition and index layers.
package evaluation
