package main

import (
	"errors"
	"flag"
	"os"

	"github.com/divijg19/GH-Analyzer/internal/storage"
)

func runDataset(args []string) error {
	fs := flag.NewFlagSet("dataset", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printDatasetHelp(fs.Output()) }
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")
	jsonOutput := fs.Bool("json", false, "output JSON")

	stop, err := parseFlagsOrHelp(fs, args)
	if err != nil {
		return err
	}
	if stop {
		return nil
	}

	indexData, err := storage.Load(*datasetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return missingDatasetError(*datasetPath)
		}
		return err
	}
	if *jsonOutput {
		return writeJSON(indexData)
	}

	printDatasetSummary(*datasetPath, indexData)
	return nil
}
