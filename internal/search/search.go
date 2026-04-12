package search

import (
	"fmt"
	"math"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	"github.com/divijg19/GH-Analyzer/internal/index"
)

type Options struct {
	Preset string
	Limit  int
}

func Search(idx index.Index, input string, opts ...Options) ([]engine.Result, error) {
	options := Options{}
	if len(opts) > 0 {
		options = opts[0]
	}

	if options.Limit < 0 {
		return nil, fmt.Errorf("invalid limit: must be >= 0")
	}

	query, err := queryFromOptions(input, options)
	if err != nil {
		return nil, err
	}

	results := engine.Execute(idx, query, engine.WeightedRanking{})
	normalizeScores(results)

	return results, nil
}

func queryFromOptions(input string, options Options) (engine.Query, error) {
	preset := strings.ToLower(strings.TrimSpace(options.Preset))

	query := engine.Query{}
	var err error

	if preset != "" {
		query, err = queryFromPreset(preset)
	} else {
		query, err = MapIntent(input)
	}
	if err != nil {
		return engine.Query{}, err
	}

	if options.Limit > 0 {
		query.Limit = options.Limit
	}

	return query, nil
}

func normalizeScores(results []engine.Result) {
	if len(results) == 0 {
		return
	}

	maxScore := results[0].Score
	for _, result := range results {
		if result.Score > maxScore {
			maxScore = result.Score
		}
	}

	if maxScore == 0 {
		return
	}

	for i := range results {
		normalized := results[i].Score / maxScore
		results[i].Score = clamp01(normalized)
	}
}

func clamp01(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}
