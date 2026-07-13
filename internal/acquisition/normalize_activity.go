package acquisition

import (
	"time"

	obs "github.com/divijg19/Atlas/internal/observations"
)

// normalizeActivityProfile converts a contributionsCollection DTO to a slice of
// canonical observations.ActivityObservation values for the 1-year activity window.
func normalizeActivityProfile(dto *contributionsCollectionDTO, username string, now time.Time) []obs.ActivityObservation {
	if dto == nil {
		return nil
	}

	recentStart := now.UTC().AddDate(0, 0, -365)
	recentEnd := now.UTC()
	observations := make([]obs.ActivityObservation, 0, 8)

	// Per-type contribution counts
	observations = append(observations,
		obs.ActivityObservation{
			Kind:       obs.ActivityKindCommit,
			Count:      dto.TotalCommitContributions,
			Actor:      username,
			OccurredAt: now,
			Metadata: obs.ActivityMetadata{
				WindowStart: recentStart,
				WindowEnd:   recentEnd,
			},
		},
		obs.ActivityObservation{
			Kind:       obs.ActivityKindPullRequest,
			Count:      dto.TotalPullRequestContributions,
			Actor:      username,
			OccurredAt: now,
			Metadata: obs.ActivityMetadata{
				WindowStart: recentStart,
				WindowEnd:   recentEnd,
			},
		},
		obs.ActivityObservation{
			Kind:       obs.ActivityKindReview,
			Count:      dto.TotalPullRequestReviewContributions,
			Actor:      username,
			OccurredAt: now,
			Metadata: obs.ActivityMetadata{
				WindowStart: recentStart,
				WindowEnd:   recentEnd,
			},
		},
		obs.ActivityObservation{
			Kind:       obs.ActivityKindIssue,
			Count:      dto.TotalIssueContributions,
			Actor:      username,
			OccurredAt: now,
			Metadata: obs.ActivityMetadata{
				WindowStart: recentStart,
				WindowEnd:   recentEnd,
			},
		},
		obs.ActivityObservation{
			Kind:       obs.ActivityKindAggregate,
			Count:      dto.RestrictedContributionsCount,
			Actor:      username,
			OccurredAt: now,
			Metadata: obs.ActivityMetadata{
				WindowStart:     recentStart,
				WindowEnd:       recentEnd,
				RestrictedCount: dto.RestrictedContributionsCount,
			},
		},
	)

	// Active days from contribution calendar
	activeDays := 0
	totalDays := 0
	for _, week := range dto.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			totalDays++
			if day.ContributionCount > 0 {
				activeDays++
			}
		}
	}
	observations = append(observations, obs.ActivityObservation{
		Kind:       obs.ActivityKindActiveDay,
		Count:      activeDays,
		Actor:      username,
		OccurredAt: now,
		Metadata: obs.ActivityMetadata{
			WindowStart: recentStart,
			WindowEnd:   recentEnd,
			ActiveDays:  activeDays,
			TotalDays:   totalDays,
		},
	})

	// Per-repo contribution breakdown
	for _, repo := range dto.CommitContributionsByRepository {
		if repo.Repository.NameWithOwner == "" {
			continue
		}
		observations = append(observations, obs.ActivityObservation{
			Kind:       obs.ActivityKindContributionByRepo,
			Count:      repo.Contributions.TotalCount,
			Repository: repo.Repository.NameWithOwner,
			Actor:      username,
			OccurredAt: now,
			Metadata: obs.ActivityMetadata{
				WindowStart:     recentStart,
				WindowEnd:       recentEnd,
				RepoCommitCount: repo.Contributions.TotalCount,
			},
		})
	}

	return observations
}

// normalizeYearContrib converts a single year's contribution DTO to per-type
// observations.ActivityObservation values.
func normalizeYearContrib(dto *yearContribDTO, username string, year int, now time.Time) []obs.ActivityObservation {
	if dto == nil {
		return nil
	}
	yearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	if year == now.UTC().Year() {
		yearEnd = now.UTC()
	}

	return []obs.ActivityObservation{
		{
			Kind:       obs.ActivityKindCommit,
			Count:      dto.TotalCommitContributions,
			Actor:      username,
			OccurredAt: now,
			Metadata:   obs.ActivityMetadata{WindowStart: yearStart, WindowEnd: yearEnd, Year: year},
		},
		{
			Kind:       obs.ActivityKindPullRequest,
			Count:      dto.TotalPullRequestContributions,
			Actor:      username,
			OccurredAt: now,
			Metadata:   obs.ActivityMetadata{WindowStart: yearStart, WindowEnd: yearEnd, Year: year},
		},
		{
			Kind:       obs.ActivityKindReview,
			Count:      dto.TotalPullRequestReviewContributions,
			Actor:      username,
			OccurredAt: now,
			Metadata:   obs.ActivityMetadata{WindowStart: yearStart, WindowEnd: yearEnd, Year: year},
		},
		{
			Kind:       obs.ActivityKindIssue,
			Count:      dto.TotalIssueContributions,
			Actor:      username,
			OccurredAt: now,
			Metadata:   obs.ActivityMetadata{WindowStart: yearStart, WindowEnd: yearEnd, Year: year},
		},
		{
			Kind:       obs.ActivityKindAggregate,
			Count:      dto.RestrictedContributionsCount,
			Actor:      username,
			OccurredAt: now,
			Metadata: obs.ActivityMetadata{
				WindowStart:     yearStart,
				WindowEnd:       yearEnd,
				Year:            year,
				RestrictedCount: dto.RestrictedContributionsCount,
			},
		},
	}
}
