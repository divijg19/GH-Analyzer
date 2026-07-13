// Package acquisition owns data fetching and normalization.
//
// It fetches external data via REST and GraphQL, produces DTOs that mirror the
// API schema, and normalizes those DTOs into domain observations
// (observations.RepositoryVestige, observations.ActivityObservation,
// profile.UserMetadata, contributions.Summary).
//
// Acquisition never derives facts, computes indicators, generates evidence,
// evaluates candidates, or performs presentation. It owns GitHub's schema,
// not the domain's.
//
// Consumed by: index, live, cmd/*.
package acquisition

import (
	"context"
	"net/http"

	"github.com/divijg19/Atlas/internal/github"
)

const defaultBaseURL = "https://api.github.com"

// Client is the GitHub REST acquisition client. It builds requests against a
// configurable base URL and executes them through the shared transport.
type Client struct {
	baseURL string
}

// NewClient returns an acquisition Client targeting the GitHub REST API.
func NewClient() *Client {
	return &Client{baseURL: defaultBaseURL}
}

// NewClientAt returns an acquisition Client targeting a custom base URL. It is
// intended for tests that serve the GitHub API from an httptest server.
func NewClientAt(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

// get executes an authenticated GET request through the transport layer.
func (c *Client) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return github.Do(ctx, req)
}
