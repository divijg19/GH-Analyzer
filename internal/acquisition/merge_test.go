package acquisition

import (
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/signals"
)

func TestMergeVestiges(t *testing.T) {
	base := signals.RepositoryVestige{
		Name:  "test-repo",
		Stars: 100,
		Forks: 20,
	}

	enrichment := signals.RepositoryVestige{
		LanguageDistribution:   map[string]int64{"Rust": 30000},
		ReleaseCount:           5,
		PullRequestCount:       30,
		DiscussionEnabled:      true,
		ParentRepository:       "owner/parent",
		CollaboratorCount:      8,
		DefaultBranchProtected: true,
	}

	result := mergeVestiges(base, enrichment)

	// REST-authoritative fields must be preserved
	if result.Name != "test-repo" {
		t.Errorf("expected name preserved, got %q", result.Name)
	}
	if result.Stars != 100 {
		t.Errorf("expected stars preserved, got %d", result.Stars)
	}
	if result.Forks != 20 {
		t.Errorf("expected forks preserved, got %d", result.Forks)
	}

	// GraphQL-authoritative fields must be applied from enrichment
	if result.LanguageDistribution["Rust"] != 30000 {
		t.Errorf("expected Rust=30000, got %d", result.LanguageDistribution["Rust"])
	}
	if result.ReleaseCount != 5 {
		t.Errorf("expected 5 releases, got %d", result.ReleaseCount)
	}
	if result.PullRequestCount != 30 {
		t.Errorf("expected 30 PRs, got %d", result.PullRequestCount)
	}
	if !result.DiscussionEnabled {
		t.Errorf("expected discussions enabled")
	}
	if result.ParentRepository != "owner/parent" {
		t.Errorf("expected owner/parent, got %q", result.ParentRepository)
	}
	if result.CollaboratorCount != 8 {
		t.Errorf("expected 8 collaborators, got %d", result.CollaboratorCount)
	}
	if !result.DefaultBranchProtected {
		t.Errorf("expected branch protection enabled")
	}
}

func TestMergeVestigesEmptyEnrichment(t *testing.T) {
	base := signals.RepositoryVestige{
		Name:  "test-repo",
		Stars: 100,
	}

	// Zero-valued enrichment — no GraphQL data observed.
	// GraphQL-authoritative fields should be explicitly set to their
	// Go zero values, representing "not yet observed".
	enrichment := signals.RepositoryVestige{}

	result := mergeVestiges(base, enrichment)

	if result.Name != "test-repo" || result.Stars != 100 {
		t.Errorf("expected REST fields preserved, got Name=%q Stars=%d", result.Name, result.Stars)
	}

	if result.LanguageDistribution != nil {
		t.Errorf("expected nil language distribution for empty enrichment")
	}
	if result.ReleaseCount != 0 {
		t.Errorf("expected 0 releases for empty enrichment, got %d", result.ReleaseCount)
	}
	if !result.LatestReleaseAt.IsZero() {
		t.Error("expected zero release time for empty enrichment")
	}
	if result.PullRequestCount != 0 {
		t.Errorf("expected 0 PRs for empty enrichment, got %d", result.PullRequestCount)
	}
	if result.DiscussionEnabled {
		t.Error("expected discussions disabled for empty enrichment")
	}
	if result.ParentRepository != "" {
		t.Errorf("expected empty parent for empty enrichment, got %q", result.ParentRepository)
	}
	if result.CollaboratorCount != 0 {
		t.Errorf("expected 0 collaborators for empty enrichment, got %d", result.CollaboratorCount)
	}
	if result.DefaultBranchProtected {
		t.Error("expected no branch protection for empty enrichment")
	}
}

func TestMergeVestigesReplacesPreviousEnrichment(t *testing.T) {
	base := signals.RepositoryVestige{
		Name:  "existing",
		Stars: 50,
	}

	enrichment := signals.RepositoryVestige{
		LanguageDistribution: map[string]int64{"Go": 20000},
		ReleaseCount:         7,
		LatestReleaseAt:      time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
	}

	result := mergeVestiges(base, enrichment)

	if result.LanguageDistribution["Go"] != 20000 {
		t.Errorf("expected Go=20000, got %d", result.LanguageDistribution["Go"])
	}
	if len(result.LanguageDistribution) != 1 {
		t.Errorf("expected 1 language, got %d", len(result.LanguageDistribution))
	}
	if result.ReleaseCount != 7 {
		t.Errorf("expected 7 releases, got %d", result.ReleaseCount)
	}
	if result.LatestReleaseAt.IsZero() {
		t.Errorf("expected non-zero latest release time")
	}
	if result.Stars != 50 {
		t.Errorf("expected stars preserved, got %d", result.Stars)
	}
}

func TestMergeVestigesOverwritesEnrichedFields(t *testing.T) {
	base := signals.RepositoryVestige{
		Name:                 "replace-test",
		Stars:                50,
		LanguageDistribution: map[string]int64{"Python": 10000},
		ReleaseCount:         3,
		LatestReleaseAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	enrichment := signals.RepositoryVestige{
		LanguageDistribution: map[string]int64{"Go": 20000},
		ReleaseCount:         7,
		LatestReleaseAt:      time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
	}

	result := mergeVestiges(base, enrichment)

	if result.LanguageDistribution["Go"] != 20000 {
		t.Errorf("expected Go=20000, got %d", result.LanguageDistribution["Go"])
	}
	if len(result.LanguageDistribution) != 1 {
		t.Errorf("expected 1 language after overwrite, got %d", len(result.LanguageDistribution))
	}
	if result.ReleaseCount != 7 {
		t.Errorf("expected release count 7, got %d", result.ReleaseCount)
	}
	if result.LatestReleaseAt.Year() != 2024 || result.LatestReleaseAt.Month() != 3 {
		t.Errorf("expected March 2024, got %v", result.LatestReleaseAt)
	}
	if result.Stars != 50 {
		t.Errorf("expected stars preserved, got %d", result.Stars)
	}
}
