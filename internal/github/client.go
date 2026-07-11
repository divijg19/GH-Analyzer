// Package github implements the Transport layer of the Atlas Intelligence
// Ontology (see docs/INTELLIGENCE.md). It owns HTTP communication with the
// GitHub REST API. It never parses domain types or interprets responses.
package github

import (
	"context"
	"net/http"
	"os"
	"time"
)

const (
	defaultUserAgent = "atlas"
	defaultTimeout   = 30 * time.Second
)

var sharedClient = &http.Client{Timeout: defaultTimeout}

func client() *http.Client {
	return sharedClient
}

func authHeader() (string, string) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", ""
	}
	return "Authorization", "Bearer " + token
}

func setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", defaultUserAgent)
	if key, val := authHeader(); key != "" {
		req.Header.Set(key, val)
	}
}

// Do executes an authenticated HTTP request through the shared transport,
// attaching the Atlas User-Agent and GitHub auth header.
func Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	setHeaders(req)
	return client().Do(req)
}
