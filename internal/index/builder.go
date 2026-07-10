package index

import (
	"context"
	"fmt"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/acquisition"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

func Build(ctx context.Context, usernames []string) (Index, error) {
	idx := Index{
		Profiles: make([]Profile, 0, len(usernames)),
	}

	client := acquisition.NewClient()

	for _, rawUsername := range usernames {
		username := strings.TrimSpace(rawUsername)
		if username == "" {
			continue
		}

		repoDTOs, err := client.FetchRepos(ctx, username)
		if err != nil {
			return Index{}, fmt.Errorf("fetch repos for %q: %w", username, err)
		}

		repos := acquisition.NormalizeRepos(repoDTOs)
		facts := signals.FromRepos(repos)

		userDTO, err := client.FetchUser(ctx, username)
		if err != nil {
			return Index{}, fmt.Errorf("fetch metadata for %q: %w", username, err)
		}

		meta := acquisition.NormalizeUser(userDTO)

		contribDTO, err := client.FetchContributions(ctx, username)
		if err != nil {
			return Index{}, fmt.Errorf("fetch contributions for %q: %w", username, err)
		}

		contribSummary := acquisition.NormalizeContributions(contribDTO)

		signalValues := signals.ExtractSignalsFromFacts(facts)

		idx.Add(Profile{
			Username:      username,
			Signals:       signals.SignalsToMap(signalValues),
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
