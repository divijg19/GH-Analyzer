// Package github implements the Transport layer of the Atlas Intelligence
// Ontology (see docs/INTELLIGENCE.md). It owns HTTP communication with the
// GitHub REST API. It never parses domain types or interprets responses.
//
// The Transport is provider-independent by design: token management, failover,
// retries, and request coalescing are internal concerns, and the layer depends
// only on the Go standard library. See .opencode/research/engineering/standards.md
// (provider independence, retry policy, graceful degradation).
package github

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	defaultUserAgent = "atlas"
	defaultTimeout   = 30 * time.Second

	// benchRemaining: a token with fewer than this many points left is benched
	// until its reset so in-flight requests don't blow through zero.
	benchRemaining = 200
	// benchFallback: how long to bench a token when a limited response carries no
	// usable reset/Retry-After.
	benchFallback = 120 * time.Second

	defaultMaxRetries  = 4
	defaultRetryBudget = 60 * time.Second
	defaultBaseBackoff = 500 * time.Millisecond
	defaultMaxBackoff  = 8 * time.Second
)

// Transport executes authenticated HTTP requests against GitHub with token
// pooling, health-aware failover, bounded retries, and single-flight coalescing.
type Transport struct {
	client *http.Client
	pool   *TokenPool
	retry  *RetryPolicy
	group  singleflight.Group
}

// Default is the process-wide Transport. It is initialized lazily and reads
// GITHUB_TOKENS (comma-separated, distinct accounts) or GITHUB_TOKEN (pool of
// one) from the environment. A deployment with no token configured still works:
// requests are sent unauthenticated.
var Default = newDefaultTransport()

func newDefaultTransport() *Transport {
	return &Transport{
		client: &http.Client{Timeout: defaultTimeout},
		pool:   NewTokenPool(envTokens()),
		retry:  NewRetryPolicy(),
	}
}

// envTokens reads the token configuration. GITHUB_TOKENS (multiple accounts)
// wins; GITHUB_TOKEN alone is treated as a single-token pool.
func envTokens() []string {
	if multi := os.Getenv("GITHUB_TOKENS"); multi != "" {
		out := make([]string, 0, 1)
		for _, t := range strings.Split(multi, ",") {
			if t = strings.TrimSpace(t); t != "" {
				out = append(out, t)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	if single := os.Getenv("GITHUB_TOKEN"); single != "" {
		return []string{strings.TrimSpace(single)}
	}
	return nil
}

// Do executes an authenticated request. Token selection is a stateless hash-shard
// on the login (when present), so every instance independently spreads users
// evenly and pins a login to one token (composing with retries and failover).
// Retries apply only to transient or rate-limited failures. Concurrent identical
// requests are coalesced via single-flight; the request body is buffered once so
// it can be replayed across retries and shared safely across coalesced callers.
func (t *Transport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Buffer the body once so it is replayable for retries and independently
	// readable by each coalesced caller. GET requests have a nil body.
	var bodyBytes []byte
	var err error
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	req.ContentLength = int64(len(bodyBytes))

	login := loginFromURL(req.URL.Path)
	key := requestKey(req)

	doOnce := func() (any, error) {
		// Runs at most once per key. Capture the full response here (under
		// single-flight) so each coalesced caller gets an independent copy.
		resp, err := t.executeWithRetry(ctx, req, login, bodyBytes)
		if err != nil {
			return nil, err
		}
		raw, berr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if berr != nil {
			return nil, ErrTransient
		}
		return capturedResponse{
			statusCode: resp.StatusCode,
			status:     resp.Status,
			header:     resp.Header,
			body:       raw,
		}, nil
	}

	res, err, _ := t.group.Do(key, doOnce)
	if err != nil {
		return nil, err
	}
	cap, ok := res.(capturedResponse)
	if !ok {
		return nil, ErrTransient
	}
	// Each coalesced caller receives an independent response with its own body
	// reader, so concurrent read/close cannot race.
	return &http.Response{
		StatusCode: cap.statusCode,
		Status:     cap.status,
		Header:     cap.header,
		Body:       io.NopCloser(bytes.NewReader(cap.body)),
		Request:    req,
	}, nil
}

// capturedResponse is the single-flight result: a self-contained snapshot of an
// HTTP response so each coalesced caller can build an independent http.Response.
type capturedResponse struct {
	statusCode int
	status     string
	header     http.Header
	body       []byte
}

func (t *Transport) executeWithRetry(ctx context.Context, req *http.Request, login string, body []byte) (*http.Response, error) {
	return t.retry.Do(ctx, func(ctx context.Context) (*http.Response, error) {
		tok, err := t.pickToken(ctx, login)
		if err != nil {
			return nil, err
		}
		// Replay the buffered body for each attempt.
		cloned := cloneWithToken(req, tok.value)
		cloned.Body = io.NopCloser(bytes.NewReader(body))
		cloned.ContentLength = int64(len(body))

		resp, err := t.client.Do(cloned)
		if err != nil {
			// Network failure is transient; bench the token so the next pick can
			// fail over, then surface ErrTransient for retry.
			t.pool.Bench(tok.idx, time.Time{})
			return nil, ErrTransient
		}
		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			// Always rate limited: bench the token and retry.
			t.pool.Bench(tok.idx, time.Time{})
			drain(resp)
			return nil, fmt.Errorf("%w: GitHub API returned status %d", ErrRateLimited, resp.StatusCode)
		case resp.StatusCode == http.StatusForbidden:
			// A bare 403 is authorization/forbidden, NOT throttling. Only treat
			// it as rate-limited when the body confirms throttling; otherwise
			// return the response so the caller classifies it. We neither bench
			// the token nor waste retries on a deterministic auth failure.
			bodyBytes, rerr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if rerr == nil && isRateLimited(resp.StatusCode, string(bodyBytes)) {
				t.pool.Bench(tok.idx, time.Time{})
				return nil, fmt.Errorf("%w: GitHub API returned status %d", ErrRateLimited, resp.StatusCode)
			}
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			return resp, nil
		case resp.StatusCode >= 500:
			// Server-side transient failure. Retry without benching the token:
			// the credential is healthy, GitHub is the one struggling.
			return nil, ErrTransient
		case resp.StatusCode >= 400:
			// Other deterministic client errors (404, 422, bare 403): surface to
			// the caller, which classifies them. Not retried.
			return resp, nil
		}
		t.pool.Record(tok.idx, resp.Header)
		return resp, nil
	})
}

// pickToken selects the stateless hash-shard token, or fails over to the
// healthiest other token when the sharded token is benched or exhausted.
func (t *Transport) pickToken(ctx context.Context, login string) (token, error) {
	if t.pool == nil || t.pool.Len() == 0 {
		return token{}, nil // unauthenticated deployment
	}
	primary := t.pool.Pick(login)
	if !t.pool.IsBenched(primary.idx) {
		return primary, nil
	}
	if failover, ok := t.pool.PickFailover(primary.idx); ok {
		return failover, nil
	}
	// Every other token benched: we still try the primary (best effort) but the
	// caller will likely receive a rate-limited response and retry/budget out.
	return primary, nil
}

// Do is the package-level entry point used by the acquisition layer. It delegates
// to the process-wide Default transport.

func cloneWithToken(req *http.Request, token string) *http.Request {
	clone := req.Clone(req.Context())
	setHeaders(clone)
	if token != "" {
		clone.Header.Set("Authorization", "Bearer "+token)
	}
	return clone
}

func setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", defaultUserAgent)
}

// Do is the package-level entry point used by the acquisition layer. It delegates
// to the process-wide Default transport.
func Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return Default.Do(ctx, req)
}

// requestKey derives a single-flight key from method + URL + body. The body is
// read once; callers must not have consumed it (the acquisition layer never
// does before calling Do).
func requestKey(req *http.Request) string {
	var b bytes.Buffer
	if req.Body != nil {
		if _, err := io.Copy(&b, req.Body); err != nil {
			// On copy failure, fall back to a key without the body.
			return req.Method + " " + req.URL.String()
		}
		req.Body = io.NopCloser(&b)
	}
	return req.Method + " " + req.URL.String() + " " + b.String()
}

// drain discards and closes a response body so the underlying connection is
// reusable. Safe to call on any response.
func drain(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

func loginFromURL(path string) string {
	// Expects /users/{login}... or /users/{login}/repos; returns the login or "".
	const prefix = "/users/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(path, prefix)
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		rest = rest[:i]
	}
	return rest
}
