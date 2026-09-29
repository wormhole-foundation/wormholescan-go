package wormholescan

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

// RateLimit is the server's rate-limit state as reported by the
// X-Ratelimit-* headers of the most recent response.
type RateLimit struct {
	// Limit is the number of requests allowed per window (1000 per minute
	// on api.wormholescan.io at the time of writing).
	Limit int
	// Remaining is the number of requests left in the current window.
	Remaining int
	// Reset is how long the current window had left when the response was
	// received.
	Reset time.Duration
	// ObservedAt is when the response carrying these values was received.
	ObservedAt time.Time
}

// rateLimitHeaders are the response headers the server uses. Reset is in
// seconds.
const (
	rateLimitLimitHeader     = "X-Ratelimit-Limit"
	rateLimitRemainingHeader = "X-Ratelimit-Remaining"
	rateLimitResetHeader     = "X-Ratelimit-Reset"
)

// rateLimitState holds the last observed RateLimit for concurrent readers
// and one writer per response.
type rateLimitState struct {
	last atomic.Pointer[RateLimit]
}

// observe records the rate-limit headers of resp, if it carries them.
func (s *rateLimitState) observe(resp *http.Response, now time.Time) {
	if resp == nil {
		return
	}
	limit, ok := headerInt(resp.Header, rateLimitLimitHeader)
	if !ok {
		return
	}
	remaining, _ := headerInt(resp.Header, rateLimitRemainingHeader)
	resetSeconds, _ := headerInt(resp.Header, rateLimitResetHeader)
	s.last.Store(&RateLimit{
		Limit:      limit,
		Remaining:  remaining,
		Reset:      time.Duration(resetSeconds) * time.Second,
		ObservedAt: now,
	})
}

// headerInt parses header key as a base-10 integer.
func headerInt(h http.Header, key string) (int, bool) {
	raw := h.Get(key)
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return n, true
}

// RateLimit returns the rate-limit state reported by the most recent response
// from the server, and false if no response has carried rate-limit headers
// yet. Requests made through [Client.API] are included.
func (c *Client) RateLimit() (RateLimit, bool) {
	last := c.rateLimit.last.Load()
	if last == nil {
		return RateLimit{}, false
	}
	return *last, true
}
