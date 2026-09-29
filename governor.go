package wormholescan

import (
	"context"
	"fmt"
	"iter"
	"net/http"

	"github.com/wormhole-foundation/wormholescan-go/api"
)

const (
	opGovernorConfig                      = "governor-config"
	opGovernorConfigByGuardianAddress     = "governor-config-by-guardian-address"
	opGovernorStatus                      = "governor-status"
	opGovernorStatusByGuardianAddress     = "governor-status-by-guardian-address"
	opGovernorNotionalLimit               = "governor-notional-limit"
	opGovernorNotionalLimitDetail         = "governor-notional-limit-detail"
	opGovernorNotionalLimitDetailByChain  = "governor-notional-limit-detail-by-chain"
	opGovernorNotionalAvailable           = "governor-notional-available"
	opGovernorNotionalAvailableByChain    = "governor-notional-available-by-chain"
	opGovernorMaxNotionalAvailableByChain = "governor-max-notional-available-by-chain"
	opGovernorEnqueuedVaas                = "governor-enqueued-vaas"
	opGuardiansEnqueuedVaasByChain        = "guardians-enqueued-vaas-by-chain"
	opGovernorVaas                        = "governor-vaas"
)

// ListGovernorConfigs returns one page of governor configs.
func (c *Client) ListGovernorConfigs(
	ctx context.Context,
	opts GovernorConfigListOptions,
) (Page[GovernorConfig], error) {
	params := &api.GovernorConfigParams{
		Page:     intParam(opts.Page),
		PageSize: intParam(opts.PageSize),
	}
	rsp, err := c.api.GovernorConfigWithResponse(ctx, params)
	if err != nil {
		return Page[GovernorConfig]{}, wrapDecodeError(opGovernorConfig, err)
	}
	if err = finishOK(opGovernorConfig, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return Page[GovernorConfig]{}, err
	}
	return Page[GovernorConfig]{
		Items:    mapSlice(deref(rsp.JSON200.Data), fromAPIGovernorConfig),
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// GovernorConfigs iterates governor configs starting at opts.Page.
func (c *Client) GovernorConfigs(
	ctx context.Context,
	opts GovernorConfigListOptions,
) iter.Seq2[GovernorConfig, error] {
	fetch := func(ctx context.Context, page PageOptions) (Page[GovernorConfig], error) {
		next := opts
		next.PageOptions = page
		return c.ListGovernorConfigs(ctx, next)
	}
	return paginate(ctx, opts.PageOptions, fetch)
}

// GetGovernorConfig returns the governor config for one guardian.
func (c *Client) GetGovernorConfig(ctx context.Context, guardian GuardianAddress) (GovernorConfig, error) {
	rsp, err := c.api.GovernorConfigByGuardianAddressWithResponse(ctx, string(guardian))
	if err != nil {
		return GovernorConfig{}, wrapDecodeError(opGovernorConfigByGuardianAddress, err)
	}
	if err = finishOK(opGovernorConfigByGuardianAddress, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return GovernorConfig{}, err
	}
	if rsp.JSON200.Data == nil {
		return GovernorConfig{}, emptyDecodeError(opGovernorConfigByGuardianAddress)
	}
	return fromAPIGovernorConfig(*rsp.JSON200.Data), nil
}

// ListGovernorStatuses returns one page of governor statuses.
func (c *Client) ListGovernorStatuses(
	ctx context.Context,
	opts GovernorStatusListOptions,
) (Page[GovernorStatus], error) {
	params := &api.GovernorStatusParams{
		Page:     intParam(opts.Page),
		PageSize: intParam(opts.PageSize),
	}
	rsp, err := c.api.GovernorStatusWithResponse(ctx, params)
	if err != nil {
		return Page[GovernorStatus]{}, wrapDecodeError(opGovernorStatus, err)
	}
	if err = finishOK(opGovernorStatus, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return Page[GovernorStatus]{}, err
	}
	return Page[GovernorStatus]{
		Items:    mapSlice(deref(rsp.JSON200.Data), fromAPIGovernorStatus),
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// GovernorStatuses iterates governor statuses starting at opts.Page.
func (c *Client) GovernorStatuses(
	ctx context.Context,
	opts GovernorStatusListOptions,
) iter.Seq2[GovernorStatus, error] {
	fetch := func(ctx context.Context, page PageOptions) (Page[GovernorStatus], error) {
		next := opts
		next.PageOptions = page
		return c.ListGovernorStatuses(ctx, next)
	}
	return paginate(ctx, opts.PageOptions, fetch)
}

// GetGovernorStatus returns the governor status for one guardian.
func (c *Client) GetGovernorStatus(ctx context.Context, guardian GuardianAddress) (GovernorStatus, error) {
	rsp, err := c.api.GovernorStatusByGuardianAddressWithResponse(ctx, string(guardian), nil)
	if err != nil {
		return GovernorStatus{}, wrapDecodeError(opGovernorStatusByGuardianAddress, err)
	}
	if err = finishOK(opGovernorStatusByGuardianAddress, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return GovernorStatus{}, err
	}
	if rsp.JSON200.Data == nil {
		return GovernorStatus{}, emptyDecodeError(opGovernorStatusByGuardianAddress)
	}
	return fromAPIGovernorStatus(*rsp.JSON200.Data), nil
}

// ListGovernorLimits returns one page of aggregated per-chain governor limits.
func (c *Client) ListGovernorLimits(
	ctx context.Context,
	opts GovernorLimitListOptions,
) (Page[GovernorLimit], error) {
	params := &api.GovernorNotionalLimitParams{
		Page:     intParam(opts.Page),
		PageSize: intParam(opts.PageSize),
	}
	rsp, err := c.api.GovernorNotionalLimitWithResponse(ctx, params)
	if err != nil {
		return Page[GovernorLimit]{}, wrapDecodeError(opGovernorNotionalLimit, err)
	}
	if err = finishOK(opGovernorNotionalLimit, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return Page[GovernorLimit]{}, err
	}
	return Page[GovernorLimit]{
		Items:    mapSlice(deref(rsp.JSON200.Data), fromAPIGovernorLimit),
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// GovernorLimits iterates aggregated governor limits starting at opts.Page.
func (c *Client) GovernorLimits(
	ctx context.Context,
	opts GovernorLimitListOptions,
) iter.Seq2[GovernorLimit, error] {
	fetch := func(ctx context.Context, page PageOptions) (Page[GovernorLimit], error) {
		next := opts
		next.PageOptions = page
		return c.ListGovernorLimits(ctx, next)
	}
	return paginate(ctx, opts.PageOptions, fetch)
}

// ListNotionalLimits returns one page of detailed notional limits.
func (c *Client) ListNotionalLimits(
	ctx context.Context,
	opts NotionalLimitListOptions,
) (Page[NotionalLimitDetail], error) {
	params := &api.GovernorNotionalLimitDetailParams{
		Page:     intParam(opts.Page),
		PageSize: intParam(opts.PageSize),
	}
	rsp, err := c.api.GovernorNotionalLimitDetailWithResponse(ctx, params)
	if err != nil {
		return Page[NotionalLimitDetail]{}, wrapDecodeError(opGovernorNotionalLimitDetail, err)
	}
	if err = finishOK(opGovernorNotionalLimitDetail, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return Page[NotionalLimitDetail]{}, err
	}
	return Page[NotionalLimitDetail]{
		Items:    mapSlice(deref(rsp.JSON200.Data), fromAPINotionalLimitDetail),
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// NotionalLimits iterates detailed notional limits starting at opts.Page.
func (c *Client) NotionalLimits(
	ctx context.Context,
	opts NotionalLimitListOptions,
) iter.Seq2[NotionalLimitDetail, error] {
	fetch := func(ctx context.Context, page PageOptions) (Page[NotionalLimitDetail], error) {
		next := opts
		next.PageOptions = page
		return c.ListNotionalLimits(ctx, next)
	}
	return paginate(ctx, opts.PageOptions, fetch)
}

// ListNotionalLimitsByChain returns one page of per-guardian limits for chain.
func (c *Client) ListNotionalLimitsByChain(
	ctx context.Context,
	chain ChainID,
	opts NotionalLimitListOptions,
) (Page[NotionalLimitDetail], error) {
	params := &api.GovernorNotionalLimitDetailByChainParams{
		Page:     intParam(opts.Page),
		PageSize: intParam(opts.PageSize),
	}
	rsp, err := c.api.GovernorNotionalLimitDetailByChainWithResponse(ctx, int(chain), params)
	if err != nil {
		return Page[NotionalLimitDetail]{}, wrapDecodeError(opGovernorNotionalLimitDetailByChain, err)
	}
	if err = finishOK(opGovernorNotionalLimitDetailByChain, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return Page[NotionalLimitDetail]{}, err
	}
	return Page[NotionalLimitDetail]{
		Items:    mapSlice(deref(rsp.JSON200.Data), fromAPINotionalLimitDetail),
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// ListNotionalAvailable returns one page of remaining notional by chain.
func (c *Client) ListNotionalAvailable(
	ctx context.Context,
	opts NotionalAvailableListOptions,
) (Page[NotionalAvailable], error) {
	params := &api.GovernorNotionalAvailableParams{
		Page:     intParam(opts.Page),
		PageSize: intParam(opts.PageSize),
	}
	if opts.Sort != "" {
		sort := api.GovernorNotionalAvailableParamsSortOrder(opts.Sort)
		params.SortOrder = &sort
	}
	rsp, err := c.api.GovernorNotionalAvailableWithResponse(ctx, params)
	if err != nil {
		return Page[NotionalAvailable]{}, wrapDecodeError(opGovernorNotionalAvailable, err)
	}
	if err = finishOK(opGovernorNotionalAvailable, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return Page[NotionalAvailable]{}, err
	}
	return Page[NotionalAvailable]{
		Items:    mapSlice(deref(rsp.JSON200.Data), fromAPINotionalAvailable),
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// NotionalAvailable iterates remaining notional by chain starting at opts.Page.
func (c *Client) NotionalAvailable(
	ctx context.Context,
	opts NotionalAvailableListOptions,
) iter.Seq2[NotionalAvailable, error] {
	fetch := func(ctx context.Context, page PageOptions) (Page[NotionalAvailable], error) {
		next := opts
		next.PageOptions = page
		return c.ListNotionalAvailable(ctx, next)
	}
	return paginate(ctx, opts.PageOptions, fetch)
}

// ListNotionalAvailableByChain returns one page of per-guardian availability for chain.
func (c *Client) ListNotionalAvailableByChain(
	ctx context.Context,
	chain ChainID,
	opts NotionalAvailableListOptions,
) (Page[NotionalAvailableDetail], error) {
	params := &api.GovernorNotionalAvailableByChainParams{
		Page:     intParam(opts.Page),
		PageSize: intParam(opts.PageSize),
	}
	rsp, err := c.api.GovernorNotionalAvailableByChainWithResponse(ctx, int(chain), params)
	if err != nil {
		return Page[NotionalAvailableDetail]{}, wrapDecodeError(opGovernorNotionalAvailableByChain, err)
	}
	if err = finishOK(opGovernorNotionalAvailableByChain, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return Page[NotionalAvailableDetail]{}, err
	}
	return Page[NotionalAvailableDetail]{
		Items:    mapSlice(deref(rsp.JSON200.Data), fromAPINotionalAvailableDetail),
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// GetMaxNotionalAvailable returns the guardian with the most remaining notional on chain.
func (c *Client) GetMaxNotionalAvailable(ctx context.Context, chain ChainID) (MaxNotionalAvailable, error) {
	rsp, err := c.api.GovernorMaxNotionalAvailableByChainWithResponse(ctx, int(chain))
	if err != nil {
		return MaxNotionalAvailable{}, wrapDecodeError(opGovernorMaxNotionalAvailableByChain, err)
	}
	if err = finishOK(opGovernorMaxNotionalAvailableByChain, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return MaxNotionalAvailable{}, err
	}
	if rsp.JSON200.Data == nil {
		return MaxNotionalAvailable{}, emptyDecodeError(opGovernorMaxNotionalAvailableByChain)
	}
	return fromAPIMaxNotionalAvailable(*rsp.JSON200.Data), nil
}

// EnqueuedVAAs returns every enqueued VAA across chains.
//
// The server does not page this endpoint.
func (c *Client) EnqueuedVAAs(ctx context.Context) ([]EnqueuedVAA, error) {
	rsp, err := c.api.GovernorEnqueuedVaasWithResponse(ctx, nil)
	if err != nil {
		return nil, wrapDecodeError(opGovernorEnqueuedVaas, err)
	}
	if err = finishOK(opGovernorEnqueuedVaas, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return nil, err
	}
	return fromAPIEnqueuedVaaGroups(deref(rsp.JSON200.Data)), nil
}

// EnqueuedVAAsByChain returns every enqueued VAA for chain.
//
// The server does not page this endpoint.
func (c *Client) EnqueuedVAAsByChain(ctx context.Context, chain ChainID) ([]EnqueuedVAA, error) {
	rsp, err := c.api.GuardiansEnqueuedVaasByChainWithResponse(ctx, int(chain), nil)
	if err != nil {
		return nil, wrapDecodeError(opGuardiansEnqueuedVaasByChain, err)
	}
	if err = finishOK(opGuardiansEnqueuedVaasByChain, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return nil, err
	}
	return mapSlice(deref(rsp.JSON200.Data), fromAPIEnqueuedVaaDetail), nil
}

// GovernorVAAs returns every VAA currently tracked by the governor.
//
// The server returns a bare array and does not page this endpoint.
func (c *Client) GovernorVAAs(ctx context.Context) ([]GovernorVAA, error) {
	rsp, err := c.api.GovernorVaasWithResponse(ctx)
	if err != nil {
		return nil, wrapDecodeError(opGovernorVaas, err)
	}
	if err = finishOK(opGovernorVaas, rsp.HTTPResponse, rsp.Body, rsp.JSON200); err != nil {
		return nil, err
	}
	return mapSlice(deref(rsp.JSON200), fromAPIGovernorVAA), nil
}

// finishOK returns an API or decode error for a generated 2xx JSON body.
func finishOK(op string, httpResp *http.Response, body []byte, json200 any) error {
	if err := checkResponse(httpResp, body); err != nil {
		return err
	}
	if json200 == nil {
		return emptyDecodeError(op)
	}
	return nil
}

// wrapDecodeError prefixes a generated-client error with the operation id.
func wrapDecodeError(op string, err error) error {
	return fmt.Errorf("%sdecode %s: %w", errPrefix, op, err)
}

// emptyDecodeError reports a 2xx response with no decoded JSON body.
func emptyDecodeError(op string) error {
	return fmt.Errorf("%sdecode %s: empty body", errPrefix, op)
}

// intParam returns n when non-zero so omitted query params stay omitted.
func intParam(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}
