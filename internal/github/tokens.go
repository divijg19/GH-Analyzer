package github

import (
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// token is a selected pool member: its index (for health tracking) and its
// bearer value. An empty value signals an unauthenticated deployment.
type token struct {
	idx   int
	value string
}

// TokenPool manages a set of GitHub tokens (distinct accounts multiply the
// rate-limit ceiling). Selection is a stateless FNV-1a hash-shard on the login,
// so every worker independently pins a login to one token and spreads users
// evenly. Health is tracked in-memory with a TTL; a token is benched when its
// rate-limit budget is low so in-flight requests don't exhaust it. Failover
// picks the healthiest other token.
//
// The pool is advisory: missing or stale health is treated as healthy, and a
// read/parse failure never blocks the hot path.
type TokenPool struct {
	mu     sync.RWMutex
	tokens []string
	health map[int]tokenHealth
}

type tokenHealth struct {
	remaining int
	reset     time.Time
}

// NewTokenPool builds a pool from the given token values. A nil/empty slice is
// valid: Pick returns an unauthenticated token.
func NewTokenPool(tokens []string) *TokenPool {
	return &TokenPool{
		tokens: tokens,
		health: make(map[int]tokenHealth),
	}
}

// Len returns the number of tokens in the pool.
func (p *TokenPool) Len() int {
	if p == nil {
		return 0
	}
	return len(p.tokens)
}

// Pick returns the stateless hash-shard token for a login. The login is
// lowercased so casing GitHub ignores cannot shard the same user onto different
// tokens. Returns an unauthenticated token when the pool is empty.
func (p *TokenPool) Pick(login string) token {
	if p == nil || len(p.tokens) == 0 {
		return token{}
	}
	idx := int(fnvHash(strings.ToLower(login)) % uint32(len(p.tokens)))
	return token{idx: idx, value: p.tokens[idx]}
}

// IsBenched reports whether a token is currently benched (its reset is still in
// the future and its remaining budget is below the bench threshold).
func (p *TokenPool) IsBenched(idx int) bool {
	if p == nil {
		return false
	}
	p.mu.RLock()
	h, ok := p.health[idx]
	p.mu.RUnlock()
	if !ok {
		return false
	}
	now := time.Now()
	if now.After(h.reset) {
		return false
	}
	return h.remaining < benchRemaining
}

// PickFailover returns the healthiest token other than excludeIdx. Tokens with
// unknown or expired health are treated as fully healthy. Benched tokens are
// skipped. Returns ok=false when no eligible token exists.
func (p *TokenPool) PickFailover(excludeIdx int) (token, bool) {
	if p == nil || len(p.tokens) == 0 {
		return token{}, false
	}
	now := time.Now()
	best := -1
	bestRemaining := -1
	for i := range p.tokens {
		if i == excludeIdx {
			continue
		}
		remaining := p.remainingFor(i, now)
		if remaining < benchRemaining {
			continue
		}
		if remaining > bestRemaining {
			best = i
			bestRemaining = remaining
		}
	}
	if best < 0 {
		return token{}, false
	}
	return token{idx: best, value: p.tokens[best]}, true
}

// remainingFor returns the known remaining budget for a token, treating missing
// or expired health as a full (infinite) quota.
func (p *TokenPool) remainingFor(idx int, now time.Time) int {
	p.mu.RLock()
	h, ok := p.health[idx]
	p.mu.RUnlock()
	if !ok {
		return int(^uint(0) >> 1) // effectively infinite
	}
	if now.After(h.reset) {
		return int(^uint(0) >> 1)
	}
	return h.remaining
}

// Record updates a token's health from GitHub's rate-limit response headers. It
// is advisory: unreadable or missing headers are ignored.
func (p *TokenPool) Record(idx int, headers http.Header) {
	if p == nil || len(p.tokens) == 0 {
		return
	}
	remaining := headers.Get("x-ratelimit-remaining")
	reset := headers.Get("x-ratelimit-reset")
	if remaining == "" || reset == "" {
		return
	}
	r, err1 := strconv.Atoi(remaining)
	s, err2 := strconv.ParseInt(reset, 10, 64)
	if err1 != nil || err2 != nil {
		return
	}
	p.mu.Lock()
	p.health[idx] = tokenHealth{remaining: r, reset: time.Unix(s, 0)}
	p.mu.Unlock()
}

// Bench marks a token as temporarily unavailable. until, when non-zero, is the
// time the bench lifts; otherwise benchFallback is used.
func (p *TokenPool) Bench(idx int, until time.Time) {
	if p == nil || len(p.tokens) == 0 {
		return
	}
	if until.IsZero() {
		until = time.Now().Add(benchFallback)
	}
	p.mu.Lock()
	p.health[idx] = tokenHealth{remaining: 0, reset: until}
	p.mu.Unlock()
}

// fnvHash is FNV-1a over the (already lowercased) login: tiny, stable, and
// well-spread for short strings.
func fnvHash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
