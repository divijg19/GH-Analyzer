package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/divijg19/Atlas/internal/acquisition"
	"github.com/divijg19/Atlas/internal/index"
	"github.com/divijg19/Atlas/internal/intelligence"
	"github.com/divijg19/Atlas/internal/projection"
)

type analyzeOptions struct {
	Username string
	JSON     bool
}

func runAnalyze(args []string) error {
	options, showHelp, err := parseAnalyzeArgs(args)
	if err != nil {
		return err
	}
	if showHelp {
		printAnalyzeHelp(os.Stdout)
		return nil
	}

	proj, err := analyzeUser(options.Username)
	if err != nil {
		return err
	}

	if options.JSON {
		return writeJSON(proj)
	}

	output, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("Analysis: %s\n\n", proj.Username)
	fmt.Println(string(output))

	return nil
}

func parseAnalyzeArgs(args []string) (analyzeOptions, bool, error) {
	options := analyzeOptions{}

	for _, arg := range args {
		value := strings.TrimSpace(arg)
		if value == "" {
			continue
		}

		switch value {
		case "--help", "-h":
			return analyzeOptions{}, true, nil
		case "--json":
			options.JSON = true
		default:
			if strings.HasPrefix(value, "-") {
				return analyzeOptions{}, false, fmt.Errorf("unknown analyze option %q", value)
			}
			if options.Username != "" {
				return analyzeOptions{}, false, fmt.Errorf("unexpected extra argument %q", value)
			}
			options.Username = value
		}
	}

	if options.Username == "" {
		return analyzeOptions{}, false, fmt.Errorf("missing GitHub username")
	}

	return options, false, nil
}

// analyzeUser builds the AnalyzeProjection (signals, scores, top repositories)
// and attaches the Candidate Intelligence view built from the same Profile.
// Profile assembly uses the single canonical owner (index.BuildProfile).
func analyzeUser(username string) (projection.AnalyzeProjection, error) {
	if strings.TrimSpace(username) == "" {
		return projection.AnalyzeProjection{}, fmt.Errorf("missing GitHub username")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := acquisition.NewClient()

	// Existing projection path (signals, scores, top repositories).
	repos, err := client.FetchReposNormalized(ctx, username)
	if err != nil {
		return projection.AnalyzeProjection{}, err
	}
	proj, err := projection.BuildAnalyzeProjection(username, repos, time.Now())
	if err != nil {
		return projection.AnalyzeProjection{}, err
	}

	// Candidate Intelligence: canonical semantic interpretation of the Profile.
	profile, err := index.BuildProfile(ctx, client, username, time.Now())
	if err != nil {
		return projection.AnalyzeProjection{}, err
	}
	ci, err := intelligence.BuildCandidateIntelligence(ctx, &profile, time.Now())
	if err != nil {
		return projection.AnalyzeProjection{}, err
	}
	proj.Intelligence = projection.IntelligenceView(ci)

	return proj, nil
}
