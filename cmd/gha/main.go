package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	ghanalyzer "github.com/divijg19/GH-Analyzer"
)

const (
	minReposForFullScore         = 3
	smallSampleOverallMultiplier = 0.7
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "error: missing GitHub username")
		os.Exit(1)
	}

	username := os.Args[1]

	repos, err := ghanalyzer.FetchRepos(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	signals := ghanalyzer.ExtractSignals(repos)

	scores := ghanalyzer.ScoreSignals(signals)
	if len(repos) < minReposForFullScore {
		scores.Overall = int(math.Round(float64(scores.Overall) * smallSampleOverallMultiplier))
	}
	report := ghanalyzer.BuildReport(username, scores)

	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
