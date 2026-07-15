package acquisition

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/divijg19/Atlas/internal/observations"
)

const repoPageSize = 100

// validateUsername rejects input that cannot be a GitHub login before it reaches
// the network. GitHub logins are non-empty, have no whitespace, and are not
// written with an @ prefix; rejecting early is deterministic and cheap.
func validateUsername(username string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username is required")
	}
	if strings.ContainsAny(username, " \t@/") {
		return fmt.Errorf("invalid username %q", username)
	}
	return nil
}

// parseNextPage extracts the URL of the next page from a GitHub REST Link header,
// if present (rel="next"). Returns "" when there is no next page.
func parseNextPage(link string) string {
	for _, part := range strings.Split(link, ",") {
		segments := strings.Split(part, ";")
		if len(segments) < 2 {
			continue
		}
		urlPart := strings.TrimSpace(segments[0])
		urlPart = strings.Trim(urlPart, "<>")
		rel := strings.TrimSpace(segments[1])
		if strings.Contains(rel, `rel="next"`) {
			return urlPart
		}
	}
	return ""
}

// licenseRef is the minimal license reference returned by the GitHub REST API.
type licenseRef struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// RepoDTO mirrors a repository object from the GitHub REST API. Field names and
// types follow GitHub's schema; timestamps remain strings and are parsed during
// normalization.
type RepoDTO struct {
	Name          string      `json:"name"`
	Fork          bool        `json:"fork"`
	Size          int         `json:"size"`
	UpdatedAt     string      `json:"updated_at"`
	Visibility    string      `json:"visibility"`
	Archived      bool        `json:"archived"`
	IsTemplate    bool        `json:"is_template"`
	License       *licenseRef `json:"license"`
	Topics        []string    `json:"topics"`
	Stars         int         `json:"stargazers_count"`
	Forks         int         `json:"forks_count"`
	Watchers      int         `json:"watchers_count"`
	OpenIssues    int         `json:"open_issues_count"`
	CreatedAt     string      `json:"created_at"`
	PushedAt      string      `json:"pushed_at"`
	DefaultBranch string      `json:"default_branch"`
}

// FetchRepos retrieves all of a user's public repositories from the GitHub REST
// API, paginating until exhaustion. It never silently truncates: every page is
// fetched (per_page=100) until the response carries no rel="next" Link header.
func (c *Client) FetchRepos(ctx context.Context, username string) ([]RepoDTO, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	var repos []RepoDTO
	pageURL := fmt.Sprintf("%s/users/%s/repos?per_page=%d", c.baseURL, url.PathEscape(username), repoPageSize)

	for pageURL != "" {
		resp, err := c.get(ctx, pageURL)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			message := resp.Status
			var githubError struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(body, &githubError); err == nil && strings.TrimSpace(githubError.Message) != "" {
				message = strings.TrimSpace(githubError.Message)
			} else if isRateLimited(resp.StatusCode, string(body)) {
				return nil, fmt.Errorf("%w: %w", ErrRateLimited, APIError{StatusCode: resp.StatusCode, Message: message})
			}
			return nil, APIError{
				StatusCode: resp.StatusCode,
				Message:    message,
			}
		}

		var page []RepoDTO
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		repos = append(repos, page...)

		// Stop when the page was not full: there cannot be a following page.
		if len(page) < repoPageSize {
			break
		}
		pageURL = parseNextPage(resp.Header.Get("Link"))
	}

	return repos, nil
}

// FetchReposNormalized retrieves a user's public repositories and returns them
// as domain models. It is equivalent to FetchRepos followed by NormalizeRepos.
func (c *Client) FetchReposNormalized(ctx context.Context, username string) ([]observations.RepositoryVestige, error) {
	dtos, err := c.FetchRepos(ctx, username)
	if err != nil {
		return nil, err
	}

	return NormalizeRepos(dtos), nil
}
