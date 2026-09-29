package wormholescan

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		want    Version
		wantErr string
	}{
		{
			name: "live shape",
			body: `{"build_date":"20260902053048","build":"9c56ed7bec04195a8f98300d5a8868a656496ba0","version":"0.0.91-rc1-test"}`,
			want: Version{
				Version:   "0.0.91-rc1-test",
				Build:     "9c56ed7bec04195a8f98300d5a8868a656496ba0",
				BuildDate: time.Date(2026, time.September, 2, 5, 30, 48, 0, time.UTC),
			},
		},
		{
			name: "no build date",
			body: `{"build":"abc","version":"0.0.1"}`,
			want: Version{Version: "0.0.1", Build: "abc"},
		},
		{
			name:    "unparsable build date",
			body:    `{"build_date":"yesterday","build":"abc","version":"0.0.1"}`,
			wantErr: `build_date "yesterday"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newTestClientNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/version", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			})
			got, err := c.GetVersion(context.Background())
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRateLimit(t *testing.T) {
	t.Parallel()

	t.Run("none observed before a response carries headers", func(t *testing.T) {
		t.Parallel()
		c := newTestClientNoRetry(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":"0.0.1"}`))
		})
		_, ok := c.RateLimit()
		assert.False(t, ok)
		_, err := c.GetVersion(context.Background())
		require.NoError(t, err)
		_, ok = c.RateLimit()
		assert.False(t, ok, "no X-RateLimit headers were sent")
	})

	t.Run("latest response wins, including through API()", func(t *testing.T) {
		t.Parallel()
		remaining := 998
		c := newTestClientNoRetry(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("X-Ratelimit-Limit", "1000")
			w.Header().Set("X-Ratelimit-Remaining", itoa(remaining))
			w.Header().Set("X-Ratelimit-Reset", "59")
			w.Header().Set("Content-Type", "application/json")
			remaining--
			_, _ = w.Write([]byte(`{"version":"0.0.1"}`))
		})
		before := time.Now()
		_, err := c.GetVersion(context.Background())
		require.NoError(t, err)
		rl, ok := c.RateLimit()
		require.True(t, ok)
		assert.Equal(t, 1000, rl.Limit)
		assert.Equal(t, 998, rl.Remaining)
		assert.Equal(t, 59*time.Second, rl.Reset)
		assert.False(t, rl.ObservedAt.Before(before))

		_, err = c.API().GetVersionWithResponse(context.Background())
		require.NoError(t, err)
		rl, ok = c.RateLimit()
		require.True(t, ok)
		assert.Equal(t, 997, rl.Remaining)
	})
}

// itoa formats n for a header value.
func itoa(n int) string {
	return strconv.Itoa(n)
}
