package guardian

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wormhole-foundation/wormholescan-go"
)

func TestGetSignedVAA(t *testing.T) {
	t.Parallel()

	body := loadFixture(t, "signed_vaa.json")
	c := newJSONClient(t, http.StatusOK, body)
	got, err := c.GetSignedVAA(t.Context(), sampleVAAID())
	require.NoError(t, err)

	var payload struct {
		VaaBytes string `json:"vaaBytes"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	want, err := base64.StdEncoding.DecodeString(payload.VaaBytes)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetSignedVAANotFound(t *testing.T) {
	t.Parallel()

	c := newJSONClient(t, http.StatusNotFound, loadFixture(t, "signed_vaa_404.json"))
	_, err := c.GetSignedVAA(t.Context(), sampleVAAID())
	require.ErrorIs(t, err, wormholescan.ErrNotFound)

	var apiErr *wormholescan.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "996a486e-3dc0-4599-843d-9e36273ae91b", apiErr.RequestID)
}

func TestGetSignedBatchVAANotFound(t *testing.T) {
	t.Parallel()

	c := newJSONClient(t, http.StatusNotFound, loadFixture(t, "signed_batch_vaa_404.json"))
	_, err := c.GetSignedBatchVAA(t.Context(), sampleVAAID())
	require.ErrorIs(t, err, wormholescan.ErrNotFound)
}

func sampleVAAID() wormholescan.VAAID {
	return wormholescan.VAAID{
		Chain:    wormholescan.ChainIDSolana,
		Emitter:  "19671a08a9cef6f3a04314ed478fc332a4966f41ad3e6fea76933dede9c6cdfe",
		Sequence: 755119,
	}
}
