package acquisition

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/divijg19/Atlas/internal/github"
	"github.com/divijg19/Atlas/internal/signals"
)

// queryRepo executes a single GraphQL repository query and returns the DTO.
func (c *Client) queryRepo(ctx context.Context, owner, name string) (*graphQLRepo, error) {
	variables := map[string]any{
		"owner": owner,
		"name":  name,
	}

	body, err := json.Marshal(graphQLRequest{
		Query:     repoQuery,
		Variables: variables,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := github.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL API returned status %d", resp.StatusCode)
	}

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode GraphQL response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %s", gqlResp.Errors[0].Message)
	}

	if gqlResp.Data.Repository == nil {
		return nil, fmt.Errorf("repository %s/%s not found", owner, name)
	}

	return gqlResp.Data.Repository, nil
}

// fetchGraphQLVestiges queries the GitHub GraphQL API for each repository and
// returns a slice of partial vestiges aligned with the input repos. Each
// partial vestige contains only GraphQL-authoritative fields (see
// OBSERVATION_SPECIFICATION.md). Repos that fail the query produce an empty
// partial vestige.
//
// The returned value has already crossed the DTO→domain boundary via
// normalizeGraphQLRepo. It is not raw GraphQL data — it is a normalized
// partial vestige ready for merge.
//
// fetchGraphQLVestiges does NOT merge with REST data. It is the GraphQL
// executor only: query → DTO → normalize. Merging is the responsibility of
// mergeVestiges in merge.go.
func (c *Client) fetchGraphQLVestiges(ctx context.Context, owner string, repos []signals.RepositoryVestige) []signals.RepositoryVestige {
	partials := make([]signals.RepositoryVestige, len(repos))

	for i, v := range repos {
		if v.Name == "" {
			continue
		}

		repo, err := c.queryRepo(ctx, owner, v.Name)
		if err != nil {
			continue
		}

		partials[i] = normalizeGraphQLRepo(repo)
	}

	return partials
}

// FetchReposEnriched retrieves repository vestiges using both REST (core
// observations) and GraphQL (enriched observations). The returned vestiges
// contain all observations Atlas currently understands.
//
// Orchestration:
//  1. REST baseline via FetchReposNormalized
//  2. GraphQL enrichment via fetchGraphQLVestiges
//  3. Merged via mergeVestiges
//
// GraphQL enrichment is additive. When GraphQL is unavailable, REST-only
// vestiges are returned with enriched fields at their documented defaults.
func (c *Client) FetchReposEnriched(ctx context.Context, username string) ([]signals.RepositoryVestige, error) {
	base, err := c.FetchReposNormalized(ctx, username)
	if err != nil {
		return nil, err
	}

	enriched := c.fetchGraphQLVestiges(ctx, username, base)

	result := make([]signals.RepositoryVestige, len(base))
	for i := range base {
		result[i] = mergeVestiges(base[i], enriched[i])
	}

	return result, nil
}
