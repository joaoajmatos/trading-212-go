package trading212

import (
	"net/http"
	"strconv"
	"time"
)

// RateLimit holds rate-limiting metadata parsed from Trading 212 response
// headers. The zero value represents an absent or unparseable rate-limit header.
type RateLimit struct {
	// Limit is the total number of requests allowed in the current period
	// (X-RateLimit-Limit).
	Limit int

	// Remaining is the number of requests still available in the current period
	// (X-RateLimit-Remaining).
	Remaining int

	// Used is the number of requests already consumed in the current period
	// (X-RateLimit-Used).
	Used int

	// Reset is the time at which the rate-limit counter resets
	// (X-RateLimit-Reset, Unix timestamp).
	Reset time.Time

	// RetryAfter is the suggested wait duration before retrying after a 429
	// response (Retry-After header, seconds).
	RetryAfter time.Duration
}

// RateLimitFromResponse parses Trading 212 rate-limit headers from an HTTP
// response. It is exported for callers that wrap the HTTP transport and need
// to inspect rate-limit state independently.
func RateLimitFromResponse(r *http.Response) RateLimit {
	var rl RateLimit
	if r == nil {
		return rl
	}
	rl.Limit = parseIntHeader(r, "X-Ratelimit-Limit")
	rl.Remaining = parseIntHeader(r, "X-Ratelimit-Remaining")
	rl.Used = parseIntHeader(r, "X-Ratelimit-Used")

	if ts := parseIntHeader(r, "X-Ratelimit-Reset"); ts > 0 {
		rl.Reset = time.Unix(int64(ts), 0)
	}
	if secs := parseIntHeader(r, "Retry-After"); secs > 0 {
		rl.RetryAfter = time.Duration(secs) * time.Second
	}
	return rl
}

func parseIntHeader(r *http.Response, key string) int {
	v := r.Header.Get(key)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}
