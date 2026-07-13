package projection

import (
	"sort"

	"github.com/divijg19/Atlas/internal/observations"
)

type topRepository struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// extractTopRepositories returns the top N non-fork repositories sorted by size descending
func extractTopRepositories(repos []observations.RepositoryVestige, limit int) []topRepository {
	reposBySize := make([]topRepository, 0, len(repos))

	for _, repo := range repos {
		if repo.Fork {
			continue
		}
		reposBySize = append(reposBySize, topRepository{
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
