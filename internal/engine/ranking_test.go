package engine

import (
	"math"
	"testing"

	"github.com/divijg19/Atlas/internal/evaluation"
)

func TestWeightedRankingSkipsMissingSignals(t *testing.T) {
	ranking := evaluation.RankingPolicy{}
	signals := map[string]float64{
		"consistency": 0.8,
		// ownership is intentionally missing
		"depth": 0.5,
	}

	got := ranking.Score(signals)
	want := (0.8 * 0.4) + (0.5 * 0.3)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("expected %.2f, got %.2f", want, got)
	}
}
