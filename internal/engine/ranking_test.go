package engine

import (
	"testing"

	"github.com/divijg19/GH-Analyzer/internal/index"
)

func TestWeightedRankingSkipsMissingSignals(t *testing.T) {
	ranking := WeightedRanking{}
	profile := index.Profile{Signals: map[string]float64{
		"consistency": 0.8,
		// ownership is intentionally missing
		"depth": 0.5,
	}}

	got := ranking.Score(profile)
	want := (0.8 * 0.5) + (0.5 * 0.2)
	if got != want {
		t.Fatalf("expected %.2f, got %.2f", want, got)
	}
}
