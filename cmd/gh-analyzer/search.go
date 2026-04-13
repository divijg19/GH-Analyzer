package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	searchpkg "github.com/divijg19/GH-Analyzer/internal/search"
)

func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printSearchHelp(fs.Output()) }
	preset := fs.String("preset", "", "preset name: strong, consistent, deep")
	limit := fs.Int("limit", defaultQueryLimit, "max results")
	kLimit := fs.Int("k", defaultQueryLimit, "max results (alias)")
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")
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

	resolvedLimit, err := resolveLimitFlag(fs, *limit, *kLimit)
	if err != nil {
		return err
	}
	if resolvedLimit < 0 {
		return fmt.Errorf("invalid --limit: must be >= 0")
	}

	indexData, err := loadDataset(*datasetPath)
	if err != nil {
		return err
	}

	allResults, err := searchpkg.Search(indexData, input, searchpkg.Options{
		Preset: strings.ToLower(strings.TrimSpace(*preset)),
		Limit:  0,
	})
	if err != nil {
		return err
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
