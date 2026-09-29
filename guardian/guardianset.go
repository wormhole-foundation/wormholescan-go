package guardian

import (
	"context"

	"github.com/wormhole-foundation/wormholescan-go"
	"github.com/wormhole-foundation/wormholescan-go/api"
)

const opGuardianSet = "guardian set"

// GuardianSet is the currently active guardian set from /v1/guardianset/current.
type GuardianSet struct { //nolint:revive // name matches the /v1/guardianset payload
	// Index is the guardian set index.
	Index uint32
	// Addresses is the ordered list of guardian addresses.
	Addresses []wormholescan.GuardianAddress
}

// CurrentGuardianSet returns the currently active guardian set.
func (c *Client) CurrentGuardianSet(ctx context.Context) (GuardianSet, error) {
	rsp, err := c.wc.API().GuardianSetWithResponse(ctx)
	if err != nil {
		return GuardianSet{}, wrapCall(opGuardianSet, err)
	}
	if err = wormholescan.CheckResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return GuardianSet{}, err
	}
	if rsp.JSON200 == nil {
		return GuardianSet{}, emptyBody(opGuardianSet)
	}
	return guardianSetFromAPI(rsp.JSON200), nil
}

// guardianSetFromAPI maps a generated guardian-set envelope to a [GuardianSet].
func guardianSetFromAPI(resp *api.GuardianGuardianSetResponse) GuardianSet {
	set := deref(resp.GuardianSet)
	return GuardianSet{
		Index:     uint32FromInt(set.Index),
		Addresses: guardianAddressesFromAPI(set.Addresses),
	}
}

// guardianAddressesFromAPI maps generated address strings to [wormholescan.GuardianAddress] values.
func guardianAddressesFromAPI(addresses *[]string) []wormholescan.GuardianAddress {
	if addresses == nil {
		return nil
	}
	out := make([]wormholescan.GuardianAddress, 0, len(*addresses))
	for _, addr := range *addresses {
		out = append(out, wormholescan.GuardianAddress(addr))
	}
	return out
}

// uint32FromInt converts a generated integer pointer to uint32.
func uint32FromInt(v *int) uint32 {
	if v == nil || *v < 0 {
		return 0
	}
	return uint32(*v) //nolint:gosec // guardian set indexes are small non-negative integers
}
