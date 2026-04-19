package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	indexpkg "github.com/divijg19/GH-Analyzer/internal/index"
	"github.com/divijg19/GH-Analyzer/internal/signals"
)

const (
	maxLiveRepoResults = 30
	maxLiveUsers       = 20
)

var (
	serverLiveRepoSearchURL = "https://api.github.com/search/repositories"
	serverHTTPClient        = http.DefaultClient
	serverFetchUserRepos    = signals.FetchRepos
)

type serverRepoSearchPayload struct {
	Items []serverRepoItem `json:"items"`
}

type serverRepoItem struct {
	Owner serverRepoOwner `json:"owner"`
}

type serverRepoOwner struct {
	Login string `json:"login"`
}

func buildLiveIndexForServer(query string) (indexpkg.Index, error) {
	usernames, err := fetchLiveUsernamesForServer(query)
	if err != nil {
		return indexpkg.Index{}, err
	}

	idx := indexpkg.Index{Profiles: make([]indexpkg.Profile, 0, len(usernames))}
	for _, username := range usernames {
		repos, err := serverFetchUserRepos(username)
		if err != nil {
			continue
		}

		signalValues := signals.ExtractSignals(repos)
		scores := signals.ScoreSignals(signalValues)
		report := signals.BuildReport(username, scores, repos)
		idx.Add(indexpkg.Profile{
			Username: username,
			Signals:  signals.SignalsFromReport(report),
		})
	}

	return idx, nil
}

func fetchLiveUsernamesForServer(query string) ([]string, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf("%s?q=%s&per_page=%d", serverLiveRepoSearchURL, url.QueryEscape(trimmed), maxLiveRepoResults)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}
	req.Header.Set("User-Agent", "gh-analyzer")

	resp, err := serverHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}

	var payload serverRepoSearchPayload
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
