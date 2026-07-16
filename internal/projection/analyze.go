package projection

import (
	"github.com/divijg19/Atlas/internal/evaluation"
	"github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/indicators"
)

// analyzeTopRepoLimit is the number of repositories surfaced in the
// AnalyzeProjection's top-repos view.
const analyzeTopRepoLimit = 3

// AnalyzeProjection includes indicator scores and a penalized overall score.
// Used for deep-dive single-user analysis.
type AnalyzeProjection struct {
	Username string              `json:"username"`
	Signals  indicators.RawScore `json:"signals"`
	TopRepos []topRepository     `json:"top_repos"`
	Overall  int                 `json:"overall"`

	// Intelligence is the Candidate Intelligence view, built from the same
	// Profile. Projection renders it; it never recomputes meaning.
	Intelligence []DimensionView `json:"intelligence,omitempty"`
}

// BuildAnalyzeProjection renders the Analyze view from an already-assembled
// canonical Profile. It consumes the Profile's signals (the single canonical
// derivation produced by index.BuildProfile) and never re-derives facts,
// indicators, or evaluation from raw observations. The overall score and small
// sample penalty are the analyze view's scoring policy, applied to the
// canonical signals.
func BuildAnalyzeProjection(p index.Profile) AnalyzeProjection {
	signals := indicators.FromMap(p.Signals)
	scores := indicators.ScoreSignals(signals)

	// Overall score via the evaluation layer (single owner of scoring policy),
	// applied to the canonical signals.
	overall := evaluation.OverallScore(scores)
	overall = evaluation.ApplySmallSamplePenalty(overall, len(p.Repositories))

	topRepos := extractTopRepositories(p.Repositories, analyzeTopRepoLimit)

	return AnalyzeProjection{
		Username: p.Username,
		Signals:  scores,
		TopRepos: topRepos,
		Overall:  overall,
	}
}
