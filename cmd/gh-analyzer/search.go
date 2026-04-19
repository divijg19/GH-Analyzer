package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/engine"
	searchpkg "github.com/divijg19/GH-Analyzer/internal/search"
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
		idx, err := loadLiveDataset(input)
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
			return writeJSON([]engine.Result{})
		}
		fmt.Println("No candidates found.")
		return nil
	}

	results := allResults
	if resolvedLimit > 0 && len(results) > resolvedLimit {
		results = results[:resolvedLimit]
	}

	if *jsonOutput {
		return writeJSON(results)
	}

	printMatchSummary(len(allResults), len(results), resolvedLimit)
	if *compactOutput {
		printCompactResults(results)
	} else {
		printGroupedCandidates(results)
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
