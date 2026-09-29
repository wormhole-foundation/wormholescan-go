package wormholescan

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormholescan-go/api"
)

// GovernorConfig is one guardian's governor configuration.
type GovernorConfig struct {
	// GuardianAddress is the guardian id as the server sends it (no 0x prefix).
	GuardianAddress GuardianAddress
	// NodeName is the guardian node name.
	NodeName string
	// Counter is the config generation counter.
	Counter int
	// Chains is the per-chain notional configuration.
	Chains []GovernorChainConfig
	// Tokens is the priced token list this guardian reports.
	Tokens []GovernorToken
	// CreatedAt is when the server first stored this config.
	CreatedAt time.Time
	// UpdatedAt is when the server last updated this config.
	UpdatedAt time.Time
}

// GovernorChainConfig is the notional limit configuration for one chain.
type GovernorChainConfig struct {
	// Chain is the Wormhole chain this limit applies to.
	Chain ChainID
	// NotionalLimit is the chain notional limit as an integer.
	NotionalLimit uint64
	// BigTransactionSize is the large-transaction threshold as an integer.
	BigTransactionSize uint64
}

// GovernorToken is a priced origin token in a governor config.
type GovernorToken struct {
	// OriginChain is the token's origin chain.
	OriginChain ChainID
	// OriginAddress is the token address on OriginChain.
	OriginAddress string
	// Price is the token price the governor uses.
	Price float64
}

// GovernorConfigListOptions filters a page of governor configs.
type GovernorConfigListOptions struct {
	// PageOptions selects the page of configs to return.
	PageOptions
}

// GovernorStatus is one guardian's live governor status.
type GovernorStatus struct {
	// GuardianAddress is the guardian id as the server sends it (no 0x prefix).
	GuardianAddress GuardianAddress
	// NodeName is the guardian node name.
	NodeName string
	// Chains is the per-chain remaining notional and enqueued VAAs.
	Chains []GovernorStatusChain
	// CreatedAt is when the server first stored this status.
	CreatedAt time.Time
	// UpdatedAt is when the server last updated this status.
	UpdatedAt time.Time
}

// GovernorStatusChain is remaining notional and emitters for one chain.
type GovernorStatusChain struct {
	// Chain is the Wormhole chain.
	Chain ChainID
	// RemainingAvailableNotional is leftover notional capacity as an integer.
	RemainingAvailableNotional uint64
	// Emitters is the per-emitter enqueue state on this chain.
	Emitters []GovernorStatusEmitter
}

// GovernorStatusEmitter is enqueue state for one emitter on a chain.
type GovernorStatusEmitter struct {
	// Emitter is the emitter address as the server sends it.
	Emitter EmitterAddress
	// TotalEnqueuedVAAs is the number of VAAs waiting on this emitter.
	TotalEnqueuedVAAs int
	// EnqueuedVAAs is the waiting VAAs. Empty when the server sends null.
	EnqueuedVAAs []EnqueuedVAA
}

// GovernorStatusListOptions filters a page of governor statuses.
type GovernorStatusListOptions struct {
	// PageOptions selects the page of statuses to return.
	PageOptions
}

// GovernorLimit is the aggregated notional limit for one chain.
type GovernorLimit struct {
	// Chain is the Wormhole chain.
	Chain ChainID
	// AvailableNotional is remaining notional capacity as an integer.
	AvailableNotional uint64
	// NotionalLimit is the chain notional limit as an integer.
	NotionalLimit uint64
	// MaxTransactionSize is the large-transaction threshold as an integer.
	MaxTransactionSize uint64
}

// GovernorLimitListOptions filters a page of aggregated governor limits.
type GovernorLimitListOptions struct {
	// PageOptions selects the page of limits to return.
	PageOptions
}

// NotionalLimitDetail is a per-guardian notional limit for one chain.
type NotionalLimitDetail struct {
	// GuardianAddress is the guardian id when the server sends it.
	GuardianAddress GuardianAddress
	// NodeName is the guardian node name when the server sends it.
	NodeName string
	// Chain is the Wormhole chain.
	Chain ChainID
	// NotionalLimit is the chain notional limit as an integer.
	NotionalLimit uint64
	// MaxTransactionSize is the large-transaction threshold as an integer.
	MaxTransactionSize uint64
	// CreatedAt is when the server first stored this row. Zero when omitted.
	CreatedAt time.Time
	// UpdatedAt is when the server last updated this row. Zero when omitted.
	UpdatedAt time.Time
}

// NotionalLimitListOptions filters a page of notional limit details.
type NotionalLimitListOptions struct {
	// PageOptions selects the page of limit details to return.
	PageOptions
}

// NotionalAvailable is remaining notional capacity for one chain.
type NotionalAvailable struct {
	// Chain is the Wormhole chain.
	Chain ChainID
	// AvailableNotional is remaining notional capacity as an integer.
	AvailableNotional uint64
}

// NotionalAvailableDetail is a per-guardian remaining notional for one chain.
type NotionalAvailableDetail struct {
	// GuardianAddress is the guardian id as the server sends it.
	GuardianAddress GuardianAddress
	// NodeName is the guardian node name.
	NodeName string
	// Chain is the Wormhole chain.
	Chain ChainID
	// AvailableNotional is remaining notional capacity as an integer.
	AvailableNotional uint64
	// CreatedAt is when the server first stored this row.
	CreatedAt time.Time
	// UpdatedAt is when the server last updated this row.
	UpdatedAt time.Time
}

// NotionalAvailableListOptions filters a page of notional availability.
type NotionalAvailableListOptions struct {
	// PageOptions selects the page of availability rows to return.
	PageOptions
}

// MaxNotionalAvailable is the guardian with the most remaining notional on a chain.
type MaxNotionalAvailable struct {
	// GuardianAddress is the guardian id as the server sends it.
	GuardianAddress GuardianAddress
	// NodeName is the guardian node name.
	NodeName string
	// Chain is the Wormhole chain.
	Chain ChainID
	// AvailableNotional is remaining notional capacity as an integer.
	AvailableNotional uint64
	// Emitters is the enqueue state this guardian reports for the chain.
	Emitters []GovernorStatusEmitter
	// CreatedAt is when the server first stored this row.
	CreatedAt time.Time
	// UpdatedAt is when the server last updated this row.
	UpdatedAt time.Time
}

// EnqueuedVAA is a VAA waiting in the governor.
type EnqueuedVAA struct {
	// Chain is the emitter chain. Zero when the nested status payload omits it.
	Chain ChainID
	// Emitter is the emitter address as the server sends it.
	Emitter EmitterAddress
	// NotionalValue is the notional value as an integer.
	NotionalValue uint64
	// Sequence is the VAA sequence.
	Sequence uint64
	// TxHash is the source transaction hash as the server sends it.
	TxHash TxHash
	// ReleaseTime is when the VAA may be released. Zero when the server has no value.
	ReleaseTime time.Time
}

// EnqueuedVAAListOptions filters a page of enqueued VAAs.
type EnqueuedVAAListOptions struct {
	// PageOptions selects the page of enqueued VAAs to return.
	PageOptions
}

// GovernorVAA is a VAA currently tracked by the governor.
type GovernorVAA struct {
	// VAAID is the chain/emitter/sequence identifier as the server sends it.
	VAAID string
	// Chain is the emitter chain.
	Chain ChainID
	// Emitter is the emitter address as the server sends it.
	Emitter EmitterAddress
	// Sequence is the VAA sequence as a decimal string.
	Sequence string
	// TxHash is the source transaction hash as the server sends it.
	TxHash TxHash
	// ReleaseTime is when the VAA may be released. Zero when the server has no value.
	ReleaseTime time.Time
	// Amount is the notional amount as an integer.
	Amount uint64
	// Status is the governor VAA status. Other values may appear.
	Status GovernorVAAStatus
}

// GovernorVAAStatus is a governor VAA status string. Other values may appear.
type GovernorVAAStatus string

const (
	// GovernorVAAStatusIssued is a VAA the governor has issued.
	GovernorVAAStatusIssued GovernorVAAStatus = "issued"
	// GovernorVAAStatusPending is a VAA still held by the governor.
	GovernorVAAStatusPending GovernorVAAStatus = "pending"
)

// GovernorVAAListOptions filters a page of governor VAAs.
type GovernorVAAListOptions struct {
	// PageOptions selects the page of governor VAAs to return.
	PageOptions
}

// fromAPIGovernorConfig maps a generated governor config document.
func fromAPIGovernorConfig(src api.GovernorGovConfig) GovernorConfig {
	return GovernorConfig{
		GuardianAddress: GuardianAddress(deref(src.Id)),
		NodeName:        deref(src.NodeName),
		Counter:         deref(src.Counter),
		Chains:          mapSlice(deref(src.Chains), fromAPIGovernorChainConfig),
		Tokens:          mapSlice(deref(src.Tokens), fromAPIGovernorToken),
		CreatedAt:       derefTime(src.CreatedAt),
		UpdatedAt:       derefTime(src.UpdatedAt),
	}
}

// fromAPIGovernorChainConfig maps a generated per-chain config row.
func fromAPIGovernorChainConfig(src api.GovernorGovConfigChains) GovernorChainConfig {
	return GovernorChainConfig{
		Chain:              chainIDFromAPI(src.ChainId),
		NotionalLimit:      uint64FromInt(src.NotionalLimit),
		BigTransactionSize: uint64FromInt(src.BigTransactionSize),
	}
}

// fromAPIGovernorToken maps a generated priced token.
func fromAPIGovernorToken(src api.GovernorGovConfigfTokens) GovernorToken {
	return GovernorToken{
		OriginChain:   chainIDFromInt(src.OriginChainId),
		OriginAddress: deref(src.OriginAddress),
		Price:         deref(src.Price),
	}
}

// fromAPIGovernorStatus maps a generated governor status document.
func fromAPIGovernorStatus(src api.GovernorGovStatus) GovernorStatus {
	return GovernorStatus{
		GuardianAddress: GuardianAddress(deref(src.Id)),
		NodeName:        deref(src.NodeName),
		Chains:          mapSlice(deref(src.Chains), fromAPIGovernorStatusChain),
		CreatedAt:       derefTime(src.CreatedAt),
		UpdatedAt:       derefTime(src.UpdatedAt),
	}
}

// fromAPIGovernorStatusChain maps a generated per-chain status row.
func fromAPIGovernorStatusChain(src api.GovernorGovStatusChains) GovernorStatusChain {
	chain := chainIDFromAPI(src.ChainId)
	return GovernorStatusChain{
		Chain:                      chain,
		RemainingAvailableNotional: uint64FromInt(src.RemainingAvailableNotional),
		Emitters: mapSlice(deref(src.Emitters), func(em api.GovernorGovStatusChainEmitter) GovernorStatusEmitter {
			return fromAPIStatusEmitter(em, chain)
		}),
	}
}

// fromAPIStatusEmitter maps a generated status emitter, including enqueued VAAs.
func fromAPIStatusEmitter(src api.GovernorGovStatusChainEmitter, chain ChainID) GovernorStatusEmitter {
	addr := EmitterAddress(deref(src.EmitterAddress))
	vaas := fromAPIStatusEnqueuedVAAs(src.EnqueuedVaas)
	for i := range vaas {
		vaas[i].Chain = chain
		if vaas[i].Emitter == "" {
			vaas[i].Emitter = addr
		}
	}
	return GovernorStatusEmitter{
		Emitter:           addr,
		TotalEnqueuedVAAs: deref(src.TotalEnqueuedVaas),
		EnqueuedVAAs:      vaas,
	}
}

// fromAPIStatusEnqueuedVAAs maps status enqueued VAAs from the generated any value.
//
// Live responses send either null, an array of objects, or a MongoDB Key/Value dump.
func fromAPIStatusEnqueuedVAAs(raw any) []EnqueuedVAA {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]EnqueuedVAA, 0, len(items))
	for _, item := range items {
		if vaa, ok := fromAPIStatusEnqueuedVAA(item); ok {
			out = append(out, vaa)
		}
	}
	return out
}

// fromAPIStatusEnqueuedVAA maps one status enqueued VAA from an object or BSON dump.
func fromAPIStatusEnqueuedVAA(item any) (EnqueuedVAA, bool) {
	switch v := item.(type) {
	case map[string]any:
		return enqueuedVAAFromObject(v), true
	case []any:
		return enqueuedVAAFromBSON(v), true
	default:
		return EnqueuedVAA{}, false
	}
}

// enqueuedVAAFromBSON maps a live MongoDB Key/Value dump of one enqueued VAA.
func enqueuedVAAFromBSON(pairs []any) EnqueuedVAA {
	obj := make(map[string]any, len(pairs))
	for _, pair := range pairs {
		m, ok := pair.(map[string]any)
		if !ok {
			continue
		}
		key, _ := m["Key"].(string)
		obj[strings.ToLower(key)] = m["Value"]
	}
	return enqueuedVAAFromObject(obj)
}

// enqueuedVAAFromObject maps a JSON object of enqueued VAA fields.
func enqueuedVAAFromObject(obj map[string]any) EnqueuedVAA {
	return EnqueuedVAA{
		NotionalValue: anyUint64(lookupAny(obj, "notionalValue", "notionalvalue")),
		Sequence:      anyUint64(lookupAny(obj, "sequence")),
		TxHash:        TxHash(anyString(lookupAny(obj, "txHash", "txhash"))),
		ReleaseTime:   anyTime(lookupAny(obj, "releaseTime", "releasetime")),
		Emitter:       EmitterAddress(anyString(lookupAny(obj, "emitterAddress", "emitteraddress"))),
		Chain:         chainIDFromUint64(anyUint64(lookupAny(obj, "chainId", "chainid"))),
	}
}

// fromAPIGovernorLimit maps a generated aggregated governor limit.
func fromAPIGovernorLimit(src api.GovernorGovernorLimit) GovernorLimit {
	return GovernorLimit{
		Chain:              chainIDFromAPI(src.ChainId),
		AvailableNotional:  uint64FromInt(src.AvailableNotional),
		NotionalLimit:      uint64FromInt(src.NotionalLimit),
		MaxTransactionSize: uint64FromInt(src.MaxTransactionSize),
	}
}

// fromAPINotionalLimitDetail maps a generated per-guardian notional limit.
func fromAPINotionalLimitDetail(src api.GovernorNotionalLimitDetail) NotionalLimitDetail {
	return NotionalLimitDetail{
		GuardianAddress:    GuardianAddress(deref(src.Id)),
		NodeName:           deref(src.NodeName),
		Chain:              chainIDFromAPI(src.ChainId),
		NotionalLimit:      uint64FromInt(src.NotionalLimit),
		MaxTransactionSize: uint64FromInt(src.MaxTransactionSize),
		CreatedAt:          derefTime(src.CreatedAt),
		UpdatedAt:          derefTime(src.UpdatedAt),
	}
}

// fromAPINotionalAvailable maps a generated chain availability row.
func fromAPINotionalAvailable(src api.GovernorNotionalAvailable) NotionalAvailable {
	return NotionalAvailable{
		Chain:             chainIDFromAPI(src.ChainId),
		AvailableNotional: uint64FromInt(src.AvailableNotional),
	}
}

// fromAPINotionalAvailableDetail maps a generated per-guardian availability row.
func fromAPINotionalAvailableDetail(src api.GovernorNotionalAvailableDetail) NotionalAvailableDetail {
	return NotionalAvailableDetail{
		GuardianAddress:   GuardianAddress(deref(src.Id)),
		NodeName:          deref(src.NodeName),
		Chain:             chainIDFromAPI(src.ChainId),
		AvailableNotional: uint64FromInt(src.AvailableNotional),
		CreatedAt:         derefTime(src.CreatedAt),
		UpdatedAt:         derefTime(src.UpdatedAt),
	}
}

// fromAPIMaxNotionalAvailable maps a generated max-available record.
func fromAPIMaxNotionalAvailable(src api.GovernorMaxNotionalAvailableRecord) MaxNotionalAvailable {
	chain := chainIDFromAPI(src.ChainId)
	return MaxNotionalAvailable{
		GuardianAddress:   GuardianAddress(deref(src.Id)),
		NodeName:          deref(src.NodeName),
		Chain:             chain,
		AvailableNotional: uint64FromInt(src.AvailableNotional),
		Emitters: mapSlice(deref(src.Emitters), func(em api.GovernorEmitter) GovernorStatusEmitter {
			return fromAPIGovernorEmitter(em, chain)
		}),
		CreatedAt: derefTime(src.CreatedAt),
		UpdatedAt: derefTime(src.UpdatedAt),
	}
}

// fromAPIGovernorEmitter maps a generated emitter with typed enqueued VAAs.
func fromAPIGovernorEmitter(src api.GovernorEmitter, chain ChainID) GovernorStatusEmitter {
	addr := EmitterAddress(deref(src.EmitterAddress))
	vaas := mapSlice(deref(src.EnqueuedVaas), fromAPINestedEnqueuedVAA)
	for i := range vaas {
		vaas[i].Chain = chain
		vaas[i].Emitter = addr
	}
	return GovernorStatusEmitter{
		Emitter:           addr,
		TotalEnqueuedVAAs: deref(src.TotalEnqueuedVaas),
		EnqueuedVAAs:      vaas,
	}
}

// fromAPINestedEnqueuedVAA maps a generated nested enqueued VAA object.
func fromAPINestedEnqueuedVAA(src api.GovernorEnqueuedVAA) EnqueuedVAA {
	return EnqueuedVAA{
		NotionalValue: uint64FromInt(src.NotionalValue),
		Sequence:      uint64FromString(deref(src.Sequence)),
		TxHash:        TxHash(deref(src.TxHash)),
		ReleaseTime:   derefTime(src.ReleaseTime),
	}
}

// fromAPIEnqueuedVaa maps a generated enqueued VAA list item.
func fromAPIEnqueuedVaa(src api.GovernorEnqueuedVaa) EnqueuedVAA {
	return EnqueuedVAA{
		Chain:         chainIDFromAPI(src.ChainId),
		Emitter:       EmitterAddress(deref(src.EmitterAddress)),
		NotionalValue: uint64FromInt(src.NotionalValue),
		Sequence:      uint64FromInt64(src.Sequence),
		TxHash:        TxHash(deref(src.TxHash)),
	}
}

// fromAPIEnqueuedVaaDetail maps a generated per-chain enqueued VAA.
func fromAPIEnqueuedVaaDetail(src api.GovernorEnqueuedVaaDetail) EnqueuedVAA {
	return EnqueuedVAA{
		Chain:         chainIDFromAPI(src.ChainId),
		Emitter:       EmitterAddress(deref(src.EmitterAddress)),
		NotionalValue: uint64FromInt(src.NotionalValue),
		Sequence:      uint64FromInt64(src.Sequence),
		TxHash:        TxHash(deref(src.TxHash)),
		ReleaseTime:   unixTime(src.ReleaseTime),
	}
}

// fromAPIEnqueuedVaaGroups flattens per-chain enqueued VAA groups.
func fromAPIEnqueuedVaaGroups(groups []api.GovernorEnqueuedVaas) []EnqueuedVAA {
	var out []EnqueuedVAA
	for _, group := range groups {
		out = append(out, mapSlice(deref(group.EnqueuedVaas), fromAPIEnqueuedVaa)...)
	}
	return out
}

// fromAPIGovernorVAA maps a generated governor VAA.
func fromAPIGovernorVAA(src api.GovernorGovernorVaasResponse) GovernorVAA {
	return GovernorVAA{
		VAAID:       deref(src.VaaId),
		Chain:       chainIDFromAPI(src.ChainId),
		Emitter:     EmitterAddress(deref(src.EmitterAddress)),
		Sequence:    deref(src.Sequence),
		TxHash:      TxHash(deref(src.TxHash)),
		ReleaseTime: derefTime(src.ReleaseTime),
		Amount:      uint64FromInt(src.Amount),
		Status:      GovernorVAAStatus(deref(src.Status)),
	}
}

// mapSlice maps in through f. A nil input yields a nil output.
func mapSlice[A, B any](in []A, f func(A) B) []B {
	if in == nil {
		return nil
	}
	out := make([]B, len(in))
	for i := range in {
		out[i] = f(in[i])
	}
	return out
}

// chainIDFromAPI converts a generated chain id pointer.
func chainIDFromAPI(id *api.VaaChainID) ChainID {
	if id == nil {
		return 0
	}
	return chainIDFromInt64(int64(*id))
}

// chainIDFromInt converts a generated integer chain id pointer.
func chainIDFromInt(id *int) ChainID {
	if id == nil {
		return 0
	}
	return chainIDFromInt64(int64(*id))
}

// chainIDFromInt64 converts a signed integer to ChainID when it fits in uint16.
func chainIDFromInt64(n int64) ChainID {
	if n < 0 || n > math.MaxUint16 {
		return 0
	}
	return ChainID(uint16(n))
}

// chainIDFromUint64 converts an unsigned integer to ChainID when it fits in uint16.
func chainIDFromUint64(n uint64) ChainID {
	if n > math.MaxUint16 {
		return 0
	}
	return ChainID(uint16(n))
}

// uint64FromInt converts a generated integer pointer to uint64.
func uint64FromInt(p *int) uint64 {
	if p == nil || *p < 0 {
		return 0
	}
	return uint64(*p)
}

// uint64FromInt64 converts a generated int64 pointer to uint64.
func uint64FromInt64(p *int64) uint64 {
	if p == nil || *p < 0 {
		return 0
	}
	return uint64(*p)
}

// uint64FromString parses a decimal string as uint64. Invalid input yields 0.
func uint64FromString(s string) uint64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// unixTime converts a unix-seconds pointer to UTC. Nil or non-positive is zero.
func unixTime(p *int) time.Time {
	if p == nil || *p <= 0 {
		return time.Time{}
	}
	return time.Unix(int64(*p), 0).UTC()
}

// lookupAny returns the first present key in obj.
func lookupAny(obj map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := obj[key]; ok {
			return v
		}
	}
	return nil
}

// anyString returns v as a string, or empty when v is not a string.
func anyString(v any) string {
	s, _ := v.(string)
	return s
}

// anyUint64 converts a JSON any value to uint64.
func anyUint64(v any) uint64 {
	switch n := v.(type) {
	case float64:
		if n < 0 {
			return 0
		}
		return uint64(n)
	case string:
		return uint64FromString(n)
	default:
		return 0
	}
}

// anyTime converts a JSON any value to UTC time. Zero when absent or null.
func anyTime(v any) time.Time {
	switch t := v.(type) {
	case string:
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339Nano, t)
			if err != nil {
				return time.Time{}
			}
		}
		return parsed.UTC()
	case float64:
		if t <= 0 {
			return time.Time{}
		}
		return time.Unix(int64(t), 0).UTC()
	default:
		return time.Time{}
	}
}

// slicePage returns the opts page of items. PageSize 0 returns items unchanged.
func slicePage[T any](items []T, opts PageOptions) []T {
	if opts.PageSize <= 0 {
		return items
	}
	start := opts.Page * opts.PageSize
	if start >= len(items) {
		return []T{}
	}
	end := min(start+opts.PageSize, len(items))
	return items[start:end]
}
