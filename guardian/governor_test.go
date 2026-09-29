package guardian

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wormhole-foundation/wormholescan-go"
)

func TestAvailableNotionalByChain(t *testing.T) {
	t.Parallel()

	c := newJSONClient(t, http.StatusOK, loadFixture(t, "available_notional_by_chain.json"))
	got, err := c.AvailableNotionalByChain(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, wormholescan.ChainIDSolana, got[0].Chain)
	assert.Equal(t, "13164057", got[0].RemainingAvailableNotional)
	assert.Equal(t, "20000000", got[0].NotionalLimit)
	assert.Equal(t, "1000000", got[0].BigTransactionSize)
}

func TestEnqueuedVAAs(t *testing.T) {
	t.Parallel()

	c := newJSONClient(t, http.StatusOK, loadFixture(t, "enqueued_vaas.json"))
	got, err := c.EnqueuedVAAs(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, wormholescan.ChainIDSolana, got[0].EmitterChain)
	assert.Equal(t, "0xec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5", got[0].EmitterAddress)
	assert.Equal(t, uint64(1384028), got[0].Sequence)
	assert.Equal(t, time.Unix(1785089717, 0).UTC(), got[0].ReleaseTime)
	assert.Equal(t, "1263008", got[0].NotionalValue)
}

func TestIsVAAEnqueued(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture string
		want    bool
	}{
		{name: "true", fixture: "is_vaa_enqueued_true.json", want: true},
		{name: "false", fixture: "is_vaa_enqueued_false.json", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newJSONClient(t, http.StatusOK, loadFixture(t, tt.fixture))
			got, err := c.IsVAAEnqueued(t.Context(), sampleVAAID())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGovernorTokenList(t *testing.T) {
	t.Parallel()

	c := newJSONClient(t, http.StatusOK, loadFixture(t, "token_list.json"))
	got, err := c.GovernorTokenList(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, wormholescan.ChainIDSolana, got[0].OriginChain)
	assert.Equal(t, "0x00587a642600d6f2bd3bbfdf116dbd613c45cfa659b55a1fbc7cbc28e6092699", got[0].OriginAddress)
	assert.InDelta(t, 8.26e-06, float64(got[0].Price), 1e-12)
}
