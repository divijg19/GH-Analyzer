package index

import (
	"fmt"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/contributions"
	"github.com/divijg19/GH-Analyzer/internal/profile"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

func Build(usernames []string) (Index, error) {
	idx := Index{
		Profiles: make([]Profile, 0, len(usernames)),
	}

	for _, rawUsername := range usernames {
		username := strings.TrimSpace(rawUsername)
		if username == "" {
			continue
		}

		repos, err := signals.FetchRepos(username)
		if err != nil {
			return Index{}, fmt.Errorf("fetch repos for %q: %w", username, err)
		}

		facts := signals.FromRepos(repos)

		meta, err := profile.FetchUserMetadata(username)
		if err != nil {
			return Index{}, fmt.Errorf("fetch metadata for %q: %w", username, err)
		}

		contribSummary, err := contributions.FetchContributions(username)
		if err != nil {
			return Index{}, fmt.Errorf("fetch contributions for %q: %w", username, err)
		}

		signalValues := signals.ExtractSignalsFromFacts(facts)
		scores := signals.ScoreSignals(signalValues)
		report := signals.BuildReport(username, scores, repos)

		idx.Add(Profile{
			Username:      username,
			Signals:       signals.SignalsFromReport(report),
			Facts:         &facts,
			Metadata:      meta,
			Contributions: contribSummary,
		})
	}

	if len(idx.Profiles) == 0 {
		return Index{}, fmt.Errorf("no usernames provided")
	}

	return idx, nil
}
