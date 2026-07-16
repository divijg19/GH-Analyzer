package acquisition_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/divijg19/Atlas/internal/acquisition"
)

func newClient(baseURL string) *acquisition.Client {
	return acquisition.NewClientAt(baseURL)
}

func TestFetchContributionsSuccess(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.RawQuery == "q=testuser+type:pr&per_page=1" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"total_count": 85,
				"items":       []interface{}{},
			})
		} else if r.URL.RawQuery == "q=testuser+type:issue&per_page=1" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"total_count": 65,
				"items":       []interface{}{},
			})
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "unexpected query"})
		}
	}))
	defer server.Close()

	dto, err := newClient(server.URL).FetchContributions(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto == nil {
		t.Fatal("expected dto, got nil")
	}

	if callCount != 2 {
		t.Fatalf("expected 2 API calls, got %d", callCount)
	}
	if dto.PullRequests != 85 {
		t.Fatalf("PullRequests = %d, want 85", dto.PullRequests)
	}
	if dto.Issues != 65 {
		t.Fatalf("Issues = %d, want 65", dto.Issues)
	}
}

func TestFetchContributionsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count": 0,
			"items":       []interface{}{},
		})
	}))
	defer server.Close()

	dto, err := newClient(server.URL).FetchContributions(context.Background(), "emptyuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto == nil {
		t.Fatal("expected dto, got nil")
	}
	if dto.PullRequests != 0 || dto.Issues != 0 {
		t.Fatalf("expected zero counts, got %+v", dto)
	}
}

func TestFetchContributionsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
	}))
	defer server.Close()

	_, err := newClient(server.URL).FetchContributions(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestFetchContributionsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := newClient(server.URL).FetchContributions(context.Background(), "testuser")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestFetchContributionsBadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	_, err := newClient(server.URL).FetchContributions(context.Background(), "testuser")
	if err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
}

func TestFetchContributionsEmptyUsername(t *testing.T) {
	_, err := newClient("http://example.invalid").FetchContributions(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}
}

// TestFetchContributionsTransientServerErrorRecovers verifies a transient 5xx
// on the first call is retried by the transport and the request recovers.
func TestFetchContributionsTransientServerErrorRecovers(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"total_count": 10,
			"items":       []interface{}{},
		})
	}))
	defer server.Close()

	dto, err := newClient(server.URL).FetchContributions(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("expected transient 5xx to be retried and recovered, got: %v", err)
	}
	if dto == nil {
		t.Fatal("expected dto, got nil")
	}
}
