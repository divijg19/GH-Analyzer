package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/storage"
)

func runInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	datasetPath := fs.String("dataset", defaultDatasetPath, "dataset file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(fs.Args()) < 1 {
		return fmt.Errorf("username is required")
	}

	username := strings.TrimSpace(fs.Args()[0])
	if username == "" {
		return fmt.Errorf("username is required")
	}

	indexData, err := storage.Load(*datasetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return missingDatasetError(*datasetPath)
		}
		return err
	}

	profile, found := findProfile(indexData, username)
	if !found {
		return fmt.Errorf("profile %q not found", username)
	}

	fmt.Printf("Profile: %s\n", profile.Username)
	fmt.Println("Signals:")

	names := make([]string, 0, len(profile.Signals))
	for name := range profile.Signals {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("- %s: %.2f\n", name, profile.Signals[name])
	}

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
