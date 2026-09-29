package wormholescan

import (
	"context"
	"fmt"
	"iter"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormholescan-go/api"
)

const (
	opGetVAA            = "get VAA"
	opGetDuplicatedVAAs = "get duplicated VAAs"
	opListVAAs          = "list VAAs"
	opListVAAsByChain   = "list VAAs by chain"
	opListVAAsByEmitter = "list VAAs by emitter"
)

// VAA is a Wormhole verified action approval as stored by Wormholescan.
type VAA struct {
	// ID is the chain/emitter/sequence identifier.
	ID VAAID
	// EmitterChain is the chain that emitted the VAA.
	EmitterChain ChainID
	// EmitterAddress is the 32-byte hex emitter address.
	EmitterAddress EmitterAddress
	// EmitterNativeAddress is the emitter in the source chain's native format.
	// Empty when the server has no value.
	EmitterNativeAddress string
	// Sequence is the VAA sequence number.
	Sequence uint64
	// GuardianSetIndex is the guardian set that signed the VAA.
	GuardianSetIndex uint32
	// Version is the VAA protocol version.
	Version uint8
	// Digest is the hex digest of the VAA.
	Digest Digest
	// TxHash is the source-chain transaction hash. Empty when the server has no value.
	TxHash TxHash
	// Raw is the signed VAA bytes.
	Raw []byte
	// Timestamp is the VAA timestamp in UTC. Zero when the server has no value.
	Timestamp time.Time
	// IndexedAt is when Wormholescan indexed the VAA. Zero when the server has no value.
	IndexedAt time.Time
	// UpdatedAt is when Wormholescan last updated the VAA. Zero when the server has no value.
	UpdatedAt time.Time
	// IsDuplicated reports whether Wormholescan stored more than one VAA for this ID.
	IsDuplicated bool
	// IsSolanaShim reports whether the VAA was emitted through an SVM shim program.
	IsSolanaShim bool
	// Payload is the parsed payload when parsedPayload=true was requested.
	// Nil when the server omitted it.
	Payload map[string]any
}

// VAAGetOptions configures [Client.GetVAA].
type VAAGetOptions struct {
	// ParsedPayload requests the decoded payload when the server has one.
	ParsedPayload bool
}

// VAAListOptions configures VAA list endpoints.
type VAAListOptions struct {
	PageOptions
	// TxHash filters find-all-vaas by transaction hash. Other list methods ignore it.
	TxHash TxHash
	// ParsedPayload requests decoded payloads on endpoints that support it.
	ParsedPayload bool
}

// GetVAA returns the VAA identified by id.
func (c *Client) GetVAA(ctx context.Context, id VAAID, opts VAAGetOptions) (VAA, error) {
	seq, err := pathSequence(id.Sequence)
	if err != nil {
		return VAA{}, err
	}
	var params *api.FindVaaByIdParams
	if opts.ParsedPayload {
		params = &api.FindVaaByIdParams{ParsedPayload: &opts.ParsedPayload}
	}
	rsp, err := c.api.FindVaaByIdWithResponse(ctx, pathChain(id.Chain), string(id.Emitter), seq, params)
	if err != nil {
		return VAA{}, callErr(opGetVAA, err)
	}
	if err = checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return VAA{}, err
	}
	if rsp.JSON200 == nil || rsp.JSON200.Data == nil {
		return VAA{}, emptyBody(opGetVAA)
	}
	return vaaFromAPI(*rsp.JSON200.Data)
}

// GetDuplicatedVAAs returns duplicate VAA documents stored for id.
func (c *Client) GetDuplicatedVAAs(ctx context.Context, id VAAID) ([]VAA, error) {
	seq, err := pathSequence(id.Sequence)
	if err != nil {
		return nil, err
	}
	rsp, err := c.api.FindDuplicatedVaaByIdWithResponse(ctx, pathChain(id.Chain), string(id.Emitter), seq)
	if err != nil {
		return nil, callErr(opGetDuplicatedVAAs, err)
	}
	if err = checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return nil, err
	}
	if rsp.JSON200 == nil {
		return nil, emptyBody(opGetDuplicatedVAAs)
	}
	return vaasFromDuplicateAPI(deref(rsp.JSON200.Data))
}

// ListVAAs returns one page of VAAs.
func (c *Client) ListVAAs(ctx context.Context, opts VAAListOptions) (Page[VAA], error) {
	rsp, err := c.api.FindAllVaasWithResponse(ctx, findAllVaasParams(opts))
	if err != nil {
		return Page[VAA]{}, callErr(opListVAAs, err)
	}
	return vaasFromResponse(opListVAAs, rsp.HTTPResponse, rsp.Body, rsp.JSON200, opts.PageOptions)
}

// VAAs iterates VAAs starting at opts.Page until the last page.
func (c *Client) VAAs(ctx context.Context, opts VAAListOptions) iter.Seq2[VAA, error] {
	return paginate(ctx, opts.PageOptions, func(ctx context.Context, page PageOptions) (Page[VAA], error) {
		o := opts
		o.PageOptions = page
		return c.ListVAAs(ctx, o)
	})
}

// ListVAAsByChain returns one page of VAAs emitted on chain.
func (c *Client) ListVAAsByChain(ctx context.Context, chain ChainID, opts VAAListOptions) (Page[VAA], error) {
	rsp, err := c.api.FindVaasByChainWithResponse(ctx, pathChain(chain), findVaasByChainParams(opts))
	if err != nil {
		return Page[VAA]{}, callErr(opListVAAsByChain, err)
	}
	return vaasFromResponse(opListVAAsByChain, rsp.HTTPResponse, rsp.Body, rsp.JSON200, opts.PageOptions)
}

// VAAsByChain iterates VAAs emitted on chain starting at opts.Page.
func (c *Client) VAAsByChain(ctx context.Context, chain ChainID, opts VAAListOptions) iter.Seq2[VAA, error] {
	return paginate(ctx, opts.PageOptions, func(ctx context.Context, page PageOptions) (Page[VAA], error) {
		o := opts
		o.PageOptions = page
		return c.ListVAAsByChain(ctx, chain, o)
	})
}

// ListVAAsByEmitter returns one page of VAAs from emitter on chain.
func (c *Client) ListVAAsByEmitter(
	ctx context.Context,
	chain ChainID,
	emitter EmitterAddress,
	opts VAAListOptions,
) (Page[VAA], error) {
	rsp, err := c.api.FindVaasByEmitterWithResponse(
		ctx,
		pathChain(chain),
		string(emitter),
		findVaasByEmitterParams(opts),
	)
	if err != nil {
		return Page[VAA]{}, callErr(opListVAAsByEmitter, err)
	}
	return vaasFromResponse(opListVAAsByEmitter, rsp.HTTPResponse, rsp.Body, rsp.JSON200, opts.PageOptions)
}

// VAAsByEmitter iterates VAAs from emitter on chain starting at opts.Page.
func (c *Client) VAAsByEmitter(
	ctx context.Context,
	chain ChainID,
	emitter EmitterAddress,
	opts VAAListOptions,
) iter.Seq2[VAA, error] {
	return paginate(ctx, opts.PageOptions, func(ctx context.Context, page PageOptions) (Page[VAA], error) {
		o := opts
		o.PageOptions = page
		return c.ListVAAsByEmitter(ctx, chain, emitter, o)
	})
}

// vaasFromResponse maps a paginated VAA envelope into a [Page].
func vaasFromResponse(
	op string,
	httpResp *http.Response,
	body []byte,
	json200 *api.ResponseResponseArrayVaaVaaDoc,
	opts PageOptions,
) (Page[VAA], error) {
	if err := checkResponse(httpResp, body); err != nil {
		return Page[VAA]{}, err
	}
	if json200 == nil {
		return Page[VAA]{}, emptyBody(op)
	}
	return vaaPage(deref(json200.Data), opts)
}

// vaaPage converts generated VAA documents into a [Page].
func vaaPage(docs []api.VaaVaaDoc, opts PageOptions) (Page[VAA], error) {
	items, err := vaasFromAPI(docs)
	if err != nil {
		return Page[VAA]{}, err
	}
	return Page[VAA]{Items: items, Page: opts.Page, PageSize: opts.PageSize}, nil
}

// vaasFromAPI converts generated VAA documents.
func vaasFromAPI(docs []api.VaaVaaDoc) ([]VAA, error) {
	out := make([]VAA, 0, len(docs))
	for i := range docs {
		v, err := vaaFromAPI(docs[i])
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// vaasFromDuplicateAPI converts duplicated-VAA documents.
func vaasFromDuplicateAPI(docs []api.VaaDuplicateVaaDoc) ([]VAA, error) {
	out := make([]VAA, 0, len(docs))
	for i := range docs {
		v, err := vaaFromDuplicateAPI(docs[i])
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// vaaFromAPI maps [api.VaaVaaDoc] onto [VAA].
func vaaFromAPI(doc api.VaaVaaDoc) (VAA, error) {
	seq := uint64FromInt64(deref(doc.Sequence))
	id, err := vaaIDFromDoc(deref(doc.Id), deref(doc.EmitterChain), deref(doc.EmitterAddr), seq)
	if err != nil {
		return VAA{}, err
	}
	return VAA{
		ID:                   id,
		EmitterChain:         chainFromAPI(deref(doc.EmitterChain)),
		EmitterAddress:       EmitterAddress(deref(doc.EmitterAddr)),
		EmitterNativeAddress: deref(doc.EmitterNativeAddr),
		Sequence:             seq,
		GuardianSetIndex:     uint32FromInt(deref(doc.GuardianSetIndex)),
		Version:              uint8FromInt(deref(doc.Version)),
		Digest:               Digest(deref(doc.Digest)),
		TxHash:               TxHash(deref(doc.TxHash)),
		Raw:                  deref(doc.Vaa),
		Timestamp:            derefTime(doc.Timestamp),
		IndexedAt:            derefTime(doc.IndexedAt),
		UpdatedAt:            derefTime(doc.UpdatedAt),
		IsDuplicated:         deref(doc.IsDuplicated),
		IsSolanaShim:         deref(doc.IsSolanaShim),
		Payload:              payloadFromAPI(doc.Payload),
	}, nil
}

// vaaFromDuplicateAPI maps [api.VaaDuplicateVaaDoc] onto [VAA].
func vaaFromDuplicateAPI(doc api.VaaDuplicateVaaDoc) (VAA, error) {
	seq := uint64FromDec(deref(doc.Sequence))
	id, err := vaaIDFromDoc(deref(doc.Id), deref(doc.EmitterChain), deref(doc.EmitterAddr), seq)
	if err != nil {
		return VAA{}, err
	}
	return VAA{
		ID:                   id,
		EmitterChain:         chainFromAPI(deref(doc.EmitterChain)),
		EmitterAddress:       EmitterAddress(deref(doc.EmitterAddr)),
		EmitterNativeAddress: deref(doc.EmitterNativeAddr),
		Sequence:             seq,
		GuardianSetIndex:     uint32FromInt(deref(doc.GuardianSetIndex)),
		Version:              uint8FromInt(deref(doc.Version)),
		Digest:               Digest(deref(doc.Digest)),
		Raw:                  deref(doc.Vaa),
		Timestamp:            derefTime(doc.Timestamp),
		IndexedAt:            derefTime(doc.IndexedAt),
		UpdatedAt:            derefTime(doc.UpdatedAt),
	}, nil
}

// vaaIDFromDoc parses id, or builds one from the document fields when id is empty.
func vaaIDFromDoc(id string, chain api.VaaChainID, emitter string, seq uint64) (VAAID, error) {
	if id == "" {
		return VAAID{
			Chain:    chainFromAPI(chain),
			Emitter:  EmitterAddress(emitter),
			Sequence: seq,
		}, nil
	}
	return ParseVAAID(id)
}

// findAllVaasParams maps list options onto generated find-all-vaas params.
func findAllVaasParams(opts VAAListOptions) *api.FindAllVaasParams {
	params := &api.FindAllVaasParams{
		Page:     pagePtr(opts.PageOptions),
		PageSize: pageSizePtr(opts.PageOptions),
	}
	if opts.Sort != "" {
		sort := api.FindAllVaasParamsSortOrder(opts.Sort)
		params.SortOrder = &sort
	}
	if opts.TxHash != "" {
		tx := string(opts.TxHash)
		params.TxHash = &tx
	}
	if opts.ParsedPayload {
		params.ParsedPayload = &opts.ParsedPayload
	}
	return params
}

// findVaasByChainParams maps list options onto generated find-vaas-by-chain params.
func findVaasByChainParams(opts VAAListOptions) *api.FindVaasByChainParams {
	params := &api.FindVaasByChainParams{
		Page:     pagePtr(opts.PageOptions),
		PageSize: pageSizePtr(opts.PageOptions),
	}
	if opts.Sort != "" {
		sort := api.FindVaasByChainParamsSortOrder(opts.Sort)
		params.SortOrder = &sort
	}
	return params
}

// findVaasByEmitterParams maps list options onto generated find-vaas-by-emitter params.
func findVaasByEmitterParams(opts VAAListOptions) *api.FindVaasByEmitterParams {
	params := &api.FindVaasByEmitterParams{
		Page:     pagePtr(opts.PageOptions),
		PageSize: pageSizePtr(opts.PageOptions),
	}
	if opts.Sort != "" {
		sort := api.FindVaasByEmitterParamsSortOrder(opts.Sort)
		params.SortOrder = &sort
	}
	return params
}

// callErr wraps err with the wormholescan prefix and operation name.
func callErr(op string, err error) error {
	return fmt.Errorf("%s%s: %w", errPrefix, op, err)
}

// emptyBody is returned when a 200 response has no decoded JSON body.
func emptyBody(op string) error {
	return fmt.Errorf("%sdecode %s: empty body", errPrefix, op)
}

// pagePtr returns opts.Page, or nil when it is the zero value.
func pagePtr(opts PageOptions) *int {
	if opts.Page == 0 {
		return nil
	}
	return &opts.Page
}

// pageSizePtr returns opts.PageSize, or nil when it is the zero value.
func pageSizePtr(opts PageOptions) *int {
	if opts.PageSize == 0 {
		return nil
	}
	return &opts.PageSize
}

// pathChain converts a [ChainID] into the integer path parameter the generated client uses.
func pathChain(id ChainID) int {
	return int(id)
}

// pathSequence converts a sequence number into the integer path parameter the generated client uses.
func pathSequence(seq uint64) (int, error) {
	if seq > uint64(math.MaxInt) {
		return 0, fmt.Errorf("%ssequence %d: exceeds path integer range", errPrefix, seq)
	}
	return int(seq), nil
}

// chainFromAPI converts a generated chain id.
func chainFromAPI(id api.VaaChainID) ChainID {
	if id < 0 || id > api.VaaChainID(math.MaxUint16) {
		return 0
	}
	return ChainID(id)
}

// uint32FromInt converts n to uint32, clamping negative values to zero.
func uint32FromInt(n int) uint32 {
	if n < 0 {
		return 0
	}
	if uint64(n) > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(n)
}

// uint8FromInt converts n to uint8, clamping out-of-range values.
func uint8FromInt(n int) uint8 {
	if n < 0 {
		return 0
	}
	if n > math.MaxUint8 {
		return math.MaxUint8
	}
	return uint8(n)
}

// uint64FromInt64 converts n to uint64, treating negatives as zero.
func uint64FromInt64(n int64) uint64 {
	if n < 0 {
		return 0
	}
	return uint64(n)
}

// uint64FromDec parses a decimal string as uint64. Invalid or empty input is zero.
func uint64FromDec(s string) uint64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// payloadFromAPI copies a generated payload map.
func payloadFromAPI(p *map[string]interface{}) map[string]any {
	if p == nil {
		return nil
	}
	return *p
}
