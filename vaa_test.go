package wormholescan

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fixtureVAAID     = "1/19671a08a9cef6f3a04314ed478fc332a4966f41ad3e6fea76933dede9c6cdfe/755119"
	fixtureEmitter   = "19671a08a9cef6f3a04314ed478fc332a4966f41ad3e6fea76933dede9c6cdfe"
	fixtureVAARawLen = 1127
	fixturePageSize  = 2
)

func TestGetVAA(t *testing.T) {
	t.Parallel()

	id, err := ParseVAAID(fixtureVAAID)
	require.NoError(t, err)

	tests := []struct {
		name      string
		status    int
		fixture   string
		opts      VAAGetOptions
		assertErr func(*testing.T, error)
		assertVAA func(*testing.T, VAA)
	}{
		{
			name:    "returns typed fields matching the fixture",
			status:  http.StatusOK,
			fixture: "vaa_by_id.json",
			opts:    VAAGetOptions{ParsedPayload: true},
			assertErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
			assertVAA: func(t *testing.T, got VAA) {
				assert.Equal(t, id.String(), got.ID.String())
				assert.Equal(t, fixtureVAARawLen, len(got.Raw))
				assert.False(t, got.Timestamp.IsZero())
				assert.Equal(t, time.UTC, got.Timestamp.Location())
				assert.Equal(t, Digest("7633abba1546352893c55ede120a8bfb6c8bf6ec8c30f86738b6ba8ff91f4933"), got.Digest)
				assert.True(t, got.IsSolanaShim)
				assert.NotEmpty(t, got.Payload)
			},
		},
		{
			name:    "returns ErrNotFound on 404",
			status:  http.StatusNotFound,
			fixture: "vaa_by_id_404.json",
			assertErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrNotFound)
				var apiErr *APIError
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, "8eaa486e-3dc0-4599-843d-9e36273ae91b", apiErr.RequestID)
			},
			assertVAA: func(t *testing.T, got VAA) {
				assert.Equal(t, VAA{}, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				writeFixtureStatus(t, w, tt.status, tt.fixture)
			})
			got, err := c.GetVAA(t.Context(), id, tt.opts)
			tt.assertErr(t, err)
			tt.assertVAA(t, got)
		})
	}
}

func TestVAAsIteratorWalksTwoPages(t *testing.T) {
	t.Parallel()

	var pages []string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		pages = append(pages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "", "0":
			writeFixture(t, w, "vaa_list.json")
		default:
			writeFixture(t, w, "vaa_list_page1.json")
		}
	})

	var got []string
	for vaa, err := range c.VAAs(t.Context(), VAAListOptions{
		PageOptions: PageOptions{PageSize: fixturePageSize},
	}) {
		require.NoError(t, err)
		got = append(got, vaa.ID.String())
	}

	assert.Equal(t, []string{"", "1"}, pages)
	assert.Equal(t, 3, len(got))
	assert.Equal(t, "1/"+fixtureEmitter+"/755601", got[0])
	assert.Equal(t, "1/"+fixtureEmitter+"/755600", got[1])
}

func TestListVAAsByEmitterPathAndQuery(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotQuery url.Values
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		writeFixture(t, w, "vaa_list.json")
	})

	page, err := c.ListVAAsByEmitter(
		t.Context(),
		ChainIDSolana,
		EmitterAddress(fixtureEmitter),
		VAAListOptions{PageOptions: PageOptions{PageSize: fixturePageSize, Sort: SortAsc}},
	)
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/vaas/1/"+fixtureEmitter, gotPath)
	assert.Equal(t, "2", gotQuery.Get("pageSize"))
	assert.Equal(t, "ASC", gotQuery.Get("sortOrder"))
	assert.Equal(t, 2, len(page.Items))
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(WithBaseURL(srv.URL), WithoutRetry())
	require.NoError(t, err)
	return c
}

func writeFixture(t *testing.T, w http.ResponseWriter, name string) {
	t.Helper()
	writeFixtureStatus(t, w, http.StatusOK, name)
}

func writeFixtureStatus(t *testing.T, w http.ResponseWriter, status int, name string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(testdataBytes(t, name))
	require.NoError(t, err)
}
func testdataBytes(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return body
}
