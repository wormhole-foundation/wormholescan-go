package guardian

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormholescan-go"
	"github.com/wormhole-foundation/wormholescan-go/api"
)

const (
	opAvailableNotional = "available notional"
	opEnqueuedVAAs      = "enqueued VAAs"
	opIsVAAEnqueued     = "is VAA enqueued"
	opTokenList         = "token list"
)

// GovernorAvailableNotional is one chain's available notional from
// /v1/governor/available_notional_by_chain.
type GovernorAvailableNotional struct {
	// Chain is the Wormhole chain id.
	Chain wormholescan.ChainID
	// RemainingAvailableNotional is the remaining capacity as a decimal string.
	RemainingAvailableNotional string
	// NotionalLimit is the chain notional limit as a decimal string.
	NotionalLimit string
	// BigTransactionSize is the big-transaction threshold as a decimal string.
	BigTransactionSize string
}

// GovernorEnqueuedVAA is one VAA waiting in the governor from
// /v1/governor/enqueued_vaas.
type GovernorEnqueuedVAA struct {
	// EmitterAddress is the emitter address as the server sent it.
	EmitterAddress string
	// EmitterChain is the emitter chain id.
	EmitterChain wormholescan.ChainID
	// NotionalValue is the USD notional as a decimal string.
	NotionalValue string
	// ReleaseTime is the unix-second release time. Zero when the server has no value.
	ReleaseTime time.Time
	// Sequence is the VAA sequence number.
	Sequence uint64
	// TxHash is the source transaction hash as the server sent it.
	TxHash wormholescan.TxHash
}

// GovernorTokenList is one priced origin token from /v1/governor/token_list.
type GovernorTokenList struct {
	// OriginAddress is the token address on the origin chain.
	OriginAddress string
	// OriginChain is the origin chain id.
	OriginChain wormholescan.ChainID
	// Price is the USD price used by the governor.
	Price float32
}

// AvailableNotionalByChain returns remaining governor capacity per chain.
func (c *Client) AvailableNotionalByChain(ctx context.Context) ([]GovernorAvailableNotional, error) {
	rsp, err := c.wc.API().GovernorAvailableNotionalByChainWithResponse(ctx)
	if err != nil {
		return nil, wrapCall(opAvailableNotional, err)
	}
	if err = wormholescan.CheckResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return nil, err
	}
	if rsp.JSON200 == nil {
		return nil, emptyBody(opAvailableNotional)
	}
	return availableNotionalFromAPI(rsp.JSON200), nil
}

// EnqueuedVAAs returns VAAs currently held by the governor.
func (c *Client) EnqueuedVAAs(ctx context.Context) ([]GovernorEnqueuedVAA, error) {
	rsp, err := c.wc.API().GuardiansEnqueuedVaasWithResponse(ctx)
	if err != nil {
		return nil, wrapCall(opEnqueuedVAAs, err)
	}
	if err = wormholescan.CheckResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return nil, err
	}
	if rsp.JSON200 == nil {
		return nil, emptyBody(opEnqueuedVAAs)
	}
	return enqueuedVAAsFromAPI(rsp.JSON200), nil
}

// IsVAAEnqueued reports whether id is currently held by the governor.
func (c *Client) IsVAAEnqueued(ctx context.Context, id wormholescan.VAAID) (bool, error) {
	chainID, emitter, seq := vaaPathParams(id)
	rsp, err := c.wc.API().GuardiansIsVaaEnqueuedWithResponse(ctx, chainID, emitter, seq)
	if err != nil {
		return false, wrapCall(opIsVAAEnqueued, err)
	}
	if err = wormholescan.CheckResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return false, err
	}
	if rsp.JSON200 == nil || rsp.JSON200.IsEnqueued == nil {
		return false, emptyBody(opIsVAAEnqueued)
	}
	return *rsp.JSON200.IsEnqueued, nil
}

// GovernorTokenList returns the governor's priced origin-token list.
func (c *Client) GovernorTokenList(ctx context.Context) ([]GovernorTokenList, error) {
	rsp, err := c.wc.API().GuardiansTokenListWithResponse(ctx)
	if err != nil {
		return nil, wrapCall(opTokenList, err)
	}
	if err = wormholescan.CheckResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return nil, err
	}
	if rsp.JSON200 == nil {
		return nil, emptyBody(opTokenList)
	}
	return tokenListFromAPI(rsp.JSON200), nil
}

// availableNotionalFromAPI maps a generated available-notional envelope.
func availableNotionalFromAPI(resp *api.GovernorAvailableNotionalResponse) []GovernorAvailableNotional {
	if resp.Entries == nil {
		return nil
	}
	out := make([]GovernorAvailableNotional, 0, len(*resp.Entries))
	for _, item := range *resp.Entries {
		out = append(out, GovernorAvailableNotional{
			Chain:                      chainIDFromVAA(item.ChainId),
			RemainingAvailableNotional: deref(item.RemainingAvailableNotional),
			NotionalLimit:              deref(item.NotionalLimit),
			BigTransactionSize:         deref(item.BigTransactionSize),
		})
	}
	return out
}

// enqueuedVAAsFromAPI maps a generated enqueued-VAA envelope.
func enqueuedVAAsFromAPI(resp *api.GovernorEnqueuedVaaResponse) []GovernorEnqueuedVAA {
	if resp.Entries == nil {
		return nil
	}
	out := make([]GovernorEnqueuedVAA, 0, len(*resp.Entries))
	for _, item := range *resp.Entries {
		out = append(out, GovernorEnqueuedVAA{
			EmitterAddress: deref(item.EmitterAddress),
			EmitterChain:   chainIDFromVAA(item.EmitterChain),
			NotionalValue:  deref(item.NotionalValue),
			ReleaseTime:    timeFromUnixSeconds(item.ReleaseTime),
			Sequence:       deref(item.Sequence),
			TxHash:         wormholescan.TxHash(deref(item.TxHash)),
		})
	}
	return out
}

// tokenListFromAPI maps a generated token-list envelope.
func tokenListFromAPI(resp *api.GuardianTokenListResponse) []GovernorTokenList {
	if resp.Entries == nil {
		return nil
	}
	out := make([]GovernorTokenList, 0, len(*resp.Entries))
	for _, item := range *resp.Entries {
		out = append(out, GovernorTokenList{
			OriginAddress: deref(item.OriginAddress),
			OriginChain:   chainIDFromVAA(item.OriginChainId),
			Price:         deref(item.Price),
		})
	}
	return out
}

// timeFromUnixSeconds converts a generated unix-second pointer to UTC time.
func timeFromUnixSeconds(v *int) time.Time {
	if v == nil {
		return time.Time{}
	}
	return time.Unix(int64(*v), 0).UTC()
}
