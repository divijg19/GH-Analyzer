package projection

import (
	"sort"

	"github.com/divijg19/Atlas/internal/signals"
)

type TopRepository struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// ExtractTopRepositories returns the top N non-fork repositories sorted by size descending
func ExtractTopRepositories(repos []signals.RepositoryVestige, limit int) []TopRepository {
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
