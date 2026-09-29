package guardian

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wormhole-foundation/wormholescan-go"
	"github.com/wormhole-foundation/wormholescan-go/api"
)

const errPrefix = "wormholescan: "

// Client is the hand-written client for the guardiand public API (/v1).
type Client struct {
	wc *wormholescan.Client
}

// Option configures a [Client].
type Option = wormholescan.Option

// New constructs a [Client] pointed at Wormholescan mainnet with default retry
// and timeout.
func New(opts ...Option) (*Client, error) {
	wc, err := wormholescan.New(opts...)
	if err != nil {
		return nil, err
	}
	return &Client{wc: wc}, nil
}

// WithBaseURL sets the API origin. The value must include a scheme and host.
// A trailing slash is trimmed.
func WithBaseURL(raw string) Option {
	return wormholescan.WithBaseURL(raw)
}

// WithHTTPClient uses hc for HTTP requests. New wraps a shallow copy so hc is
// not mutated. A nil client is rejected.
func WithHTTPClient(hc *http.Client) Option {
	return wormholescan.WithHTTPClient(hc)
}

// WithUserAgent sets the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return wormholescan.WithUserAgent(ua)
}

// WithRetry sets the maximum number of attempts (including the first) and the
// base delay used for exponential full-jitter backoff.
func WithRetry(maxAttempts int, backoff time.Duration) Option {
	return wormholescan.WithRetry(maxAttempts, backoff)
}

// WithoutRetry disables retries. The client makes a single attempt per request.
func WithoutRetry() Option {
	return wormholescan.WithoutRetry()
}

// deref returns the value pointed to by p, or the zero value of T when p is nil.
func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// chainIDFromInt converts a generated integer chain id pointer to [wormholescan.ChainID].
func chainIDFromInt(id *int) wormholescan.ChainID {
	if id == nil {
		return 0
	}
	return wormholescan.ChainID(*id) //nolint:gosec // Wormhole chain ids fit uint16
}

// chainIDFromVAA converts a generated [api.VaaChainID] pointer to [wormholescan.ChainID].
func chainIDFromVAA(id *api.VaaChainID) wormholescan.ChainID {
	if id == nil {
		return 0
	}
	return wormholescan.ChainID(*id) //nolint:gosec // Wormhole chain ids fit uint16
}

// wrapCall prefixes a transport error with the operation name.
func wrapCall(op string, err error) error {
	return fmt.Errorf("%s%s: %w", errPrefix, op, err)
}

// emptyBody is returned when a 2xx response has no decoded JSON body.
func emptyBody(op string) error {
	return fmt.Errorf("%sdecode %s: empty body", errPrefix, op)
}
