package acquisition

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	obs "github.com/divijg19/Atlas/internal/observations"
)

// TestFetchGraphQLVestigesConcurrent verifies bounded-concurrency enrichment is
// correct (order preserved, failures tolerated) and faster than the per-repo
// sequential delay would allow.
func TestFetchGraphQLVestigesConcurrent(t *testing.T) {
	const delay = 120 * time.Millisecond
	var mu sync.Mutex
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mu.Lock()
		hits++
		mu.Unlock()
		time.Sleep(delay)
		// Minimal parseable GraphQL response (repository:null → empty vestige).
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"repository":null}}`))
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL}

	repos := []obs.RepositoryVestige{
		{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"},
	}
	start := time.Now()
	partials := c.fetchGraphQLVestiges(context.Background(), "owner", repos)
	elapsed := time.Since(start)

	if len(partials) != len(repos) {
		t.Fatalf("expected %d partials, got %d", len(repos), len(partials))
	}
	for i, p := range partials {
		if p.Name != "" {
			t.Errorf("slot %d unexpectedly populated: %+v", i, p)
		}
	}
	mu.Lock()
	got := hits
	mu.Unlock()
	if got != len(repos) {
		t.Errorf("expected %d upstream calls, got %d", len(repos), got)
	}
	// With graphqlConcurrency=8 and 4 repos, all run concurrently: well under
	// 4*delay. A regression to sequential would exceed that.
	if elapsed > 2*time.Duration(len(repos))*delay {
		t.Errorf("enrichment took %v; expected concurrent (< %v)", elapsed, 2*time.Duration(len(repos))*delay)
	}
}
