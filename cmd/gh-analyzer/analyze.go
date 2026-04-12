package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/signals"
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

	report, err := analyzeUser(options.Username)
	if err != nil {
		return err
	}

	if options.JSON {
		return writeJSON(report)
	}

	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("Report: %s\n\n", report.Username)
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

func analyzeUser(username string) (signals.Report, error) {
	if strings.TrimSpace(username) == "" {
		return signals.Report{}, fmt.Errorf("missing GitHub username")
	}

	repos, err := signals.FetchRepos(username)
	if err != nil {
		return signals.Report{}, err
	}

	signalValues := signals.ExtractSignals(repos)
	scores := signals.ScoreSignals(signalValues)
	if len(repos) < minReposForFullScore {
		scores.Overall = int(math.Round(float64(scores.Overall) * smallSampleOverallMultiplier))
	}

	report := signals.BuildReport(username, scores, repos)
	return report, nil
}
