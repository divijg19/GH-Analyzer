package ghanalyzer

import (
	"fmt"
	"strings"
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

		repos, err := FetchRepos(username)
		if err != nil {
			return Index{}, fmt.Errorf("fetch repos for %q: %w", username, err)
		}

		signals := ExtractSignals(repos)
		scores := ScoreSignals(signals)
		report := BuildReport(username, scores, repos)

		idx.Add(Profile{
			Username: username,
			Signals:  SignalsFromReport(report),
		})
	}

	if len(idx.Profiles) == 0 {
		return Index{}, fmt.Errorf("no usernames provided")
	}

	return idx, nil
}
