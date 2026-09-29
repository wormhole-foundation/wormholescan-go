package wormholescan

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

const (
	errPrefix         = "wormholescan: "
	maxErrorBodyBytes = 64 * 1024
)

// ErrNotFound is returned when the API responds with 404.
var ErrNotFound = errors.New("wormholescan: not found")

// ErrRateLimited is returned when the API responds with 429.
var ErrRateLimited = errors.New("wormholescan: rate limited")

// ErrBadRequest is returned when the API responds with 400.
var ErrBadRequest = errors.New("wormholescan: bad request")

// APIError is a non-2xx Wormholescan API response.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Code is the API error code from the JSON body, or 0 when absent.
	Code int
	// Message is the API error message from the JSON body, or empty when absent.
	Message string
	// RequestID is the request id from the JSON details, or empty when absent.
	RequestID string
	// Body is the raw response body, truncated to 64 KiB.
	Body []byte
}

// Error returns a summary of the API error.
func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	out := errPrefix + strconv.Itoa(e.StatusCode)
	if msg != "" {
		out += " " + msg
	}
	if e.RequestID != "" {
		out += " (request_id " + e.RequestID + ")"
	}
	return out
}

// Is reports whether target is the sentinel that matches e's status code.
func (e *APIError) Is(target error) bool {
	switch e.StatusCode {
	case http.StatusNotFound:
		return target == ErrNotFound
	case http.StatusTooManyRequests:
		return target == ErrRateLimited
	case http.StatusBadRequest:
		return target == ErrBadRequest
	default:
		return false
	}
}

type apiErrorBody struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Details []apiErrorDetail `json:"details"`
}

type apiErrorDetail struct {
	RequestID string `json:"request_id"`
}

// checkResponse returns nil for 2xx responses and an [APIError] otherwise.
func checkResponse(rsp *http.Response, body []byte) error {
	if rsp.StatusCode >= http.StatusOK && rsp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	apiErr := &APIError{
		StatusCode: rsp.StatusCode,
		Body:       clipBody(body),
	}
	fillAPIErrorFromBody(apiErr, body)
	return apiErr
}

// fillAPIErrorFromBody copies code, message, and request id out of a JSON body.
func fillAPIErrorFromBody(apiErr *APIError, body []byte) {
	if !json.Valid(body) {
		return
	}
	var parsed apiErrorBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return
	}
	apiErr.Code = parsed.Code
	apiErr.Message = parsed.Message
	if len(parsed.Details) > 0 {
		apiErr.RequestID = parsed.Details[0].RequestID
	}
}

// clipBody returns body truncated to maxErrorBodyBytes.
func clipBody(body []byte) []byte {
	if len(body) <= maxErrorBodyBytes {
		return body
	}
	return body[:maxErrorBodyBytes]
}
