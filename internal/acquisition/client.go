// Package acquisition owns the Acquisition and Normalization layers of the
// Atlas Intelligence Ontology (see docs/INTELLIGENCE.md).
//
// Acquisition fetches external data and produces DTOs that mirror the API
// schema. Normalization maps those DTOs to domain vestiges (RepositoryVestige,
// UserMetadata, Contributions.Summary). This package never interprets,
// evaluates, or derives intelligence beyond the vestige boundary.
//
// Architectural rules:
//   - Acquisition owns GitHub's schema, not the domain's. DTOs mirror GitHub
//     responses exactly and never leak upward.
//   - Acquisition performs no signal computation, no fact derivation, and no
//     presentation. Mapping GitHub DTOs to domain models is normalization
//     (see normalize.go).
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
