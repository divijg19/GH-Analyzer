package facts

import (
	"reflect"
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/observations"
)

// ref is a fixed reference time so maintenance-bucket classification is stable.
var portfolioRef = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func repoFixture() []observations.RepositoryVestige {
	return []observations.RepositoryVestige{
		{
			Name:                 "a",
			Fork:                 false,
			Topics:               []string{"cli", "go", "go"},
			LanguageDistribution: map[string]int64{"Go": 100, "Shell": 10},
			CreatedAt:            time.Date(2019, 3, 1, 0, 0, 0, 0, time.UTC),
			PushedAt:             time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:                 "b",
			Fork:                 false,
			Topics:               []string{"cli", "web"},
			LanguageDistribution: map[string]int64{"TypeScript": 80},
			CreatedAt:            time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			PushedAt:             time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:                 "c",
			Fork:                 true,
			ParentRepository:     "parent/upstream",
			Topics:               []string{"experimental"},
			LanguageDistribution: map[string]int64{"Rust": 40},
			CreatedAt:            time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			PushedAt:             time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:                 "d",
			Fork:                 true,
			ParentRepository:     "parent/upstream",
			Topics:               []string{"experimental"},
			LanguageDistribution: map[string]int64{"Rust": 40},
			CreatedAt:            time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			PushedAt:             time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
}

// TestFromReposRankedTopics verifies B1: topics are unionized across the
// portfolio and ordered by descending frequency with a lexical tiebreak.
func TestFromReposRankedTopics(t *testing.T) {
	got := FromRepos(repoFixture(), portfolioRef)
	want := []string{"cli", "experimental", "go", "web"}
	if !reflect.DeepEqual(got.RankedTopics, want) {
		t.Fatalf("RankedTopics = %v, want %v", got.RankedTopics, want)
	}
	if got.TopicUniverse != len(want) {
		t.Fatalf("TopicUniverse = %d, want %d", got.TopicUniverse, len(want))
	}
}

// TestFromReposForkLineage verifies B2: fork concentration is aggregated by
// parent, ordered by descending fork count.
func TestFromReposForkLineage(t *testing.T) {
	got := FromRepos(repoFixture(), portfolioRef)
	want := []ForkParentFact{{Parent: "parent/upstream", Forks: 2}}
	if !reflect.DeepEqual(got.ForkLineage, want) {
		t.Fatalf("ForkLineage = %v, want %v", got.ForkLineage, want)
	}
}

// TestFromReposTechnologyTimeline verifies B4: all languages observed across
// repositories created in the same year are preserved (lossless), not collapsed
// to a single "dominant" language.
func TestFromReposTechnologyTimeline(t *testing.T) {
	got := FromRepos(repoFixture(), portfolioRef)
	want := map[int][]string{
		2019: {"Go", "Shell"},
		2021: {"TypeScript"},
		2022: {"Rust"},
		2023: {"Rust"},
	}
	if !reflect.DeepEqual(got.TechnologyTimeline, want) {
		t.Fatalf("TechnologyTimeline = %v, want %v", got.TechnologyTimeline, want)
	}
}

// TestFromReposMaintenanceBuckets verifies B5: maintenance recency is bucketed
// by last-push age relative to the reference time.
func TestFromReposMaintenanceBuckets(t *testing.T) {
	got := FromRepos(repoFixture(), portfolioRef)
	// a: pushed 2023-12-01 (<=90d) active; b: 2023-06-01 recent; c,d: 2018/2019 dormant.
	want := MaintenanceBuckets{Active: 1, Recent: 1, Dormant: 2}
	if got.MaintenanceBuckets != want {
		t.Fatalf("MaintenanceBuckets = %+v, want %+v", got.MaintenanceBuckets, want)
	}
}
