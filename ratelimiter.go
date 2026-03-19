package trading212

import (
	"context"
	"strings"
	"sync"
	"time"
)

// WithRateLimiting enables automatic client-side rate limiting. When set, the
// client waits before sending a request that would exceed Trading 212's
// documented API limits, rather than letting it fail with a 429 response.
//
// Limits are tracked per [Client] instance. If multiple clients share the same
// API key, each tracks its own limits independently — consider sharing a single
// client instead.
//
// Context cancellation is respected: if ctx is cancelled while a call is
// waiting for a rate-limit slot, the wait is abandoned and ctx.Err() is
// returned.
func WithRateLimiting() Option {
	return func(c *Client) {
		c.rateLimiter = newEndpointRateLimiter()
	}
}

// endpointRateLimiter holds a token bucket per endpoint.
type endpointRateLimiter struct {
	buckets map[string]*tokenBucket // key: "METHOD /canonical/path"
}

func newEndpointRateLimiter() *endpointRateLimiter {
	return &endpointRateLimiter{
		buckets: map[string]*tokenBucket{
			// Account
			"GET /equity/account/summary": newTokenBucket(1, 5*time.Second),

			// Orders
			"GET /equity/orders":          newTokenBucket(1, 5*time.Second),
			"GET /equity/orders/{id}":     newTokenBucket(1, time.Second),
			"POST /equity/orders/market":  newTokenBucket(50, 60*time.Second),
			"POST /equity/orders/limit":   newTokenBucket(1, 2*time.Second),
			"POST /equity/orders/stop":    newTokenBucket(1, 2*time.Second),
			"POST /equity/orders/stop_limit": newTokenBucket(1, 2*time.Second),
			"DELETE /equity/orders/{id}":  newTokenBucket(50, 60*time.Second),

			// Positions
			"GET /equity/positions": newTokenBucket(1, time.Second),

			// Metadata
			"GET /equity/metadata/instruments": newTokenBucket(1, 50*time.Second),
			"GET /equity/metadata/exchanges":   newTokenBucket(1, 30*time.Second),

			// History
			"GET /equity/history/orders":        newTokenBucket(6, 60*time.Second),
			"GET /equity/history/dividends":     newTokenBucket(6, 60*time.Second),
			"GET /equity/history/transactions":  newTokenBucket(6, 60*time.Second),
			"POST /equity/history/exports":      newTokenBucket(1, 30*time.Second),
			"GET /equity/history/exports":       newTokenBucket(1, 60*time.Second),

			// Pies
			"GET /equity/pies":                    newTokenBucket(1, 5*time.Second),
			"POST /equity/pies":                   newTokenBucket(1, 5*time.Second),
			"GET /equity/pies/{id}":               newTokenBucket(1, 5*time.Second),
			"POST /equity/pies/{id}":              newTokenBucket(1, 5*time.Second),
			"DELETE /equity/pies/{id}":            newTokenBucket(1, 30*time.Second),
			"POST /equity/pies/{id}/duplicate":    newTokenBucket(1, 30*time.Second),
		},
	}
}

// wait blocks until a token is available for the given request, or until ctx
// is cancelled.
func (r *endpointRateLimiter) wait(ctx context.Context, method, path string) error {
	// Strip query string — paths arrive as "/equity/history/orders?limit=50".
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}
	key := method + " " + canonicalizePath(path)
	bucket, ok := r.buckets[key]
	if !ok {
		return nil // unknown endpoint — don't throttle
	}
	return bucket.wait(ctx)
}

// canonicalizePath replaces purely numeric path segments with "{id}" so that
// "/equity/orders/42" maps to the same bucket as "/equity/orders/99".
func canonicalizePath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if isAllDigits(p) {
			parts[i] = "{id}"
		}
	}
	return strings.Join(parts, "/")
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Token bucket
// ---------------------------------------------------------------------------

// tokenBucket is a thread-safe token-bucket rate limiter.
//
// It starts full (burst tokens available) and refills at a constant rate.
// Callers block on wait until a token is available or ctx is cancelled.
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64 // tokens per second
	last       time.Time
}

// newTokenBucket creates a bucket that allows burst requests per interval,
// starting full.
func newTokenBucket(burst int, interval time.Duration) *tokenBucket {
	rate := float64(burst) / interval.Seconds()
	return &tokenBucket{
		tokens:     float64(burst),
		capacity:   float64(burst),
		refillRate: rate,
		last:       time.Now(),
	}
}

// wait blocks until a token is available or ctx is done.
func (b *tokenBucket) wait(ctx context.Context) error {
	for {
		b.mu.Lock()
		now := time.Now()
		// Refill tokens based on elapsed time.
		b.tokens += now.Sub(b.last).Seconds() * b.refillRate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.last = now

		if b.tokens >= 1 {
			b.tokens--
			b.mu.Unlock()
			return nil
		}

		// Calculate how long until the next token arrives.
		wait := time.Duration((1-b.tokens)/b.refillRate*1e9) * time.Nanosecond
		b.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
			// Loop and try again — another goroutine may have consumed the token.
		}
	}
}
