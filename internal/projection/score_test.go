package projection

import (
	"testing"
	"time"

	obs "github.com/divijg19/Atlas/internal/observations"
)

func TestExtractTopRepositories(t *testing.T) {
	repos := []obs.RepositoryVestige{
		{Name: "a", Size: 10, Fork: false},
		{Name: "b", Size: 50, Fork: false},
		{Name: "c", Size: 30, Fork: false},
		{Name: "d", Size: 100, Fork: true}, // forks are excluded
		{Name: "e", Size: 40, Fork: false},
	}

	top := extractTopRepositories(repos, 3)
	if len(top) != 3 {
		t.Fatalf("expected 3 top repos, got %d", len(top))
	}

	// Non-forks by size descending: 50(b), 40(e), 30(c).
	if top[0].Name != "b" || top[0].Size != 50 {
		t.Fatalf("top[0] = %+v, want b/50", top[0])
	}
	if top[1].Name != "e" || top[1].Size != 40 {
		t.Fatalf("top[1] = %+v, want e/40", top[1])
	}
	if top[2].Name != "c" || top[2].Size != 30 {
		t.Fatalf("top[2] = %+v, want c/30", top[2])
	}
}

func TestExtractTopRepositoriesLimitExceedsCount(t *testing.T) {
	repos := []obs.RepositoryVestige{
		{Name: "a", Size: 10, Fork: false},
		{Name: "b", Size: 50, Fork: false},
	}
	top := extractTopRepositories(repos, 10)
	if len(top) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(top))
	}
}

func TestExtractTopRepositoriesEmpty(t *testing.T) {
	if got := extractTopRepositories(nil, 3); len(got) != 0 {
		t.Fatalf("expected empty slice, got %d", len(got))
	}
}

func TestBuildAnalyzeProjection(t *testing.T) {
	now := time.Now()
	repos := []obs.RepositoryVestige{
		{Name: "big", Size: 100, Fork: false, UpdatedAt: now},
		{Name: "small", Size: 20, Fork: false, UpdatedAt: now},
		{Name: "fork", Size: 200, Fork: true, UpdatedAt: now},
	}

	proj, err := BuildAnalyzeProjection("alice", repos, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.Username != "alice" {
		t.Fatalf("username = %q, want alice", proj.Username)
	}

	// Only non-forks are surfaced; both non-forks fit within the top limit.
	if len(proj.TopRepos) != 2 {
		t.Fatalf("expected 2 top repos, got %d", len(proj.TopRepos))
	}
	if proj.TopRepos[0].Name != "big" {
		t.Fatalf("top[0] = %q, want big", proj.TopRepos[0].Name)
	}

	// With 3 repos no small-sample penalty applies; overall stays in range and
	// is deterministic across calls.
	if proj.Overall <= 0 || proj.Overall > 100 {
		t.Fatalf("overall out of range: %d", proj.Overall)
	}

	again, err := BuildAnalyzeProjection("alice", repos, now)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if again.Overall != proj.Overall {
		t.Fatalf("overall not deterministic: %d != %d", again.Overall, proj.Overall)
	}
}
