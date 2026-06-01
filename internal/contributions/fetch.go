package contributions

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/divijg19/GH-Analyzer/internal/github"
)

var githubSearchIssuesURL = func(query string) string {
	return "https://api.github.com/search/issues?q=" + query + "&per_page=1"
}

type searchResponse struct {
	TotalCount int `json:"total_count"`
}

func FetchContributions(username string) (*Summary, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	prCount, err := searchCount(username + "+type:pr")
	if err != nil {
		return nil, fmt.Errorf("fetch pull requests: %w", err)
	}

	issueCount, err := searchCount(username + "+type:issue")
	if err != nil {
		return nil, fmt.Errorf("fetch issues: %w", err)
	}

	return &Summary{
		TotalContributions: prCount + issueCount,
		TotalPullRequests:  prCount,
		IssuesOpened:       issueCount,
	}, nil
}

func searchCount(query string) (int, error) {
	url := githubSearchIssuesURL(query)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	github.SetHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var ghErr struct {
			Message string `json:"message"`
		}
		message := resp.Status
		if err := json.NewDecoder(resp.Body).Decode(&ghErr); err == nil && ghErr.Message != "" {
			message = ghErr.Message
		}
		return 0, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, message)
	}

	var res searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, fmt.Errorf("decode search response: %w", err)
	}

	return res.TotalCount, nil
}
