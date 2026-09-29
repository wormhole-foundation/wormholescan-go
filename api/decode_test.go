package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/wormhole-foundation/wormholescan-go/api"
)

// TestDecodeRecordedResponses feeds recorded live responses through the
// generated Parse*Response functions and re-encodes the result. Any key the
// server sent that does not survive the round trip, or any leaf whose value
// changes, means the generated type (and so api/spec/patch.jq) disagrees with
// what the server really sends.
func TestDecodeRecordedResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture string
		parse   func(*http.Response) (any, error)
	}{
		{
			name:    "vaas",
			fixture: "api_v1_vaas_pageSize_1.json",
			parse:   json200(api.ParseFindAllVaasResponse, func(r *api.FindAllVaasResponse) any { return r.JSON200 }),
		},
		{
			name:    "operations",
			fixture: "api_v1_operations_pageSize_1.json",
			parse: json200(
				api.ParseGetOperationsResponse,
				func(r *api.GetOperationsResponse) any { return r.JSON200 },
			),
		},
		{
			name:    "observations",
			fixture: "api_v1_observations_pageSize_1.json",
			parse: json200(
				api.ParseFindObservationsResponse,
				func(r *api.FindObservationsResponse) any { return r.JSON200 },
			),
		},
		{
			name:    "delegate observations",
			fixture: "api_v1_observations_delegate_51_8192a21f_18242.json",
			parse: json200(
				api.ParseFindDelegateObservationsBySequenceResponse,
				func(r *api.FindDelegateObservationsBySequenceResponse) any { return r.JSON200 },
			),
		},
		{
			name:    "governor status",
			fixture: "api_v1_governor_status_pageSize_1.json",
			parse: json200(
				api.ParseGovernorStatusResponse,
				func(r *api.GovernorStatusResponse) any { return r.JSON200 },
			),
		},
		{
			name:    "heartbeats",
			fixture: "v1_heartbeats.json",
			parse: json200(
				api.ParseGuardiansHearbeatsResponse,
				func(r *api.GuardiansHearbeatsResponse) any { return r.JSON200 },
			),
		},
		{
			name:    "guardian set",
			fixture: "v1_guardianset_current.json",
			parse:   json200(api.ParseGuardianSetResponse, func(r *api.GuardianSetResponse) any { return r.JSON200 }),
		},
		{
			name:    "last transactions",
			fixture: "api_v1_last-txs_timeSpan_1d_sampleRate_1h.json",
			parse: json200(
				api.ParseGetLastTransactionsResponse,
				func(r *api.GetLastTransactionsResponse) any { return r.JSON200 },
			),
		},
		{
			name:    "scorecards",
			fixture: "api_v1_scorecards.json",
			parse: json200(
				api.ParseGetScorecardsResponse,
				func(r *api.GetScorecardsResponse) any { return r.JSON200 },
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body, err := os.ReadFile("testdata/" + tt.fixture)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := tt.parse(&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader(body)),
			})
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if reflect.ValueOf(decoded).IsNil() {
				t.Fatal("no JSON200 body decoded")
			}
			for _, d := range roundTripDiffs(t, body, decoded) {
				t.Error(d)
			}
		})
	}
}

// json200 adapts a generated Parse*Response function to the table: it fails
// on a parse error instead of dereferencing a nil response, and returns the
// decoded 200 body.
func json200[R any](parse func(*http.Response) (*R, error), body func(*R) any) func(*http.Response) (any, error) {
	return func(r *http.Response) (any, error) {
		resp, err := parse(r)
		if err != nil {
			return nil, err
		}
		return body(resp), nil
	}
}

// roundTripDiffs re-encodes decoded and compares it with the original JSON.
// It reports keys present in the original but missing after re-encoding and
// leaf values that changed. Only the first element of each array is compared;
// null originals are skipped because omitted and null are indistinguishable
// on the wire.
func roundTripDiffs(t *testing.T, original []byte, decoded any) []string {
	t.Helper()
	var want, got any
	if err := json.Unmarshal(original, &want); err != nil {
		t.Fatal(err)
	}
	re, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(re, &got); err != nil {
		t.Fatal(err)
	}
	return diffValues("$", want, got)
}

// diffValues dispatches on the original value's JSON kind.
func diffValues(path string, want, got any) []string {
	switch w := want.(type) {
	case map[string]any:
		return diffObjects(path, w, got)
	case []any:
		return diffArrays(path, w, got)
	default:
		if !reflect.DeepEqual(want, got) {
			return []string{path + ": value changed"}
		}
		return nil
	}
}

// diffObjects reports keys of want that are missing from got, then recurses.
func diffObjects(path string, want map[string]any, got any) []string {
	g, ok := got.(map[string]any)
	if !ok {
		return []string{path + ": object became " + reflect.TypeOf(got).String()}
	}
	var diffs []string
	for k, v := range want {
		if v == nil {
			continue
		}
		w, ok := g[k]
		if !ok {
			diffs = append(diffs, path+"."+k+": missing after decode")
			continue
		}
		diffs = append(diffs, diffValues(path+"."+k, v, w)...)
	}
	return diffs
}

// diffArrays checks length and recurses into the first element.
func diffArrays(path string, want []any, got any) []string {
	g, ok := got.([]any)
	if !ok || len(g) != len(want) {
		return []string{path + ": array length or shape changed"}
	}
	if len(want) == 0 {
		return nil
	}
	return diffValues(path+"[0]", want[0], g[0])
}
