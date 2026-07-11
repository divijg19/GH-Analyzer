// Package index owns the Profile layer of the Atlas Intelligence Ontology
// (see docs/INTELLIGENCE.md). It assembles and stores the canonical candidate
// aggregate, combining facts, signals, metadata, and contributions.
//
// The Profile is the single source of truth for what Atlas knows about a
// candidate; it stores observations, not evaluations.
package index

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/divijg19/Atlas/internal/acquisition"
	"github.com/divijg19/Atlas/internal/signals"
)

// ProfileFetcher is the acquisition surface required to assemble a Profile.
// It is satisfied by *acquisition.Client and by test fakes.
type ProfileFetcher interface {
	FetchReposNormalized(ctx context.Context, username string) ([]signals.RepositoryVestige, error)
	FetchUser(ctx context.Context, username string) (*acquisition.UserDTO, error)
	FetchContributions(ctx context.Context, username string) (*acquisition.ContributionsDTO, error)
}

// BuildProfile assembles a single candidate Profile from the acquisition
// layer. It is the single canonical owner of profile assembly, shared by both
// the batch (Build) and live (live.BuildLiveIndex) pipelines.
func BuildProfile(ctx context.Context, fetcher ProfileFetcher, username string) (Profile, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return Profile{}, fmt.Errorf("username is required")
	}

	repos, err := fetcher.FetchReposNormalized(ctx, username)
	if err != nil {
		return Profile{}, fmt.Errorf("fetch repos for %q: %w", username, err)
	}

	facts := signals.FromRepos(repos, time.Now())

	userDTO, err := fetcher.FetchUser(ctx, username)
	if err != nil {
		return Profile{}, fmt.Errorf("fetch metadata for %q: %w", username, err)
	}

	meta := acquisition.NormalizeUser(userDTO)

	contribDTO, err := fetcher.FetchContributions(ctx, username)
	if err != nil {
		return Profile{}, fmt.Errorf("fetch contributions for %q: %w", username, err)
	}

	contribSummary := acquisition.NormalizeContributions(contribDTO)

	signalValues := signals.ExtractSignalsFromFacts(facts, time.Now())

	return Profile{
		Username:      username,
		Signals:       signals.SignalsToMap(signalValues),
		Facts:         &facts,
		Metadata:      meta,
		Contributions: contribSummary,
	}, nil
}

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

		profile, err := BuildProfile(ctx, client, username)
		if err != nil {
			return Index{}, err
		}

		idx.Add(profile)
	}

	if len(idx.Profiles) == 0 {
		return Index{}, fmt.Errorf("no usernames provided")
	}

	return idx, nil
}
