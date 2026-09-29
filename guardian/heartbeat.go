package guardian

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormholescan-go"
	"github.com/wormhole-foundation/wormholescan-go/api"
)

const opHeartbeats = "heartbeats"

// Heartbeat is one guardian's latest heartbeat from /v1/heartbeats.
type Heartbeat struct {
	// VerifiedGuardianAddress is the guardian address verified by the API.
	VerifiedGuardianAddress wormholescan.GuardianAddress
	// P2PNodeAddress is the libp2p multiaddr reported for the guardian.
	P2PNodeAddress string
	// Raw is the heartbeat payload as submitted by the guardian.
	Raw HeartbeatBody
}

// HeartbeatBody is the raw heartbeat payload inside a [Heartbeat].
type HeartbeatBody struct {
	// NodeName is the guardian node name.
	NodeName string
	// Counter is the heartbeat counter as a decimal string.
	Counter string
	// Timestamp is the heartbeat time, converted from a unix-nanosecond string.
	Timestamp time.Time
	// Networks is the per-chain height snapshot.
	Networks []HeartbeatNetwork
	// Version is the guardian software version.
	Version string
	// GuardianAddress is the address the guardian claims.
	GuardianAddress wormholescan.GuardianAddress
	// BootTimestamp is when the guardian process started, converted from a
	// unix-nanosecond string.
	BootTimestamp time.Time
	// Features is the feature list the guardian advertised.
	Features []string
	// P2PNodeID is the libp2p peer id. Empty when the server omits it.
	P2PNodeID []byte
}

// HeartbeatNetwork is one chain's height snapshot inside a [HeartbeatBody].
type HeartbeatNetwork struct {
	// Chain is the Wormhole chain id (the server's "id" field).
	Chain wormholescan.ChainID
	// Height is the latest block height as a decimal string.
	Height string
	// SafeHeight is the safe-head height as a decimal string. Empty when omitted.
	SafeHeight string
	// FinalizedHeight is the finalized-head height as a decimal string. Empty when omitted.
	FinalizedHeight string
	// ContractAddress is the core contract address on this chain.
	ContractAddress string
	// ErrorCount is the observation error count as a decimal string.
	ErrorCount string
}

// Heartbeats returns the latest heartbeat for each guardian.
func (c *Client) Heartbeats(ctx context.Context) ([]Heartbeat, error) {
	rsp, err := c.wc.API().GuardiansHearbeatsWithResponse(ctx)
	if err != nil {
		return nil, wrapCall(opHeartbeats, err)
	}
	if err = wormholescan.CheckResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return nil, err
	}
	if rsp.JSON200 == nil {
		return nil, emptyBody(opHeartbeats)
	}
	out, err := heartbeatsFromAPI(rsp.JSON200)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// heartbeatsFromAPI maps a generated heartbeats envelope to domain values.
func heartbeatsFromAPI(resp *api.HeartbeatsHeartbeatsResponse) ([]Heartbeat, error) {
	if resp.Entries == nil {
		return nil, nil
	}
	out := make([]Heartbeat, 0, len(*resp.Entries))
	for i := range *resp.Entries {
		hb, err := heartbeatFromAPI((*resp.Entries)[i])
		if err != nil {
			return nil, err
		}
		out = append(out, hb)
	}
	return out, nil
}

// heartbeatFromAPI maps one generated heartbeat entry to a [Heartbeat].
func heartbeatFromAPI(entry api.HeartbeatsHeartbeatResponse) (Heartbeat, error) {
	raw, err := heartbeatBodyFromAPI(entry.RawHeartbeat)
	if err != nil {
		return Heartbeat{}, err
	}
	return Heartbeat{
		VerifiedGuardianAddress: wormholescan.GuardianAddress(deref(entry.VerifiedGuardianAddr)),
		P2PNodeAddress:          deref(entry.P2pNodeAddr),
		Raw:                     raw,
	}, nil
}

// heartbeatBodyFromAPI maps a generated raw heartbeat to a [HeartbeatBody].
func heartbeatBodyFromAPI(raw *api.HeartbeatsRawHeartbeat) (HeartbeatBody, error) {
	if raw == nil {
		return HeartbeatBody{}, nil
	}
	ts, err := parseUnixNanos("timestamp", deref(raw.Timestamp))
	if err != nil {
		return HeartbeatBody{}, err
	}
	boot, err := parseUnixNanos("bootTimestamp", deref(raw.BootTimestamp))
	if err != nil {
		return HeartbeatBody{}, err
	}
	return HeartbeatBody{
		NodeName:        deref(raw.NodeName),
		Counter:         deref(raw.Counter),
		Timestamp:       ts,
		Networks:        heartbeatNetworksFromAPI(raw.Networks),
		Version:         deref(raw.Version),
		GuardianAddress: wormholescan.GuardianAddress(deref(raw.GuardianAddr)),
		BootTimestamp:   boot,
		Features:        deref(raw.Features),
		P2PNodeID:       deref(raw.P2pNodeId),
	}, nil
}

// heartbeatNetworksFromAPI maps generated network snapshots to [HeartbeatNetwork] values.
func heartbeatNetworksFromAPI(networks *[]api.HeartbeatsHeartbeatNetworkResponse) []HeartbeatNetwork {
	if networks == nil {
		return nil
	}
	out := make([]HeartbeatNetwork, 0, len(*networks))
	for _, n := range *networks {
		out = append(out, HeartbeatNetwork{
			Chain:           chainIDFromInt(n.Id),
			Height:          deref(n.Height),
			SafeHeight:      deref(n.SafeHeight),
			FinalizedHeight: deref(n.FinalizedHeight),
			ContractAddress: deref(n.ContractAddress),
			ErrorCount:      deref(n.ErrorCount),
		})
	}
	return out
}

// parseUnixNanos parses a decimal unix-nanosecond timestamp.
func parseUnixNanos(field, raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	ns, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("%sparse %s %q: %w", errPrefix, field, raw, err)
	}
	return time.Unix(0, ns).UTC(), nil
}
