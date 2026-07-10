package projection

import (
	"sort"

	"github.com/divijg19/GH-Analyzer/internal/signals"
)

type TopRepository struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// ApplySmallSamplePenalty applies the small sample penalty to a raw overall score.
func ApplySmallSamplePenalty(rawScore int, repoCount int) int {
	const (
		smallSampleRepoThreshold     = 3
		smallSamplePenaltyMultiplier = 7
	)

	if repoCount >= smallSampleRepoThreshold {
		return rawScore
	}

	return (rawScore * smallSamplePenaltyMultiplier) / 10
}

// ExtractTopRepositories returns the top N non-fork repositories sorted by size descending
func ExtractTopRepositories(repos []signals.Repo, limit int) []TopRepository {
	reposBySize := make([]TopRepository, 0, len(repos))

	for _, repo := range repos {
		if repo.Fork {
			continue
		}
		reposBySize = append(reposBySize, TopRepository{
			Name: repo.Name,
			Size: repo.Size,
		})
	}

	sort.Slice(reposBySize, func(i, j int) bool {
		return reposBySize[i].Size > reposBySize[j].Size
	})

	if len(reposBySize) > limit {
		return reposBySize[:limit]
	}
	return reposBySize
}
