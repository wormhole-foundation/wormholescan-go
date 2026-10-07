package guardian

import (
	"context"

	"github.com/wormhole-foundation/wormholescan-go"
)

const (
	opSignedVAA      = "signed VAA"
	opSignedBatchVAA = "signed batch VAA"
)

// SignedVAA is a signed VAA from /v1/signed_vaa.
type SignedVAA struct {
	// Bytes is the raw VAA, decoded from the server's base64 vaaBytes field.
	Bytes []byte
}

// SignedBatchVAA is a signed batch VAA from /v1/signed_batch_vaa.
type SignedBatchVAA struct {
	// Bytes is the raw batch VAA, decoded from the server's base64 vaaBytes field.
	Bytes []byte
}

// GetSignedVAA returns the signed VAA bytes for id.
func (c *Client) GetSignedVAA(ctx context.Context, id wormholescan.VAAID) ([]byte, error) {
	chainID, emitter, seq := vaaPathParams(id)
	rsp, err := c.wc.API().GuardiansFindSignedVaaWithResponse(ctx, chainID, emitter, seq)
	if err != nil {
		return nil, wrapCall(opSignedVAA, err)
	}
	if err = wormholescan.CheckResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return nil, err
	}
	if rsp.JSON200 == nil || rsp.JSON200.VaaBytes == nil {
		return nil, emptyBody(opSignedVAA)
	}
	return *rsp.JSON200.VaaBytes, nil
}

// GetSignedBatchVAA returns the signed batch VAA bytes for id.
func (c *Client) GetSignedBatchVAA(ctx context.Context, id wormholescan.VAAID) ([]byte, error) {
	chainID, emitter, seq := vaaPathParams(id)
	rsp, err := c.wc.API().GuardiansFindSignedBatchVaaWithResponse(ctx, chainID, emitter, seq)
	if err != nil {
		return nil, wrapCall(opSignedBatchVAA, err)
	}
	if err = wormholescan.CheckResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return nil, err
	}
	if rsp.JSON200 == nil || rsp.JSON200.VaaBytes == nil {
		return nil, emptyBody(opSignedBatchVAA)
	}
	return *rsp.JSON200.VaaBytes, nil
}

// vaaPathParams converts a [wormholescan.VAAID] into generated path parameters.
func vaaPathParams(id wormholescan.VAAID) (int, string, uint64) {
	return int(id.Chain), string(id.Emitter), id.Sequence
}
