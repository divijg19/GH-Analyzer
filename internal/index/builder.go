// Package index owns the canonical candidate aggregate, Profile.
//
// It assembles facts, indicators, metadata, and contributions into a single
// Profile record. The Profile is the single source of truth for what Atlas
// knows about a candidate before evaluation.
//
// Index never acquires observations, derives facts, computes indicators,
// generates evidence, evaluates candidates, or performs presentation.
//
// Consumed by: engine, projection, search, live, storage, cmd/*.
package index

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/divijg19/Atlas/internal/acquisition"
	"github.com/divijg19/Atlas/internal/facts"
	"github.com/divijg19/Atlas/internal/indicators"
	"github.com/divijg19/Atlas/internal/observations"
)

// ProfileFetcher is the acquisition surface required to assemble a Profile.
// It is satisfied by *acquisition.Client and by test fakes.
type ProfileFetcher interface {
	FetchReposNormalized(ctx context.Context, username string) ([]observations.RepositoryVestige, error)
	FetchUser(ctx context.Context, username string) (*acquisition.UserDTO, error)
	FetchContributions(ctx context.Context, username string) (*acquisition.ContributionsDTO, error)
	FetchActivityObservations(ctx context.Context, username string) []observations.ActivityObservation
}

// BuildProfile assembles a single candidate Profile from the acquisition
// layer. It is the single canonical owner of profile assembly, shared by both
// the batch (Build) and live (live.BuildLiveIndex) pipelines.
//
// referenceTime is the explicit "now" used for all time-relative derivations
// (recent windows, age, activity recency). Passing it explicitly keeps the
// assembly deterministic for a given acquisition snapshot.
func BuildProfile(ctx context.Context, fetcher ProfileFetcher, username string, referenceTime time.Time) (Profile, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return Profile{}, fmt.Errorf("username is required")
	}

	repos, err := fetcher.FetchReposNormalized(ctx, username)
	if err != nil {
		return Profile{}, fmt.Errorf("fetch repos for %q: %w", username, err)
	}

	repoFacts := facts.FromRepos(repos, referenceTime)

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

	// Activity observations → ActivityFacts
	activityObs := fetcher.FetchActivityObservations(ctx, username)
	activityFacts := facts.ActivityFactsFromObservations(activityObs, referenceTime)

	signalValues := indicators.ExtractSignalsFromFacts(repoFacts, referenceTime)

	return Profile{
		Username:      username,
		Signals:       indicators.SignalsToMap(signalValues),
		Repositories:  repos,
		Facts:         &repoFacts,
		Metadata:      meta,
		Contributions: contribSummary,
		ActivityFacts: &activityFacts,
	}, nil
}
