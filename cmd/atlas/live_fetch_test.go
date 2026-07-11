package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/divijg19/Atlas/internal/acquisition"
	"github.com/divijg19/Atlas/internal/live"
	"github.com/divijg19/Atlas/internal/signals"
)

// mockAcquisitionClient is a test seam for the live package. It returns canned
// data without performing network calls and satisfies the live.fetcher
// interface.
type mockAcquisitionClient struct {
	repos     map[string][]signals.RepositoryVestige
	usernames []string
}

func (m *mockAcquisitionClient) FetchReposNormalized(_ context.Context, username string) ([]signals.RepositoryVestige, error) {
	r, ok := m.repos[username]
	if !ok {
		return nil, context.Canceled
	}
	return r, nil
}

func (m *mockAcquisitionClient) FetchUser(_ context.Context, _ string) (*acquisition.UserDTO, error) {
	return &acquisition.UserDTO{}, nil
}

func (m *mockAcquisitionClient) FetchContributions(_ context.Context, _ string) (*acquisition.ContributionsDTO, error) {
	return &acquisition.ContributionsDTO{}, nil
}

func (m *mockAcquisitionClient) SearchRepositoryOwners(_ context.Context, _ string) ([]string, error) {
	seen := make(map[string]struct{}, len(m.usernames))
	out := make([]string, 0, len(m.usernames))
	for _, u := range m.usernames {
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out, nil
}

func setupLiveTestGlobals(mock *mockAcquisitionClient) func() {
	original := live.Client
	live.Client = mock
	return func() {
		live.Client = original
	}
}

func TestBuildLiveIndexSuccess(t *testing.T) {
	now := time.Now()
	mock := &mockAcquisitionClient{
		usernames: []string{"alice", "bob", "alice"},
		repos: map[string][]signals.RepositoryVestige{
			"alice": {{Name: "alice-repo", Size: 100, UpdatedAt: now}},
			"bob":   {{Name: "bob-repo", Size: 100, UpdatedAt: now}},
		},
	}
	restore := setupLiveTestGlobals(mock)
	defer restore()

	ctx := context.Background()
	idx, err := buildLiveIndex(ctx, "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profiles := idx.All()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 deduplicated profiles, got %d", len(profiles))
	}
	if profiles[0].Username != "alice" || profiles[1].Username != "bob" {
		t.Fatalf("unexpected profile order: %+v", profiles)
	}
}

func TestFetchLiveUsernamesEmptyResponse(t *testing.T) {
	mock := &mockAcquisitionClient{usernames: []string{}}
	restore := setupLiveTestGlobals(mock)
	defer restore()

	ctx := context.Background()
	usernames, err := fetchLiveUsernames(ctx, "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(usernames) != 0 {
		t.Fatalf("expected no usernames, got %d", len(usernames))
	}

	idx, err := buildLiveIndex(ctx, "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(idx.All()) != 0 {
		t.Fatalf("expected empty index, got %d profiles", len(idx.All()))
	}
}

func TestBuildLiveIndexPartialFailures(t *testing.T) {
	now := time.Now()
	mock := &mockAcquisitionClient{
		usernames: []string{"alice", "bob", "cara"},
		repos: map[string][]signals.RepositoryVestige{
			"alice": {{Name: "alice-repo", Size: 100, UpdatedAt: now}},
			"cara":  {{Name: "cara-repo", Size: 100, UpdatedAt: now}},
		},
	}
	restore := setupLiveTestGlobals(mock)
	defer restore()

	ctx := context.Background()
	idx, err := buildLiveIndex(ctx, "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profiles := idx.All()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles after skipping failed user, got %d", len(profiles))
	}

	if profiles[0].Username != "alice" || profiles[1].Username != "cara" {
		t.Fatalf("unexpected surviving usernames: %+v", profiles)
	}
}

func TestFetchLiveUsernamesLimitEnforced(t *testing.T) {
	items := make([]string, 0, 25)
	for i := 1; i <= 25; i++ {
		items = append(items, fmt.Sprintf(`{"owner":{"login":"user%02d"}}`, i))
	}
	payload := `{"items":[` + strings.Join(items, ",") + `]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()

	restore := setupLiveTestGlobals(&mockAcquisitionClient{})
	defer restore()

	live.Client = acquisition.NewClientAt(server.URL)

	ctx := context.Background()
	usernames, err := fetchLiveUsernames(ctx, "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(usernames) != 20 {
		t.Fatalf("expected 20 usernames, got %d", len(usernames))
	}
	if usernames[0] != "user01" || usernames[len(usernames)-1] != "user20" {
		t.Fatalf("unexpected limited range: first=%q last=%q", usernames[0], usernames[len(usernames)-1])
	}
}

func TestFetchLiveUsernamesAPIFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"rate limit"}`))
	}))
	defer server.Close()

	restore := setupLiveTestGlobals(&mockAcquisitionClient{})
	defer restore()

	live.Client = acquisition.NewClientAt(server.URL)

	ctx := context.Background()
	_, err := fetchLiveUsernames(ctx, "backend")
	if err == nil {
		t.Fatal("expected API failure error")
	}
	if err.Error() != "failed to fetch GitHub data" {
		t.Fatalf("expected canonical API error, got %q", err.Error())
	}
}
