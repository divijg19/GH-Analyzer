package evaluation

import (
	"testing"

	"github.com/divijg19/Atlas/internal/indicators"
)

func TestOverallScore(t *testing.T) {
	cases := []struct {
		name string
		rs   indicators.RawScore
		want int
	}{
		{
			name: "all max",
			rs:   indicators.RawScore{Consistency: 100, Ownership: 100, Depth: 100},
			want: 100,
		},
		{
			name: "consistency only",
			rs:   indicators.RawScore{Consistency: 100, Ownership: 0, Depth: 0},
			want: 40,
		},
		{
			name: "uniform half",
			rs:   indicators.RawScore{Consistency: 50, Ownership: 50, Depth: 50},
			want: 50,
		},
		{
			name: "zero",
			rs:   indicators.RawScore{Consistency: 0, Ownership: 0, Depth: 0},
			want: 0,
		},
		{
			name: "truncates fractional component",
			rs:   indicators.RawScore{Consistency: 33, Ownership: 33, Depth: 33},
			// 33*0.4 + 33*0.3 + 33*0.3 = 13.2 + 9.9 + 9.9 = 33.0
			want: 33,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := OverallScore(tc.rs)
			if got != tc.want {
				t.Fatalf("OverallScore(%+v) = %d, want %d", tc.rs, got, tc.want)
			}
		})
	}
}

func TestRankingPolicyScore(t *testing.T) {
	signals := map[string]float64{
		"consistency": 1.0,
		"ownership":   1.0,
		"depth":       1.0,
	}

	// Default weights sum to 1.0 for fully-saturated indicator values.
	if got := (RankingPolicy{}).Score(signals); got != 1.0 {
		t.Fatalf("default RankingPolicy.Score = %v, want 1.0", got)
	}

	// Custom weights are honoured.
	custom := RankingPolicy{Weights: map[string]float64{
		"consistency": 1.0,
		"ownership":   0.0,
		"depth":       0.0,
	}}
	if got := custom.Score(signals); got != 1.0 {
		t.Fatalf("custom RankingPolicy.Score = %v, want 1.0", got)
	}

	// Missing signal keys contribute nothing.
	sparse := map[string]float64{"depth": 0.5}
	if got := (RankingPolicy{}).Score(sparse); got != 0.5*depthWeight {
		t.Fatalf("sparse RankingPolicy.Score = %v, want %v", got, 0.5*depthWeight)
	}
}

func TestApplySmallSamplePenalty(t *testing.T) {
	cases := []struct {
		name      string
		raw       int
		repoCount int
		want      int
	}{
		{name: "at threshold no penalty", raw: 100, repoCount: 3, want: 100},
		{name: "above threshold no penalty", raw: 100, repoCount: 10, want: 100},
		{name: "below threshold penalised", raw: 100, repoCount: 2, want: 70},
		{name: "zero repos penalised", raw: 80, repoCount: 0, want: 56},
		{name: "single repo penalised", raw: 90, repoCount: 1, want: 63},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ApplySmallSamplePenalty(tc.raw, tc.repoCount)
			if got != tc.want {
				t.Fatalf("ApplySmallSamplePenalty(%d, %d) = %d, want %d", tc.raw, tc.repoCount, got, tc.want)
			}
		})
	}
}

func TestDefaultRankingWeights(t *testing.T) {
	w := defaultRankingWeights()
	if w["consistency"] != consistencyWeight {
		t.Fatalf("consistency weight = %v, want %v", w["consistency"], consistencyWeight)
	}
	if w["ownership"] != ownershipWeight {
		t.Fatalf("ownership weight = %v, want %v", w["ownership"], ownershipWeight)
	}
	if w["depth"] != depthWeight {
		t.Fatalf("depth weight = %v, want %v", w["depth"], depthWeight)
	}

	// Weights must sum to 1.0 — the overall score is a convex combination.
	var sum float64
	for _, v := range w {
		sum += v
	}
	if sum != 1.0 {
		t.Fatalf("ranking weights sum = %v, want 1.0", sum)
	}
}
