package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/intelligence"
	"github.com/divijg19/Atlas/internal/projection"
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
	referenceTime := time.Now()
	proj := projection.BuildInspectProjection(profile)
	if ci, err := intelligence.BuildCandidateIntelligence(context.Background(), &profile, referenceTime); err == nil {
		proj.Intelligence = projection.IntelligenceView(ci)
	}
	proj.RepositoryIntelligence = projection.RepositoryIntelligenceViews(profile.Repositories, referenceTime)
	if *jsonOutput {
		return writeJSON(proj)
	}
	printInspectProjection(proj)

	return nil
}

func findProfile(indexData index.Index, username string) (index.Profile, bool) {
	for _, profile := range indexData.All() {
		if strings.EqualFold(profile.Username, username) {
			return profile, true
		}
	}

	return index.Profile{}, false
}
