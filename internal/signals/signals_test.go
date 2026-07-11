package signals

import (
	"math"
	"testing"
	"time"
)

var refTime = time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

func TestExtractSignalsConsistencyStabilitySmallDataset(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(20)},
	}

	s := ExtractSignals(repos, refTime)

	if !almostEqual(s.Consistency, 0.2) {
		t.Fatalf("expected consistency 0.20, got %.2f", s.Consistency)
	}
}

func TestExtractSignalsConsistencyStabilityManyRepos(t *testing.T) {
	repos := make([]RepositoryVestige, 0, 20)
	for i := 0; i < 16; i++ {
		repos = append(repos, RepositoryVestige{Fork: false, Size: 100, UpdatedAt: daysAgo(10)})
	}
	for i := 0; i < 4; i++ {
		repos = append(repos, RepositoryVestige{Fork: false, Size: 100, UpdatedAt: daysAgo(120)})
	}

	s := ExtractSignals(repos, refTime)

	if !almostEqual(s.Consistency, 0.8) {
		t.Fatalf("expected consistency 0.80, got %.2f", s.Consistency)
	}
}

func TestExtractSignalsDepthUsesNonForkAndSizeThreshold(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 120, UpdatedAt: daysAgo(20)},
		{Fork: false, Size: 20, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 1000, UpdatedAt: daysAgo(20)},
	}

	s := ExtractSignals(repos, refTime)

	if !almostEqual(s.Depth, 0.2) {
		t.Fatalf("expected depth 0.20, got %.2f", s.Depth)
	}
}

func TestExtractSignalsOwnershipIgnoresZeroSizeRepos(t *testing.T) {
	repos := []RepositoryVestige{
		{Fork: false, Size: 0, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 0, UpdatedAt: daysAgo(20)},
		{Fork: false, Size: 120, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 120, UpdatedAt: daysAgo(20)},
	}

	s := ExtractSignals(repos, refTime)

	if !almostEqual(s.Ownership, 0.5) {
		t.Fatalf("expected ownership 0.50, got %.2f", s.Ownership)
	}
}

func TestExtractSignalsActivityGrading(t *testing.T) {
	tests := []struct {
		name string
		days int
		want float64
	}{
		{name: "10 days ago", days: 10, want: 1.0},
		{name: "60 days ago", days: 60, want: 0.7},
		{name: "150 days ago", days: 150, want: 0.4},
		{name: "300 days ago", days: 300, want: 0.1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repos := []RepositoryVestige{{Fork: false, Size: 100, UpdatedAt: daysAgo(tc.days)}}
			s := ExtractSignals(repos, refTime)

			if !almostEqual(s.Activity, tc.want) {
				t.Fatalf("expected activity %.1f, got %.1f", tc.want, s.Activity)
			}
		})
	}
}

func TestExtractSignalsEdgeCases(t *testing.T) {
	t.Run("zero repos", func(t *testing.T) {
		s := ExtractSignals(nil, refTime)
		assertZeroSignals(t, s)
	})

	t.Run("all forks", func(t *testing.T) {
		repos := []RepositoryVestige{
			{Fork: true, Size: 100, UpdatedAt: daysAgo(5)},
			{Fork: true, Size: 50, UpdatedAt: daysAgo(6)},
		}
		s := ExtractSignals(repos, refTime)

		if s.Ownership != 0 {
			t.Fatalf("expected ownership 0, got %.2f", s.Ownership)
		}
		if s.Depth != 0 {
			t.Fatalf("expected depth 0, got %.2f", s.Depth)
		}
	})

	t.Run("all size zero", func(t *testing.T) {
		repos := []RepositoryVestige{
			{Fork: false, Size: 0, UpdatedAt: daysAgo(5)},
			{Fork: false, Size: 0, UpdatedAt: daysAgo(6)},
		}
		s := ExtractSignals(repos, refTime)

		if s.Ownership != 0 {
			t.Fatalf("expected ownership 0, got %.2f", s.Ownership)
		}
		if s.Depth != 0 {
			t.Fatalf("expected depth 0, got %.2f", s.Depth)
		}
	})

	t.Run("mixed data", func(t *testing.T) {
		repos := []RepositoryVestige{
			{Fork: false, Size: 100, UpdatedAt: daysAgo(20)},
			{Fork: true, Size: 500, UpdatedAt: daysAgo(40)},
			{Fork: false, Size: 0, UpdatedAt: daysAgo(140)},
			{Fork: false, Size: 60, UpdatedAt: daysAgo(300)},
		}
		s := ExtractSignals(repos, refTime)

		if s.Ownership <= 0 || s.Ownership >= 1 {
			t.Fatalf("expected ownership in (0,1), got %.2f", s.Ownership)
		}
		if s.Depth <= 0 || s.Depth >= 1 {
			t.Fatalf("expected depth in (0,1), got %.2f", s.Depth)
		}
		if s.Activity != 1.0 {
			t.Fatalf("expected activity 1.0 from most recent repo, got %.2f", s.Activity)
		}
	})
}

func TestScoreSignalsComponents(t *testing.T) {
	scores := ScoreSignals(Signals{Consistency: 1, Ownership: 0, Depth: 0})
	if scores.Consistency != 100 {
		t.Fatalf("expected consistency 100, got %d", scores.Consistency)
	}
	if scores.Ownership != 0 {
		t.Fatalf("expected ownership 0, got %d", scores.Ownership)
	}
	if scores.Depth != 0 {
		t.Fatalf("expected depth 0, got %d", scores.Depth)
	}
}

func TestSignalsToMap(t *testing.T) {
	sig := Signals{
		Ownership:   0.65,
		Consistency: 0.82,
		Depth:       0.47,
		Activity:    0.7,
	}

	m := SignalsToMap(sig)

	if !almostEqual(m["ownership"], 0.65) {
		t.Fatalf("expected ownership 0.65, got %.2f", m["ownership"])
	}
	if !almostEqual(m["consistency"], 0.82) {
		t.Fatalf("expected consistency 0.82, got %.2f", m["consistency"])
	}
	if !almostEqual(m["depth"], 0.47) {
		t.Fatalf("expected depth 0.47, got %.2f", m["depth"])
	}
	if !almostEqual(m["activity"], 0.7) {
		t.Fatalf("expected activity 0.70, got %.2f", m["activity"])
	}
}

func TestSignalsToMapClampsValues(t *testing.T) {
	sig := Signals{
		Ownership:   1.5,
		Consistency: -0.2,
		Depth:       0.5,
		Activity:    2.0,
	}

	m := SignalsToMap(sig)

	if m["ownership"] != 1.0 {
		t.Fatalf("expected ownership clamped to 1.0, got %.2f", m["ownership"])
	}
	if m["consistency"] != 0.0 {
		t.Fatalf("expected consistency clamped to 0.0, got %.2f", m["consistency"])
	}
	if m["activity"] != 1.0 {
		t.Fatalf("expected activity clamped to 1.0, got %.2f", m["activity"])
	}
}

func daysAgo(days int) time.Time {
	return refTime.Add(-time.Duration(days) * 24 * time.Hour)
}

func almostEqual(a, b float64) bool {
	const epsilon = 1e-9
	return math.Abs(a-b) <= epsilon
}

func assertZeroSignals(t *testing.T, s Signals) {
	t.Helper()

	if s.Ownership != 0 || s.Consistency != 0 || s.Depth != 0 || s.Activity != 0 {
		t.Fatalf("expected all signals to be zero, got %+v", s)
	}
}
