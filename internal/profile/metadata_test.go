package profile

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchUserMetadataSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("User-Agent") != "gh-analyzer" {
			t.Fatalf("expected User-Agent gh-analyzer, got %s", r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":       "Monalisa Octocat",
			"bio":        "GitHub mascot",
			"location":   "San Francisco, CA",
			"company":    "@github",
			"followers":  1500,
			"following":  100,
			"created_at": "2020-03-15T12:00:00Z",
		})
	}))
	defer server.Close()

	originalURL := githubUserURL
	githubUserURL = func(username string) string {
		return server.URL
	}
	defer func() { githubUserURL = originalURL }()

	meta, err := FetchUserMetadata("octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta == nil {
		t.Fatal("expected metadata, got nil")
	}

	if meta.Name != "Monalisa Octocat" {
		t.Fatalf("Name = %q, want %q", meta.Name, "Monalisa Octocat")
	}
	if meta.Bio != "GitHub mascot" {
		t.Fatalf("Bio = %q, want %q", meta.Bio, "GitHub mascot")
	}
	if meta.Location != "San Francisco, CA" {
		t.Fatalf("Location = %q, want %q", meta.Location, "San Francisco, CA")
	}
	if meta.Company != "@github" {
		t.Fatalf("Company = %q, want %q", meta.Company, "@github")
	}
	if meta.Followers != 1500 {
		t.Fatalf("Followers = %d, want %d", meta.Followers, 1500)
	}
	if meta.Following != 100 {
		t.Fatalf("Following = %d, want %d", meta.Following, 100)
	}

	expectedCreated := time.Date(2020, 3, 15, 12, 0, 0, 0, time.UTC)
	if !meta.CreatedAt.Equal(expectedCreated) {
		t.Fatalf("CreatedAt = %v, want %v", meta.CreatedAt, expectedCreated)
	}
}

func TestFetchUserMetadataEmptyFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":       "",
			"bio":        "",
			"location":   "",
			"company":    "",
			"followers":  0,
			"following":  0,
			"created_at": "2020-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	originalURL := githubUserURL
	githubUserURL = func(username string) string {
		return server.URL
	}
	defer func() { githubUserURL = originalURL }()

	meta, err := FetchUserMetadata("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta == nil {
		t.Fatal("expected metadata, got nil")
	}

	if meta.Name != "" {
		t.Fatalf("expected empty Name, got %q", meta.Name)
	}
	if meta.Bio != "" {
		t.Fatalf("expected empty Bio, got %q", meta.Bio)
	}
	if meta.Followers != 0 {
		t.Fatalf("expected 0 Followers, got %d", meta.Followers)
	}
	if meta.Following != 0 {
		t.Fatalf("expected 0 Following, got %d", meta.Following)
	}
}

func TestFetchUserMetadataBadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	originalURL := githubUserURL
	githubUserURL = func(username string) string {
		return server.URL
	}
	defer func() { githubUserURL = originalURL }()

	_, err := FetchUserMetadata("testuser")
	if err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
}

func TestFetchUserMetadataNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Not Found",
		})
	}))
	defer server.Close()

	originalURL := githubUserURL
	githubUserURL = func(username string) string {
		return server.URL
	}
	defer func() { githubUserURL = originalURL }()

	_, err := FetchUserMetadata("nonexistent")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestFetchUserMetadataEmptyUsername(t *testing.T) {
	_, err := FetchUserMetadata("")
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}
}

func TestFetchUserMetadataServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalURL := githubUserURL
	githubUserURL = func(username string) string {
		return server.URL
	}
	defer func() { githubUserURL = originalURL }()

	_, err := FetchUserMetadata("testuser")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestFetchUserMetadataBadCreatedAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":       "Test User",
			"bio":        "",
			"location":   "",
			"company":    "",
			"followers":  0,
			"following":  0,
			"created_at": "not-a-valid-date",
		})
	}))
	defer server.Close()

	originalURL := githubUserURL
	githubUserURL = func(username string) string {
		return server.URL
	}
	defer func() { githubUserURL = originalURL }()

	meta, err := FetchUserMetadata("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !meta.CreatedAt.IsZero() {
		t.Fatalf("expected zero CreatedAt for bad date, got %v", meta.CreatedAt)
	}
}
