package wormholescan

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/wormhole-foundation/wormholescan-go/api"
)

const (
	// MainnetURL is the Wormholescan production API origin.
	MainnetURL = "https://api.wormholescan.io"
	// TestnetURL is the Wormholescan testnet API origin.
	TestnetURL = "https://api.testnet.wormholescan.io"
)

const (
	defaultHTTPTimeout   = 30 * time.Second
	defaultRetryAttempts = 3
	defaultRetryBackoff  = 500 * time.Millisecond
	maxRetryBackoff      = 10 * time.Second
	maxRetryAfterSeconds = int(math.MaxInt64 / int64(time.Second))
	userAgentProduct     = "wormholescan-go"
	develVersion         = "(devel)"
	modulePath           = "github.com/wormhole-foundation/wormholescan-go"
)

// Client is the hand-written Wormholescan API client.
type Client struct {
	// api is the generated client sharing this client's transport and headers.
	api *api.ClientWithResponses
}

// config is the resolved client configuration after applying options.
type config struct {
	// baseURL is the API origin without a trailing slash.
	baseURL string
	// userAgent is sent as the User-Agent header.
	userAgent string
	// httpClient is the caller-supplied client, or nil to use the default.
	httpClient *http.Client
	// retryAttempts is the maximum number of tries, including the first.
	retryAttempts int
	// retryBackoff is the base delay for exponential full-jitter backoff.
	retryBackoff time.Duration
}

// Option configures a [Client].
type Option func(*config) error

// New constructs a [Client] pointed at mainnet with default retry and timeout.
func New(opts ...Option) (*Client, error) {
	cfg := config{
		baseURL:       MainnetURL,
		userAgent:     defaultUserAgent(),
		retryAttempts: defaultRetryAttempts,
		retryBackoff:  defaultRetryBackoff,
	}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	httpClient := buildHTTPClient(cfg)
	apiClient, err := api.NewClientWithResponses(
		cfg.baseURL,
		api.WithHTTPClient(httpClient),
		api.WithRequestEditorFn(userAgentEditor(cfg.userAgent)),
	)
	if err != nil {
		return nil, fmt.Errorf("%screate API client: %w", errPrefix, err)
	}
	return &Client{api: apiClient}, nil
}

// WithBaseURL sets the API origin. The value must include a scheme and host.
// A trailing slash is trimmed.
func WithBaseURL(raw string) Option {
	return func(cfg *config) error {
		parsed, err := url.Parse(raw)
		if err != nil {
			return fmt.Errorf("%sbase URL: %w", errPrefix, err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("%sbase URL %q is missing scheme or host", errPrefix, raw)
		}
		cfg.baseURL = strings.TrimRight(raw, "/")
		return nil
	}
}

// WithHTTPClient uses hc for HTTP requests. New wraps a shallow copy so hc is
// not mutated. A nil client is rejected.
func WithHTTPClient(hc *http.Client) Option {
	return func(cfg *config) error {
		if hc == nil {
			return fmt.Errorf("%sHTTP client is nil", errPrefix)
		}
		cfg.httpClient = hc
		return nil
	}
}

// WithUserAgent sets the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return func(cfg *config) error {
		cfg.userAgent = ua
		return nil
	}
}

// WithRetry sets the maximum number of attempts (including the first) and the
// base delay used for exponential full-jitter backoff.
func WithRetry(maxAttempts int, backoff time.Duration) Option {
	return func(cfg *config) error {
		if maxAttempts < 1 {
			return fmt.Errorf("%sretry attempts must be at least 1", errPrefix)
		}
		if backoff < 0 {
			return fmt.Errorf("%sretry backoff must be non-negative", errPrefix)
		}
		cfg.retryAttempts = maxAttempts
		cfg.retryBackoff = backoff
		return nil
	}
}

// WithoutRetry disables retries. The client makes a single attempt per request.
func WithoutRetry() Option {
	return func(cfg *config) error {
		cfg.retryAttempts = 1
		return nil
	}
}

// API returns the generated client that shares this client's base URL, HTTP
// client, user agent, and retry transport.
func (c *Client) API() *api.ClientWithResponses {
	return c.api
}

// defaultUserAgent returns "wormholescan-go/<version>" from build info, or
// "wormholescan-go" when the version is unavailable.
func defaultUserAgent() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return userAgentProduct
	}
	version := moduleVersion(info)
	if version == "" || version == develVersion {
		return userAgentProduct
	}
	return userAgentProduct + "/" + version
}

// moduleVersion returns this module's version from build info.
func moduleVersion(info *debug.BuildInfo) string {
	if info.Main.Path == modulePath {
		return info.Main.Version
	}
	for _, dep := range info.Deps {
		if dep != nil && dep.Path == modulePath {
			return dep.Version
		}
	}
	return ""
}

// userAgentEditor returns a request editor that sets User-Agent to ua.
func userAgentEditor(ua string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", ua)
		return nil
	}
}

// buildHTTPClient returns the HTTP client used for all requests, wrapping its
// transport with retry policy.
func buildHTTPClient(cfg config) *http.Client {
	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	} else {
		clone := *httpClient
		httpClient = &clone
	}
	base := httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	httpClient.Transport = &retryTransport{
		base:     base,
		attempts: cfg.retryAttempts,
		backoff:  cfg.retryBackoff,
	}
	return httpClient
}

// retryTransport retries idempotent requests on transport errors and a small
// set of HTTP status codes.
type retryTransport struct {
	// base performs the actual request.
	base http.RoundTripper
	// attempts is the maximum number of tries, including the first.
	attempts int
	// backoff is the base delay for exponential full-jitter backoff.
	backoff time.Duration
}

// RoundTrip implements [http.RoundTripper].
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	attempt := 0
	for {
		attempt++
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		resp, err := t.base.RoundTrip(req)
		if attempt >= t.attempts || !retriable(req, resp, err) {
			return resp, err
		}
		delay := t.delay(attempt, resp)
		if exceedsDeadline(ctx, delay) {
			return resp, err
		}
		drainAndClose(resp)
		if err := sleep(ctx, delay); err != nil {
			return nil, err
		}
	}
}

// retriable reports whether the request may be tried again.
func retriable(req *http.Request, resp *http.Response, err error) bool {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return false
	}
	if err != nil {
		return req.Context().Err() == nil && transientTransportError(err)
	}
	if resp == nil {
		return false
	}
	switch resp.StatusCode {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// transientTransportError reports whether err is a transient network failure.
func transientTransportError(err error) bool {
	for err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return true
		}
		if errors.Is(err, syscall.ECONNRESET) ||
			errors.Is(err, syscall.EPIPE) ||
			errors.Is(err, syscall.ECONNABORTED) {
			return true
		}
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
			continue
		}
		var netErr net.Error
		return errors.As(err, &netErr)
	}
	return false
}

// exceedsDeadline reports whether waiting delay would miss ctx's deadline.
func exceedsDeadline(ctx context.Context, delay time.Duration) bool {
	dl, ok := ctx.Deadline()
	return ok && time.Until(dl) < delay
}

// delay returns how long to wait before the next attempt.
func (t *retryTransport) delay(failedAttempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
			return d
		}
	}
	capDelay := t.backoff
	shifts := failedAttempt - 1
	for range shifts {
		if capDelay >= maxRetryBackoff/2 {
			capDelay = maxRetryBackoff
			break
		}
		capDelay *= 2
	}
	if capDelay > maxRetryBackoff {
		capDelay = maxRetryBackoff
	}
	return fullJitter(capDelay)
}

// parseRetryAfter parses a Retry-After value as integer seconds or an HTTP date.
func parseRetryAfter(raw string) (time.Duration, bool) {
	if raw == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(raw); err == nil {
		if secs < 0 {
			return 0, true
		}
		if secs > maxRetryAfterSeconds {
			secs = maxRetryAfterSeconds
		}
		return time.Duration(secs) * time.Second, true
	}
	when, err := http.ParseTime(raw)
	if err != nil {
		return 0, false
	}
	d := time.Until(when)
	if d < 0 {
		return 0, true
	}
	return d, true
}

// fullJitter returns a delay in [0, cap).
func fullJitter(capDelay time.Duration) time.Duration {
	if capDelay <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(capDelay))) //nolint:gosec // retry jitter is not a cryptographic decision
}

// sleep waits for d or until ctx is done.
func sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// drainAndClose discards and closes a response body that will not be returned.
func drainAndClose(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.CopyN(io.Discard, resp.Body, maxErrorBodyBytes)
	_ = resp.Body.Close()
}
