package wormholescan

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	governorTestPageSize     = 2
	governorTestGuardianHex  = "7899ceab1dc961dae9defdb7a4f521269a5448fc"
	governorTestGuardianPath = "0x7899ceab1dc961dae9defdb7a4f521269a5448fc"
	governorTestNotional     = uint64(20_000_000)
	governorTestBigTx        = uint64(1_000_000)
	governorTestRequestID    = "8a51486e-3dc0-4599-843d-9e36273ae91b"
	governorTestPriceDelta   = 0.0001
)

func TestGetGovernorConfig(t *testing.T) {
	t.Parallel()

	c := newGovernorClient(t, governorFixtureHandler(t, "governor_config_guardian.json"))
	got, err := c.GetGovernorConfig(t.Context(), governorTestGuardianPath)
	require.NoError(t, err)

	assert.Equal(t, GuardianAddress(governorTestGuardianHex), got.GuardianAddress)
	assert.Equal(t, "Google Cloud", got.NodeName)
	require.Len(t, got.Chains, governorTestPageSize)
	assert.Equal(t, ChainIDSolana, got.Chains[0].Chain)
	assert.Equal(t, governorTestNotional, got.Chains[0].NotionalLimit)
	assert.Equal(t, governorTestBigTx, got.Chains[0].BigTransactionSize)
	require.Len(t, got.Tokens, governorTestPageSize)
	assert.Equal(t, ChainIDBSC, got.Tokens[0].OriginChain)
	assert.InDelta(t, 124.07, got.Tokens[0].Price, governorTestPriceDelta)
	assert.Equal(t, ChainIDBase, got.Tokens[1].OriginChain)
	assert.InDelta(t, 2782.3, got.Tokens[1].Price, governorTestPriceDelta)
}

func TestGetGovernorStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture string
		check   func(t *testing.T, got GovernorStatus)
	}{
		{
			name:    "object enqueued VAAs with null releaseTime",
			fixture: "governor_status.json",
			check: func(t *testing.T, got GovernorStatus) {
				t.Helper()
				require.Len(t, got.Chains, 1)
				assert.Equal(t, ChainIDNear, got.Chains[0].Chain)
				require.Len(t, got.Chains[0].Emitters, 1)
				vaas := got.Chains[0].Emitters[0].EnqueuedVAAs
				require.Len(t, vaas, governorTestPageSize)
				assert.True(t, vaas[0].ReleaseTime.IsZero(), "null releaseTime should map to zero")
				assert.Equal(t, uint64(100), vaas[0].NotionalValue)
				assert.Equal(t, uint64(9615), vaas[0].Sequence)
				assert.False(t, vaas[1].ReleaseTime.IsZero())
				assert.Equal(t, time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC), vaas[1].ReleaseTime)
			},
		},
		{
			name:    "live BSON Key/Value enqueued VAAs",
			fixture: "governor_status_bson.json",
			check: func(t *testing.T, got GovernorStatus) {
				t.Helper()
				require.Len(t, got.Chains, 1)
				require.Len(t, got.Chains[0].Emitters, 1)
				vaas := got.Chains[0].Emitters[0].EnqueuedVAAs
				require.Len(t, vaas, 1)
				assert.Equal(t, uint64(552792), vaas[0].NotionalValue)
				assert.Equal(t, uint64(9615), vaas[0].Sequence)
				assert.Equal(t, time.Unix(1790643850, 0).UTC(), vaas[0].ReleaseTime)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newGovernorClient(t, governorFixtureHandler(t, tt.fixture))
			got, err := c.GetGovernorStatus(t.Context(), governorTestGuardianPath)
			require.NoError(t, err)
			assert.Equal(t, GuardianAddress(governorTestGuardianHex), got.GuardianAddress)
			tt.check(t, got)
		})
	}
}

func TestEnqueuedVAAsByChainNotFound(t *testing.T) {
	t.Parallel()

	body := readGovernorFixture(t, "governor_enqueued_vaas_chain.json")
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(body)
	})
	c := newGovernorClient(t, handler)

	_, err := c.EnqueuedVAAsByChain(t.Context(), ChainIDEthereum)
	require.ErrorIs(t, err, ErrNotFound)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, governorTestRequestID, apiErr.RequestID)
}

func TestGovernorConfigsWalksTwoPages(t *testing.T) {
	t.Parallel()

	var fetches atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		n := fetches.Add(1)
		page := r.URL.Query().Get("page")
		var body string
		switch {
		case n == 1 && (page == "" || page == "0"):
			body = `{"data":[{"id":"aa"},{"id":"bb"}]}`
		case n == 2 && page == "1":
			body = `{"data":[{"id":"cc"}]}`
		default:
			body = `{"data":[]}`
		}
		_, _ = w.Write([]byte(body))
	})
	c := newGovernorClient(t, handler)

	var ids []GuardianAddress
	for item, err := range c.GovernorConfigs(t.Context(), GovernorConfigListOptions{
		PageOptions: PageOptions{PageSize: governorTestPageSize},
	}) {
		require.NoError(t, err)
		ids = append(ids, item.GuardianAddress)
	}
	assert.Equal(t, []GuardianAddress{"aa", "bb", "cc"}, ids)
	assert.Equal(t, int32(2), fetches.Load())
}

func TestGovernorVAAsBareArray(t *testing.T) {
	t.Parallel()

	c := newGovernorClient(t, governorFixtureHandler(t, "governor_vaas.json"))
	got, err := c.GovernorVAAs(t.Context())
	require.NoError(t, err)
	require.Len(t, got, governorTestPageSize)
	assert.Equal(t, GovernorVAAStatusIssued, got[0].Status)
	assert.Equal(t, ChainIDBSC, got[0].Chain)
	assert.Equal(t, "874517", got[0].Sequence)
}

func TestEnqueuedVAAsByChain(t *testing.T) {
	t.Parallel()

	c := newGovernorClient(t, governorFixtureHandler(t, "governor_enqueued_vaas_chain_1.json"))
	got, err := c.EnqueuedVAAsByChain(t.Context(), ChainIDSolana)
	require.NoError(t, err)
	require.Len(t, got, governorTestPageSize)
	assert.Equal(t, ChainIDSolana, got[0].Chain)
	assert.Equal(t, uint64(1384028), got[0].Sequence)
	assert.Equal(t, uint64(1168266), got[0].NotionalValue)
	assert.False(t, got[0].ReleaseTime.IsZero())
}

func newGovernorClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(WithBaseURL(srv.URL), WithoutRetry())
	require.NoError(t, err)
	return c
}

func governorFixtureHandler(t *testing.T, name string) http.Handler {
	t.Helper()
	body := readGovernorFixture(t, name)
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
}

func readGovernorFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return body
}
