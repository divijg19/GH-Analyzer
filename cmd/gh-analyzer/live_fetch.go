package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/divijg19/GH-Analyzer/internal/contributions"
	"github.com/divijg19/GH-Analyzer/internal/github"
	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

const (
	maxLiveRepos = 30
	maxLiveUsers = 20
)

var (
	liveRepoSearchURL        = "https://api.github.com/search/repositories"
	liveHTTPClient           = http.DefaultClient
	fetchReposUser           = signals.FetchRepos
	fetchContributionsUser   = contributions.FetchContributions
)

type repositorySearchResponse struct {
	Items []repositoryItem `json:"items"`
}

type repositoryItem struct {
	Owner repositoryOwner `json:"owner"`
}

type repositoryOwner struct {
	Login string `json:"login"`
}

func buildLiveIndex(query string) (indexpkg.Index, error) {
	usernames, err := fetchLiveUsernames(query)
	if err != nil {
		return indexpkg.Index{}, err
	}

	idx := indexpkg.Index{Profiles: make([]indexpkg.Profile, 0, len(usernames))}
	for _, username := range usernames {
		repos, err := fetchReposUser(username)
		if err != nil {
			continue
		}

		facts := signals.FromRepos(repos)

		contribSummary, err := fetchContributionsUser(username)
		if err != nil {
			continue
		}

		signalValues := signals.ExtractSignalsFromFacts(facts)
		scores := signals.ScoreSignals(signalValues)
		report := signals.BuildReport(username, scores, repos)
		profileSignals := signals.SignalsFromReport(report)

		idx.Add(indexpkg.Profile{
			Username:      username,
			Signals:       profileSignals,
			Facts:         &facts,
			Contributions: contribSummary,
		})
	}

	return idx, nil
}

func fetchLiveUsernames(query string) ([]string, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf("%s?q=%s&per_page=%d", liveRepoSearchURL, url.QueryEscape(trimmedQuery), maxLiveRepos)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}
	github.SetHeaders(req)

	resp, err := liveHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}

	var payload repositorySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}

	seen := make(map[string]struct{}, len(payload.Items))
	usernames := make([]string, 0, maxLiveUsers)

	for _, item := range payload.Items {
		login := strings.TrimSpace(item.Owner.Login)
		if login == "" {
			continue
		}

		if _, ok := seen[login]; ok {
			continue
		}

		seen[login] = struct{}{}
		usernames = append(usernames, login)
		if len(usernames) >= maxLiveUsers {
			break
		}
	}

	return usernames, nil
}
