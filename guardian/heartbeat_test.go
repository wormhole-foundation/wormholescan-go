package guardian

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wormhole-foundation/wormholescan-go"
)

const (
	recordedHeartbeatCount     = 19
	recordedHeartbeatTimestamp = int64(1790641873495327000)
	googleCloudGuardian        = wormholescan.GuardianAddress("0x7899ceab1dc961dae9defdb7a4f521269a5448fc")
)

func TestHeartbeats(t *testing.T) {
	t.Parallel()

	c := newJSONClient(t, http.StatusOK, loadFixture(t, "heartbeats.json"))
	got, err := c.Heartbeats(t.Context())
	require.NoError(t, err)
	require.Len(t, got, recordedHeartbeatCount)

	var google *Heartbeat
	for i := range got {
		if got[i].VerifiedGuardianAddress == googleCloudGuardian {
			google = &got[i]
			break
		}
	}
	require.NotNil(t, google)
	assert.Equal(t, "Google Cloud", google.Raw.NodeName)
	assert.Equal(t, time.Unix(0, recordedHeartbeatTimestamp).UTC(), google.Raw.Timestamp)
	require.NotEmpty(t, google.Raw.Networks)
	assert.Equal(t, wormholescan.ChainIDSolana, google.Raw.Networks[0].Chain)
	assert.Equal(t, "451480770", google.Raw.Networks[0].Height)
}

func TestHeartbeatsUnparsableTimestamp(t *testing.T) {
	t.Parallel()

	body := []byte(`{"entries":[{"rawHeartbeat":{"timestamp":"not-a-number"}}]}`)
	c := newJSONClient(t, http.StatusOK, body)
	_, err := c.Heartbeats(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timestamp")
}
