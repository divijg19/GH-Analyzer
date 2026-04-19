package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/divijg19/GH-Analyzer/internal/signals"
)

func TestBuildLiveIndexSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/repositories" {
			http.NotFound(w, r)
			return
		}

		if r.Header.Get("User-Agent") != "gh-analyzer" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Fprint(w, `{"items":[{"owner":{"login":"alice"}},{"owner":{"login":"bob"}},{"owner":{"login":"alice"}}]}`)
	}))
	defer server.Close()

	restore := setupLiveTestGlobals(server.URL + "/search/repositories")
	defer restore()

	fetchReposUser = func(username string) ([]signals.Repo, error) {
		return []signals.Repo{{
			Name:      username + "-repo",
			Fork:      false,
			Size:      100,
			UpdatedAt: time.Now(),
		}}, nil
	}

	idx, err := buildLiveIndex("backend")
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"items":[]}`)
	}))
	defer server.Close()

	restore := setupLiveTestGlobals(server.URL + "/search/repositories")
	defer restore()

	usernames, err := fetchLiveUsernames("backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(usernames) != 0 {
		t.Fatalf("expected no usernames, got %d", len(usernames))
	}

	idx, err := buildLiveIndex("backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(idx.All()) != 0 {
		t.Fatalf("expected empty index, got %d profiles", len(idx.All()))
	}
}

func TestBuildLiveIndexPartialFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"items":[{"owner":{"login":"alice"}},{"owner":{"login":"bob"}},{"owner":{"login":"cara"}}]}`)
	}))
	defer server.Close()

	restore := setupLiveTestGlobals(server.URL + "/search/repositories")
	defer restore()

	fetchReposUser = func(username string) ([]signals.Repo, error) {
		if username == "bob" {
			return nil, fmt.Errorf("fetch failed")
		}

		return []signals.Repo{{
			Name:      username + "-repo",
			Fork:      false,
			Size:      100,
			UpdatedAt: time.Now(),
		}}, nil
	}

	idx, err := buildLiveIndex("backend")
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
	payload := fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, payload)
	}))
	defer server.Close()

	restore := setupLiveTestGlobals(server.URL + "/search/repositories")
	defer restore()

	usernames, err := fetchLiveUsernames("backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(usernames) != maxLiveUsers {
		t.Fatalf("expected %d usernames, got %d", maxLiveUsers, len(usernames))
	}
	if usernames[0] != "user01" || usernames[len(usernames)-1] != "user20" {
		t.Fatalf("unexpected limited range: first=%q last=%q", usernames[0], usernames[len(usernames)-1])
	}
}

func TestFetchLiveUsernamesAPIFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"rate limit"}`)
	}))
	defer server.Close()

	restore := setupLiveTestGlobals(server.URL + "/search/repositories")
	defer restore()

	_, err := fetchLiveUsernames("backend")
	if err == nil {
		t.Fatal("expected API failure error")
	}
	if err.Error() != "failed to fetch GitHub data" {
		t.Fatalf("expected canonical API error, got %q", err.Error())
	}
}

func setupLiveTestGlobals(searchURL string) func() {
	originalURL := liveRepoSearchURL
	originalClient := liveHTTPClient
	originalFetchRepos := fetchReposUser

	liveRepoSearchURL = searchURL
	liveHTTPClient = http.DefaultClient
	fetchReposUser = func(username string) ([]signals.Repo, error) {
		return nil, nil
	}

	return func() {
		liveRepoSearchURL = originalURL
		liveHTTPClient = originalClient
		fetchReposUser = originalFetchRepos
	}
}
