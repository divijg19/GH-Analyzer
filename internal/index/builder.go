package index

import (
	"fmt"
	"strings"

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

		signalValues := signals.ExtractSignals(repos)
		scores := signals.ScoreSignals(signalValues)
		report := signals.BuildReport(username, scores, repos)

		idx.Add(Profile{
			Username: username,
			Signals:  signals.SignalsFromReport(report),
		})
	}

	if len(idx.Profiles) == 0 {
		return Index{}, fmt.Errorf("no usernames provided")
	}

	return idx, nil
}
