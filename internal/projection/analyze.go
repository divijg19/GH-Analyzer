// Package projection defines consumer-shaped, read-only views of the domain.
//
// It owns deterministic re-shaping and ordering for presentation (Candidate,
// Analyze, Inspect, Search projections). It never computes overall scores,
// penalties, or confidence — those are supplied by internal/evaluation.
// See docs/INTELLIGENCE.md.
package projection

import (
	"time"

	"github.com/divijg19/Atlas/internal/evaluation"
	"github.com/divijg19/Atlas/internal/signals"
)

// analyzeTopRepoLimit is the number of repositories surfaced in the
// AnalyzeProjection's top-repos view.
const analyzeTopRepoLimit = 3

// AnalyzeProjection includes signals and penalized overall score.
// Used for deep-dive single-user analysis.
type AnalyzeProjection struct {
	Username string           `json:"username"`
	Signals  signals.RawScore `json:"signals"`
	TopRepos []TopRepository  `json:"top_repos"`
	Overall  int              `json:"overall"`
}

// BuildAnalyzeProjection creates analysis-ready data from repos.
func BuildAnalyzeProjection(username string, repos []signals.RepositoryVestige) (AnalyzeProjection, error) {
	signalValues := signals.ExtractSignals(repos, time.Now())
	scores := signals.ScoreSignals(signalValues)

	// Compute overall from raw scores via the evaluation layer (single owner
	// of scoring policy).
	overall := evaluation.OverallScore(scores)

	// Apply small sample penalty
	repoCount := len(repos)
	overall = evaluation.ApplySmallSamplePenalty(overall, repoCount)

	// Extract top repositories
	topRepos := ExtractTopRepositories(repos, analyzeTopRepoLimit)

	return AnalyzeProjection{
		Username: username,
		Signals:  scores,
		TopRepos: topRepos,
		Overall:  overall,
	}, nil
}
