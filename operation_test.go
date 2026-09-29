package wormholescan

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	operationByIDFixture           = "operation_by_id.json"
	operationGlobalTxOriginFixture = "operation_global_tx_origin_only.json"
	operationByID                  = "5/00000000000000000000000027428dd2d3dd32a4d7f7c497eaaa23130d894911/300859"
	operationGlobalTxOriginOnlyID  = "2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/691371"
	operationByIDAmount            = "140075046480"
	operationByIDSymbol            = "BRZ"
	operationByIDSrcTimestamp      = "2026-09-28T19:48:27Z"
	operationByIDTgtTimestamp      = "2026-09-28T19:48:35Z"
	operationOriginTxHash          = "0xddd90145fe029aa1013540812effee9e1f28782043120747e63659e77dc7ef77"
	searchTxHash                   = "0xabc"
	relayJSON                      = `{"id":"relay-1","relayer":"wormhole","status":"completed","receivedAt":"2026-01-02T03:04:05Z","data":{"fromTxHash":"0xfrom","toTxHash":"0xto","maxAttempts":3,"instructions":{"targetChainId":30,"extraReceiverValue":{"_hex":"0x01","_isBigNumber":true}},"delivery":{"budget":"1","execution":{"status":"ok","transactionHash":"0xto"}}}}`
	operationsPageJSON             = `{"operations":[{"id":"a"},{"id":"b"}]}`
	operationsLastPageJSON         = `{"operations":[{"id":"c"}]}`
	operationsEmptyJSON            = `{"operations":[]}`
	contentTypeJSON                = "application/json"
	listPageSize                   = 2
	searchJSONBody                 = `["0xabc"]`
	listFromRFC3339                = "2024-06-07T08:09:10Z"
	listToRFC3339                  = "2024-06-07T09:09:10Z"
)

func TestGetOperationMapsFixture(t *testing.T) {
	t.Parallel()

	body := readTestdata(t, operationByIDFixture)
	c := newTestClientNoRetry(t, writeJSON(body))
	id, err := ParseVAAID(operationByID)
	require.NoError(t, err)

	got, err := c.GetOperation(t.Context(), id)
	require.NoError(t, err)

	assert.Equal(t, operationByID, got.ID)
	assert.Equal(t, ChainIDPolygon, got.EmitterChain)
	assert.Equal(t, ChainIDPolygon, got.SourceChain.ChainID)
	assert.Equal(t, ChainIDBase, got.TargetChain.ChainID)
	assert.Equal(t, TxStatusConfirmed, got.SourceChain.Status)
	assert.Equal(t, TxStatusCompleted, got.TargetChain.Status)
	assert.Equal(t, time.UTC, got.SourceChain.Timestamp.Location())
	assert.Equal(t, time.UTC, got.TargetChain.Timestamp.Location())
	assert.True(t, got.SourceChain.Timestamp.Equal(parseUTC(t, operationByIDSrcTimestamp)))
	assert.True(t, got.TargetChain.Timestamp.Equal(parseUTC(t, operationByIDTgtTimestamp)))
	assert.Equal(t, operationByIDAmount, got.Content.StandardizedProperties.Amount)
	assert.Equal(t, ChainIDPolygon, got.Content.StandardizedProperties.FromChain)
	assert.Equal(t, ChainIDBase, got.Content.StandardizedProperties.ToChain)
	assert.Contains(t, got.Content.StandardizedProperties.AppIDs, "GENERIC_RELAYER")
	assert.Equal(t, operationByIDSymbol, got.Data.Symbol)
	assert.NotEmpty(t, got.Data.TokenAmount)
}

func TestListOperationsEncodesFilters(t *testing.T) {
	t.Parallel()

	var gotQuery string
	c := newTestClientNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		writeJSONBytes(w, []byte(operationsEmptyJSON))
	})
	from := parseUTC(t, listFromRFC3339)
	to := parseUTC(t, listToRFC3339)

	_, err := c.ListOperations(t.Context(), OperationListOptions{
		PageOptions:    PageOptions{PageSize: listPageSize},
		SourceChains:   []ChainID{ChainIDSolana, ChainIDEthereum},
		TargetChains:   []ChainID{ChainIDBase},
		IncludesChains: []ChainID{ChainIDPolygon},
		From:           from,
		To:             to,
	})
	require.NoError(t, err)
	require.NotEmpty(t, gotQuery)

	query, err := url.ParseQuery(gotQuery)
	require.NoError(t, err)
	assert.Equal(t, "1,2", query.Get("sourceChain"))
	assert.Equal(t, "30", query.Get("targetChain"))
	assert.Equal(t, "5", query.Get("includesChain"))
	assert.Equal(t, listFromRFC3339, query.Get("from"))
	assert.Equal(t, listToRFC3339, query.Get("to"))
}

func TestOperationsWalksTwoPages(t *testing.T) {
	t.Parallel()

	var pages []string
	c := newTestClientNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		if page == "" || page == "0" {
			writeJSONBytes(w, []byte(operationsPageJSON))
			return
		}
		writeJSONBytes(w, []byte(operationsLastPageJSON))
	})

	var ids []string
	for op, err := range c.Operations(t.Context(), OperationListOptions{
		PageOptions: PageOptions{PageSize: listPageSize},
	}) {
		require.NoError(t, err)
		ids = append(ids, op.ID)
	}

	assert.Equal(t, []string{"a", "b", "c"}, ids)
	assert.Len(t, pages, listPageSize)
}

func TestSearchOperationsByTxHashesNotRetried(t *testing.T) {
	t.Parallel()

	var n atomic.Int32
	var gotBody []byte
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		assert.Equal(t, http.MethodPost, r.Method)
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	_, err := c.SearchOperationsByTxHashes(t.Context(), []TxHash{searchTxHash})
	require.Error(t, err)
	assert.Equal(t, int32(1), n.Load())
	assert.JSONEq(t, searchJSONBody, string(gotBody))
}

func TestGetGlobalTransactionOriginOnly(t *testing.T) {
	t.Parallel()

	body := readTestdata(t, operationGlobalTxOriginFixture)
	c := newTestClientNoRetry(t, writeJSON(body))
	id, err := ParseVAAID(operationGlobalTxOriginOnlyID)
	require.NoError(t, err)

	got, err := c.GetGlobalTransaction(t.Context(), id)
	require.NoError(t, err)
	assert.Equal(t, operationGlobalTxOriginOnlyID, got.ID)
	assert.Equal(t, TxHash(operationOriginTxHash), got.Origin.TxHash)
	assert.Equal(t, TxStatusConfirmed, got.Origin.Status)
	assert.Equal(t, DestinationTx{}, got.Destination)
}

func TestGetRelayMapsGeneratedShape(t *testing.T) {
	t.Parallel()

	c := newTestClientNoRetry(t, writeJSON([]byte(relayJSON)))
	id, err := ParseVAAID(operationByID)
	require.NoError(t, err)

	got, err := c.GetRelay(t.Context(), id)
	require.NoError(t, err)
	assert.Equal(t, "relay-1", got.ID)
	assert.Equal(t, RelayStatus("completed"), got.Status)
	assert.Equal(t, TxHash("0xfrom"), got.Data.FromTxHash)
	assert.Equal(t, ChainIDBase, got.Data.Instructions.TargetChainID)
	assert.Equal(t, "0x01", got.Data.Instructions.ExtraReceiverValue.Hex)
	assert.True(t, got.Data.Instructions.ExtraReceiverValue.IsBigNumber)
	assert.Equal(t, TxHash("0xto"), got.Data.Delivery.Execution.TransactionHash)
}

func newTestClient(t *testing.T, handler http.HandlerFunc, opts ...Option) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	all := append([]Option{WithBaseURL(srv.URL)}, opts...)
	c, err := New(all...)
	require.NoError(t, err)
	return c
}

func newTestClientNoRetry(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	return newTestClient(t, handler, WithoutRetry())
}

func writeJSON(body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSONBytes(w, body)
	}
}

func writeJSONBytes(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", contentTypeJSON)
	_, _ = w.Write(body)
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return b
}

func parseUTC(t *testing.T, raw string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, raw)
	require.NoError(t, err)
	return ts.UTC()
}
