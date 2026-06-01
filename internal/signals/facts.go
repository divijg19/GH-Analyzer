package signals

import "time"

type Facts struct {
	TotalRepos         int       `json:"total_repos"`
	OriginalRepos      int       `json:"original_repos"`
	ForkRepos          int       `json:"fork_repos"`
	RecentRepos        int       `json:"recent_repos"`
	DeepRepos          int       `json:"deep_repos"`
	ValidRepos         int       `json:"valid_repos"`
	ValidOriginalRepos int       `json:"valid_original_repos"`
	LargestRepoSize    int       `json:"largest_repo_size"`
	LatestActivity     time.Time `json:"latest_activity"`
}

func FromRepos(repos []Repo) Facts {
	if len(repos) == 0 {
		return Facts{}
	}

	now := time.Now()
	cutoff := now.AddDate(0, 0, -recentWindowDays)

	totalRepos := len(repos)
	var originalRepos int
	var forkRepos int
	var recentRepos int
	var deepRepos int
	var validRepos int
	var validOriginalRepos int
	largestRepoSize := 0
	var latestActivity time.Time

	for i, repo := range repos {
		if repo.Size > 0 {
			validRepos++
			if !repo.Fork {
				validOriginalRepos++
			}
		}

		if repo.Fork {
			forkRepos++
		} else {
			originalRepos++
			if !repo.UpdatedAt.Before(cutoff) {
				recentRepos++
			}
			if repo.Size >= minDepthRepoSize {
				deepRepos++
			}
		}

		if repo.Size > largestRepoSize {
			largestRepoSize = repo.Size
		}

		if i == 0 || repo.UpdatedAt.After(latestActivity) {
			latestActivity = repo.UpdatedAt
		}
	}

	return Facts{
		TotalRepos:         totalRepos,
		OriginalRepos:      originalRepos,
		ForkRepos:          forkRepos,
		RecentRepos:        recentRepos,
		DeepRepos:          deepRepos,
		ValidRepos:         validRepos,
		ValidOriginalRepos: validOriginalRepos,
		LargestRepoSize:    largestRepoSize,
		LatestActivity:     latestActivity,
	}
}
