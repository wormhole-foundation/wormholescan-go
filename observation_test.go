package wormholescan

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fixtureObservationID   = "1/19671a08a9cef6f3a04314ed478fc332a4966f41ad3e6fea76933dede9c6cdfe/755601/5893b5a76c3f739645648885bdccc06cd70a3cd3/d48b0791aadbe21c47b7c1415282de05708027591be2a1f0844f779492e6093e"
	fixtureObservationHash = "d48b0791aadbe21c47b7c1415282de05708027591be2a1f0844f779492e6093e"
	fixtureGuardian        = GuardianAddress("0x5893b5a76c3f739645648885bdccc06cd70a3cd3")
)

func TestListObservationsBareArray(t *testing.T) {
	t.Parallel()

	c := newTestClientNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/observations", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("pageSize"))
		writeFixture(t, w, "observation_list.json")
	})

	page, err := c.ListObservations(t.Context(), ObservationListOptions{
		PageOptions: PageOptions{PageSize: fixturePageSize},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(page.Items))
	assert.Equal(t, fixturePageSize, page.PageSize)
	assert.False(t, page.Last())
	assert.Equal(t, fixtureObservationID, page.Items[0].ID)
	assert.NotEmpty(t, page.Items[0].Hash)
	assert.NotEmpty(t, page.Items[0].TxHash)
	assert.Equal(t, fixtureGuardian, page.Items[0].GuardianAddress)
}

func TestGetObservationPathUsesHexHash(t *testing.T) {
	t.Parallel()

	id, err := ParseVAAID("1/" + fixtureEmitter + "/755601")
	require.NoError(t, err)

	var gotPath string
	c := newTestClientNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		writeFixture(t, w, "observation_by_id.json")
	})

	got, err := c.GetObservation(t.Context(), id, fixtureGuardian, fixtureObservationHash)
	require.NoError(t, err)
	assert.Equal(t, fixtureObservationID, got.ID)
	assert.Equal(
		t,
		"/api/v1/observations/1/"+fixtureEmitter+"/755601/"+string(fixtureGuardian)+"/"+fixtureObservationHash,
		gotPath,
	)
}

func TestGetDelegateObservationPath(t *testing.T) {
	t.Parallel()

	id, err := ParseVAAID(fixtureVAAID)
	require.NoError(t, err)

	var gotPath string
	c := newTestClientNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write([]byte(`[{
			"id":"delegate-1",
			"emitterChain":1,
			"emitterAddr":"` + fixtureEmitter + `",
			"sequence":"755119",
			"delegatedGuardianAddr":"` + string(fixtureGuardian) + `"
		}]`))
		require.NoError(t, writeErr)
	})

	got, err := c.GetDelegateObservation(t.Context(), id, fixtureGuardian)
	require.NoError(t, err)
	assert.Equal(t, "delegate-1", got.ID)
	assert.Equal(
		t,
		"/api/v1/observations/delegate/1/"+fixtureEmitter+"/755119/"+string(fixtureGuardian),
		gotPath,
	)
}
