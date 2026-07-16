package facts

import (
	"sort"
	"time"

	"github.com/divijg19/Atlas/internal/observations"
)

// RepositoryFacts is the canonical fact family for the repository domain.
// Facts are deterministic aggregates derived entirely from RepositoryVestige.
type RepositoryFacts struct {
	// Repository inventory
	TotalRepos         int       `json:"total_repos"`
	OriginalRepos      int       `json:"original_repos"`
	ForkRepos          int       `json:"fork_repos"`
	RecentRepos        int       `json:"recent_repos"`
	DeepRepos          int       `json:"deep_repos"`
	ValidRepos         int       `json:"valid_repos"`
	ValidOriginalRepos int       `json:"valid_original_repos"`
	LargestRepoSize    int       `json:"largest_repo_size"`
	LatestActivity     time.Time `json:"latest_activity"`

	// Repository identity and maintenance
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

	// Repository technology and structure
	MaxRepoStars         int       `json:"max_repo_stars"`
	MeanRepoStars        float64   `json:"mean_repo_stars"`
	LanguageCount        int       `json:"language_count"`
	RankedLanguages      []string  `json:"ranked_languages"`
	TotalReleases        int       `json:"total_releases"`
	LatestReleaseAt      time.Time `json:"latest_release_at"`
	ReleasedRepos        int       `json:"released_repos"`
	TotalPullRequests    int       `json:"total_pull_requests"`
	TotalCollaborators   int       `json:"total_collaborators"`
	ProtectedBranchRepos int       `json:"protected_branch_repos"`
	DiscussionRepos      int       `json:"discussion_repos"`

	// Portfolio aggregates
	ForkRatio              float64 `json:"fork_ratio"`
	LicensedRatio          float64 `json:"licensed_ratio"`
	ArchivedRatio          float64 `json:"archived_ratio"`
	PortfolioAgeDays       int     `json:"portfolio_age_days"`
	NewestRepoAgeDays      int     `json:"newest_repo_age_days"`
	DaysSinceLatestRelease int     `json:"days_since_latest_release"`
	MeanRepoSize           float64 `json:"mean_repo_size"`
	TopicBreadth           float64 `json:"topic_breadth"`

	// Portfolio intelligence (deterministic aggregations from RepositoryVestige).
	// These answer "what was observed or deterministically aggregated?", never
	// "what does it mean?". Interpretation belongs to higher layers.
	RankedTopics       []string           `json:"ranked_topics"`
	TopicUniverse      int                `json:"topic_universe"`
	ForkLineage        []ForkParentFact   `json:"fork_lineage"`
	TechnologyTimeline map[int][]string   `json:"technology_timeline"`
	MaintenanceBuckets MaintenanceBuckets `json:"maintenance_buckets"`
}

const (
	recentWindowDays  = 90
	minDepthRepoSize  = 50
	minConsistencyDen = 10
	minDepthDen       = 5
)

// Maintenance window thresholds (days since last push). These are the single
// authoritative definition of the maintenance recency windows shared by the
// portfolio MaintenanceBuckets fact and the per-repository Maintenance
// dimension in repositoryintelligence. Both must agree; this is the one place
// the boundary is defined.
const (
	MaintenanceActiveWindowDays  = 90
	MaintenanceDormantWindowDays = 365
)

// ForkParentFact captures deterministic upstream-lineage concentration: how many
// of the candidate's repositories fork a given parent. It is a pure
// aggregation of RepositoryVestige.ParentRepository; it assigns no score.
type ForkParentFact struct {
	Parent string `json:"parent"`
	Forks  int    `json:"forks"`
}

// MaintenanceBuckets is a deterministic histogram of repository maintenance
// recency derived from each repository's last-push age.
type MaintenanceBuckets struct {
	Active  int `json:"active"`
	Recent  int `json:"recent"`
	Dormant int `json:"dormant"`
}

func FromRepos(repos []observations.RepositoryVestige, referenceTime time.Time) RepositoryFacts {
	if len(repos) == 0 {
		return RepositoryFacts{}
	}

	cutoff := referenceTime.AddDate(0, 0, -recentWindowDays)

	totalRepos := len(repos)
	var originalRepos, forkRepos, recentRepos, deepRepos, validRepos, validOriginalRepos int
	largestRepoSize := 0
	var latestActivity time.Time

	var archivedRepos, templateRepos, publicRepos, privateRepos, licensedRepos int
	var totalStars, totalForks, totalWatchers, totalOpenIssues, totalTopics int
	var oldestCreated, newestCreated time.Time

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

	// Portfolio-intelligence accumulators.
	topicCounts := make(map[string]int)
	forkParents := make(map[string]int)
	technologyTimeline := make(map[int][]string)
	var maintenance MaintenanceBuckets

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
			// "Recent" is the closed window [cutoff, referenceTime]: updated
			// within the 90 days ending at the reference time. Both bounds are
			// required so replaying at a past referenceTime does not classify
			// future-dated repos as recent.
			if !repo.UpdatedAt.Before(cutoff) && !repo.UpdatedAt.After(referenceTime) {
				recentRepos++
			}
			if repo.Size >= minDepthRepoSize {
				deepRepos++
			}
		}

		// Portfolio intelligence: deterministic aggregations only.
		for _, topic := range repo.Topics {
			if topic != "" {
				topicCounts[topic]++
			}
		}
		if repo.Fork && repo.ParentRepository != "" {
			forkParents[repo.ParentRepository]++
		}
		// Technology timeline: every language observed across repositories
		// created in the same year, preserved (not collapsed to one). Recording
		// all languages keeps the data deterministic and lossless so higher
		// layers can later derive transitions, diversification, or migrations.
		if !repo.CreatedAt.IsZero() && len(repo.LanguageDistribution) > 0 {
			timelineYear := repo.CreatedAt.UTC().Year()
			for _, lang := range rankLanguages(repo.LanguageDistribution) {
				if !containsString(technologyTimeline[timelineYear], lang) {
					technologyTimeline[timelineYear] = append(technologyTimeline[timelineYear], lang)
				}
			}
		}
		pushAge := daysSince(repo.PushedAt, referenceTime)
		switch {
		case pushAge <= MaintenanceActiveWindowDays:
			maintenance.Active++
		case pushAge <= MaintenanceDormantWindowDays:
			maintenance.Recent++
		default:
			maintenance.Dormant++
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

	rankedTopics := rankedTopics(topicCounts)

	forkLineage := make([]ForkParentFact, 0, len(forkParents))
	for parent, count := range forkParents {
		forkLineage = append(forkLineage, ForkParentFact{Parent: parent, Forks: count})
	}
	sort.Slice(forkLineage, func(i, j int) bool {
		if forkLineage[i].Forks != forkLineage[j].Forks {
			return forkLineage[i].Forks > forkLineage[j].Forks
		}
		return forkLineage[i].Parent < forkLineage[j].Parent
	})

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
		TotalRepos:             totalRepos,
		OriginalRepos:          originalRepos,
		ForkRepos:              forkRepos,
		RecentRepos:            recentRepos,
		DeepRepos:              deepRepos,
		ValidRepos:             validRepos,
		ValidOriginalRepos:     validOriginalRepos,
		LargestRepoSize:        largestRepoSize,
		LatestActivity:         latestActivity,
		ArchivedRepos:          archivedRepos,
		TemplateRepos:          templateRepos,
		PublicRepos:            publicRepos,
		PrivateRepos:           privateRepos,
		LicensedRepos:          licensedRepos,
		TotalStars:             totalStars,
		TotalForks:             totalForks,
		TotalWatchers:          totalWatchers,
		TotalOpenIssues:        totalOpenIssues,
		TotalTopics:            totalTopics,
		OldestCreated:          oldestCreated,
		NewestCreated:          newestCreated,
		MaxRepoStars:           maxRepoStars,
		MeanRepoStars:          meanRepoStars,
		LanguageCount:          len(rankedLanguages),
		RankedLanguages:        rankedLanguages,
		TotalReleases:          totalReleases,
		LatestReleaseAt:        latestReleaseAt,
		ReleasedRepos:          releasedRepos,
		TotalPullRequests:      totalPullRequests,
		TotalCollaborators:     totalCollaborators,
		ProtectedBranchRepos:   protectedBranchRepos,
		DiscussionRepos:        discussionRepos,
		ForkRatio:              forkRatio,
		LicensedRatio:          licensedRatio,
		ArchivedRatio:          archivedRatio,
		PortfolioAgeDays:       daysSince(oldestCreated, referenceTime),
		NewestRepoAgeDays:      daysSince(newestCreated, referenceTime),
		DaysSinceLatestRelease: daysSince(latestReleaseAt, referenceTime),
		MeanRepoSize:           meanRepoSize,
		TopicBreadth:           topicBreadth,

		RankedTopics:       rankedTopics,
		TopicUniverse:      len(rankedTopics),
		ForkLineage:        forkLineage,
		TechnologyTimeline: technologyTimeline,
		MaintenanceBuckets: maintenance,
	}
}

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

// rankedTopics unionizes repository topics across the portfolio, ordered by
// frequency (highest first) with a lexical tiebreak. Deterministic: identical
// input yields identical output. Mirrors rankLanguages for the topic domain.
func rankedTopics(topicCounts map[string]int) []string {
	if len(topicCounts) == 0 {
		return nil
	}

	topics := make([]string, 0, len(topicCounts))
	for topic := range topicCounts {
		topics = append(topics, topic)
	}

	sort.Slice(topics, func(i, j int) bool {
		ci, cj := topicCounts[topics[i]], topicCounts[topics[j]]
		if ci != cj {
			return ci > cj
		}
		return topics[i] < topics[j]
	})

	return topics
}

// containsString reports whether s contains v. Used to deduplicate language
// entries while preserving per-year order in the technology timeline.
func containsString(s []string, v string) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

func daysSince(t time.Time, reference time.Time) int {
	if t.IsZero() {
		return 0
	}
	return int(reference.Sub(t).Hours() / 24)
}
