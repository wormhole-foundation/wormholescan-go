package wormholescan

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wormhole-foundation/wormholescan-go/api"
)

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := New()
	require.NoError(t, err)
	inner, ok := c.API().ClientInterface.(*api.Client)
	require.True(t, ok)
	assert.Equal(t, MainnetURL+"/", inner.Server)

	c, err = New(WithBaseURL(srv.URL), WithoutRetry())
	require.NoError(t, err)
	resp, err := c.API().HealthCheckWithResponse(t.Context())
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, userAgentProduct, gotUA)
}

func TestWithBaseURLRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{name: "empty", url: ""},
		{name: "no scheme", url: "nope"},
		{name: "empty host", url: "http://"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := New(WithBaseURL(tt.url))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "base URL")
		})
	}
}

func TestRetryHonoursRetryAfter(t *testing.T) {
	t.Parallel()

	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if n.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := New(WithBaseURL(srv.URL))
	require.NoError(t, err)
	start := time.Now()
	resp, err := doMethod(t.Context(), t, c, http.MethodGet, srv.URL)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(start), time.Second)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), n.Load())
}

func TestRetryAfterPastDeadlineReturnsResponse(t *testing.T) {
	t.Parallel()

	const retryAfterSeconds = 60
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n.Add(1)
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	c, err := New(WithBaseURL(srv.URL))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	t.Cleanup(cancel)
	start := time.Now()
	resp, err := doMethod(ctx, t, c, http.MethodGet, srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, int32(1), n.Load())
	assert.Less(t, time.Since(start), time.Duration(retryAfterSeconds)*time.Second)
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)
	future := time.Now().Add(5 * time.Second).UTC().Format(http.TimeFormat)

	tests := []struct {
		name    string
		raw     string
		wantOK  bool
		wantMin time.Duration
		wantMax time.Duration
	}{
		{name: "seconds", raw: "1", wantOK: true, wantMin: time.Second, wantMax: time.Second},
		{name: "negative", raw: "-1", wantOK: true, wantMin: 0, wantMax: 0},
		{name: "garbage", raw: "nope", wantOK: false},
		{name: "empty", raw: "", wantOK: false},
		{name: "past HTTP-date", raw: past, wantOK: true, wantMin: 0, wantMax: 0},
		{name: "future HTTP-date", raw: future, wantOK: true, wantMin: time.Nanosecond, wantMax: 5 * time.Second},
		{
			name:    "overflow seconds",
			raw:     "10000000000",
			wantOK:  true,
			wantMin: time.Duration(maxRetryAfterSeconds) * time.Second,
			wantMax: time.Duration(maxRetryAfterSeconds) * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseRetryAfter(tt.raw)
			assert.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				return
			}
			assert.GreaterOrEqual(t, got, tt.wantMin)
			assert.LessOrEqual(t, got, tt.wantMax)
		})
	}
}

func TestPostServiceUnavailableIsNotRetried(t *testing.T) {
	t.Parallel()

	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	c, err := New(WithBaseURL(srv.URL))
	require.NoError(t, err)
	resp, err := doMethod(t.Context(), t, c, http.MethodPost, srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, int32(1), n.Load())
}

func TestWithoutRetryMakesOneRequest(t *testing.T) {
	t.Parallel()

	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	c, err := New(WithBaseURL(srv.URL), WithoutRetry())
	require.NoError(t, err)
	resp, err := doMethod(t.Context(), t, c, http.MethodGet, srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, int32(1), n.Load())
}

func TestRetryStopsWhenContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n.Add(1)
		cancel()
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	c, err := New(WithBaseURL(srv.URL))
	require.NoError(t, err)
	_, err = doMethod(ctx, t, c, http.MethodGet, srv.URL)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, int32(1), n.Load())
}

// doMethod issues method against rawURL using c's retrying HTTP client.
func doMethod(
	ctx context.Context,
	t *testing.T,
	c *Client,
	method string,
	rawURL string,
) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, method, rawURL, http.NoBody)
	require.NoError(t, err)
	resp, err := httpClientOf(t, c).Do(req)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	})
	return resp, nil
}

// httpClientOf returns the [*http.Client] wired into c.
func httpClientOf(t *testing.T, c *Client) *http.Client {
	t.Helper()
	inner, ok := c.API().ClientInterface.(*api.Client)
	require.True(t, ok)
	hc, ok := inner.Client.(*http.Client)
	require.True(t, ok)
	return hc
}
