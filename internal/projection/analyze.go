package projection

import (
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

// AnalyzeProjection includes signals and penalized overall score.
// Used for deep-dive single-user analysis.
type AnalyzeProjection struct {
	Username string           `json:"username"`
	Signals  signals.RawScore `json:"signals"`
	TopRepos []TopRepository  `json:"top_repos"`
	Overall  int              `json:"overall"`
}

// BuildAnalyzeProjection creates analysis-ready data from repos.
func BuildAnalyzeProjection(username string, repos []signals.Repo) (AnalyzeProjection, error) {
	signalValues := signals.ExtractSignals(repos)
	scores := signals.ScoreSignals(signalValues)

	// Compute overall from raw scores (evaluation policy, not signal extraction)
	overall := computeOverall(scores)

	// Apply small sample penalty
	repoCount := len(repos)
	overall = ApplySmallSamplePenalty(overall, repoCount)

	// Extract top repositories
	topRepos := ExtractTopRepositories(repos, 3)

	return AnalyzeProjection{
		Username: username,
		Signals:  scores,
		TopRepos: topRepos,
		Overall:  overall,
	}, nil
}

// computeOverall calculates the weighted overall score from raw component scores.
// Weights: consistency 40%, ownership 30%, depth 30%.
func computeOverall(rs signals.RawScore) int {
	const (
		consistencyWeight = 0.4
		ownershipWeight   = 0.3
		depthWeight       = 0.3
	)

	return int(float64(rs.Consistency)*consistencyWeight +
		float64(rs.Ownership)*ownershipWeight +
		float64(rs.Depth)*depthWeight)
}
