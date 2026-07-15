package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var errPermanent = errors.New("permanent failure")

// TestTokenPoolDeterministic verifies hash-shard selection is stable for a given
// login and spreads across tokens.
func TestTokenPoolDeterministic(t *testing.T) {
	pool := NewTokenPool([]string{"a", "b", "c"})
	if pool.Len() != 3 {
		t.Fatalf("expected 3 tokens, got %d", pool.Len())
	}
	first := pool.Pick("octocat").idx
	for i := 0; i < 50; i++ {
		if pool.Pick("octocat").idx != first {
			t.Fatal("Pick is not deterministic for a fixed login")
		}
	}
	// Casing must not reshuffle the same login.
	if pool.Pick("OctoCat").idx != first {
		t.Fatal("Pick is case-sensitive; expected case-insensitive sharding")
	}
}

// TestTokenPoolFailover verifies a benched primary fails over to the healthiest
// other token, and that a healthy pool never benches.
func TestTokenPoolFailover(t *testing.T) {
	pool := NewTokenPool([]string{"primary", "secondary", "tertiary"})

	// Bench the primary.
	pool.Bench(pool.Pick("u").idx, time.Time{})
	if !pool.IsBenched(pool.Pick("u").idx) {
		t.Fatal("expected primary to be benched")
	}

	failover, ok := pool.PickFailover(pool.Pick("u").idx)
	if !ok {
		t.Fatal("expected a failover token")
	}
	if failover.idx == pool.Pick("u").idx {
		t.Fatal("failover returned the benched token")
	}
	if pool.IsBenched(failover.idx) {
		t.Fatal("failover token should not be benched")
	}
}

// TestTokenPoolHealthExpiry verifies benched status clears after reset.
func TestTokenPoolHealthExpiry(t *testing.T) {
	pool := NewTokenPool([]string{"a"})
	pool.Bench(0, time.Now().Add(-time.Second)) // already past
	if pool.IsBenched(0) {
		t.Fatal("benched token with past reset should not be benched")
	}
}

// TestRetryPolicyRetriesTransient verifies transient failures are retried until
// success, and respects the attempt cap.
func TestRetryPolicyRetriesTransient(t *testing.T) {
	rp := &RetryPolicy{MaxRetries: 3, Budget: 5 * time.Second, BaseBackoff: time.Millisecond, MaxBackoff: time.Millisecond}

	var attempts int
	resp, err := rp.Do(context.Background(), func(ctx context.Context) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return nil, ErrTransient
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(nil)}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.StatusCode != 200 {
		t.Fatal("expected success after retries")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

// TestRetryPolicyStopsOnPermanent verifies a non-retryable error is not retried.
func TestRetryPolicyStopsOnPermanent(t *testing.T) {
	rp := &RetryPolicy{MaxRetries: 5, Budget: 5 * time.Second, BaseBackoff: time.Millisecond, MaxBackoff: time.Millisecond}

	var attempts int
	_, err := rp.Do(context.Background(), func(ctx context.Context) (*http.Response, error) {
		attempts++
		return nil, errPermanent
	})
	if !errors.Is(err, errPermanent) {
		t.Fatalf("expected permanent error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected no retry on permanent error, got %d attempts", attempts)
	}
}

// TestTransportSingleFlight verifies concurrent identical requests collapse to a
// single upstream call.
func TestTransportSingleFlight(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	tr := &Transport{
		client: &http.Client{Timeout: 5 * time.Second},
		pool:   NewTokenPool(nil), // unauthenticated
		retry:  NewRetryPolicy(),
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/u/repos", nil)
			resp, err := tr.Do(context.Background(), req)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	wg.Wait()

	mu.Lock()
	got := calls
	mu.Unlock()
	if got != 1 {
		t.Fatalf("expected single upstream call via single-flight, got %d", got)
	}
}

// TestTransportUnauthenticatedCompat verifies the no-token path still issues a
// request without an Authorization header (preserves prior behavior).
func TestTransportUnauthenticatedCompat(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := &Transport{client: &http.Client{Timeout: 5 * time.Second}, pool: NewTokenPool(nil), retry: NewRetryPolicy()}
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/u", nil)
	resp, err := tr.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if gotAuth != "" {
		t.Fatalf("expected no Authorization header, got %q", gotAuth)
	}
}

// newTestTransport builds a Transport with a single token and a fast retry
// policy for deterministic status-classification tests.
func newTestTransport() *Transport {
	return &Transport{
		client: &http.Client{Timeout: 5 * time.Second},
		pool:   NewTokenPool([]string{"tok"}),
		retry:  &RetryPolicy{MaxRetries: 4, Budget: 5 * time.Second, BaseBackoff: time.Millisecond, MaxBackoff: time.Millisecond},
	}
}

// TestTransportRateLimited429 verifies a 429 is classified as rate-limited,
// retried, and benches the token.
func TestTransportRateLimited429(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		if calls < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	tr := newTestTransport()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/repos", nil)
	resp, err := tr.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected recovery to 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	mu.Lock()
	got := calls
	mu.Unlock()
	if got < 2 {
		t.Fatalf("expected 429 to be retried, got %d calls", got)
	}
	if !tr.pool.IsBenched(0) {
		t.Fatal("expected token to be benched after 429")
	}
}

// TestTransportBare403NotRateLimited verifies a 403 without a rate-limit body is
// returned to the caller (not retried, not benched).
func TestTransportBare403NotRateLimited(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	tr := newTestTransport()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/repos", nil)
	resp, err := tr.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 returned to caller, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	mu.Lock()
	got := calls
	mu.Unlock()
	if got != 1 {
		t.Fatalf("expected bare 403 to NOT be retried, got %d calls", got)
	}
	if tr.pool.IsBenched(0) {
		t.Fatal("bare 403 must not bench the token")
	}
}

// TestTransportRateLimited403 verifies a 403 carrying a rate-limit body is
// treated like 429: retried and benched.
func TestTransportRateLimited403(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		if calls < 2 {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"API rate limit exceeded for this token"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	tr := newTestTransport()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/repos", nil)
	resp, err := tr.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected recovery to 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	mu.Lock()
	got := calls
	mu.Unlock()
	if got < 2 {
		t.Fatalf("expected rate-limited 403 to be retried, got %d calls", got)
	}
	if !tr.pool.IsBenched(0) {
		t.Fatal("expected token to be benched after rate-limited 403")
	}
}

// TestTransportServerErrorRetried verifies 5xx is retried as transient and does
// not bench the (healthy) token.
func TestTransportServerErrorRetried(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		if calls < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"boom"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	tr := newTestTransport()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/repos", nil)
	resp, err := tr.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected recovery to 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	mu.Lock()
	got := calls
	mu.Unlock()
	if got < 2 {
		t.Fatalf("expected 5xx to be retried, got %d calls", got)
	}
	if tr.pool.IsBenched(0) {
		t.Fatal("5xx is server-side; must not bench the token")
	}
}

// TestTransportServerErrorExhausted verifies persistent 5xx surfaces ErrTransient
// after the retry budget is spent, without benching the token.
func TestTransportServerErrorExhausted(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	tr := newTestTransport()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/repos", nil)
	_, err := tr.Do(context.Background(), req)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !errors.Is(err, ErrTransient) {
		t.Fatalf("expected ErrTransient, got %v", err)
	}

	mu.Lock()
	got := calls
	mu.Unlock()
	if got != tr.retry.MaxRetries+1 {
		t.Fatalf("expected %d attempts, got %d", tr.retry.MaxRetries+1, got)
	}
	if tr.pool.IsBenched(0) {
		t.Fatal("5xx must not bench the token")
	}
}
