package wormholescan

import "time"

// TxStatus is a source-chain, target-chain, or global-transaction status.
// Other values than the constants below may appear.
type TxStatus string

const (
	// TxStatusConfirmed means the origin transaction is confirmed.
	TxStatusConfirmed TxStatus = "confirmed"
	// TxStatusCompleted means the destination transaction completed.
	TxStatusCompleted TxStatus = "completed"
)

// AddressType selects whether [OperationListOptions.Address] matches the sender
// or the recipient. Other values than the constants below may appear.
type AddressType string

const (
	// AddressTypeFrom matches the source address.
	AddressTypeFrom AddressType = "from"
	// AddressTypeTo matches the destination address.
	AddressTypeTo AddressType = "to"
)

// RelayStatus is a relay document status. Other values may appear.
type RelayStatus string

// Operation is a Wormholescan operation: a VAA plus the source and destination
// chain activity the explorer has associated with it.
type Operation struct {
	// ID is the operation id the server sent, usually chain/emitter/sequence.
	ID string
	// EmitterChain is the VAA emitter chain.
	EmitterChain ChainID
	// EmitterAddress is the VAA emitter in hex and native encodings.
	EmitterAddress OperationEmitter
	// Sequence is the VAA sequence as a decimal string. Empty when omitted.
	Sequence string
	// Content holds parsed payload, standardized properties, and executor data.
	Content OperationContent
	// Data holds token-transfer display fields when the server includes them.
	Data OperationData
	// SourceChain is origin-chain activity. Zero when the server omits it.
	SourceChain SourceChain
	// TargetChain is destination-chain activity. Zero when the server omits it.
	TargetChain TargetChain
	// VAA is the signed VAA attached to the operation. Zero when omitted.
	VAA OperationVAA
}

// OperationEmitter is an emitter address in hex and chain-native form.
type OperationEmitter struct {
	// Hex is the 32-byte hex emitter address, without a 0x prefix when the
	// server sends it that way.
	Hex EmitterAddress
	// Native is the emitter encoded in the emitter chain's native format.
	Native string
}

// OperationVAA is the signed VAA bytes attached to an operation.
type OperationVAA struct {
	// GuardianSetIndex is the guardian set used to sign the VAA.
	GuardianSetIndex int
	// IsDuplicated reports whether the explorer treats this VAA as a duplicate.
	IsDuplicated bool
	// Raw is the signed VAA bytes. Empty when the server omits it.
	Raw []byte
}

// OperationContent is the parsed payload and standardized properties of an operation.
type OperationContent struct {
	// ExecutorRequest is the executor quote/request payload when present.
	// The OpenAPI document leaves this untyped; live payloads vary by protocol
	// (ERN1, ERC2, ...) so the value is kept as a generic object.
	ExecutorRequest map[string]any
	// Payload is the parsed VAA payload. Nil when the server omits it.
	Payload map[string]any
	// StandardizedProperties is the explorer's normalized view of the payload.
	StandardizedProperties StandardizedProperties
}

// StandardizedProperties is the explorer's normalized transfer summary.
type StandardizedProperties struct {
	// Amount is the transfer amount as a decimal string.
	Amount string
	// AppIDs are the application ids the explorer associated with the operation.
	AppIDs []string
	// Fee is the protocol fee as a decimal string. Empty when the server sends no fee.
	Fee string
	// FeeAddress is the fee recipient address. Empty when omitted.
	FeeAddress string
	// FeeChain is the chain the fee is paid on. Zero when omitted.
	FeeChain ChainID
	// FromAddress is the normalized sender address. Empty when omitted.
	FromAddress string
	// FromChain is the normalized source chain. Zero when omitted.
	FromChain ChainID
	// NormalizedDecimals is the decimal count used to normalize Amount.
	// Zero when the server omits or nulls the field.
	NormalizedDecimals int
	// ToAddress is the normalized recipient address. Empty when omitted.
	ToAddress string
	// ToChain is the normalized destination chain. Zero when omitted.
	ToChain ChainID
	// TokenAddress is the token address on TokenChain. Empty when omitted.
	TokenAddress string
	// TokenChain is the chain the token originates from. Zero when omitted.
	TokenChain ChainID
}

// OperationData is token-transfer display data when the server includes it.
type OperationData struct {
	// Symbol is the token symbol.
	Symbol string
	// TokenAmount is the amount in token units as a decimal string.
	TokenAmount string
	// UsdAmount is the USD value as a decimal string.
	UsdAmount string
}

// SourceChain is origin-chain activity for an operation.
type SourceChain struct {
	// Attribute is an optional typed extension payload. Zero when omitted.
	Attribute Attribute
	// BalanceChanges are token balance changes observed on the source chain.
	BalanceChanges []BalanceChange
	// ChainID is the origin chain.
	ChainID ChainID
	// Fee is the origin transaction fee as a decimal string.
	Fee string
	// FeeUSD is the origin fee in USD as a decimal string.
	FeeUSD string
	// From is the origin transaction sender.
	From string
	// GasTokenNotional is the gas token notional as a decimal string.
	GasTokenNotional string
	// IsSolanaShim reports whether the VAA was emitted through an SVM shim program.
	IsSolanaShim bool
	// Status is the origin transaction status. Empty when omitted.
	Status TxStatus
	// Timestamp is the origin transaction time in UTC. Zero when omitted.
	Timestamp time.Time
	// Transaction is the origin transaction hashes. Zero when omitted.
	Transaction OperationTx
}

// TargetChain is destination-chain activity for an operation.
type TargetChain struct {
	// BalanceChanges are token balance changes observed on the destination chain.
	BalanceChanges []BalanceChange
	// ChainID is the destination chain.
	ChainID ChainID
	// Fee is the destination transaction fee as a decimal string.
	Fee string
	// FeeUSD is the destination fee in USD as a decimal string.
	FeeUSD string
	// From is the destination transaction sender.
	From string
	// GasTokenNotional is the gas token notional as a decimal string.
	GasTokenNotional string
	// Status is the destination transaction status. Empty when omitted.
	Status TxStatus
	// Timestamp is the destination transaction time in UTC. Zero when omitted.
	Timestamp time.Time
	// To is the destination transaction recipient.
	To string
	// Transaction is the destination transaction hashes. Zero when omitted.
	Transaction OperationTx
}

// OperationTx is a chain transaction hash pair attached to source or target activity.
type OperationTx struct {
	// TxHash is the primary transaction hash as the server sent it.
	TxHash TxHash
	// SecondTxHash is an additional hash when the server sends one.
	SecondTxHash TxHash
}

// BalanceChange is a token balance change observed on a chain.
type BalanceChange struct {
	// Amount is the signed amount as a decimal string.
	Amount string
	// Recipient is the address whose balance changed.
	Recipient string
	// TokenAddress is the token that changed.
	TokenAddress string
}

// Attribute is a typed extension payload the server attaches to a chain or transaction.
type Attribute struct {
	// Type is the attribute type name.
	Type string
	// Value is the attribute payload. Nil when omitted.
	Value map[string]any
}

// OperationListOptions filters a page of operations.
//
// [PageOptions.Sort] is ignored because get-operations has no sortOrder parameter.
type OperationListOptions struct {
	// PageOptions selects which page to return.
	PageOptions
	// Address filters by emitter or user address.
	Address string
	// TxHash filters by a source transaction hash.
	TxHash TxHash
	// SourceChains filters by origin chain ids, encoded as a comma-joined query value.
	SourceChains []ChainID
	// TargetChains filters by destination chain ids, encoded as a comma-joined query value.
	TargetChains []ChainID
	// IncludesChains matches operations whose source or target is one of the ids.
	IncludesChains []ChainID
	// AppID filters by application id.
	AppID string
	// ExclusiveAppID, when true, requires AppID to be the only application id.
	ExclusiveAppID bool
	// MinAmount is a minimum USD amount for address search. Zero means unset.
	MinAmount float64
	// AddressType selects how Address is interpreted. Empty means the server default.
	AddressType AddressType
	// From is the start of the time range in UTC. Zero means unset.
	From time.Time
	// To is the end of the time range in UTC. Zero means unset.
	To time.Time
}

// GlobalTransaction is the origin and destination transactions for one VAA.
type GlobalTransaction struct {
	// ID is the VAA id the server sent, usually chain/emitter/sequence.
	ID string
	// Origin is the origin-chain transaction. Zero when omitted.
	Origin OriginTx
	// Destination is the destination-chain transaction. Zero when the VAA is unredeemed.
	Destination DestinationTx
}

// OriginTx is the origin-chain transaction of a global transaction.
type OriginTx struct {
	// Attribute is an optional typed extension payload. Zero when omitted.
	Attribute Attribute
	// From is the origin transaction sender.
	From string
	// Status is the origin transaction status.
	Status TxStatus
	// TxHash is the origin transaction hash.
	TxHash TxHash
}

// DestinationTx is the destination-chain transaction of a global transaction.
type DestinationTx struct {
	// BlockNumber is the destination block number as a decimal string.
	BlockNumber string
	// ChainID is the destination chain.
	ChainID ChainID
	// From is the destination transaction sender.
	From string
	// Method is the destination method name, or "unknown" when the server cannot classify it.
	Method string
	// Status is the destination transaction status.
	Status TxStatus
	// Timestamp is the destination transaction time in UTC. Zero when omitted.
	Timestamp time.Time
	// To is the destination transaction recipient.
	To string
	// TxHash is the destination transaction hash.
	TxHash TxHash
	// UpdatedAt is when the explorer last updated this record. Zero when omitted.
	UpdatedAt time.Time
}

// Relay is a Wormhole Relayer delivery attempt for a VAA.
type Relay struct {
	// ID is the relay document id.
	ID string
	// Relayer is the relayer identity the server sent.
	Relayer string
	// Status is the relay status. Empty when omitted.
	Status RelayStatus
	// ReceivedAt is when the relayer received the request. Zero when omitted.
	ReceivedAt time.Time
	// CompletedAt is when the delivery completed. Zero when omitted.
	CompletedAt time.Time
	// FailedAt is when the delivery failed. Zero when omitted.
	FailedAt time.Time
	// Data is the relay request, instructions, and execution result. Zero when omitted.
	Data RelayData
}

// RelayData is the request and execution payload of a relay.
type RelayData struct {
	// FromTxHash is the source transaction hash.
	FromTxHash TxHash
	// ToTxHash is the destination transaction hash.
	ToTxHash TxHash
	// MaxAttempts is the maximum delivery attempts.
	MaxAttempts int
	// Instructions are the encoded delivery instructions. Zero when omitted.
	Instructions RelayInstructions
	// Delivery is the delivery execution result. Zero when omitted.
	Delivery RelayDelivery
}

// RelayInstructions are Wormhole Relayer delivery instructions.
type RelayInstructions struct {
	// EncodedExecutionInfo is the encoded execution info blob.
	EncodedExecutionInfo string
	// ExtraReceiverValue is additional value forwarded to the receiver.
	ExtraReceiverValue RelayBigNumber
	// RequestedReceiverValue is the requested receiver value.
	RequestedReceiverValue RelayBigNumber
	// RefundAddress is the refund recipient.
	RefundAddress string
	// RefundChainID is the chain the refund is sent to.
	RefundChainID ChainID
	// RefundDeliveryProvider is the refund delivery provider address.
	RefundDeliveryProvider string
	// SenderAddress is the message sender.
	SenderAddress string
	// SourceDeliveryProvider is the source delivery provider address.
	SourceDeliveryProvider string
	// TargetAddress is the destination contract.
	TargetAddress string
	// TargetChainID is the destination chain.
	TargetChainID ChainID
	// VaaKeys are additional VAA keys attached to the delivery. Shape is untyped.
	VaaKeys []any
}

// RelayBigNumber is an ethers-style big number object with _hex and _isBigNumber keys.
type RelayBigNumber struct {
	// Hex is the 0x-prefixed hex magnitude from the server's _hex field.
	Hex string
	// IsBigNumber is the server's _isBigNumber flag.
	IsBigNumber bool
}

// RelayDelivery is the execution result of a relay.
type RelayDelivery struct {
	// Budget is the delivery budget as a decimal string.
	Budget string
	// MaxRefund is the maximum refund as a decimal string.
	MaxRefund string
	// RelayGasUsed is the gas used by the relay.
	RelayGasUsed int
	// TargetChainDecimals is the destination chain native-token decimals.
	TargetChainDecimals int
	// Execution is the on-chain execution result. Zero when omitted.
	Execution RelayExecution
}

// RelayExecution is the on-chain result of a relay delivery.
type RelayExecution struct {
	// Detail is a human-readable execution detail.
	Detail string
	// GasUsed is the gas used as a decimal string.
	GasUsed string
	// RefundStatus is the refund outcome.
	RefundStatus string
	// RevertString is the revert reason when execution failed.
	RevertString string
	// Status is the execution status.
	Status string
	// TransactionHash is the destination transaction hash.
	TransactionHash TxHash
}
