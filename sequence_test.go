package wormholescan

import (
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// governanceVAAID is the Core upgrade to guardian set 5. Its sequence,
// like many governance sequences, is above math.MaxInt64.
const governanceVAAID = "1/0000000000000000000000000000000000000000000000000000000000000004/18220114619187442754"

// TestSequenceAboveMaxInt64 sends a sequence above 2^63-1 in the path as
// its decimal value and decodes it from the JSON number the server
// returns. The fixtures are live responses for governanceVAAID.
func TestSequenceAboveMaxInt64(t *testing.T) {
	t.Parallel()

	id, err := ParseVAAID(governanceVAAID)
	require.NoError(t, err)
	require.Greater(t, id.Sequence, uint64(math.MaxInt64))

	t.Run("GetVAA", func(t *testing.T) {
		t.Parallel()
		var path string
		c := newTestClientNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			writeFixture(t, w, "vaa_by_id_governance.json")
		})
		got, err := c.GetVAA(t.Context(), id, VAAGetOptions{})
		require.NoError(t, err)
		assert.Equal(t, "/api/v1/vaas/"+governanceVAAID, path)
		assert.Equal(t, id, got.ID)
	})

	t.Run("ListObservationsByVAA", func(t *testing.T) {
		t.Parallel()
		var path string
		c := newTestClientNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			writeFixture(t, w, "observation_list_governance.json")
		})
		page, err := c.ListObservationsByVAA(t.Context(), id, ObservationListOptions{})
		require.NoError(t, err)
		assert.Equal(t, "/api/v1/observations/"+governanceVAAID, path)
		require.NotEmpty(t, page.Items)
		for _, obs := range page.Items {
			assert.Equal(t, id.Sequence, obs.Sequence)
		}
	})
}
