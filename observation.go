package wormholescan

import (
	"context"
	"iter"
	"net/http"
	"time"

	"github.com/wormhole-foundation/wormholescan-go/api"
)

const (
	opListObservations                  = "list observations"
	opListObservationsByChain           = "list observations by chain"
	opListObservationsByEmitter         = "list observations by emitter"
	opListObservationsByVAA             = "list observations by VAA"
	opGetObservation                    = "get observation"
	opListDelegateObservationsByChain   = "list delegate observations by chain"
	opListDelegateObservationsByEmitter = "list delegate observations by emitter"
	opListDelegateObservationsByVAA     = "list delegate observations by VAA"
	opGetDelegateObservation            = "get delegate observation"
)

// Observation is a guardian observation of a VAA, as stored by Wormholescan.
type Observation struct {
	// ID is the observation identifier as the server sends it.
	ID string
	// EmitterChain is the chain that emitted the observed VAA.
	EmitterChain ChainID
	// EmitterAddress is the 32-byte hex emitter address.
	EmitterAddress EmitterAddress
	// Sequence is the observed VAA sequence number.
	Sequence uint64
	// Hash is the observation hash. The server sends this as base64 bytes.
	Hash []byte
	// TxHash is the source-chain transaction hash as raw bytes, not a string.
	TxHash []byte
	// GuardianAddress is the observing guardian's 0x-prefixed EVM address.
	GuardianAddress GuardianAddress
	// Signature is the guardian signature bytes.
	Signature []byte
	// IndexedAt is when Wormholescan indexed the observation. Zero when absent.
	IndexedAt time.Time
	// UpdatedAt is when Wormholescan last updated the observation. Zero when absent.
	UpdatedAt time.Time
}

// ObservationListOptions configures observation list endpoints.
type ObservationListOptions struct {
	PageOptions

	// TxHash filters find-observations by transaction hash.
	// Chain, emitter, VAA, and delegate list methods ignore it.
	TxHash TxHash
}

// DelegateObservation is a delegated guardian observation.
type DelegateObservation struct {
	// ID is the observation identifier as the server sends it.
	ID string
	// EmitterChain is the chain that emitted the observed VAA.
	EmitterChain ChainID
	// EmitterAddress is the 32-byte hex emitter address.
	EmitterAddress EmitterAddress
	// Sequence is the observed VAA sequence number.
	Sequence uint64
	// Hash is the observation hash. The server sends this as base64 bytes.
	Hash []byte
	// TxHash is the source-chain transaction hash as raw bytes.
	TxHash []byte
	// DelegatedGuardianAddr is the delegated guardian address.
	// The JSON name is not a typo; the server spells it this way.
	DelegatedGuardianAddr GuardianAddress
	// Signature is the guardian signature bytes.
	Signature []byte
	// Payload is the observation payload bytes.
	Payload []byte
	// ConsistencyLevel is the observation consistency level.
	ConsistencyLevel int
	// Nonce is the observation nonce.
	Nonce int
	// IsReobservation reports whether this is a reobservation.
	IsReobservation bool
	// Unreliable reports whether the observation is marked unreliable.
	Unreliable bool
	// VerificationState is the server-assigned verification state.
	VerificationState int
	// SentTimestamp is the sent timestamp as the server sends it.
	// Empty when the server has no value.
	SentTimestamp string
	// Timestamp is the observation timestamp in UTC. Zero when absent.
	Timestamp time.Time
	// IndexedAt is when Wormholescan indexed the observation. Zero when absent.
	IndexedAt time.Time
	// UpdatedAt is when Wormholescan last updated the observation. Zero when absent.
	UpdatedAt time.Time
}

// ListObservations returns one page of observations.
//
// The server sends a bare JSON array, not the pagination envelope. [Page.Last]
// still uses the item count versus PageSize.
func (c *Client) ListObservations(ctx context.Context, opts ObservationListOptions) (Page[Observation], error) {
	rsp, err := c.api.FindObservationsWithResponse(ctx, findObservationsParams(opts))
	if err != nil {
		return Page[Observation]{}, callErr(opListObservations, err)
	}
	return observationsFromResponse(
		opListObservations,
		rsp.HTTPResponse,
		rsp.Body,
		rsp.JSON200,
		opts.PageOptions,
	)
}

// Observations iterates observations starting at opts.Page until the last page.
func (c *Client) Observations(ctx context.Context, opts ObservationListOptions) iter.Seq2[Observation, error] {
	return paginate(ctx, opts.PageOptions, func(ctx context.Context, page PageOptions) (Page[Observation], error) {
		o := opts
		o.PageOptions = page
		return c.ListObservations(ctx, o)
	})
}

// ListObservationsByChain returns one page of observations for chain.
func (c *Client) ListObservationsByChain(
	ctx context.Context,
	chain ChainID,
	opts ObservationListOptions,
) (Page[Observation], error) {
	rsp, err := c.api.FindObservationsByChainWithResponse(
		ctx,
		pathChain(chain),
		findObservationsByChainParams(opts),
	)
	if err != nil {
		return Page[Observation]{}, callErr(opListObservationsByChain, err)
	}
	return observationsFromResponse(
		opListObservationsByChain,
		rsp.HTTPResponse,
		rsp.Body,
		rsp.JSON200,
		opts.PageOptions,
	)
}

// ObservationsByChain iterates observations for chain starting at opts.Page.
func (c *Client) ObservationsByChain(
	ctx context.Context,
	chain ChainID,
	opts ObservationListOptions,
) iter.Seq2[Observation, error] {
	return paginate(ctx, opts.PageOptions, func(ctx context.Context, page PageOptions) (Page[Observation], error) {
		o := opts
		o.PageOptions = page
		return c.ListObservationsByChain(ctx, chain, o)
	})
}

// ListObservationsByEmitter returns one page of observations for emitter on chain.
func (c *Client) ListObservationsByEmitter(
	ctx context.Context,
	chain ChainID,
	emitter EmitterAddress,
	opts ObservationListOptions,
) (Page[Observation], error) {
	rsp, err := c.api.FindObservationsByEmitterWithResponse(
		ctx,
		pathChain(chain),
		string(emitter),
		findObservationsByEmitterParams(opts),
	)
	if err != nil {
		return Page[Observation]{}, callErr(opListObservationsByEmitter, err)
	}
	return observationsFromResponse(
		opListObservationsByEmitter,
		rsp.HTTPResponse,
		rsp.Body,
		rsp.JSON200,
		opts.PageOptions,
	)
}

// ObservationsByEmitter iterates observations for emitter on chain starting at opts.Page.
func (c *Client) ObservationsByEmitter(
	ctx context.Context,
	chain ChainID,
	emitter EmitterAddress,
	opts ObservationListOptions,
) iter.Seq2[Observation, error] {
	return paginate(ctx, opts.PageOptions, func(ctx context.Context, page PageOptions) (Page[Observation], error) {
		o := opts
		o.PageOptions = page
		return c.ListObservationsByEmitter(ctx, chain, emitter, o)
	})
}

// ListObservationsByVAA returns one page of observations for id.
func (c *Client) ListObservationsByVAA(
	ctx context.Context,
	id VAAID,
	opts ObservationListOptions,
) (Page[Observation], error) {
	seq, err := pathSequence(id.Sequence)
	if err != nil {
		return Page[Observation]{}, err
	}
	rsp, err := c.api.FindObservationsBySequenceWithResponse(
		ctx,
		pathChain(id.Chain),
		string(id.Emitter),
		seq,
		findObservationsBySequenceParams(opts),
	)
	if err != nil {
		return Page[Observation]{}, callErr(opListObservationsByVAA, err)
	}
	return observationsFromResponse(
		opListObservationsByVAA,
		rsp.HTTPResponse,
		rsp.Body,
		rsp.JSON200,
		opts.PageOptions,
	)
}

// ObservationsByVAA iterates observations for id starting at opts.Page.
func (c *Client) ObservationsByVAA(
	ctx context.Context,
	id VAAID,
	opts ObservationListOptions,
) iter.Seq2[Observation, error] {
	return paginate(ctx, opts.PageOptions, func(ctx context.Context, page PageOptions) (Page[Observation], error) {
		o := opts
		o.PageOptions = page
		return c.ListObservationsByVAA(ctx, id, o)
	})
}

// GetObservation returns the observation for id signed by signer whose hash matches hash.
//
// hash is the lowercase hex encoding of the observation hash with no 0x prefix,
// the last segment of [Observation.ID].
//
// The live API currently answers 404 for every known hash encoding (hex from
// Observation.ID, base64 Observation.Hash, URL-encoded raw hash bytes, and the
// corresponding txHash forms). This method is kept for when upstream fixes
// find-observations-by-id.
func (c *Client) GetObservation(
	ctx context.Context,
	id VAAID,
	signer GuardianAddress,
	hash string,
) (Observation, error) {
	seq, err := pathSequence(id.Sequence)
	if err != nil {
		return Observation{}, err
	}
	rsp, err := c.api.FindObservationsByIdWithResponse(
		ctx,
		pathChain(id.Chain),
		string(id.Emitter),
		seq,
		string(signer),
		hash,
		nil,
	)
	if err != nil {
		return Observation{}, callErr(opGetObservation, err)
	}
	if err = checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return Observation{}, err
	}
	if rsp.JSON200 == nil {
		return Observation{}, emptyBody(opGetObservation)
	}
	return observationFromAPI(*rsp.JSON200), nil
}

// ListDelegateObservationsByChain returns one page of delegate observations for chain.
func (c *Client) ListDelegateObservationsByChain(
	ctx context.Context,
	chain ChainID,
	opts ObservationListOptions,
) (Page[DelegateObservation], error) {
	rsp, err := c.api.FindDelegateObservationsByChainWithResponse(
		ctx,
		pathChain(chain),
		findDelegateObservationsByChainParams(opts),
	)
	if err != nil {
		return Page[DelegateObservation]{}, callErr(opListDelegateObservationsByChain, err)
	}
	return delegateObservationsFromResponse(
		opListDelegateObservationsByChain,
		rsp.HTTPResponse,
		rsp.Body,
		rsp.JSON200,
		opts.PageOptions,
	)
}

// DelegateObservationsByChain iterates delegate observations for chain starting at opts.Page.
func (c *Client) DelegateObservationsByChain(
	ctx context.Context,
	chain ChainID,
	opts ObservationListOptions,
) iter.Seq2[DelegateObservation, error] {
	return paginate(
		ctx,
		opts.PageOptions,
		func(ctx context.Context, page PageOptions) (Page[DelegateObservation], error) {
			o := opts
			o.PageOptions = page
			return c.ListDelegateObservationsByChain(ctx, chain, o)
		},
	)
}

// ListDelegateObservationsByEmitter returns one page of delegate observations for emitter on chain.
func (c *Client) ListDelegateObservationsByEmitter(
	ctx context.Context,
	chain ChainID,
	emitter EmitterAddress,
	opts ObservationListOptions,
) (Page[DelegateObservation], error) {
	rsp, err := c.api.FindDelegateObservationsByEmitterWithResponse(
		ctx,
		pathChain(chain),
		string(emitter),
		findDelegateObservationsByEmitterParams(opts),
	)
	if err != nil {
		return Page[DelegateObservation]{}, callErr(opListDelegateObservationsByEmitter, err)
	}
	return delegateObservationsFromResponse(
		opListDelegateObservationsByEmitter,
		rsp.HTTPResponse,
		rsp.Body,
		rsp.JSON200,
		opts.PageOptions,
	)
}

// DelegateObservationsByEmitter iterates delegate observations for emitter on chain.
func (c *Client) DelegateObservationsByEmitter(
	ctx context.Context,
	chain ChainID,
	emitter EmitterAddress,
	opts ObservationListOptions,
) iter.Seq2[DelegateObservation, error] {
	return paginate(
		ctx,
		opts.PageOptions,
		func(ctx context.Context, page PageOptions) (Page[DelegateObservation], error) {
			o := opts
			o.PageOptions = page
			return c.ListDelegateObservationsByEmitter(ctx, chain, emitter, o)
		},
	)
}

// ListDelegateObservationsByVAA returns one page of delegate observations for id.
func (c *Client) ListDelegateObservationsByVAA(
	ctx context.Context,
	id VAAID,
	opts ObservationListOptions,
) (Page[DelegateObservation], error) {
	seq, err := pathSequence(id.Sequence)
	if err != nil {
		return Page[DelegateObservation]{}, err
	}
	rsp, err := c.api.FindDelegateObservationsBySequenceWithResponse(
		ctx,
		pathChain(id.Chain),
		string(id.Emitter),
		seq,
		findDelegateObservationsBySequenceParams(opts),
	)
	if err != nil {
		return Page[DelegateObservation]{}, callErr(opListDelegateObservationsByVAA, err)
	}
	return delegateObservationsFromResponse(
		opListDelegateObservationsByVAA,
		rsp.HTTPResponse,
		rsp.Body,
		rsp.JSON200,
		opts.PageOptions,
	)
}

// DelegateObservationsByVAA iterates delegate observations for id starting at opts.Page.
func (c *Client) DelegateObservationsByVAA(
	ctx context.Context,
	id VAAID,
	opts ObservationListOptions,
) iter.Seq2[DelegateObservation, error] {
	return paginate(
		ctx,
		opts.PageOptions,
		func(ctx context.Context, page PageOptions) (Page[DelegateObservation], error) {
			o := opts
			o.PageOptions = page
			return c.ListDelegateObservationsByVAA(ctx, id, o)
		},
	)
}

// GetDelegateObservation returns the delegate observation for id recorded by guardian.
//
// The request path is /api/v1/observations/delegate/{chain}/{emitter}/{sequence}/{guardian}.
func (c *Client) GetDelegateObservation(
	ctx context.Context,
	id VAAID,
	guardian GuardianAddress,
) (DelegateObservation, error) {
	seq, err := pathSequence(id.Sequence)
	if err != nil {
		return DelegateObservation{}, err
	}
	rsp, err := c.api.FindDelegateObservationsByGuardianWithResponse(
		ctx,
		pathChain(id.Chain),
		string(id.Emitter),
		seq,
		string(guardian),
		nil,
	)
	if err != nil {
		return DelegateObservation{}, callErr(opGetDelegateObservation, err)
	}
	if err = checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return DelegateObservation{}, err
	}
	if rsp.JSON200 == nil {
		return DelegateObservation{}, emptyBody(opGetDelegateObservation)
	}
	docs := *rsp.JSON200
	if len(docs) == 0 {
		return DelegateObservation{}, emptyBody(opGetDelegateObservation)
	}
	return delegateObservationFromAPI(docs[0]), nil
}

// observationsFromResponse maps a bare observation array into a [Page].
func observationsFromResponse(
	op string,
	httpResp *http.Response,
	body []byte,
	json200 *[]api.ObservationsObservationDoc,
	opts PageOptions,
) (Page[Observation], error) {
	if err := checkResponse(httpResp, body); err != nil {
		return Page[Observation]{}, err
	}
	if json200 == nil {
		return Page[Observation]{}, emptyBody(op)
	}
	return observationsPage(*json200, opts), nil
}

// observationsPage converts generated observation documents into a [Page].
func observationsPage(docs []api.ObservationsObservationDoc, opts PageOptions) Page[Observation] {
	items := make([]Observation, 0, len(docs))
	for i := range docs {
		items = append(items, observationFromAPI(docs[i]))
	}
	return Page[Observation]{Items: items, Page: opts.Page, PageSize: opts.PageSize}
}

// observationFromAPI maps [api.ObservationsObservationDoc] onto [Observation].
func observationFromAPI(doc api.ObservationsObservationDoc) Observation {
	return Observation{
		ID:              deref(doc.Id),
		EmitterChain:    chainFromAPI(deref(doc.EmitterChain)),
		EmitterAddress:  EmitterAddress(deref(doc.EmitterAddr)),
		Sequence:        uint64FromInt64(deref(doc.Sequence)),
		Hash:            deref(doc.Hash),
		TxHash:          deref(doc.TxHash),
		GuardianAddress: GuardianAddress(deref(doc.GuardianAddr)),
		Signature:       deref(doc.Signature),
		IndexedAt:       derefTime(doc.IndexedAt),
		UpdatedAt:       derefTime(doc.UpdatedAt),
	}
}

// delegateObservationsFromResponse maps a bare delegate-observation array into a [Page].
func delegateObservationsFromResponse(
	op string,
	httpResp *http.Response,
	body []byte,
	json200 *[]api.DelegateObservationsDelegateObservationDoc,
	opts PageOptions,
) (Page[DelegateObservation], error) {
	if err := checkResponse(httpResp, body); err != nil {
		return Page[DelegateObservation]{}, err
	}
	if json200 == nil {
		return Page[DelegateObservation]{}, emptyBody(op)
	}
	return delegateObservationsPage(*json200, opts), nil
}

// delegateObservationsPage converts generated delegate documents into a [Page].
func delegateObservationsPage(
	docs []api.DelegateObservationsDelegateObservationDoc,
	opts PageOptions,
) Page[DelegateObservation] {
	items := make([]DelegateObservation, 0, len(docs))
	for i := range docs {
		items = append(items, delegateObservationFromAPI(docs[i]))
	}
	return Page[DelegateObservation]{Items: items, Page: opts.Page, PageSize: opts.PageSize}
}

// delegateObservationFromAPI maps a generated delegate document onto [DelegateObservation].
func delegateObservationFromAPI(doc api.DelegateObservationsDelegateObservationDoc) DelegateObservation {
	return DelegateObservation{
		ID:                    deref(doc.Id),
		EmitterChain:          chainFromAPI(deref(doc.EmitterChain)),
		EmitterAddress:        EmitterAddress(deref(doc.EmitterAddr)),
		Sequence:              uint64FromInt64(deref(doc.Sequence)),
		Hash:                  deref(doc.Hash),
		TxHash:                deref(doc.TxHash),
		DelegatedGuardianAddr: GuardianAddress(deref(doc.DelegatedGuardianAddr)),
		Signature:             deref(doc.Signature),
		Payload:               deref(doc.Payload),
		ConsistencyLevel:      deref(doc.ConsistencyLevel),
		Nonce:                 deref(doc.Nonce),
		IsReobservation:       deref(doc.IsReobservation),
		Unreliable:            deref(doc.Unreliable),
		VerificationState:     deref(doc.VerificationState),
		SentTimestamp:         deref(doc.SentTimestamp),
		Timestamp:             derefTime(doc.Timestamp),
		IndexedAt:             derefTime(doc.IndexedAt),
		UpdatedAt:             derefTime(doc.UpdatedAt),
	}
}

// findObservationsParams maps list options onto generated find-observations params.
func findObservationsParams(opts ObservationListOptions) *api.FindObservationsParams {
	params := &api.FindObservationsParams{
		Page:     pagePtr(opts.PageOptions),
		PageSize: pageSizePtr(opts.PageOptions),
	}
	if opts.Sort != "" {
		sort := api.FindObservationsParamsSortOrder(opts.Sort)
		params.SortOrder = &sort
	}
	if opts.TxHash != "" {
		tx := string(opts.TxHash)
		params.TxHash = &tx
	}
	return params
}

// findObservationsByChainParams maps list options onto generated by-chain params.
func findObservationsByChainParams(opts ObservationListOptions) *api.FindObservationsByChainParams {
	params := &api.FindObservationsByChainParams{
		Page:     pagePtr(opts.PageOptions),
		PageSize: pageSizePtr(opts.PageOptions),
	}
	if opts.Sort != "" {
		sort := api.FindObservationsByChainParamsSortOrder(opts.Sort)
		params.SortOrder = &sort
	}
	return params
}

// findObservationsByEmitterParams maps list options onto generated by-emitter params.
func findObservationsByEmitterParams(opts ObservationListOptions) *api.FindObservationsByEmitterParams {
	params := &api.FindObservationsByEmitterParams{
		Page:     pagePtr(opts.PageOptions),
		PageSize: pageSizePtr(opts.PageOptions),
	}
	if opts.Sort != "" {
		sort := api.FindObservationsByEmitterParamsSortOrder(opts.Sort)
		params.SortOrder = &sort
	}
	return params
}

// findObservationsBySequenceParams maps list options onto generated by-sequence params.
func findObservationsBySequenceParams(opts ObservationListOptions) *api.FindObservationsBySequenceParams {
	params := &api.FindObservationsBySequenceParams{
		Page:     pagePtr(opts.PageOptions),
		PageSize: pageSizePtr(opts.PageOptions),
	}
	if opts.Sort != "" {
		sort := api.FindObservationsBySequenceParamsSortOrder(opts.Sort)
		params.SortOrder = &sort
	}
	return params
}

// findDelegateObservationsByChainParams maps list options onto generated delegate-by-chain params.
func findDelegateObservationsByChainParams(opts ObservationListOptions) *api.FindDelegateObservationsByChainParams {
	return &api.FindDelegateObservationsByChainParams{
		Page:      pagePtr(opts.PageOptions),
		PageSize:  pageSizePtr(opts.PageOptions),
		SortOrder: sortStringPtr(opts.PageOptions),
	}
}

// findDelegateObservationsByEmitterParams maps list options onto generated delegate-by-emitter params.
func findDelegateObservationsByEmitterParams(
	opts ObservationListOptions,
) *api.FindDelegateObservationsByEmitterParams {
	return &api.FindDelegateObservationsByEmitterParams{
		Page:      pagePtr(opts.PageOptions),
		PageSize:  pageSizePtr(opts.PageOptions),
		SortOrder: sortStringPtr(opts.PageOptions),
	}
}

// findDelegateObservationsBySequenceParams maps list options onto generated delegate-by-sequence params.
func findDelegateObservationsBySequenceParams(
	opts ObservationListOptions,
) *api.FindDelegateObservationsBySequenceParams {
	return &api.FindDelegateObservationsBySequenceParams{
		Page:      pagePtr(opts.PageOptions),
		PageSize:  pageSizePtr(opts.PageOptions),
		SortOrder: sortStringPtr(opts.PageOptions),
	}
}

// sortStringPtr returns opts.Sort as a string pointer, or nil when unset.
func sortStringPtr(opts PageOptions) *string {
	if opts.Sort == "" {
		return nil
	}
	s := string(opts.Sort)
	return &s
}
