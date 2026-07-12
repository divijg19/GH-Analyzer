package acquisition

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeGraphQLRepoFull(t *testing.T) {
	repo := &graphQLRepo{
		Languages: &graphQLLanguages{
			Edges: []graphQLLanguageEdge{
				{Size: 20000, Node: graphQLLanguage{Name: "Go"}},
				{Size: 5000, Node: graphQLLanguage{Name: "TypeScript"}},
			},
		},
		Releases: &graphQLCountWithNodes{
			TotalCount: 12,
			Nodes: []graphQLNodeTime{
				{CreatedAt: "2024-06-15T10:00:00Z"},
			},
		},
		PullRequests:          &graphQLCount{TotalCount: 45},
		HasDiscussionsEnabled: boolPtr(true),
		Parent: &graphQLParent{
			Name:  "upstream-repo",
			Owner: graphQLOwner{Login: "upstream-owner"},
		},
		Collaborators:         &graphQLCount{TotalCount: 8},
		BranchProtectionRules: &graphQLCount{TotalCount: 3},
	}

	result := normalizeGraphQLRepo(repo)

	if result.LanguageDistribution == nil {
		t.Fatal("expected language distribution")
	}
	if result.LanguageDistribution["Go"] != 20000 {
		t.Errorf("expected Go=20000, got %d", result.LanguageDistribution["Go"])
	}
	if result.LanguageDistribution["TypeScript"] != 5000 {
		t.Errorf("expected TypeScript=5000, got %d", result.LanguageDistribution["TypeScript"])
	}
	if result.ReleaseCount != 12 {
		t.Errorf("expected 12 releases, got %d", result.ReleaseCount)
	}
	if result.LatestReleaseAt.IsZero() {
		t.Errorf("expected non-zero release time")
	}
	if result.PullRequestCount != 45 {
		t.Errorf("expected 45 PRs, got %d", result.PullRequestCount)
	}
	if !result.DiscussionEnabled {
		t.Errorf("expected discussions enabled")
	}
	if result.ParentRepository != "upstream-owner/upstream-repo" {
		t.Errorf("expected upstream-owner/upstream-repo, got %q", result.ParentRepository)
	}
	if result.CollaboratorCount != 8 {
		t.Errorf("expected 8 collaborators, got %d", result.CollaboratorCount)
	}
	if !result.DefaultBranchProtected {
		t.Errorf("expected branch protection enabled")
	}
}

func TestNormalizeGraphQLRepoNilLanguages(t *testing.T) {
	repo := &graphQLRepo{
		Languages:             nil,
		Releases:              &graphQLCountWithNodes{TotalCount: 3},
		PullRequests:          &graphQLCount{TotalCount: 10},
		HasDiscussionsEnabled: boolPtr(false),
		Parent:                nil,
		Collaborators:         &graphQLCount{TotalCount: 2},
		BranchProtectionRules: &graphQLCount{TotalCount: 0},
	}

	result := normalizeGraphQLRepo(repo)

	if result.LanguageDistribution != nil {
		t.Errorf("expected nil language distribution for nil languages")
	}
	if result.ReleaseCount != 3 {
		t.Errorf("expected 3 releases, got %d", result.ReleaseCount)
	}
	if result.PullRequestCount != 10 {
		t.Errorf("expected 10 PRs, got %d", result.PullRequestCount)
	}
	if result.DiscussionEnabled {
		t.Errorf("expected discussions disabled")
	}
	if result.ParentRepository != "" {
		t.Errorf("expected empty parent, got %q", result.ParentRepository)
	}
	if result.CollaboratorCount != 2 {
		t.Errorf("expected 2 collaborators, got %d", result.CollaboratorCount)
	}
	if result.DefaultBranchProtected {
		t.Errorf("expected no branch protection")
	}
}

func TestNormalizeGraphQLRepoAllNil(t *testing.T) {
	result := normalizeGraphQLRepo(nil)

	if result.LanguageDistribution != nil {
		t.Error("expected nil language distribution for nil repo")
	}
	if result.ReleaseCount != 0 {
		t.Errorf("expected 0 releases, got %d", result.ReleaseCount)
	}
	if !result.LatestReleaseAt.IsZero() {
		t.Error("expected zero release time")
	}
	if result.PullRequestCount != 0 {
		t.Errorf("expected 0 PRs, got %d", result.PullRequestCount)
	}
	if result.DiscussionEnabled {
		t.Error("expected discussions disabled")
	}
	if result.ParentRepository != "" {
		t.Errorf("expected empty parent, got %q", result.ParentRepository)
	}
	if result.CollaboratorCount != 0 {
		t.Errorf("expected 0 collaborators, got %d", result.CollaboratorCount)
	}
	if result.DefaultBranchProtected {
		t.Error("expected no branch protection")
	}
}

func TestNormalizeGraphQLRepoEmptyLanguages(t *testing.T) {
	repo := &graphQLRepo{
		Languages: &graphQLLanguages{
			Edges: []graphQLLanguageEdge{},
		},
	}

	result := normalizeGraphQLRepo(repo)

	if result.LanguageDistribution == nil {
		t.Fatal("expected non-nil language distribution")
	}
	if len(result.LanguageDistribution) != 0 {
		t.Errorf("expected empty distribution, got %d entries", len(result.LanguageDistribution))
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestFetchReposEnrichedDegradation(t *testing.T) {
	// When GraphQL is unavailable (no auth or test server), FetchReposEnriched
	// should degrade gracefully to REST-only vestiges.
	//
	// We test this by using a REST mock that returns data and a GraphQL endpoint
	// that will fail (unreachable default endpoint). The REST vestiges should
	// still be returned.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/testuser/repos" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"name":"repo1","size":100,"updated_at":"2024-01-01T00:00:00Z"}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := testClient(srv.URL)
	vestiges, err := client.FetchReposEnriched(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vestiges) != 1 {
		t.Fatalf("expected 1 vestige, got %d", len(vestiges))
	}
	if vestiges[0].Name != "repo1" || vestiges[0].Size != 100 {
		t.Errorf("unexpected vestige: %+v", vestiges[0])
	}
}
