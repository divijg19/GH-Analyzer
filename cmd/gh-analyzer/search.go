package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	searchpkg "github.com/divijg19/GH-Analyzer/internal/search"
	"github.com/divijg19/GH-Analyzer/internal/storage"
)

func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printSearchHelp(fs.Output()) }
	preset := fs.String("preset", "", "preset name: strong, consistent, deep")
	limit := fs.Int("limit", defaultQueryLimit, "max results")
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")
	jsonOutput := fs.Bool("json", false, "output JSON")

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

	indexData, err := storage.Load(*datasetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return missingDatasetError(*datasetPath)
		}
		return err
	}

	results, err := searchpkg.Search(indexData, input, searchpkg.Options{
		Preset: strings.ToLower(strings.TrimSpace(*preset)),
		Limit:  *limit,
	})
	if err != nil {
		return err
	}

	if *jsonOutput {
		return writeJSON(results)
	}

	printTopCandidates(results)
	return nil
}
