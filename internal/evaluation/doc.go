// Package evaluation owns interpretation of indicators into candidate
// assessment.
//
// It defines ranking policy (RankingPolicy), confidence classification
// (ClassifyConfidence), overall score computation (OverallScore), and
// small-sample penalty (ApplySmallSamplePenalty).
//
// Evaluation never acquires observations, derives facts, computes indicators,
// generates evidence, or performs presentation.
//
// Invariants:
//   - RankingPolicy, OverallScore, and ApplySmallSamplePenalty are pure; the
//     ranking weights are the single source of ranking policy.
//   - RankingPolicy.Score takes a signal map, never a profile, so evaluation
//     stays independent of the acquisition and index layers.
//
// Consumed by: engine, projection.
package evaluation
