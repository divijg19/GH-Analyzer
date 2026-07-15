package github

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"regexp"
	"time"
)

// rateLimitMessage matches GitHub's explicit rate-limit error bodies, which may
// arrive with a 403 (or a 200-surrounded GraphQL payload).
var rateLimitMessage = regexp.MustCompile(`(?i)api rate limit exceeded|rate limit|secondary rate`)

// isRateLimited reports whether a response (status + body) indicates rate
// limiting. A 429 is always rate limiting; a 403 is only rate limiting when the
// body confirms throttling (a bare 403 is authorization, not throttling). This
// mirrors the acquisition layer's classifier; this package intentionally keeps
// its own copy rather than importing acquisition, preserving the layering
// (github is the lower transport layer and must not depend on acquisition).
func isRateLimited(statusCode int, body string) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode == http.StatusForbidden {
		return rateLimitMessage.MatchString(body)
	}
	return false
}

// Typed transport error taxonomy. The acquisition layer maps these to its own
// sentinels (see internal/acquisition/errors.go); this package never parses
// message strings to classify.
var (
	// ErrTransient means the failure is non-deterministic (network error, 5xx,
	// context deadline) and may succeed on retry.
	ErrTransient = errors.New("github transport: transient failure")

	// ErrRateLimited means GitHub rejected the request due to rate limiting.
	ErrRateLimited = errors.New("github transport: rate limited")
)

// RetryPolicy bounds retry attempts by count and a total time budget. It retries
// only retryable failures (ErrTransient, ErrRateLimited) and always respects
// context cancellation. Backoff is exponential with full jitter.
type RetryPolicy struct {
	MaxRetries  int
	Budget      time.Duration
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
}

// NewRetryPolicy returns the default transport retry policy.
func NewRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:  defaultMaxRetries,
		Budget:      defaultRetryBudget,
		BaseBackoff: defaultBaseBackoff,
		MaxBackoff:  defaultMaxBackoff,
	}
}

// Do runs op, retrying retryable failures until success, the attempt cap, the
// time budget, or context cancellation. The final error (possibly wrapped) is
// returned if all attempts fail.
func (rp *RetryPolicy) Do(ctx context.Context, op func(context.Context) (*http.Response, error)) (*http.Response, error) {
	deadline := time.Now().Add(rp.Budget)
	var lastErr error
	for attempt := 0; attempt <= rp.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, err
		}
		if attempt > 0 && time.Now().After(deadline) {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, ErrTransient
		}

		resp, err := op(ctx)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetryable(err) {
			return nil, err
		}
		if attempt == rp.MaxRetries {
			return nil, err
		}

		backoff := rp.BaseBackoff
		for i := 0; i < attempt; i++ {
			backoff *= 2
		}
		if backoff > rp.MaxBackoff {
			backoff = rp.MaxBackoff
		}
		backoff = jitter(backoff)

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func isRetryable(err error) bool {
	return errors.Is(err, ErrTransient) || errors.Is(err, ErrRateLimited)
}

// jitter returns a random duration in [0, d], giving exponential backoff full
// jitter so concurrent retries spread out.
func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(d) + 1))
}
