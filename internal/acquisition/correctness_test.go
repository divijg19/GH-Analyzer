package acquisition

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestFetchReposPagination verifies FetchRepos never silently truncates: every
// page is fetched until exhaustion (per_page=100, stop on a short final page or
// on absence of a rel="next" Link header).
func TestFetchReposPagination(t *testing.T) {
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		wantCount int
	}{
		{
			name: "single short page",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("per_page") != "100" {
					t.Errorf("expected per_page=100, got %q", r.URL.Query().Get("per_page"))
				}
				_, _ = w.Write([]byte(`[{"name":"r1"},{"name":"r2"}]`))
			},
			wantCount: 2,
		},
		{
			name: "two full pages via Link header",
			handler: func(w http.ResponseWriter, r *http.Request) {
				page := r.URL.Query().Get("page")
				switch page {
				case "", "1":
					w.Header().Set("Link", `<http://`+r.Host+`/users/u/repos?page=2>; rel="next"`)
					writeRepos(w, 100)
				case "2":
					writeRepos(w, 3)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			},
			wantCount: 103,
		},
		{
			name: "short final page terminates without Link",
			handler: func(w http.ResponseWriter, r *http.Request) {
				writeRepos(w, 7)
			},
			wantCount: 7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			repos, err := testClient(srv.URL).FetchRepos(context.Background(), "u")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(repos) != tc.wantCount {
				t.Fatalf("expected %d repos, got %d", tc.wantCount, len(repos))
			}
		})
	}
}

func writeRepos(w http.ResponseWriter, n int) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("["))
	for i := 0; i < n; i++ {
		if i > 0 {
			_, _ = w.Write([]byte(","))
		}
		_, _ = fmt.Fprintf(w, `{"name":"r%d"}`, i)
	}
	_, _ = w.Write([]byte("]"))
}

// TestFetchReposUsernameValidation verifies malformed usernames are rejected
// before any network call.
func TestFetchReposUsernameValidation(t *testing.T) {
	client := testClient("http://example.invalid")
	for _, bad := range []string{"", "  ", "@octocat", "has space", "a/b"} {
		if _, err := client.FetchRepos(context.Background(), bad); err == nil {
			t.Errorf("expected validation error for username %q", bad)
		}
	}
}

// TestClassifyGraphQLError verifies the GraphQL error classifier returns typed
// sentinels: rate-limited (429, or 403 with a rate-limit body), not-found (404),
// and preserves the cause for everything else.
func TestClassifyGraphQLError(t *testing.T) {
	cause := errors.New("wrapped")
	tests := []struct {
		name         string
		status       int
		body         string
		wantSentinel error
	}{
		{name: "not found", status: http.StatusNotFound, body: `{"message":"Not Found"}`, wantSentinel: ErrNotFound},
		{name: "rate limited 429", status: http.StatusTooManyRequests, body: "", wantSentinel: ErrRateLimited},
		{name: "rate limited 403 with body", status: http.StatusForbidden, body: "API rate limit exceeded", wantSentinel: ErrRateLimited},
		{name: "bare 403 is not rate limit", status: http.StatusForbidden, body: "forbidden", wantSentinel: nil},
		{name: "server error preserved", status: http.StatusInternalServerError, body: "boom", wantSentinel: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := classifyGraphQLError(tc.status, tc.body, cause)
			if tc.wantSentinel == nil {
				if !errors.Is(err, cause) {
					t.Fatalf("expected cause preserved, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantSentinel) {
				t.Fatalf("expected sentinel %v, got %v", tc.wantSentinel, err)
			}
		})
	}
}

// TestQueryActivityProfileDeadline verifies the activity profile query enforces
// its own bounded deadline (activityTimeout) independent of the caller context.
func TestQueryActivityProfileDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the internal 8s deadline so the query must abort on
		// the internal deadline, not on the (open) parent context.
		time.Sleep(20 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"user":{"recent":{}}}}`))
	}))
	defer srv.Close()

	start := time.Now()
	_, _ = testClient(srv.URL).queryActivityProfile(context.Background(), "u")
	elapsed := time.Since(start)

	// Must return near the internal deadline, far below the handler's 20s sleep.
	if elapsed > 15*time.Second {
		t.Fatalf("query did not honor its internal deadline; elapsed %v", elapsed)
	}
}

// TestFetchActivityUsernameValidation verifies activity acquisition validates the
// username up front and returns no observations for invalid input.
func TestFetchActivityUsernameValidation(t *testing.T) {
	if obs := testClient("http://example.invalid").FetchActivityObservations(context.Background(), "@bad"); len(obs) != 0 {
		t.Errorf("expected no observations for invalid username, got %d", len(obs))
	}
}

// TestIsRateLimited exercises the centralized rate-limit classifier used by both
// REST and GraphQL paths.
func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		status int
		body   string
		want   bool
	}{
		{http.StatusForbidden, "API rate limit exceeded", true},
		{http.StatusTooManyRequests, "", true},
		{http.StatusForbidden, "secondary rate limit", true},
		{403, "forbidden", false},
		{200, "ok", false},
		{500, "boom", false},
	}
	for _, tc := range tests {
		if got := isRateLimited(tc.status, tc.body); got != tc.want {
			t.Errorf("isRateLimited(%d,%q)=%v, want %v", tc.status, tc.body, got, tc.want)
		}
	}
}
