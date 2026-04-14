package engine

import (
	"math"
	"sort"
	"testing"

	"github.com/divijg19/GH-Analyzer/internal/index"
)

func TestBuildDistributionSorted(t *testing.T) {
	profiles := []index.Profile{
		{Username: "c", Signals: map[string]float64{"consistency": 0.2}},
		{Username: "a", Signals: map[string]float64{"consistency": 0.9}},
		{Username: "b", Signals: map[string]float64{"consistency": 0.5}},
	}

	d := BuildDistribution(profiles, rankingByUsername(map[string]float64{
		"a": 0.9,
		"b": 0.5,
		"c": 0.2,
	}))

	want := []float64{0.2, 0.5, 0.9}
	if len(d.Overall) != len(want) {
		t.Fatalf("expected %d values, got %d", len(want), len(d.Overall))
	}

	for i := range want {
		if math.Abs(d.Overall[i]-want[i]) > 1e-9 {
			t.Fatalf("expected sorted[%d]=%.2f, got %.2f", i, want[i], d.Overall[i])
		}
	}

	if profiles[0].Username != "c" || profiles[1].Username != "a" || profiles[2].Username != "b" {
		t.Fatal("BuildDistribution mutated input profile order")
	}
}

func TestPercentile(t *testing.T) {
	sorted := []float64{0.2, 0.5, 0.8}

	if got := Percentile(sorted, 0.2); math.Abs(got-(1.0/3.0)) > 1e-9 {
		t.Fatalf("expected ~0.33, got %.8f", got)
	}
	if got := Percentile(sorted, 0.5); math.Abs(got-(2.0/3.0)) > 1e-9 {
		t.Fatalf("expected ~0.66, got %.8f", got)
	}
	if got := Percentile(sorted, 0.8); math.Abs(got-1.0) > 1e-9 {
		t.Fatalf("expected 1.0, got %.8f", got)
	}
}

func TestPercentileEdgeCases(t *testing.T) {
	sorted := []float64{0.2, 0.5, 0.5, 0.8}

	if got := Percentile(sorted, 0.5); math.Abs(got-0.75) > 1e-9 {
		t.Fatalf("expected duplicate percentile 0.75, got %.8f", got)
	}
	if got := Percentile(sorted, -1.0); math.Abs(got-0.25) > 1e-9 {
		t.Fatalf("expected below-range percentile 0.25, got %.8f", got)
	}
	if got := Percentile(sorted, 1.2); math.Abs(got-1.0) > 1e-9 {
		t.Fatalf("expected above-range percentile 1.0, got %.8f", got)
	}
}

func TestExecuteCalibrationStabilityPreservesOrdering(t *testing.T) {
	idx := index.Index{Profiles: []index.Profile{
		{Username: "u01", Signals: map[string]float64{"consistency": 0.10, "ownership": 0.10, "depth": 0.10}},
		{Username: "u02", Signals: map[string]float64{"consistency": 0.20, "ownership": 0.20, "depth": 0.20}},
		{Username: "u03", Signals: map[string]float64{"consistency": 0.30, "ownership": 0.30, "depth": 0.30}},
		{Username: "u04", Signals: map[string]float64{"consistency": 0.40, "ownership": 0.40, "depth": 0.40}},
		{Username: "u05", Signals: map[string]float64{"consistency": 0.50, "ownership": 0.50, "depth": 0.50}},
		{Username: "u06", Signals: map[string]float64{"consistency": 0.60, "ownership": 0.60, "depth": 0.60}},
		{Username: "u07", Signals: map[string]float64{"consistency": 0.70, "ownership": 0.70, "depth": 0.70}},
		{Username: "u08", Signals: map[string]float64{"consistency": 0.80, "ownership": 0.80, "depth": 0.80}},
		{Username: "u09", Signals: map[string]float64{"consistency": 0.90, "ownership": 0.90, "depth": 0.90}},
		{Username: "u10", Signals: map[string]float64{"consistency": 1.00, "ownership": 1.00, "depth": 1.00}},
	}}

	ranking := WeightedRanking{}
	results := Execute(idx, Query{}, ranking)

	type raw struct {
		username string
		score    float64
	}
	rawScores := make([]raw, 0, len(idx.Profiles))
	for _, p := range idx.Profiles {
		rawScores = append(rawScores, raw{username: p.Username, score: ranking.Score(p)})
	}
	sort.SliceStable(rawScores, func(i, j int) bool {
		if rawScores[i].score == rawScores[j].score {
			return rawScores[i].username < rawScores[j].username
		}

		return rawScores[i].score > rawScores[j].score
	})

	if len(results) != len(rawScores) {
		t.Fatalf("expected %d results, got %d", len(rawScores), len(results))
	}

	for i := range results {
		if results[i].Profile.Username != rawScores[i].username {
			t.Fatalf("ordering changed at position %d: expected %s, got %s", i, rawScores[i].username, results[i].Profile.Username)
		}
	}
}

func TestExecuteSmallDatasetReturnsRawScores(t *testing.T) {
	idx := index.Index{Profiles: []index.Profile{
		{Username: "a", Signals: map[string]float64{"consistency": 0.9, "ownership": 0.7, "depth": 0.8}},
		{Username: "b", Signals: map[string]float64{"consistency": 0.4, "ownership": 0.5, "depth": 0.6}},
		{Username: "c", Signals: map[string]float64{"consistency": 0.1, "ownership": 0.3, "depth": 0.2}},
	}}

	ranking := WeightedRanking{}
	results := Execute(idx, Query{}, ranking)

	for _, result := range results {
		raw := ranking.Score(result.Profile)
		if math.Abs(result.Score-raw) > 1e-9 {
			t.Fatalf("expected raw score %.8f for %s, got %.8f", raw, result.Profile.Username, result.Score)
		}
	}
}

type rankingByUsername map[string]float64

func (r rankingByUsername) Score(p index.Profile) float64 {
	return r[p.Username]
}
