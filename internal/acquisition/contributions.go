package acquisition

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ContributionsDTO holds the raw contribution counts acquired from the GitHub
// search API. It aggregates two separate search calls (pull requests and
// issues); assembling the domain summary is normalization's responsibility.
type ContributionsDTO struct {
	PullRequests int
	Issues       int
}

// searchCountDTO mirrors the GitHub search API response envelope.
type searchCountDTO struct {
	TotalCount int `json:"total_count"`
}

// FetchContributions retrieves a user's pull request and issue counts from the
// GitHub search API.
func (c *Client) FetchContributions(ctx context.Context, username string) (*ContributionsDTO, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	prCount, err := c.searchCount(ctx, username+"+type:pr")
	if err != nil {
		return nil, fmt.Errorf("fetch pull requests: %w", err)
	}

	issueCount, err := c.searchCount(ctx, username+"+type:issue")
	if err != nil {
		return nil, fmt.Errorf("fetch issues: %w", err)
	}

	return &ContributionsDTO{
		PullRequests: prCount,
		Issues:       issueCount,
	}, nil
}

func (c *Client) searchCount(ctx context.Context, query string) (int, error) {
	resp, err := c.get(ctx, c.baseURL+"/search/issues?q="+query+"&per_page=1")
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

	var res searchCountDTO
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, fmt.Errorf("decode search response: %w", err)
	}

	return res.TotalCount, nil
}
