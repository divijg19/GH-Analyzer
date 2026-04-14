package search

import (
	"fmt"
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
