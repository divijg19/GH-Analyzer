package projection

import (
	"time"

	"github.com/divijg19/Atlas/internal/evaluation"
	"github.com/divijg19/Atlas/internal/indicators"
	"github.com/divijg19/Atlas/internal/observations"
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
}

// BuildAnalyzeProjection creates analysis-ready data from repos.
//
// referenceTime is the explicit "now" used for time-relative signal derivation,
// keeping the projection deterministic for a given repository snapshot.
func BuildAnalyzeProjection(username string, repos []observations.RepositoryVestige, referenceTime time.Time) (AnalyzeProjection, error) {
	signalValues := indicators.ExtractSignals(repos, referenceTime)
	scores := indicators.ScoreSignals(signalValues)

	// Compute overall from raw scores via the evaluation layer (single owner
	// of scoring policy).
	overall := evaluation.OverallScore(scores)

	// Apply small sample penalty
	repoCount := len(repos)
	overall = evaluation.ApplySmallSamplePenalty(overall, repoCount)

	// Extract top repositories
	topRepos := extractTopRepositories(repos, analyzeTopRepoLimit)

	return AnalyzeProjection{
		Username: username,
		Signals:  scores,
		TopRepos: topRepos,
		Overall:  overall,
	}, nil
}
