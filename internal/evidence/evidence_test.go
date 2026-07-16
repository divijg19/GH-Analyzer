package evidence

import (
	"reflect"
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/indicators"
)

// TestGenerateEvidenceGroupOrder verifies the documented invariant: the
// evidence group order is fixed (ownership, consistency, depth, activity) and
// never depends on input ordering.
func TestGenerateEvidenceGroupOrder(t *testing.T) {
	f := facts.RepositoryFacts{
		ValidRepos:         10,
		ValidOriginalRepos: 7,
		OriginalRepos:      9,
		RecentRepos:        6,
		DeepRepos:          4,
		LargestRepoSize:    2048,
		LatestActivity:     mustTime(2026, 6, 1),
	}
	s := indicators.Signals{Ownership: 0.7, Consistency: 0.6, Depth: 0.4, Activity: 1.0}

	groups := GenerateEvidence(f, s)
	want := []string{indicators.SignalOwnership, indicators.SignalConsistency, indicators.SignalDepth, indicators.SignalActivity, PortfolioGroup}
	if len(groups) != len(want) {
		t.Fatalf("expected %d groups, got %d", len(want), len(groups))
	}
	for i, g := range groups {
		if g.Signal != want[i] {
			t.Fatalf("group %d = %q, want %q", i, g.Signal, want[i])
		}
	}
}

// TestGenerateEvidenceDeterministic verifies the documented invariant:
// GenerateEvidence is a pure function of (facts, signals).
func TestGenerateEvidenceDeterministic(t *testing.T) {
	f := facts.RepositoryFacts{
		ValidRepos:         10,
		ValidOriginalRepos: 7,
		OriginalRepos:      9,
		RecentRepos:        6,
		DeepRepos:          4,
		LargestRepoSize:    2048,
		LatestActivity:     mustTime(2026, 6, 1),
	}
	s := indicators.Signals{Ownership: 0.7, Consistency: 0.6, Depth: 0.4, Activity: 1.0}

	first := GenerateEvidence(f, s)
	for i := 0; i < 1000; i++ {
		if got := GenerateEvidence(f, s); !reflect.DeepEqual(first, got) {
			t.Fatalf("GenerateEvidence not deterministic on iteration %d", i)
		}
	}
}

func mustTime(year int, month int, day int) time.Time {
	return time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC)
}
