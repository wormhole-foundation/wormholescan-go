package wormholescan

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckResponseNotFoundJSON(t *testing.T) {
	t.Parallel()

	const requestID = "req-123"
	body := []byte(`{"code":5,"message":"NOT FOUND","details":[{"request_id":"` + requestID + `"}]}`)
	rsp := &http.Response{StatusCode: http.StatusNotFound}
	err := checkResponse(rsp, body)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, requestID, apiErr.RequestID)
	assert.Equal(t, "NOT FOUND", apiErr.Message)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestCheckResponseNotFoundNonJSON(t *testing.T) {
	t.Parallel()

	body := []byte("plain text not found")
	rsp := &http.Response{StatusCode: http.StatusNotFound}
	err := checkResponse(rsp, body)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Empty(t, apiErr.RequestID)
	assert.Equal(t, body, apiErr.Body)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestCheckResponseOK(t *testing.T) {
	t.Parallel()

	err := checkResponse(&http.Response{StatusCode: http.StatusOK}, []byte(`{"ok":true}`))
	require.NoError(t, err)
}

func TestAPIErrorIsSentinels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		target error
	}{
		{name: "not found", status: http.StatusNotFound, target: ErrNotFound},
		{name: "rate limited", status: http.StatusTooManyRequests, target: ErrRateLimited},
		{name: "bad request", status: http.StatusBadRequest, target: ErrBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := &APIError{StatusCode: tt.status}
			require.ErrorIs(t, err, tt.target)
		})
	}
}
