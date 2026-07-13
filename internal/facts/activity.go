package facts

import (
	"math"
	"time"

	"github.com/divijg19/Atlas/internal/observations"
)

// ActivityFacts is the canonical fact family for the Activity Intelligence domain.
type ActivityFacts struct {
	RecentCommits         int     `json:"recent_commits"`
	RecentPullRequests    int     `json:"recent_pull_requests"`
	RecentReviews         int     `json:"recent_reviews"`
	RecentIssues          int     `json:"recent_issues"`
	RecentPrivate         int     `json:"recent_private"`
	LifetimeCommits       int     `json:"lifetime_commits"`
	LifetimePullRequests  int     `json:"lifetime_pull_requests"`
	LifetimeReviews       int     `json:"lifetime_reviews"`
	LifetimeIssues        int     `json:"lifetime_issues"`
	LifetimePrivate       int     `json:"lifetime_private"`
	LifetimeTotal         int     `json:"lifetime_total"`
	ContributionBreadth   int     `json:"contribution_breadth"`
	RepositoryBreadth     int     `json:"repository_breadth"`
	ActiveDays            int     `json:"active_days"`
	ActivityCadence       float64 `json:"activity_cadence"`
	ContributionFrequency float64 `json:"contribution_frequency"`
	RepositoryDepth       float64 `json:"repository_depth"`
	YearCount             int     `json:"year_count"`
	CommitCadence         float64 `json:"commit_cadence"`
	ContributionRecency   float64 `json:"contribution_recency"`
}

func ActivityFactsFromObservations(observationsSlice []observations.ActivityObservation, referenceTime time.Time) ActivityFacts {
	if len(observationsSlice) == 0 {
		return ActivityFacts{}
	}

	var facts ActivityFacts
	var recentCommitCount, recentPRCount, recentReviewCount, recentIssueCount, recentPrivateCount int
	var lifetimeCommitCount, lifetimePRCount, lifetimeReviewCount, lifetimeIssueCount, lifetimePrivateCount int
	var activeDays, totalDays int
	reposByCommit := make(map[string]int)
	years := make(map[int]bool)

	for _, obs := range observationsSlice {
		isRecent := obs.Metadata.Year == 0 || obs.Metadata.Year == referenceTime.UTC().Year()

		switch obs.Kind {
		case observations.ActivityKindCommit:
			if isRecent {
				recentCommitCount += obs.Count
			}
			lifetimeCommitCount += obs.Count
		case observations.ActivityKindPullRequest:
			if isRecent {
				recentPRCount += obs.Count
			}
			lifetimePRCount += obs.Count
		case observations.ActivityKindReview:
			if isRecent {
				recentReviewCount += obs.Count
			}
			lifetimeReviewCount += obs.Count
		case observations.ActivityKindIssue:
			if isRecent {
				recentIssueCount += obs.Count
			}
			lifetimeIssueCount += obs.Count
		case observations.ActivityKindAggregate:
			if isRecent {
				recentPrivateCount += obs.Count
			}
			lifetimePrivateCount += obs.Count
		case observations.ActivityKindActiveDay:
			activeDays = obs.Metadata.ActiveDays
			totalDays = obs.Metadata.TotalDays
		case observations.ActivityKindContributionByRepo:
			if obs.Repository != "" {
				reposByCommit[obs.Repository] += obs.Count
			}
		}
		if obs.Metadata.Year > 0 {
			years[obs.Metadata.Year] = true
		}
	}

	facts.RecentCommits = recentCommitCount
	facts.RecentPullRequests = recentPRCount
	facts.RecentReviews = recentReviewCount
	facts.RecentIssues = recentIssueCount
	facts.RecentPrivate = recentPrivateCount
	facts.LifetimeCommits = lifetimeCommitCount
	facts.LifetimePullRequests = lifetimePRCount
	facts.LifetimeReviews = lifetimeReviewCount
	facts.LifetimeIssues = lifetimeIssueCount
	facts.LifetimePrivate = lifetimePrivateCount
	facts.LifetimeTotal = lifetimeCommitCount + lifetimePRCount + lifetimeReviewCount + lifetimeIssueCount + lifetimePrivateCount

	breadth := 0
	if recentCommitCount > 0 {
		breadth++
	}
	if recentPRCount > 0 {
		breadth++
	}
	if recentReviewCount > 0 {
		breadth++
	}
	if recentIssueCount > 0 {
		breadth++
	}
	if recentPrivateCount > 0 {
		breadth++
	}
	facts.ContributionBreadth = breadth
	facts.RepositoryBreadth = len(reposByCommit)
	if facts.RepositoryBreadth > 0 {
		total := 0
		for _, c := range reposByCommit {
			total += c
		}
		facts.RepositoryDepth = float64(total) / float64(facts.RepositoryBreadth)
	}
	facts.ActiveDays = activeDays
	if totalDays > 0 {
		facts.ActivityCadence = float64(activeDays) / float64(totalDays)
	}
	totalRecent := recentCommitCount + recentPRCount + recentReviewCount + recentIssueCount + recentPrivateCount
	if activeDays > 0 {
		facts.ContributionFrequency = float64(totalRecent) / float64(activeDays)
	}
	facts.YearCount = len(years)
	if facts.YearCount > 0 {
		facts.CommitCadence = float64(lifetimeCommitCount) / float64(facts.YearCount)
	}
	expectedRecent := 0
	if facts.YearCount > 0 {
		expectedRecent = facts.LifetimeTotal / facts.YearCount
	}
	if expectedRecent > 0 {
		facts.ContributionRecency = float64(totalRecent) / float64(expectedRecent)
	}

	facts.ActivityCadence = math.Round(facts.ActivityCadence*10000) / 10000
	facts.ContributionFrequency = math.Round(facts.ContributionFrequency*10000) / 10000
	facts.RepositoryDepth = math.Round(facts.RepositoryDepth*10000) / 10000
	facts.CommitCadence = math.Round(facts.CommitCadence*10000) / 10000
	facts.ContributionRecency = math.Round(facts.ContributionRecency*10000) / 10000

	return facts
}
