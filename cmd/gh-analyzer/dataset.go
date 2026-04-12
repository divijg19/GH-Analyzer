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
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	indexData, err := storage.Load(*datasetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return missingDatasetError(*datasetPath)
		}
		return err
	}

	printDatasetSummary(*datasetPath, indexData)
	return nil
}
