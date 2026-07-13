package engine

import (
	"math"
	"sort"

	"github.com/divijg19/Atlas/internal/index"
)

const minCalibrationDatasetSize = 10

type distribution struct {
	Overall []float64
}

func buildDistribution(profiles []index.Profile, ranking rankingStrategy) distribution {
	overall := make([]float64, 0, len(profiles))
	for _, profile := range profiles {
		overall = append(overall, ranking.Score(profile.Signals))
	}

	sortedOverall := append([]float64(nil), overall...)
	sort.Float64s(sortedOverall)

	return distribution{Overall: sortedOverall}
}

func percentile(sorted []float64, value float64) float64 {
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

func calibrateScore(distribution distribution, rawScore float64) float64 {
	if len(distribution.Overall) < minCalibrationDatasetSize {
		return rawScore
	}

	return percentile(distribution.Overall, rawScore)
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
