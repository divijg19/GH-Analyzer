package signals

import (
	"math"
	"testing"
	"time"
)

func TestExtractSignalsConsistencyStabilitySmallDataset(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(20)},
	}

	s := ExtractSignals(repos)

	if !almostEqual(s.Consistency, 0.2) {
		t.Fatalf("expected consistency 0.20, got %.2f", s.Consistency)
	}
}

func TestExtractSignalsConsistencyStabilityManyRepos(t *testing.T) {
	repos := make([]Repo, 0, 20)
	for i := 0; i < 16; i++ {
		repos = append(repos, Repo{Fork: false, Size: 100, UpdatedAt: daysAgo(10)})
	}
	for i := 0; i < 4; i++ {
		repos = append(repos, Repo{Fork: false, Size: 100, UpdatedAt: daysAgo(120)})
	}

	s := ExtractSignals(repos)

	if !almostEqual(s.Consistency, 0.8) {
		t.Fatalf("expected consistency 0.80, got %.2f", s.Consistency)
	}
}

func TestExtractSignalsDepthUsesNonForkAndSizeThreshold(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 120, UpdatedAt: daysAgo(20)},
		{Fork: false, Size: 20, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 1000, UpdatedAt: daysAgo(20)},
	}

	s := ExtractSignals(repos)

	if !almostEqual(s.Depth, 0.2) {
		t.Fatalf("expected depth 0.20, got %.2f", s.Depth)
	}
}

func TestExtractSignalsOwnershipIgnoresZeroSizeRepos(t *testing.T) {
	repos := []Repo{
		{Fork: false, Size: 0, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 0, UpdatedAt: daysAgo(20)},
		{Fork: false, Size: 120, UpdatedAt: daysAgo(20)},
		{Fork: true, Size: 120, UpdatedAt: daysAgo(20)},
	}

	s := ExtractSignals(repos)

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
			repos := []Repo{{Fork: false, Size: 100, UpdatedAt: daysAgo(tc.days)}}
			s := ExtractSignals(repos)

			if !almostEqual(s.Activity, tc.want) {
				t.Fatalf("expected activity %.1f, got %.1f", tc.want, s.Activity)
			}
		})
	}
}

func TestExtractSignalsEdgeCases(t *testing.T) {
	t.Run("zero repos", func(t *testing.T) {
		s := ExtractSignals(nil)
		assertZeroSignals(t, s)
	})

	t.Run("all forks", func(t *testing.T) {
		repos := []Repo{
			{Fork: true, Size: 100, UpdatedAt: daysAgo(5)},
			{Fork: true, Size: 50, UpdatedAt: daysAgo(6)},
		}
		s := ExtractSignals(repos)

		if s.Ownership != 0 {
			t.Fatalf("expected ownership 0, got %.2f", s.Ownership)
		}
		if s.Depth != 0 {
			t.Fatalf("expected depth 0, got %.2f", s.Depth)
		}
	})

	t.Run("all size zero", func(t *testing.T) {
		repos := []Repo{
			{Fork: false, Size: 0, UpdatedAt: daysAgo(5)},
			{Fork: false, Size: 0, UpdatedAt: daysAgo(6)},
		}
		s := ExtractSignals(repos)

		if s.Ownership != 0 {
			t.Fatalf("expected ownership 0, got %.2f", s.Ownership)
		}
		if s.Depth != 0 {
			t.Fatalf("expected depth 0, got %.2f", s.Depth)
		}
	})

	t.Run("mixed data", func(t *testing.T) {
		repos := []Repo{
			{Fork: false, Size: 100, UpdatedAt: daysAgo(20)},
			{Fork: true, Size: 500, UpdatedAt: daysAgo(40)},
			{Fork: false, Size: 0, UpdatedAt: daysAgo(140)},
			{Fork: false, Size: 60, UpdatedAt: daysAgo(300)},
		}
		s := ExtractSignals(repos)

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

func TestScoreSignalsWeightsAndPenalty(t *testing.T) {
	scores := ScoreSignals(Signals{Consistency: 1, Ownership: 0, Depth: 0})
	if scores.Overall != 40 {
		t.Fatalf("expected overall 40 with 0.4 consistency weight, got %d", scores.Overall)
	}

	reportWithPenalty := BuildReport("user", Scores{Consistency: 100, Ownership: 100, Depth: 100, Overall: 100}, []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(5)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
	})
	if reportWithPenalty.Scores.Overall != 70 {
		t.Fatalf("expected penalized overall 70 for dataset < 3 repos, got %d", reportWithPenalty.Scores.Overall)
	}

	reportAlreadyPenalized := BuildReport("user", Scores{Consistency: 100, Ownership: 100, Depth: 100, Overall: 70}, []Repo{
		{Fork: false, Size: 100, UpdatedAt: daysAgo(5)},
		{Fork: false, Size: 100, UpdatedAt: daysAgo(10)},
	})
	if reportAlreadyPenalized.Scores.Overall != 70 {
		t.Fatalf("expected overall to remain 70 when already penalized, got %d", reportAlreadyPenalized.Scores.Overall)
	}
}

func TestSignalsFromReportUsesRawSignalValues(t *testing.T) {
	report := Report{
		Scores: Scores{Consistency: 100, Ownership: 100, Depth: 100},
		SignalValues: Signals{
			Consistency: 0.23,
			Ownership:   0.61,
			Depth:       0.47,
			Activity:    0.7,
		},
		HasSignalValues: true,
	}

	s := SignalsFromReport(report)
	if !almostEqual(s["consistency"], 0.23) {
		t.Fatalf("expected consistency 0.23, got %.2f", s["consistency"])
	}
	if !almostEqual(s["ownership"], 0.61) {
		t.Fatalf("expected ownership 0.61, got %.2f", s["ownership"])
	}
	if !almostEqual(s["depth"], 0.47) {
		t.Fatalf("expected depth 0.47, got %.2f", s["depth"])
	}
	if !almostEqual(s["activity"], 0.7) {
		t.Fatalf("expected activity 0.70, got %.2f", s["activity"])
	}
}

func TestSignalsFromReportFallbackNormalization(t *testing.T) {
	report := Report{
		Scores: Scores{Consistency: 120, Ownership: -10, Depth: 50},
	}

	s := SignalsFromReport(report)

	if s["consistency"] != 1.0 {
		t.Fatalf("expected consistency 1.0, got %.2f", s["consistency"])
	}
	if s["ownership"] != 0.0 {
		t.Fatalf("expected ownership 0.0, got %.2f", s["ownership"])
	}
	if s["depth"] != 0.5 {
		t.Fatalf("expected depth 0.5, got %.2f", s["depth"])
	}
	if s["activity"] != 1.0 {
		t.Fatalf("expected fallback activity 1.0, got %.2f", s["activity"])
	}
}

func daysAgo(days int) time.Time {
	return time.Now().Add(-time.Duration(days) * 24 * time.Hour)
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
