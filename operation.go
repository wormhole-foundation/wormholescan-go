package wormholescan

import (
	"context"
	"fmt"
	"iter"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormholescan-go/api"
)

const (
	// opGetOperationByID is the get-operation-by-id operation id.
	opGetOperationByID = "get-operation-by-id"
	// opGetOperations is the get-operations operation id.
	opGetOperations = "get-operations"
	// opSearchOperations is the search-operations operation id.
	opSearchOperations = "search-operations"
	// opFindRelay is the find-relay-by-vaa-id operation id.
	opFindRelay = "find-relay-by-vaa-id"
	// opFindGlobalTx is the find-global-transaction-by-id operation id.
	opFindGlobalTx = "find-global-transaction-by-id"
	// chainListSep joins chain ids in list query parameters.
	chainListSep = ","
)

// GetOperation fetches one operation by VAA id (get-operation-by-id).
func (c *Client) GetOperation(ctx context.Context, id VAAID) (Operation, error) {
	chain, emitter, seq, err := vaaIDPath(id)
	if err != nil {
		return Operation{}, err
	}
	rsp, err := c.api.GetOperationByIdWithResponse(ctx, chain, emitter, seq)
	if err != nil {
		return Operation{}, wrapCall(opGetOperationByID, err)
	}
	if err := checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return Operation{}, err
	}
	if rsp.JSON200 == nil {
		return Operation{}, emptyDecode(opGetOperationByID)
	}
	return fromAPIOperation(*rsp.JSON200), nil
}

// ListOperations fetches one page of operations (get-operations).
//
// The list response has no pagination envelope; PageSize is taken from opts.
func (c *Client) ListOperations(ctx context.Context, opts OperationListOptions) (Page[Operation], error) {
	rsp, err := c.api.GetOperationsWithResponse(ctx, operationListParams(opts))
	if err != nil {
		return Page[Operation]{}, wrapCall(opGetOperations, err)
	}
	if err := checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return Page[Operation]{}, err
	}
	if rsp.JSON200 == nil {
		return Page[Operation]{}, emptyDecode(opGetOperations)
	}
	return Page[Operation]{
		Items:    fromAPIOperations(deref(rsp.JSON200.Operations)),
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// Operations walks operations page by page starting at opts.Page until the last page.
func (c *Client) Operations(ctx context.Context, opts OperationListOptions) iter.Seq2[Operation, error] {
	return paginate(ctx, opts.PageOptions, func(ctx context.Context, page PageOptions) (Page[Operation], error) {
		call := opts
		call.PageOptions = page
		return c.ListOperations(ctx, call)
	})
}

// SearchOperationsByTxHashes looks up operations by source transaction hashes (search-operations).
//
// This is a POST and is not retried.
func (c *Client) SearchOperationsByTxHashes(ctx context.Context, hashes []TxHash) ([]Operation, error) {
	body := make([]string, len(hashes))
	for i, h := range hashes {
		body[i] = string(h)
	}
	rsp, err := c.api.SearchOperationsWithResponse(ctx, body)
	if err != nil {
		return nil, wrapCall(opSearchOperations, err)
	}
	if err := checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return nil, err
	}
	if rsp.JSON200 == nil {
		return nil, emptyDecode(opSearchOperations)
	}
	return fromAPIOperations(*rsp.JSON200), nil
}

// GetRelay fetches Wormhole Relayer delivery info for a VAA (find-relay-by-vaa-id).
func (c *Client) GetRelay(ctx context.Context, id VAAID) (Relay, error) {
	chain, emitter, seq, err := vaaIDPath(id)
	if err != nil {
		return Relay{}, err
	}
	rsp, err := c.api.FindRelayByVaaIdWithResponse(ctx, chain, emitter, seq)
	if err != nil {
		return Relay{}, wrapCall(opFindRelay, err)
	}
	if err := checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return Relay{}, err
	}
	if rsp.JSON200 == nil {
		return Relay{}, emptyDecode(opFindRelay)
	}
	return fromAPIRelay(*rsp.JSON200), nil
}

// GetGlobalTransaction fetches the origin and destination transactions for a VAA
// (find-global-transaction-by-id). Destination is zero when the VAA is unredeemed.
func (c *Client) GetGlobalTransaction(ctx context.Context, id VAAID) (GlobalTransaction, error) {
	chain, emitter, seq, err := vaaIDPath(id)
	if err != nil {
		return GlobalTransaction{}, err
	}
	rsp, err := c.api.FindGlobalTransactionByIdWithResponse(ctx, chain, emitter, seq)
	if err != nil {
		return GlobalTransaction{}, wrapCall(opFindGlobalTx, err)
	}
	if err := checkResponse(rsp.HTTPResponse, rsp.Body); err != nil {
		return GlobalTransaction{}, err
	}
	if rsp.JSON200 == nil {
		return GlobalTransaction{}, emptyDecode(opFindGlobalTx)
	}
	return fromAPIGlobalTransaction(*rsp.JSON200), nil
}

// operationListParams maps list options onto generated query parameters.
func operationListParams(opts OperationListOptions) *api.GetOperationsParams {
	page := opts.Page
	params := &api.GetOperationsParams{Page: &page}
	if opts.PageSize > 0 {
		pageSize := opts.PageSize
		params.PageSize = &pageSize
	}
	if opts.Address != "" {
		params.Address = &opts.Address
	}
	if opts.TxHash != "" {
		h := string(opts.TxHash)
		params.TxHash = &h
	}
	if s := joinChainIDs(opts.SourceChains); s != "" {
		params.SourceChain = &s
	}
	if s := joinChainIDs(opts.TargetChains); s != "" {
		params.TargetChain = &s
	}
	if s := joinChainIDs(opts.IncludesChains); s != "" {
		params.IncludesChain = &s
	}
	if opts.AppID != "" {
		params.AppId = &opts.AppID
	}
	if opts.ExclusiveAppID {
		exclusive := opts.ExclusiveAppID
		params.ExclusiveAppId = &exclusive
	}
	if opts.MinAmount != 0 {
		minAmount := float32(opts.MinAmount)
		params.MinAmount = &minAmount
	}
	if opts.AddressType != "" {
		addressType := api.GetOperationsParamsAddressType(opts.AddressType)
		params.AddressType = &addressType
	}
	if !opts.From.IsZero() {
		from := opts.From.UTC().Format(time.RFC3339)
		params.From = &from
	}
	if !opts.To.IsZero() {
		to := opts.To.UTC().Format(time.RFC3339)
		params.To = &to
	}
	return params
}

// joinChainIDs encodes chain ids as a comma-separated query value.
func joinChainIDs(ids []ChainID) string {
	if len(ids) == 0 {
		return ""
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatUint(uint64(id), 10)
	}
	return strings.Join(parts, chainListSep)
}

// vaaIDPath converts a VAA id into generated path parameters.
func vaaIDPath(id VAAID) (int, string, int, error) {
	if id.Sequence > uint64(math.MaxInt) {
		return 0, "", 0, fmt.Errorf("%sVAA sequence %d exceeds path parameter range", errPrefix, id.Sequence)
	}
	return int(id.Chain), string(id.Emitter), int(id.Sequence), nil //nolint:gosec // sequence checked against MaxInt; chain is uint16
}

// wrapCall prefixes a generated-client error with the operation id.
func wrapCall(op string, err error) error {
	return fmt.Errorf("%s%s: %w", errPrefix, op, err)
}

// emptyDecode returns a decode error when a 200 response had no JSON body.
func emptyDecode(op string) error {
	return fmt.Errorf("%sdecode %s: empty body", errPrefix, op)
}

// fromAPIOperations maps a generated operation slice.
func fromAPIOperations(in []api.OperationsOperationResponse) []Operation {
	if in == nil {
		return nil
	}
	out := make([]Operation, len(in))
	for i := range in {
		out[i] = fromAPIOperation(in[i])
	}
	return out
}

// fromAPIOperation maps a generated operation document.
func fromAPIOperation(in api.OperationsOperationResponse) Operation {
	return Operation{
		ID:             deref(in.Id),
		EmitterChain:   chainIDFromAPI(in.EmitterChain),
		EmitterAddress: fromAPIOperationEmitter(in.EmitterAddress),
		Sequence:       deref(in.Sequence),
		Content:        fromAPIOperationContent(in.Content),
		Data:           fromAPIOperationData(in.Data),
		SourceChain:    fromAPISourceChain(in.SourceChain),
		TargetChain:    fromAPITargetChain(in.TargetChain),
		VAA:            fromAPIOperationVAA(in.Vaa),
	}
}

// fromAPIOperationEmitter maps a generated emitter address.
func fromAPIOperationEmitter(p *api.OperationsEmitterAddress) OperationEmitter {
	if p == nil {
		return OperationEmitter{}
	}
	return OperationEmitter{
		Hex:    EmitterAddress(deref(p.Hex)),
		Native: deref(p.Native),
	}
}

// fromAPIOperationVAA maps a generated operation VAA.
func fromAPIOperationVAA(p *api.OperationsVaa) OperationVAA {
	if p == nil {
		return OperationVAA{}
	}
	return OperationVAA{
		GuardianSetIndex: deref(p.GuardianSetIndex),
		IsDuplicated:     deref(p.IsDuplicated),
		Raw:              deref(p.Raw),
	}
}

// fromAPIOperationContent maps generated operation content.
func fromAPIOperationContent(p *api.OperationsContent) OperationContent {
	if p == nil {
		return OperationContent{}
	}
	return OperationContent{
		ExecutorRequest:        fromAPIExecutorRequest(p.ExecutorRequest),
		Payload:                fromAPIAnyMap(p.Payload),
		StandardizedProperties: fromAPIStandardizedProperties(p.StandarizedProperties),
	}
}

// fromAPIStandardizedProperties maps generated standardized properties.
func fromAPIStandardizedProperties(p *api.OperationsStandardizedProperties) StandardizedProperties {
	if p == nil {
		return StandardizedProperties{}
	}
	return StandardizedProperties{
		Amount:             deref(p.Amount),
		AppIDs:             deref(p.AppIds),
		Fee:                deref(p.Fee),
		FeeAddress:         deref(p.FeeAddress),
		FeeChain:           chainIDFromAPI(p.FeeChain),
		FromAddress:        deref(p.FromAddress),
		FromChain:          chainIDFromAPI(p.FromChain),
		NormalizedDecimals: deref(p.NormalizedDecimals),
		ToAddress:          deref(p.ToAddress),
		ToChain:            chainIDFromAPI(p.ToChain),
		TokenAddress:       deref(p.TokenAddress),
		TokenChain:         chainIDFromAPI(p.TokenChain),
	}
}

// fromAPIOperationData maps generated token-transfer display data.
func fromAPIOperationData(p *map[string]interface{}) OperationData {
	if p == nil {
		return OperationData{}
	}
	m := *p
	return OperationData{
		Symbol:      anyString(m["symbol"]),
		TokenAmount: anyString(m["tokenAmount"]),
		UsdAmount:   anyString(m["usdAmount"]),
	}
}

// fromAPISourceChain maps generated source-chain activity.
func fromAPISourceChain(p *api.OperationsSourceChain) SourceChain {
	if p == nil {
		return SourceChain{}
	}
	return SourceChain{
		Attribute:        fromAPIAttribute(p.Attribute),
		BalanceChanges:   fromAPIBalanceChanges(p.BalanceChanges),
		ChainID:          chainIDFromAPI(p.ChainId),
		Fee:              deref(p.Fee),
		FeeUSD:           deref(p.FeeUSD),
		From:             deref(p.From),
		GasTokenNotional: deref(p.GasTokenNotional),
		IsSolanaShim:     deref(p.IsSolanaShim),
		Status:           TxStatus(deref(p.Status)),
		Timestamp:        derefTime(p.Timestamp),
		Transaction:      fromAPIOperationTx(p.Transaction),
	}
}

// fromAPITargetChain maps generated target-chain activity.
func fromAPITargetChain(p *api.OperationsTargetChain) TargetChain {
	if p == nil {
		return TargetChain{}
	}
	return TargetChain{
		BalanceChanges:   fromAPIBalanceChanges(p.BalanceChanges),
		ChainID:          chainIDFromAPI(p.ChainId),
		Fee:              deref(p.Fee),
		FeeUSD:           deref(p.FeeUSD),
		From:             deref(p.From),
		GasTokenNotional: deref(p.GasTokenNotional),
		Status:           TxStatus(deref(p.Status)),
		Timestamp:        derefTime(p.Timestamp),
		To:               deref(p.To),
		Transaction:      fromAPIOperationTx(p.Transaction),
	}
}

// fromAPIOperationTx maps a generated operation transaction.
func fromAPIOperationTx(p *api.OperationsTransaction) OperationTx {
	if p == nil {
		return OperationTx{}
	}
	return OperationTx{
		TxHash:       TxHash(deref(p.TxHash)),
		SecondTxHash: TxHash(deref(p.SecondTxHash)),
	}
}

// fromAPIAttribute maps a generated operations.Data attribute.
func fromAPIAttribute(p *api.OperationsData) Attribute {
	if p == nil {
		return Attribute{}
	}
	return Attribute{
		Type:  deref(p.Type),
		Value: fromAPIAnyMap(p.Value),
	}
}

// fromAPITxAttribute maps a generated transaction attribute.
func fromAPITxAttribute(p *api.TransactionsAttributeDoc) Attribute {
	if p == nil {
		return Attribute{}
	}
	return Attribute{
		Type:  deref(p.Type),
		Value: fromAPIAnyMap(p.Value),
	}
}

// fromAPIBalanceChanges maps generated balance changes.
func fromAPIBalanceChanges(
	p *[]api.GithubComWormholeFoundationWormholeExplorerApiRoutesWormscanOperationsBalanceChanges,
) []BalanceChange {
	if p == nil {
		return nil
	}
	out := make([]BalanceChange, len(*p))
	for i, bc := range *p {
		out[i] = BalanceChange{
			Amount:       deref(bc.Amount),
			Recipient:    deref(bc.Recipient),
			TokenAddress: deref(bc.TokenAddress),
		}
	}
	return out
}

// fromAPIGlobalTransaction maps a generated global transaction document.
func fromAPIGlobalTransaction(in api.TransactionsGlobalTransactionDoc) GlobalTransaction {
	return GlobalTransaction{
		ID:          deref(in.Id),
		Origin:      fromAPIOriginTx(in.OriginTx),
		Destination: fromAPIDestinationTx(in.DestinationTx),
	}
}

// fromAPIOriginTx maps a generated origin transaction.
func fromAPIOriginTx(p *api.TransactionsOriginTx) OriginTx {
	if p == nil {
		return OriginTx{}
	}
	return OriginTx{
		Attribute: fromAPITxAttribute(p.Attribute),
		From:      deref(p.From),
		Status:    TxStatus(deref(p.Status)),
		TxHash:    TxHash(deref(p.TxHash)),
	}
}

// fromAPIDestinationTx maps a generated destination transaction.
func fromAPIDestinationTx(p *api.TransactionsDestinationTx) DestinationTx {
	if p == nil {
		return DestinationTx{}
	}
	return DestinationTx{
		BlockNumber: deref(p.BlockNumber),
		ChainID:     chainIDFromAPI(p.ChainId),
		From:        deref(p.From),
		Method:      deref(p.Method),
		Status:      TxStatus(deref(p.Status)),
		Timestamp:   derefTime(p.Timestamp),
		To:          deref(p.To),
		TxHash:      TxHash(deref(p.TxHash)),
		UpdatedAt:   derefTime(p.UpdatedAt),
	}
}

// fromAPIRelay maps a generated relay document.
func fromAPIRelay(in api.RelaysRelayResponse) Relay {
	return Relay{
		ID:          deref(in.Id),
		Relayer:     deref(in.Relayer),
		Status:      RelayStatus(deref(in.Status)),
		ReceivedAt:  derefTime(in.ReceivedAt),
		CompletedAt: derefTime(in.CompletedAt),
		FailedAt:    derefTime(in.FailedAt),
		Data:        fromAPIRelayData(in.Data),
	}
}

// fromAPIRelayData maps generated relay data.
func fromAPIRelayData(p *api.RelaysRelayDataResponse) RelayData {
	if p == nil {
		return RelayData{}
	}
	return RelayData{
		FromTxHash:   TxHash(deref(p.FromTxHash)),
		ToTxHash:     TxHash(deref(p.ToTxHash)),
		MaxAttempts:  deref(p.MaxAttempts),
		Instructions: fromAPIRelayInstructions(p.Instructions),
		Delivery:     fromAPIRelayDelivery(p.Delivery),
	}
}

// fromAPIRelayInstructions maps generated relay instructions.
func fromAPIRelayInstructions(p *api.RelaysInstructionsResponse) RelayInstructions {
	if p == nil {
		return RelayInstructions{}
	}
	out := RelayInstructions{
		EncodedExecutionInfo:   deref(p.EncodedExecutionInfo),
		RefundAddress:          deref(p.RefundAddress),
		RefundChainID:          chainIDFromInt(p.RefundChainId),
		RefundDeliveryProvider: deref(p.RefundDeliveryProvider),
		SenderAddress:          deref(p.SenderAddress),
		SourceDeliveryProvider: deref(p.SourceDeliveryProvider),
		TargetAddress:          deref(p.TargetAddress),
		TargetChainID:          chainIDFromInt(p.TargetChainId),
		VaaKeys:                deref(p.VaaKeys),
	}
	if p.ExtraReceiverValue != nil {
		out.ExtraReceiverValue = RelayBigNumber{
			Hex:         deref(p.ExtraReceiverValue.UnderscoreHex),
			IsBigNumber: deref(p.ExtraReceiverValue.UnderscoreIsBigNumber),
		}
	}
	if p.RequestedReceiverValue != nil {
		out.RequestedReceiverValue = RelayBigNumber{
			Hex:         deref(p.RequestedReceiverValue.UnderscoreHex),
			IsBigNumber: deref(p.RequestedReceiverValue.UnderscoreIsBigNumber),
		}
	}
	return out
}

// fromAPIRelayDelivery maps generated relay delivery.
func fromAPIRelayDelivery(p *api.RelaysDeliveryReponse) RelayDelivery {
	if p == nil {
		return RelayDelivery{}
	}
	return RelayDelivery{
		Budget:              deref(p.Budget),
		MaxRefund:           deref(p.MaxRefund),
		RelayGasUsed:        deref(p.RelayGasUsed),
		TargetChainDecimals: deref(p.TargetChainDecimals),
		Execution:           fromAPIRelayExecution(p.Execution),
	}
}

// fromAPIRelayExecution maps generated relay execution.
func fromAPIRelayExecution(p *api.RelaysResultExecutionResponse) RelayExecution {
	if p == nil {
		return RelayExecution{}
	}
	return RelayExecution{
		Detail:          deref(p.Detail),
		GasUsed:         deref(p.GasUsed),
		RefundStatus:    deref(p.RefundStatus),
		RevertString:    deref(p.RevertString),
		Status:          deref(p.Status),
		TransactionHash: TxHash(deref(p.TransactionHash)),
	}
}

// chainIDFromAPI converts a generated chain id pointer.
func chainIDFromAPI(id *api.VaaChainID) ChainID {
	if id == nil {
		return 0
	}
	return ChainID(*id) //nolint:gosec // Wormhole chain ids fit uint16
}

// chainIDFromInt converts a generated integer chain id pointer.
func chainIDFromInt(id *int) ChainID {
	if id == nil {
		return 0
	}
	return ChainID(*id) //nolint:gosec // Wormhole chain ids fit uint16
}

// fromAPIAnyMap copies a generated additionalProperties object.
func fromAPIAnyMap(p *map[string]interface{}) map[string]any {
	if p == nil {
		return nil
	}
	return *p
}

// fromAPIExecutorRequest copies an untyped executor request payload.
func fromAPIExecutorRequest(v any) map[string]any {
	switch x := v.(type) {
	case nil:
		return nil
	case map[string]any:
		return x
	default:
		return map[string]any{"value": x}
	}
}

// anyString returns v if it is a string, or empty otherwise.
func anyString(v any) string {
	s, _ := v.(string)
	return s
}
