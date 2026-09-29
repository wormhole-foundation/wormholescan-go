package guardian

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithBaseURLHeartbeats(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "heartbeats.json"))
	}))
	t.Cleanup(srv.Close)

	c, err := New(WithBaseURL(srv.URL), WithoutRetry())
	require.NoError(t, err)

	got, err := c.Heartbeats(t.Context())
	require.NoError(t, err)
	require.Len(t, got, recordedHeartbeatCount)
	require.Equal(t, "/v1/heartbeats", gotPath)
}

func TestNewRejectsBadBaseURL(t *testing.T) {
	t.Parallel()

	_, err := New(WithBaseURL("not-a-url"))
	require.Error(t, err)
}

// loadFixture reads a committed testdata file.
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return body
}

// newJSONClient serves body with status from every path.
func newJSONClient(t *testing.T, status int, body []byte) *Client {
	t.Helper()
	return newClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
}

// newClient points a [Client] at handler.
func newClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(WithBaseURL(srv.URL), WithoutRetry())
	require.NoError(t, err)
	return c
}
