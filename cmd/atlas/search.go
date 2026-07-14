package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/divijg19/Atlas/internal/engine"
	"github.com/divijg19/Atlas/internal/evaluation"
	"github.com/divijg19/Atlas/internal/intelligence"
	"github.com/divijg19/Atlas/internal/projection"
	searchpkg "github.com/divijg19/Atlas/internal/search"
)

var loadLiveDataset = buildLiveIndex

func runSearch(args []string) error {
	args, liveFromPosition := stripLiveFlag(args)

	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printSearchHelp(fs.Output()) }
	preset := fs.String("preset", "", "preset name: strong, consistent, deep")
	limit := fs.Int("limit", defaultQueryLimit, "max results")
	kLimit := fs.Int("k", defaultQueryLimit, "max results (alias)")
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")
	live := fs.Bool("live", false, "fetch temporary candidates from live GitHub search")
	jsonOutput := fs.Bool("json", false, "output JSON")
	compactOutput := fs.Bool("compact", false, "compact output (no explanations)")

	stop, err := parseFlagsOrHelp(fs, args)
	if err != nil {
		return err
	}
	if stop {
		return nil
	}

	input := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if input == "" && strings.TrimSpace(*preset) == "" {
		return fmt.Errorf("search input is required")
	}

	liveMode := *live || liveFromPosition
	if liveMode && input == "" {
		return fmt.Errorf("search input is required")
	}

	resolvedLimit, err := resolveLimitFlag(fs, *limit, *kLimit)
	if err != nil {
		return err
	}
	if resolvedLimit < 0 {
		return fmt.Errorf("invalid --limit: must be >= 0")
	}

	var allResults []engine.Result
	liveCandidateCount := 0

	if liveMode {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		idx, err := loadLiveDataset(ctx, input)
		if err != nil {
			return err
		}
		liveCandidateCount = len(idx.All())

		allResults, err = searchpkg.Search(idx, input, searchpkg.Options{
			Preset: strings.ToLower(strings.TrimSpace(*preset)),
			Limit:  0,
		})
		if err != nil {
			return err
		}
	} else {
		idx, err := loadDataset(*datasetPath)
		if err != nil {
			return err
		}

		allResults, err = searchpkg.Search(idx, input, searchpkg.Options{
			Preset: strings.ToLower(strings.TrimSpace(*preset)),
			Limit:  0,
		})
		if err != nil {
			return err
		}
	}

	if liveMode && liveCandidateCount == 0 {
		if *jsonOutput {
			return writeJSON([]projection.SearchProjection{})
		}
		fmt.Println("No candidates found.")
		return nil
	}

	results := allResults
	if resolvedLimit > 0 && len(results) > resolvedLimit {
		results = results[:resolvedLimit]
	}

	projections := make([]projection.SearchProjection, len(results))
	for i, result := range results {
		projections[i] = projection.BuildSearchProjection(
			result.Profile,
			result.Score,
			evaluation.ClassifyConfidence(result.Score),
			result.Reasons,
		)
		if ci, err := intelligence.BuildCandidateIntelligence(context.Background(), &result.Profile, time.Now()); err == nil {
			projections[i].Intelligence = projection.IntelligenceView(ci)
		}
	}

	if *jsonOutput {
		return writeJSON(projections)
	}

	printMatchSummary(len(allResults), len(results), resolvedLimit)
	if *compactOutput {
		printCompactSearchProjections(projections)
	} else {
		printGroupedSearchProjections(projections)
	}
	return nil
}

func stripLiveFlag(args []string) ([]string, bool) {
	result := make([]string, 0, len(args))
	live := false

	for _, arg := range args {
		if strings.TrimSpace(arg) == "--live" {
			live = true
			continue
		}
		result = append(result, arg)
	}

	return result, live
}
