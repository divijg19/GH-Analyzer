package acquisition

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	maxRepoResults = 30
	maxUsers       = 20
)

// RepoSearchDTO mirrors the GitHub repository search API response.
type RepoSearchDTO struct {
	Items []struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"items"`
}

// SearchRepositoryOwners searches GitHub repositories for the given query and
// returns deduplicated owner logins, up to MaxUsers.
func (c *Client) SearchRepositoryOwners(ctx context.Context, query string) ([]string, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf("%s/search/repositories?q=%s&per_page=%d", c.baseURL, url.QueryEscape(trimmed), maxRepoResults)
	resp, err := c.get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}

	var payload RepoSearchDTO
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub data")
	}

	seen := make(map[string]struct{}, len(payload.Items))
	usernames := make([]string, 0, maxUsers)

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
		if len(usernames) >= maxUsers {
			break
		}
	}

	return usernames, nil
}
