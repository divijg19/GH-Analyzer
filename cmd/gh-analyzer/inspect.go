package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
)

func runInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { printInspectHelp(fs.Output()) }
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")
	jsonOutput := fs.Bool("json", false, "output JSON")

	stop, err := parseFlagsOrHelp(fs, args)
	if err != nil {
		return err
	}
	if stop {
		return nil
	}

	if len(fs.Args()) < 1 {
		return fmt.Errorf("username is required")
	}

	username := strings.TrimSpace(fs.Args()[0])
	if username == "" {
		return fmt.Errorf("username is required")
	}

	indexData, err := loadDataset(*datasetPath)
	if err != nil {
		return err
	}

	profile, found := findProfile(indexData, username)
	if !found {
		return fmt.Errorf("profile %q not found", username)
	}
	if *jsonOutput {
		return writeJSON(profile)
	}
	printProfileSignals(profile)

	return nil
}

func findProfile(indexData indexpkg.Index, username string) (indexpkg.Profile, bool) {
	for _, profile := range indexData.All() {
		if strings.EqualFold(profile.Username, username) {
			return profile, true
		}
	}

	return indexpkg.Profile{}, false
}
