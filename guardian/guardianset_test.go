package guardian

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wormhole-foundation/wormholescan-go"
)

const (
	recordedGuardianSetIndex = 7
	recordedGuardianSetSize  = 19
)

func TestCurrentGuardianSet(t *testing.T) {
	t.Parallel()

	c := newJSONClient(t, http.StatusOK, loadFixture(t, "guardianset_current.json"))
	got, err := c.CurrentGuardianSet(t.Context())
	require.NoError(t, err)
	assert.Equal(t, uint32(recordedGuardianSetIndex), got.Index)
	require.Len(t, got.Addresses, recordedGuardianSetSize)
	assert.Equal(
		t,
		wormholescan.GuardianAddress("0x5893B5A76c3f739645648885bDCcC06cd70a3Cd3"),
		got.Addresses[0],
	)
}
