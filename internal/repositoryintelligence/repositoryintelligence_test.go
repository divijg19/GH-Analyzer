package repositoryintelligence

import (
	"context"
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/observations"
)

func sampleRepo() observations.RepositoryVestige {
	return observations.RepositoryVestige{
		Name:                   "atlas",
		Visibility:             "public",
		CreatedAt:              time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:              time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		PushedAt:               time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		LatestReleaseAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ReleaseCount:           3,
		License:                "MIT",
		Topics:                 []string{"cli", "git"},
		DefaultBranch:          "main",
		DefaultBranchProtected: true,
		LanguageDistribution:   map[string]int64{"Go": 1000, "Shell": 200},
		OpenIssues:             5,
		Stars:                  120,
		Forks:                  12,
		Watchers:               60,
		PullRequestCount:       8,
		Size:                   500,
		DiscussionEnabled:      true,
		CollaboratorCount:      4,
	}
}

func TestBuildRepositoryIntelligenceDeterministic(t *testing.T) {
	ref := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	a := BuildRepositoryIntelligence(context.Background(), sampleRepo(), ref)
	b := BuildRepositoryIntelligence(context.Background(), sampleRepo(), ref)

	if a.Identity.Summary() != b.Identity.Summary() {
		t.Fatalf("non-deterministic identity summary")
	}
	if a.Risk.Summary() != b.Risk.Summary() {
		t.Fatalf("non-deterministic risk summary")
	}
	if len(a.Dimensions()) != 13 {
		t.Fatalf("expected 13 dimensions, got %d", len(a.Dimensions()))
	}
}

func TestBuildRepositoryIntelligenceContract(t *testing.T) {
	ref := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	ri := BuildRepositoryIntelligence(context.Background(), sampleRepo(), ref)

	seen := make(map[string]bool)
	for _, dim := range ri.Dimensions() {
		if dim.Name() == "" {
			t.Fatalf("dimension with empty name")
		}
		if seen[dim.Name()] {
			t.Fatalf("duplicate dimension name %q", dim.Name())
		}
		seen[dim.Name()] = true
		if len(dim.Evidence()) == 0 {
			t.Fatalf("dimension %q has no evidence", dim.Name())
		}
		if dim.Summary() == "" {
			t.Fatalf("dimension %q has empty summary", dim.Name())
		}
	}
	if len(seen) != 13 {
		t.Fatalf("expected 13 unique dimensions, got %d", len(seen))
	}
}

func TestBuildRepositoryIntelligenceEmpty(t *testing.T) {
	ref := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	ri := BuildRepositoryIntelligence(context.Background(), observations.RepositoryVestige{}, ref)
	if ri == nil {
		t.Fatal("nil intelligence for empty repo")
	}
	for _, dim := range ri.Dimensions() {
		if dim.Summary() == "" {
			t.Fatalf("dimension %q empty summary for empty repo", dim.Name())
		}
	}
}

func TestBuildRepositoryIntelligenceReferenceTimeSensitive(t *testing.T) {
	repo := sampleRepo()
	recent := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	ancient := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	riRecent := BuildRepositoryIntelligence(context.Background(), repo, recent)
	riAncient := BuildRepositoryIntelligence(context.Background(), repo, ancient)

	if riRecent.Health.Level == riAncient.Health.Level {
		t.Fatalf("health level should change with reference time: recent=%s ancient=%s", riRecent.Health.Level, riAncient.Health.Level)
	}
}

// TestEveryDimensionCarriesObservationProvenance verifies the v0.9.1 invariant:
// every repository dimension's evidence traces back to the repository's own
// observation identity and the specific observed fields it interpreted.
func TestEveryDimensionCarriesObservationProvenance(t *testing.T) {
	ref := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	repo := sampleRepo()
	ri := BuildRepositoryIntelligence(context.Background(), repo, ref)

	for _, d := range ri.Dimensions() {
		groups := d.Evidence()
		if len(groups) == 0 {
			t.Fatalf("dimension %q has no evidence", d.Name())
		}
		for _, g := range groups {
			if len(g.Provenance.Observations) == 0 {
				t.Fatalf("dimension %q evidence lacks observation provenance", d.Name())
			}
			for _, o := range g.Provenance.Observations {
				if o.ID != repo.ObservationID() {
					t.Fatalf("dimension %q references observation id %q, want %q", d.Name(), o.ID, repo.ObservationID())
				}
				if o.Field == "" || o.Source == "" {
					t.Fatalf("dimension %q has incomplete observation ref: %+v", d.Name(), o)
				}
			}
		}
	}
}
