package ghanalyzer

func defaultRankingWeights() map[string]float64 {
	return map[string]float64{
		"consistency": 0.5,
		"ownership":   0.3,
		"depth":       0.2,
	}
}

type RankingStrategy interface {
	Score(Profile) float64
}

type WeightedRanking struct {
	Weights map[string]float64
}

func (w WeightedRanking) Score(p Profile) float64 {
	weights := w.Weights
	if len(weights) == 0 {
		weights = defaultRankingWeights()
	}

	score := 0.0
	for signal, weight := range weights {
		score += p.Signals[signal] * weight
	}

	return score
}
