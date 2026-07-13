package facts

import (
	"reflect"
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/observations"
)

// TestFromReposDeterministic verifies the core invariant documented on the
// facts package: identical vestiges plus an identical referenceTime always
// yield byte-identical facts. Repeated calls must not drift.
func TestFromReposDeterministic(t *testing.T) {
	ref := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	repos := []observations.RepositoryVestige{
		{
			Name:          "active",
			Fork:          false,
			Size:          120,
			UpdatedAt:     ref.AddDate(0, 0, -10),
			Visibility:    "public",
			Archived:      false,
			Template:      false,
			License:       "mit",
			Stars:         30,
			Forks:         5,
			Watchers:      3,
			OpenIssues:    2,
			CreatedAt:     ref.AddDate(-3, 0, 0),
			PushedAt:      ref.AddDate(0, 0, -4),
			DefaultBranch: "main",
		},
		{
			Name:       "fork",
			Fork:       true,
			Size:       200,
			UpdatedAt:  ref.AddDate(0, 0, -200),
			Visibility: "private",
		},
	}

	first := FromRepos(repos, ref)
	for i := 0; i < 1000; i++ {
		got := FromRepos(repos, ref)
		if !reflect.DeepEqual(first, got) {
			t.Fatalf("FromRepos not deterministic on iteration %d", i)
		}
	}
}

// TestFromReposReferenceTimeSensitive verifies that referenceTime is honoured,
// not silently ignored: shifting the reference changes the recent-window
// classification and therefore the derived facts.
func TestFromReposReferenceTimeSensitive(t *testing.T) {
	recent := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	old := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)

	repos := []observations.RepositoryVestige{
		{
			Name:       "edge",
			Fork:       false,
			Size:       120,
			UpdatedAt:  recent.AddDate(0, 0, -80), // inside 90-day window at `recent`, outside at `old`
			Visibility: "public",
		},
	}

	recentFacts := FromRepos(repos, recent)
	oldFacts := FromRepos(repos, old)

	if recentFacts.RecentRepos == oldFacts.RecentRepos {
		t.Fatalf("referenceTime ignored: RecentRepos equal (%d) for differing references", recentFacts.RecentRepos)
	}
	if recentFacts.RecentRepos != 1 {
		t.Fatalf("expected 1 recent repo at `recent`, got %d", recentFacts.RecentRepos)
	}
	if oldFacts.RecentRepos != 0 {
		t.Fatalf("expected 0 recent repos at `old`, got %d", oldFacts.RecentRepos)
	}
}

// TestActivityFactsFromObservationsDeterministic verifies the activity facts
// derivation is pure and stable across repeated calls.
func TestActivityFactsFromObservationsDeterministic(t *testing.T) {
	ref := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	obs := []observations.ActivityObservation{
		{Kind: observations.ActivityKindCommit, Count: 10, Repository: "a/b", OccurredAt: ref.AddDate(0, 0, -5)},
		{Kind: observations.ActivityKindPullRequest, Count: 4, Repository: "a/b", OccurredAt: ref.AddDate(0, 0, -5)},
		{Kind: observations.ActivityKindReview, Count: 2, Repository: "a/c", OccurredAt: ref.AddDate(0, 0, -5)},
		{Kind: observations.ActivityKindContributionByRepo, Count: 6, Repository: "a/b", OccurredAt: ref.AddDate(0, 0, -5)},
		{Kind: observations.ActivityKindActiveDay, Count: 1, Metadata: observations.ActivityMetadata{ActiveDays: 40, TotalDays: 120}},
	}

	first := ActivityFactsFromObservations(obs, ref)
	for i := 0; i < 1000; i++ {
		got := ActivityFactsFromObservations(obs, ref)
		if !reflect.DeepEqual(first, got) {
			t.Fatalf("ActivityFactsFromObservations not deterministic on iteration %d", i)
		}
	}
}

// TestFromReposEmpty verifies the documented empty-input contract.
func TestFromReposEmpty(t *testing.T) {
	if got := FromRepos(nil, time.Now()); !reflect.DeepEqual(got, RepositoryFacts{}) {
		t.Fatalf("FromRepos(nil) = %+v, want zero RepositoryFacts", got)
	}
	if got := FromRepos([]observations.RepositoryVestige{}, time.Now()); !reflect.DeepEqual(got, RepositoryFacts{}) {
		t.Fatalf("FromRepos(empty) = %+v, want zero RepositoryFacts", got)
	}
}
