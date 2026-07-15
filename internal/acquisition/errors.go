package acquisition

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

// rateLimitMessage matches GitHub's explicit rate-limit error bodies, which may
// arrive with a 200-surrounded GraphQL payload or a non-403 status.
var rateLimitMessage = regexp.MustCompile(`(?i)api rate limit exceeded|rate limit|secondary rate`)

// APIError represents a non-successful GitHub REST API response. It carries the
// HTTP status code so callers can map GitHub failures to their own responses.
type APIError struct {
	StatusCode int
	Message    string
}

func (e APIError) Error() string {
	return e.Message
}

// Typed error taxonomy shared across the REST and GraphQL acquisition paths.
// Callers classify failures by these sentinels (via errors.Is) rather than by
// parsing message strings. See research/engineering/standards.md (typed errors).
var (
	// ErrNotFound means the requested GitHub resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrRateLimited means GitHub rejected the request due to rate limiting
	// (HTTP 403/429 or an explicit "API rate limit exceeded" GraphQL message).
	ErrRateLimited = errors.New("github rate limit exceeded")

	// ErrTransient means the failure is non-deterministic and may succeed on
	// retry (network error, 5xx, timeout).
	ErrTransient = errors.New("transient acquisition failure")
)

// isRateLimited reports whether a GitHub response (status + body) indicates rate
// limiting. A 429 is always rate limiting; a 403 is only rate limiting when the
// body confirms it (a bare 403 is more often authorization, not throttling).
// Kept centralized so REST and GraphQL classification agree.
func isRateLimited(statusCode int, body string) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode == http.StatusForbidden {
		return rateLimitMessage.MatchString(body)
	}
	return false
}

// classifyGraphQLError maps a failed GraphQL response to a typed sentinel when
// the failure is classifiable (rate-limited, not-found), preserving the original
// error as the wrapped cause for non-classifiable failures.
func classifyGraphQLError(statusCode int, body string, cause error) error {
	if isRateLimited(statusCode, body) {
		return fmt.Errorf("%w: %v", ErrRateLimited, cause)
	}
	if statusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %v", ErrNotFound, cause)
	}
	return cause
}
