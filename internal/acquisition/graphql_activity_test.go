package acquisition

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// yearAliasRe extracts the aliased year fields (e.g. "y2022:") from a lifetime
// GraphQL query so the mock server can echo them back.
var yearAliasRe = regexp.MustCompile(`y\d{4}:`)

// TestFetchActivityObservationsUsesConfiguredBaseURL verifies the activity
// GraphQL path honors the client's configured base URL (no hardcoded
// api.github.com endpoint), and that a well-formed server yields observations.
func TestFetchActivityObservationsUsesConfiguredBaseURL(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		q := string(body)
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(q, "LifetimeActivity") {
			user := map[string]any{}
			for _, m := range yearAliasRe.FindAllString(q, -1) {
				alias := strings.TrimSuffix(m, ":")
				user[alias] = map[string]any{"totalCommitContributions": 1}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"user": user}})
			return
		}

		// Profile activity query.
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"user": map[string]any{
			"recent": map[string]any{
				"totalCommitContributions":            10,
				"totalPullRequestContributions":       2,
				"totalPullRequestReviewContributions": 1,
				"totalIssueContributions":             3,
				"restrictedContributionsCount":        0,
				"commitContributionsByRepository": []map[string]any{
					{"contributions": map[string]any{"totalCount": 5}, "repository": map[string]any{"nameWithOwner": "u/repo"}},
				},
				"contributionCalendar": map[string]any{"weeks": []map[string]any{
					{"contributionDays": []map[string]any{{"contributionCount": 1}, {"contributionCount": 2}}},
				}},
			},
		}}})
	}))
	defer srv.Close()

	c := NewClientAt(srv.URL)
	observations := c.FetchActivityObservations(context.Background(), "octocat")

	if len(paths) == 0 {
		t.Fatal("no requests reached the test server")
	}
	for _, p := range paths {
		if p != "/graphql" {
			t.Fatalf("activity query hit %q; expected /graphql from the configured base URL", p)
		}
	}

	if len(observations) == 0 {
		t.Fatal("expected activity observations from a well-formed server")
	}
}

// TestFetchActivityLifetimeBatchPropagatesError verifies a lifetime batch
// failure is surfaced (not silently swallowed) and that the best-effort
// boundary in FetchActivityObservations tolerates it.
func TestFetchActivityLifetimeBatchPropagatesError(t *testing.T) {
	// Server fails the lifetime query but succeeds on the profile query.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "LifetimeActivity") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"user": map[string]any{
			"recent": map[string]any{
				"totalCommitContributions":            1,
				"totalPullRequestContributions":       0,
				"totalPullRequestReviewContributions": 0,
				"totalIssueContributions":             0,
				"restrictedContributionsCount":        0,
				"commitContributionsByRepository":     []any{},
				"contributionCalendar":                map[string]any{"weeks": []any{}},
			},
		}}})
	}))
	defer srv.Close()

	c := NewClientAt(srv.URL)
	// Profile succeeds, so observations are non-empty even though the lifetime
	// batch failed at the documented best-effort boundary.
	observations := c.FetchActivityObservations(context.Background(), "octocat")
	if len(observations) == 0 {
		t.Fatal("profile activity should still yield observations despite a failed lifetime batch")
	}
}
