package live

import (
	"context"
	"fmt"

	"github.com/divijg19/GH-Analyzer/internal/acquisition"
	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

// fetcher is the acquisition surface the live pipeline depends on. It is
// satisfied by *acquisition.Client and overridable in tests.
type fetcher interface {
	FetchReposNormalized(ctx context.Context, username string) ([]signals.Repo, error)
	FetchUser(ctx context.Context, username string) (*acquisition.UserDTO, error)
	FetchContributions(ctx context.Context, username string) (*acquisition.ContributionsDTO, error)
	SearchRepositoryOwners(ctx context.Context, query string) ([]string, error)
}

// Client is the acquisition client used for live fetching. Tests may override
// it with a fetcher implementation; production code uses the default
// GitHub-backed client.
var Client fetcher = acquisition.NewClient()

// BuildLiveIndex fetches live data from GitHub for a search query and
// returns a complete index with profiles containing signals, facts,
// metadata, and contributions.
func BuildLiveIndex(ctx context.Context, query string) (indexpkg.Index, error) {
	usernames, err := FetchLiveUsernames(ctx, query)
	if err != nil {
		return indexpkg.Index{}, err
	}

	idx := indexpkg.Index{Profiles: make([]indexpkg.Profile, 0, len(usernames))}
	for _, username := range usernames {
		profile, err := buildLiveProfile(ctx, username)
		if err != nil {
			continue
		}
		idx.Add(profile)
	}

	return idx, nil
}

func buildLiveProfile(ctx context.Context, username string) (indexpkg.Profile, error) {
	repos, err := Client.FetchReposNormalized(ctx, username)
	if err != nil {
		return indexpkg.Profile{}, fmt.Errorf("fetch repos for %q: %w", username, err)
	}

	facts := signals.FromRepos(repos)

	userDTO, err := Client.FetchUser(ctx, username)
	if err != nil {
		return indexpkg.Profile{}, fmt.Errorf("fetch metadata for %q: %w", username, err)
	}

	meta := acquisition.NormalizeUser(userDTO)

	contribDTO, err := Client.FetchContributions(ctx, username)
	if err != nil {
		return indexpkg.Profile{}, fmt.Errorf("fetch contributions for %q: %w", username, err)
	}

	contribSummary := acquisition.NormalizeContributions(contribDTO)

	signalValues := signals.ExtractSignalsFromFacts(facts)

	return indexpkg.Profile{
		Username:      username,
		Signals:       signals.SignalsToMap(signalValues),
		Facts:         &facts,
		Metadata:      meta,
		Contributions: contribSummary,
	}, nil
}
