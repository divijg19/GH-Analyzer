package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/signals"
)

func runAnalyze(username string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("missing GitHub username")
	}

	repos, err := signals.FetchRepos(username)
	if err != nil {
		return err
	}

	signalValues := signals.ExtractSignals(repos)
	scores := signals.ScoreSignals(signalValues)
	if len(repos) < minReposForFullScore {
		scores.Overall = int(math.Round(float64(scores.Overall) * smallSampleOverallMultiplier))
	}

	report := signals.BuildReport(username, scores, repos)
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("Report for %s\n\n", report.Username)
	fmt.Println(string(output))

	return nil
}
