package engine

import "github.com/divijg19/GH-Analyzer/internal/index"

func defaultRankingWeights() map[string]float64 {
	return map[string]float64{
		"consistency": 0.4,
		"ownership":   0.3,
		"depth":       0.3,
	}
}

type RankingStrategy interface {
	Score(index.Profile) float64
}

type WeightedRanking struct {
	Weights map[string]float64
}

func (w WeightedRanking) Score(p index.Profile) float64 {
	weights := w.Weights
	if len(weights) == 0 {
		weights = defaultRankingWeights()
	}

	score := 0.0
	for signal, weight := range weights {
		value, ok := p.Signals[signal]
		if !ok {
			continue
		}

		score += value * weight
	}

	return score
}
