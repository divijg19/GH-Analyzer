package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/contributions"
	"github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/profile"
	"github.com/divijg19/Atlas/internal/signals"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "dataset.json")

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	input := index.Index{Profiles: []index.Profile{
		{
			Username: "alice",
			Signals:  map[string]float64{"consistency": 0.8, "ownership": 0.7, "depth": 0.6, "activity": 0.9},
			Facts: &signals.RepositoryFacts{
				TotalRepos:         42,
				OriginalRepos:      38,
				ForkRepos:          4,
				RecentRepos:        35,
				DeepRepos:          12,
				ValidRepos:         39,
				ValidOriginalRepos: 35,
				LargestRepoSize:    2048,
				LatestActivity:     now,
			},
			Metadata: &profile.UserMetadata{
				Name:      "Alice User",
				Bio:       "Go developer",
				Location:  "San Francisco",
				Company:   "@acme",
				Followers: 150,
				Following: 80,
				CreatedAt: now.AddDate(-3, 0, 0),
			},
			Contributions: &contributions.Summary{
				TotalContributions: 200,
				TotalPullRequests:  120,
				IssuesOpened:       80,
			},
		},
		{
			Username: "bob",
			Signals:  map[string]float64{"consistency": 0.5, "ownership": 0.9, "depth": 0.4, "activity": 0.2},
		},
	}}

	if err := Save(path, input); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if !reflect.DeepEqual(input, loaded) {
		t.Fatalf("round trip mismatch:\n  expected %+v\n  got      %+v", input, loaded)
	}
}

func TestLoadNilProfilesDefaultsToEmptySlice(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "dataset.json")

	if err := os.WriteFile(path, []byte(`{"profiles":null}`), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Profiles == nil {
		t.Fatal("expected non-nil empty profiles slice")
	}
	if len(loaded.Profiles) != 0 {
		t.Fatalf("expected empty profiles, got %d", len(loaded.Profiles))
	}
}
