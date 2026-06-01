package contributions

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchContributionsSuccess(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.RawQuery == "q=testuser+type:pr&per_page=1" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total_count": 85,
				"items":       []interface{}{},
			})
		} else if r.URL.RawQuery == "q=testuser+type:issue&per_page=1" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total_count": 65,
				"items":       []interface{}{},
			})
		} else {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "unexpected query"})
		}
	}))
	defer server.Close()

	originalURL := githubSearchIssuesURL
	githubSearchIssuesURL = func(query string) string {
		return server.URL + "?" + "q=" + query + "&per_page=1"
	}
	defer func() { githubSearchIssuesURL = originalURL }()

	summary, err := FetchContributions("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected summary, got nil")
	}

	if callCount != 2 {
		t.Fatalf("expected 2 API calls, got %d", callCount)
	}
	if summary.TotalContributions != 150 {
		t.Fatalf("TotalContributions = %d, want 150", summary.TotalContributions)
	}
	if summary.TotalPullRequests != 85 {
		t.Fatalf("TotalPullRequests = %d, want 85", summary.TotalPullRequests)
	}
	if summary.IssuesOpened != 65 {
		t.Fatalf("IssuesOpened = %d, want 65", summary.IssuesOpened)
	}
}

func TestFetchContributionsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count": 0,
			"items":       []interface{}{},
		})
	}))
	defer server.Close()

	originalURL := githubSearchIssuesURL
	githubSearchIssuesURL = func(query string) string {
		return server.URL + "?" + "q=" + query + "&per_page=1"
	}
	defer func() { githubSearchIssuesURL = originalURL }()

	summary, err := FetchContributions("emptyuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Fatal("expected summary, got nil")
	}
	if summary.TotalContributions != 0 {
		t.Fatalf("TotalContributions = %d, want 0", summary.TotalContributions)
	}
	if summary.TotalPullRequests != 0 {
		t.Fatalf("TotalPullRequests = %d, want 0", summary.TotalPullRequests)
	}
	if summary.IssuesOpened != 0 {
		t.Fatalf("IssuesOpened = %d, want 0", summary.IssuesOpened)
	}
}

func TestFetchContributionsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
	}))
	defer server.Close()

	originalURL := githubSearchIssuesURL
	githubSearchIssuesURL = func(query string) string {
		return server.URL + "?" + "q=" + query + "&per_page=1"
	}
	defer func() { githubSearchIssuesURL = originalURL }()

	_, err := FetchContributions("nonexistent")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestFetchContributionsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalURL := githubSearchIssuesURL
	githubSearchIssuesURL = func(query string) string {
		return server.URL + "?" + "q=" + query + "&per_page=1"
	}
	defer func() { githubSearchIssuesURL = originalURL }()

	_, err := FetchContributions("testuser")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestFetchContributionsBadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	originalURL := githubSearchIssuesURL
	githubSearchIssuesURL = func(query string) string {
		return server.URL + "?" + "q=" + query + "&per_page=1"
	}
	defer func() { githubSearchIssuesURL = originalURL }()

	_, err := FetchContributions("testuser")
	if err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
}

func TestFetchContributionsEmptyUsername(t *testing.T) {
	_, err := FetchContributions("")
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}
}

func TestFetchContributionsFirstCallFailsSecondSucceeds(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count": 10,
			"items":       []interface{}{},
		})
	}))
	defer server.Close()

	originalURL := githubSearchIssuesURL
	githubSearchIssuesURL = func(query string) string {
		return server.URL + "?" + "q=" + query + "&per_page=1"
	}
	defer func() { githubSearchIssuesURL = originalURL }()

	_, err := FetchContributions("testuser")
	if err == nil {
		t.Fatal("expected error when first call fails")
	}
}
