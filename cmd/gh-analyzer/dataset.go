package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func runDataset(args []string) error {
	if len(args) > 0 && strings.EqualFold(strings.TrimSpace(args[0]), "stats") {
		return runDatasetStats(args[1:])
	}
	if len(args) > 0 && strings.EqualFold(strings.TrimSpace(args[0]), "info") {
		return runDatasetInfo(args[1:])
	}
	if len(args) > 0 && strings.EqualFold(strings.TrimSpace(args[0]), "preview") {
		return runDatasetPreview(args[1:])
	}

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

	indexData, err := loadDataset(*datasetPath)
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writeJSON(indexData)
	}

	printDatasetSummary(*datasetPath, indexData)
	return nil
}

func runDatasetStats(args []string) error {
	fs := flag.NewFlagSet("dataset stats", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printDatasetHelp(fs.Output()) }
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")

	stop, err := parseFlagsOrHelp(fs, args)
	if err != nil {
		return err
	}
	if stop {
		return nil
	}
	if len(fs.Args()) > 0 {
		return fmt.Errorf("unexpected dataset stats argument %q", fs.Args()[0])
	}

	indexData, err := loadDataset(*datasetPath)
	if err != nil {
		return err
	}

	printDatasetStats(*datasetPath, indexData)
	return nil
}

func runDatasetInfo(args []string) error {
	fs := flag.NewFlagSet("dataset info", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printDatasetHelp(fs.Output()) }
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")

	stop, err := parseFlagsOrHelp(fs, args)
	if err != nil {
		return err
	}
	if stop {
		return nil
	}
	if len(fs.Args()) > 0 {
		return fmt.Errorf("unexpected dataset info argument %q", fs.Args()[0])
	}

	indexData, err := loadDataset(*datasetPath)
	if err != nil {
		return err
	}

	printDatasetInfo(*datasetPath, indexData)
	return nil
}

func runDatasetPreview(args []string) error {
	fs := flag.NewFlagSet("dataset preview", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printDatasetHelp(fs.Output()) }
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")

	stop, err := parseFlagsOrHelp(fs, args)
	if err != nil {
		return err
	}
	if stop {
		return nil
	}
	if len(fs.Args()) > 0 {
		return fmt.Errorf("unexpected dataset preview argument %q", fs.Args()[0])
	}

	indexData, err := loadDataset(*datasetPath)
	if err != nil {
		return err
	}

	printDatasetPreview(*datasetPath, indexData)
	return nil
}
