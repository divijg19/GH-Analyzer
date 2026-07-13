package signals

import (
	"testing"
	"time"
)

func TestActivityFactsFromObservationsEmpty(t *testing.T) {
	now := time.Now()

	f := ActivityFactsFromObservations(nil, now)
	if f.RecentCommits != 0 || f.LifetimeTotal != 0 || f.ContributionBreadth != 0 {
		t.Fatalf("expected zero values for empty input, got %+v", f)
	}
}

func TestActivityFactsFromObservationsOneYearCounts(t *testing.T) {
	now := time.Now()

	observations := []ActivityObservation{
		{Kind: ActivityKindCommit, Count: 100},
		{Kind: ActivityKindPullRequest, Count: 50},
		{Kind: ActivityKindReview, Count: 25},
		{Kind: ActivityKindIssue, Count: 10},
		{Kind: ActivityKindAggregate, Count: 5, Metadata: ActivityMetadata{RestrictedCount: 5}},
		{Kind: ActivityKindActiveDay, Metadata: ActivityMetadata{ActiveDays: 200, TotalDays: 365}},
	}

	f := ActivityFactsFromObservations(observations, now)

	if f.RecentCommits != 100 {
		t.Errorf("expected RecentCommits 100, got %d", f.RecentCommits)
	}
	if f.RecentPullRequests != 50 {
		t.Errorf("expected RecentPullRequests 50, got %d", f.RecentPullRequests)
	}
	if f.RecentReviews != 25 {
		t.Errorf("expected RecentReviews 25, got %d", f.RecentReviews)
	}
	if f.RecentIssues != 10 {
		t.Errorf("expected RecentIssues 10, got %d", f.RecentIssues)
	}
	if f.RecentPrivate != 5 {
		t.Errorf("expected RecentPrivate 5, got %d", f.RecentPrivate)
	}
	if f.ContributionBreadth != 5 {
		t.Errorf("expected ContributionBreadth 5, got %d", f.ContributionBreadth)
	}
	if f.ActiveDays != 200 {
		t.Errorf("expected ActiveDays 200, got %d", f.ActiveDays)
	}
}

func TestActivityFactsFromObservationsLifetimeCounts(t *testing.T) {
	now := time.Now()
	year := now.UTC().Year()

	observations := []ActivityObservation{
		{Kind: ActivityKindCommit, Count: 100, Metadata: ActivityMetadata{Year: year}},
		{Kind: ActivityKindCommit, Count: 150, Metadata: ActivityMetadata{Year: year - 1}},
		{Kind: ActivityKindPullRequest, Count: 50, Metadata: ActivityMetadata{Year: year}},
		{Kind: ActivityKindAggregate, Count: 10, Metadata: ActivityMetadata{Year: year - 1, RestrictedCount: 10}},
	}

	f := ActivityFactsFromObservations(observations, now)

	if f.LifetimeCommits != 250 {
		t.Errorf("expected LifetimeCommits 250, got %d", f.LifetimeCommits)
	}
	if f.LifetimePullRequests != 50 {
		t.Errorf("expected LifetimePullRequests 50, got %d", f.LifetimePullRequests)
	}
	if f.LifetimeTotal != 310 {
		t.Errorf("expected LifetimeTotal 310, got %d", f.LifetimeTotal)
	}
	if f.YearCount != 2 {
		t.Errorf("expected YearCount 2, got %d", f.YearCount)
	}
	if f.CommitCadence != 125.0 {
		t.Errorf("expected CommitCadence 125.0, got %f", f.CommitCadence)
	}
}

func TestActivityFactsFromObservationsRepositoryBreadth(t *testing.T) {
	now := time.Now()

	observations := []ActivityObservation{
		{Kind: ActivityKindCommit, Count: 10},
		{Kind: ActivityKindContributionByRepo, Repository: "owner/repo1"},
		{Kind: ActivityKindContributionByRepo, Repository: "owner/repo2"},
		{Kind: ActivityKindContributionByRepo, Repository: "owner/repo1"}, // duplicate, should merge
	}

	f := ActivityFactsFromObservations(observations, now)

	if f.RepositoryBreadth != 2 {
		t.Errorf("expected RepositoryBreadth 2, got %d", f.RepositoryBreadth)
	}
}

func TestActivityFactsFromObservationsContributionFrequency(t *testing.T) {
	now := time.Now()

	observations := []ActivityObservation{
		{Kind: ActivityKindCommit, Count: 365},
		{Kind: ActivityKindPullRequest, Count: 100},
		{Kind: ActivityKindReview, Count: 50},
		{Kind: ActivityKindIssue, Count: 20},
		{Kind: ActivityKindAggregate, Count: 10},
		{Kind: ActivityKindActiveDay, Metadata: ActivityMetadata{ActiveDays: 200, TotalDays: 365}},
	}

	f := ActivityFactsFromObservations(observations, now)

	// Total recent = 365 + 100 + 50 + 20 + 10 = 545
	// Active days = 200
	// Frequency = 545 / 200 = 2.725
	expected := 2.725
	if !closeFloat(f.ContributionFrequency, expected) {
		t.Errorf("expected ContributionFrequency %.4f, got %.4f", expected, f.ContributionFrequency)
	}
}

func TestActivityFactsFromObservationsActivityCadence(t *testing.T) {
	now := time.Now()

	observations := []ActivityObservation{
		{Kind: ActivityKindActiveDay, Metadata: ActivityMetadata{ActiveDays: 200, TotalDays: 365}},
	}

	f := ActivityFactsFromObservations(observations, now)

	// Active days = 200, Total days = 365
	// Cadence = 200 / 365 = 0.5479...
	expected := 0.5479
	if !closeFloat(f.ActivityCadence, expected) {
		t.Errorf("expected ActivityCadence %.4f, got %.4f", expected, f.ActivityCadence)
	}
}

func TestActivityFactsFromObservationsContributionRecency(t *testing.T) {
	now := time.Now()
	year := now.UTC().Year()

	// Reduced recent activity relative to the lifetime baseline.
	observations := []ActivityObservation{
		{Kind: ActivityKindCommit, Count: 100},                                             // recent
		{Kind: ActivityKindCommit, Count: 100, Metadata: ActivityMetadata{Year: year - 1}}, // yearly
		{Kind: ActivityKindPullRequest, Count: 50},
		{Kind: ActivityKindPullRequest, Count: 50, Metadata: ActivityMetadata{Year: year - 1}},
		{Kind: ActivityKindActiveDay, Metadata: ActivityMetadata{ActiveDays: 200}},
	}

	f := ActivityFactsFromObservations(observations, now)

	// Lifetime = 300, years = 1 (one explicit yearly bucket), expected per year = 300
	// Recent total = 100+50 = 150.
	// Recency = 150 / 300 = 0.5
	expected := 0.5
	recentTotal := 150
	expectedRecent := 300
	if !closeFloat(f.ContributionRecency, expected) {
		t.Errorf("expected ContributionRecency %.4f, got %.4f (recent=%d, lifetime=%d, years=%d, expectedRecent=%d)", expected, f.ContributionRecency, recentTotal, f.LifetimeTotal, f.YearCount, expectedRecent)
	}
}

func TestActivityFactsFromObservationsZeroYears(t *testing.T) {
	now := time.Now()

	observations := []ActivityObservation{
		{Kind: ActivityKindCommit, Count: 100},
		{Kind: ActivityKindPullRequest, Count: 50},
	}

	f := ActivityFactsFromObservations(observations, now)

	if f.YearCount != 0 {
		t.Errorf("expected YearCount 0, got %d", f.YearCount)
	}
	if f.CommitCadence != 0 {
		t.Errorf("expected CommitCadence 0 for no years, got %f", f.CommitCadence)
	}
	if f.ContributionRecency != 0 {
		t.Errorf("expected ContributionRecency 0 for no years, got %f", f.ContributionRecency)
	}
}

func closeFloat(a, b float64) bool {
	const epsilon = 0.001
	return abs(a-b) <= epsilon
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
