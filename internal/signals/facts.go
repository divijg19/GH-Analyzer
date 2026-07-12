package signals

import (
	"sort"
	"time"
)

// Facts of the Repository intelligence domain. Currently the only implemented
// fact family; future releases introduce ContributionFacts, TechnologyFacts,
// BehaviourFacts, and CollaborationFacts as documented placeholders.
//
// Facts are deterministic aggregates derived entirely from Vestiges. They never
// observe, fetch, or evaluate — they only count, sum, classify, and normalize.
//
// Two semantic classes of facts coexist in this struct (distinguished inline
// below). Both are derived solely from RepositoryVestige; the distinction is
// documentary, not structural:
//
//   - Aggregate observations — sums, counts, maxima, and extrema of raw
//     vestige fields (e.g. TotalPullRequests, ProtectedBranchRepos,
//     MaxRepoStars). They answer "how many / how much".
//   - Derived intelligence — ratios, means, and age spans that normalize or
//     relate observations (e.g. ForkRatio, MeanRepoStars, PortfolioAgeDays).
//     They answer "what does the aggregate imply".
//
// Indicators (Ownership/Consistency/Depth/Activity) are computed separately in
// analyzer.go and never recompute these facts. See
// docs/REPOSITORY_INTELLIGENCE.md for the normative derivation contract.
type RepositoryFacts struct {
	// --- Portfolio size (aggregate observations) ---
	TotalRepos         int       `json:"total_repos"`
	OriginalRepos      int       `json:"original_repos"`
	ForkRepos          int       `json:"fork_repos"`
	RecentRepos        int       `json:"recent_repos"`
	DeepRepos          int       `json:"deep_repos"`
	ValidRepos         int       `json:"valid_repos"`
	ValidOriginalRepos int       `json:"valid_original_repos"`
	LargestRepoSize    int       `json:"largest_repo_size"`
	LatestActivity     time.Time `json:"latest_activity"`

	// --- Repository composition (aggregate observations) ---
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

	// --- Repository maintenance (aggregate observations) ---
	TotalPullRequests    int `json:"total_pull_requests"`
	TotalCollaborators   int `json:"total_collaborators"`
	ProtectedBranchRepos int `json:"protected_branch_repos"`
	DiscussionRepos      int `json:"discussion_repos"`

	// --- Repository releases (aggregate observations) ---
	TotalReleases   int       `json:"total_releases"`
	ReleasedRepos   int       `json:"released_repos"`
	LatestReleaseAt time.Time `json:"latest_release_at"`

	// --- Technology (derived intelligence from LanguageDistribution) ---
	// RankedLanguages is ordered by aggregate byte share descending; ties break
	// lexicographically by language name. The ordering is a total order, so it
	// is deterministic across Go versions and independent of map iteration.
	LanguageCount   int      `json:"language_count"`
	RankedLanguages []string `json:"ranked_languages"`

	// --- Star distribution (derived intelligence from Stars) ---
	MaxRepoStars  int     `json:"max_repo_stars"`  // extremum aggregate
	MeanRepoStars float64 `json:"mean_repo_stars"` // mean (TotalStars/TotalRepos)

	// --- Ratios & averages (derived intelligence, deterministic in [0, 1]) ---
	ForkRatio     float64 `json:"fork_ratio"`     // ForkRepos/TotalRepos
	LicensedRatio float64 `json:"licensed_ratio"` // LicensedRepos/TotalRepos
	ArchivedRatio float64 `json:"archived_ratio"` // ArchivedRepos/TotalRepos
	MeanRepoSize  float64 `json:"mean_repo_size"` // mean Size
	TopicBreadth  float64 `json:"topic_breadth"`  // TotalTopics/TotalRepos

	// --- Age & freshness (derived intelligence; deterministic given referenceTime) ---
	PortfolioAgeDays       int `json:"portfolio_age_days"`        // referenceTime - OldestCreated
	NewestRepoAgeDays      int `json:"newest_repo_age_days"`      // referenceTime - NewestCreated
	DaysSinceLatestRelease int `json:"days_since_latest_release"` // referenceTime - LatestReleaseAt
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

	var maxRepoStars int
	var totalSize int64
	languageBytes := make(map[string]int64)

	var totalReleases int
	var latestReleaseAt time.Time
	var releasedRepos int
	var totalPullRequests int
	var totalCollaborators int
	var protectedBranchRepos int
	var discussionRepos int

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

		if repo.Stars > maxRepoStars {
			maxRepoStars = repo.Stars
		}
		totalSize += int64(repo.Size)

		for lang, bytes := range repo.LanguageDistribution {
			languageBytes[lang] += bytes
		}

		totalReleases += repo.ReleaseCount
		if repo.ReleaseCount > 0 {
			releasedRepos++
		}
		if !repo.LatestReleaseAt.IsZero() && repo.LatestReleaseAt.After(latestReleaseAt) {
			latestReleaseAt = repo.LatestReleaseAt
		}
		totalPullRequests += repo.PullRequestCount
		totalCollaborators += repo.CollaboratorCount
		if repo.DefaultBranchProtected {
			protectedBranchRepos++
		}
		if repo.DiscussionEnabled {
			discussionRepos++
		}
	}

	rankedLanguages := rankLanguages(languageBytes)

	var meanRepoStars float64
	var forkRatio float64
	var licensedRatio float64
	var archivedRatio float64
	var meanRepoSize float64
	var topicBreadth float64
	if totalRepos > 0 {
		meanRepoStars = float64(totalStars) / float64(totalRepos)
		forkRatio = float64(forkRepos) / float64(totalRepos)
		licensedRatio = float64(licensedRepos) / float64(totalRepos)
		archivedRatio = float64(archivedRepos) / float64(totalRepos)
		meanRepoSize = float64(totalSize) / float64(totalRepos)
		topicBreadth = float64(totalTopics) / float64(totalRepos)
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

		MaxRepoStars:  maxRepoStars,
		MeanRepoStars: meanRepoStars,

		LanguageCount:   len(rankedLanguages),
		RankedLanguages: rankedLanguages,

		TotalReleases:        totalReleases,
		LatestReleaseAt:      latestReleaseAt,
		ReleasedRepos:        releasedRepos,
		TotalPullRequests:    totalPullRequests,
		TotalCollaborators:   totalCollaborators,
		ProtectedBranchRepos: protectedBranchRepos,
		DiscussionRepos:      discussionRepos,

		ForkRatio:     forkRatio,
		LicensedRatio: licensedRatio,
		ArchivedRatio: archivedRatio,

		PortfolioAgeDays:       daysSince(oldestCreated, referenceTime),
		NewestRepoAgeDays:      daysSince(newestCreated, referenceTime),
		DaysSinceLatestRelease: daysSince(latestReleaseAt, referenceTime),
		MeanRepoSize:           meanRepoSize,
		TopicBreadth:           topicBreadth,
	}
}

// rankLanguages returns the distinct language universe ordered by aggregate
// byte share descending, with a deterministic ascending-name tiebreak.
func rankLanguages(languageBytes map[string]int64) []string {
	if len(languageBytes) == 0 {
		return nil
	}

	langs := make([]string, 0, len(languageBytes))
	for lang := range languageBytes {
		langs = append(langs, lang)
	}

	sort.Slice(langs, func(i, j int) bool {
		bi, bj := languageBytes[langs[i]], languageBytes[langs[j]]
		if bi != bj {
			return bi > bj
		}
		return langs[i] < langs[j]
	})

	return langs
}

// daysSince returns the whole-day distance from t to reference, or 0 if t is
// the zero time.
func daysSince(t time.Time, reference time.Time) int {
	if t.IsZero() {
		return 0
	}
	return int(reference.Sub(t).Hours() / 24)
}
