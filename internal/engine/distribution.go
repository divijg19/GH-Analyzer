package engine

import (
	"math"
	"sort"

	"github.com/divijg19/GH-Analyzer/internal/index"
)

const minCalibrationDatasetSize = 10

type Distribution struct {
	Overall []float64
}

func BuildDistribution(profiles []index.Profile, ranking RankingStrategy) Distribution {
	if ranking == nil {
		ranking = WeightedRanking{}
	}

	overall := make([]float64, 0, len(profiles))
	for _, profile := range profiles {
		overall = append(overall, ranking.Score(profile))
	}

	sortedOverall := append([]float64(nil), overall...)
	sort.Float64s(sortedOverall)

	return Distribution{Overall: sortedOverall}
}

func Percentile(sorted []float64, value float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	if value <= sorted[0] {
		return clamp01(float64(1) / float64(len(sorted)))
	}
	if value >= sorted[len(sorted)-1] {
		return 1.0
	}

	// Upper-bound index gives a monotonic percentile rank for duplicates.
	count := sort.Search(len(sorted), func(i int) bool {
		return sorted[i] > value
	})

	return clamp01(float64(count) / float64(len(sorted)))
}

func CalibrateScore(distribution Distribution, rawScore float64) float64 {
	if len(distribution.Overall) < minCalibrationDatasetSize {
		return rawScore
	}

	return Percentile(distribution.Overall, rawScore)
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}

	if math.IsNaN(value) {
		return 0
	}

	return value
}
