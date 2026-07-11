package signals

import "time"

// Facts of the Repository intelligence domain. Currently the only implemented
// fact family; future releases introduce ContributionFacts, TechnologyFacts,
// BehaviourFacts, and CollaborationFacts as documented placeholders.
//
// Facts are deterministic aggregates derived entirely from Vestiges. They never
// observe, fetch, or evaluate — they only count, sum, and classify.
type RepositoryFacts struct {
	TotalRepos         int       `json:"total_repos"`
	OriginalRepos      int       `json:"original_repos"`
	ForkRepos          int       `json:"fork_repos"`
	RecentRepos        int       `json:"recent_repos"`
	DeepRepos          int       `json:"deep_repos"`
	ValidRepos         int       `json:"valid_repos"`
	ValidOriginalRepos int       `json:"valid_original_repos"`
	LargestRepoSize    int       `json:"largest_repo_size"`
	LatestActivity     time.Time `json:"latest_activity"`

	// Repository metadata facts (Phase 9 enrichment). Deterministic aggregates
	// derived from the new observation fields; no new indicators are added.
	ArchivedRepos   int       `json:"archived_repos"`
	TemplateRepos   int       `json:"template_repos"`
	PublicRepos     int       `json:"public_repos"`
	PrivateRepos    int       `json:"private_repos"`
	LicensedRepos   int       `json:"licensed_repos"`
	TotalStars      int       `json:"total_stars"`
	TotalForks      int       `json:"total_forks"`
	TotalWatchers   int       `json:"total_watchers"`
	TotalOpenIssues int       `json:"total_open_issues"`
	TotalTopics     int       `json:"total_topics"`
	OldestCreated   time.Time `json:"oldest_created"`
	NewestCreated   time.Time `json:"newest_created"`
}

// ContributionFacts is a documented placeholder for the Contribution
// intelligence domain. Facts derived from contribution vestiges (pull
// requests, issues, reviews) will be added in a future release.
type ContributionFacts struct{}

// TechnologyFacts is a documented placeholder for the Technology intelligence
// domain. Facts derived from language, framework, and toolchain vestiges will
// be added in a future release.
type TechnologyFacts struct{}

// BehaviourFacts is a documented placeholder for the Behaviour intelligence
// domain. Facts derived from commit cadence, review patterns, and ownership
// evolution will be added in a future release.
type BehaviourFacts struct{}

// CollaborationFacts is a documented placeholder for the Collaboration
// intelligence domain. Facts derived from review metrics, team interactions,
// and contribution distribution will be added in a future release.
type CollaborationFacts struct{}

func FromRepos(repos []RepositoryVestige, referenceTime time.Time) RepositoryFacts {
	if len(repos) == 0 {
		return RepositoryFacts{}
	}

	cutoff := referenceTime.AddDate(0, 0, -recentWindowDays)

	totalRepos := len(repos)
	var originalRepos int
	var forkRepos int
	var recentRepos int
	var deepRepos int
	var validRepos int
	var validOriginalRepos int
	largestRepoSize := 0
	var latestActivity time.Time

	var archivedRepos int
	var templateRepos int
	var publicRepos int
	var privateRepos int
	var licensedRepos int
	var totalStars int
	var totalForks int
	var totalWatchers int
	var totalOpenIssues int
	var totalTopics int
	var oldestCreated time.Time
	var newestCreated time.Time

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

		if repo.Archived {
			archivedRepos++
		}
		if repo.Template {
			templateRepos++
		}
		if repo.Visibility == "public" {
			publicRepos++
		} else if repo.Visibility == "private" {
			privateRepos++
		}
		if repo.License != "" {
			licensedRepos++
		}

		totalStars += repo.Stars
		totalForks += repo.Forks
		totalWatchers += repo.Watchers
		totalOpenIssues += repo.OpenIssues
		totalTopics += len(repo.Topics)

		if !repo.CreatedAt.IsZero() {
			if oldestCreated.IsZero() || repo.CreatedAt.Before(oldestCreated) {
				oldestCreated = repo.CreatedAt
			}
			if repo.CreatedAt.After(newestCreated) {
				newestCreated = repo.CreatedAt
			}
		}
	}

	return RepositoryFacts{
		TotalRepos:         totalRepos,
		OriginalRepos:      originalRepos,
		ForkRepos:          forkRepos,
		RecentRepos:        recentRepos,
		DeepRepos:          deepRepos,
		ValidRepos:         validRepos,
		ValidOriginalRepos: validOriginalRepos,
		LargestRepoSize:    largestRepoSize,
		LatestActivity:     latestActivity,

		ArchivedRepos:   archivedRepos,
		TemplateRepos:   templateRepos,
		PublicRepos:     publicRepos,
		PrivateRepos:    privateRepos,
		LicensedRepos:   licensedRepos,
		TotalStars:      totalStars,
		TotalForks:      totalForks,
		TotalWatchers:   totalWatchers,
		TotalOpenIssues: totalOpenIssues,
		TotalTopics:     totalTopics,
		OldestCreated:   oldestCreated,
		NewestCreated:   newestCreated,
	}
}
