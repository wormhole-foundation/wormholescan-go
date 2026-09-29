package wormholescan

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const (
	vaaIDNumParts     = 3
	emitterByteLength = 32
	hexPrefix         = "0x"
	chainIDBitSize    = 16
	sequenceBitSize   = 64
)

// ChainID is a Wormhole chain identifier.
type ChainID uint16

// Wormhole chain identifiers, mirrored from the generated api.VaaChainID enum.
const (
	ChainIDAlgorand        ChainID = 8
	ChainIDAptos           ChainID = 22
	ChainIDArbitrum        ChainID = 23
	ChainIDArbitrumSepolia ChainID = 10003
	ChainIDArc             ChainID = 71
	ChainIDAvalanche       ChainID = 6
	ChainIDAztec           ChainID = 56
	ChainIDBOB             ChainID = 42
	ChainIDBSC             ChainID = 4
	ChainIDBase            ChainID = 30
	ChainIDBaseSepolia     ChainID = 10004
	ChainIDBerachain       ChainID = 39
	ChainIDBtc             ChainID = 29
	ChainIDCelestia        ChainID = 4004
	ChainIDCelo            ChainID = 14
	ChainIDCodex           ChainID = 54
	ChainIDConverge        ChainID = 53
	ChainIDCosmoshub       ChainID = 4000
	ChainIDCreditCoin      ChainID = 59
	ChainIDDogecoin        ChainID = 65
	ChainIDDymension       ChainID = 4007
	ChainIDEclipse         ChainID = 41
	ChainIDEthereum        ChainID = 2
	ChainIDEvmos           ChainID = 4001
	ChainIDFileCoin        ChainID = 31
	ChainIDFogo            ChainID = 51
	ChainIDGnosis          ChainID = 25
	ChainIDHolesky         ChainID = 10006
	ChainIDHyperCore       ChainID = 65000
	ChainIDHyperEVM        ChainID = 47
	ChainIDInjective       ChainID = 19
	ChainIDInk             ChainID = 46
	ChainIDKlaytn          ChainID = 13
	ChainIDKujira          ChainID = 4002
	ChainIDLinea           ChainID = 38
	ChainIDMegaETH         ChainID = 64
	ChainIDMezo            ChainID = 50
	ChainIDMoca            ChainID = 63
	ChainIDMonad           ChainID = 48
	ChainIDMonadTestnet    ChainID = 10009
	ChainIDMoonbeam        ChainID = 16
	ChainIDMovement        ChainID = 49
	ChainIDNear            ChainID = 15
	ChainIDNeutron         ChainID = 4003
	ChainIDNexus           ChainID = 69
	ChainIDNoble           ChainID = 4009
	ChainIDOptimism        ChainID = 24
	ChainIDOptimismSepolia ChainID = 10005
	ChainIDOsmosis         ChainID = 20
	ChainIDPlasma          ChainID = 58
	ChainIDPlume           ChainID = 55
	ChainIDPolygon         ChainID = 5
	ChainIDPolygonSepolia  ChainID = 10007
	ChainIDProvenance      ChainID = 4008
	ChainIDPythNet         ChainID = 26
	ChainIDRootstock       ChainID = 33
	ChainIDSeda            ChainID = 4006
	ChainIDSei             ChainID = 32
	ChainIDSeiEVM          ChainID = 40
	ChainIDSepolia         ChainID = 10002
	ChainIDSolana          ChainID = 1
	ChainIDSonic           ChainID = 52
	ChainIDStacks          ChainID = 60
	ChainIDStargaze        ChainID = 4005
	ChainIDStellar         ChainID = 61
	ChainIDSui             ChainID = 21
	ChainIDTON             ChainID = 62
	ChainIDTempo           ChainID = 68
	ChainIDTerra2          ChainID = 18
	ChainIDTron            ChainID = 70
	ChainIDUnichain        ChainID = 44
	ChainIDUnset           ChainID = 0
	ChainIDWorldchain      ChainID = 45
	ChainIDWormchain       ChainID = 3104
	ChainIDXRPL            ChainID = 66
	ChainIDXRPLEVM         ChainID = 57
	ChainIDZeroGravity     ChainID = 67
)

// String returns the chain name for a known id, or "chain(N)" otherwise.
func (id ChainID) String() string { //nolint:gocyclo,cyclop,funlen // one case per api.VaaChainID member
	switch id {
	case ChainIDAlgorand:
		return "Algorand"
	case ChainIDAptos:
		return "Aptos"
	case ChainIDArbitrum:
		return "Arbitrum"
	case ChainIDArbitrumSepolia:
		return "ArbitrumSepolia"
	case ChainIDArc:
		return "Arc"
	case ChainIDAvalanche:
		return "Avalanche"
	case ChainIDAztec:
		return "Aztec"
	case ChainIDBOB:
		return "BOB"
	case ChainIDBSC:
		return "BSC"
	case ChainIDBase:
		return "Base"
	case ChainIDBaseSepolia:
		return "BaseSepolia"
	case ChainIDBerachain:
		return "Berachain"
	case ChainIDBtc:
		return "Btc"
	case ChainIDCelestia:
		return "Celestia"
	case ChainIDCelo:
		return "Celo"
	case ChainIDCodex:
		return "Codex"
	case ChainIDConverge:
		return "Converge"
	case ChainIDCosmoshub:
		return "Cosmoshub"
	case ChainIDCreditCoin:
		return "CreditCoin"
	case ChainIDDogecoin:
		return "Dogecoin"
	case ChainIDDymension:
		return "Dymension"
	case ChainIDEclipse:
		return "Eclipse"
	case ChainIDEthereum:
		return "Ethereum"
	case ChainIDEvmos:
		return "Evmos"
	case ChainIDFileCoin:
		return "FileCoin"
	case ChainIDFogo:
		return "Fogo"
	case ChainIDGnosis:
		return "Gnosis"
	case ChainIDHolesky:
		return "Holesky"
	case ChainIDHyperCore:
		return "HyperCore"
	case ChainIDHyperEVM:
		return "HyperEVM"
	case ChainIDInjective:
		return "Injective"
	case ChainIDInk:
		return "Ink"
	case ChainIDKlaytn:
		return "Klaytn"
	case ChainIDKujira:
		return "Kujira"
	case ChainIDLinea:
		return "Linea"
	case ChainIDMegaETH:
		return "MegaETH"
	case ChainIDMezo:
		return "Mezo"
	case ChainIDMoca:
		return "Moca"
	case ChainIDMonad:
		return "Monad"
	case ChainIDMonadTestnet:
		return "MonadTestnet"
	case ChainIDMoonbeam:
		return "Moonbeam"
	case ChainIDMovement:
		return "Movement"
	case ChainIDNear:
		return "Near"
	case ChainIDNeutron:
		return "Neutron"
	case ChainIDNexus:
		return "Nexus"
	case ChainIDNoble:
		return "Noble"
	case ChainIDOptimism:
		return "Optimism"
	case ChainIDOptimismSepolia:
		return "OptimismSepolia"
	case ChainIDOsmosis:
		return "Osmosis"
	case ChainIDPlasma:
		return "Plasma"
	case ChainIDPlume:
		return "Plume"
	case ChainIDPolygon:
		return "Polygon"
	case ChainIDPolygonSepolia:
		return "PolygonSepolia"
	case ChainIDProvenance:
		return "Provenance"
	case ChainIDPythNet:
		return "PythNet"
	case ChainIDRootstock:
		return "Rootstock"
	case ChainIDSeda:
		return "Seda"
	case ChainIDSei:
		return "Sei"
	case ChainIDSeiEVM:
		return "SeiEVM"
	case ChainIDSepolia:
		return "Sepolia"
	case ChainIDSolana:
		return "Solana"
	case ChainIDSonic:
		return "Sonic"
	case ChainIDStacks:
		return "Stacks"
	case ChainIDStargaze:
		return "Stargaze"
	case ChainIDStellar:
		return "Stellar"
	case ChainIDSui:
		return "Sui"
	case ChainIDTON:
		return "TON"
	case ChainIDTempo:
		return "Tempo"
	case ChainIDTerra2:
		return "Terra2"
	case ChainIDTron:
		return "Tron"
	case ChainIDUnichain:
		return "Unichain"
	case ChainIDUnset:
		return "Unset"
	case ChainIDWorldchain:
		return "Worldchain"
	case ChainIDWormchain:
		return "Wormchain"
	case ChainIDXRPL:
		return "XRPL"
	case ChainIDXRPLEVM:
		return "XRPLEVM"
	case ChainIDZeroGravity:
		return "ZeroGravity"
	default:
		return "chain(" + strconv.FormatUint(uint64(id), 10) + ")"
	}
}

// EmitterAddress is a 32-byte hex emitter address, lowercase, with no 0x prefix.
type EmitterAddress string

// TxHash is a transaction hash as the server sends it (chain-specific encoding).
// It is not normalized.
type TxHash string

// Digest is a hex digest with no 0x prefix.
type Digest string

// GuardianAddress is a 0x-prefixed lowercase EVM address, as heartbeats send it.
type GuardianAddress string

// VAAID is the triple that uniquely identifies a VAA.
type VAAID struct {
	// Chain is the emitter chain.
	Chain ChainID
	// Emitter is the emitter address.
	Emitter EmitterAddress
	// Sequence is the VAA sequence number.
	Sequence uint64
}

// String returns the canonical "chain/emitter/sequence" form.
func (id VAAID) String() string {
	return strconv.FormatUint(uint64(id.Chain), 10) + "/" + string(id.Emitter) + "/" +
		strconv.FormatUint(id.Sequence, 10)
}

// ParseVAAID parses a "chain/emitter/sequence" identifier.
func ParseVAAID(s string) (VAAID, error) {
	parts := strings.Split(s, "/")
	if len(parts) != vaaIDNumParts {
		return VAAID{}, fmt.Errorf("%sVAA ID %q: want chain/emitter/sequence", errPrefix, s)
	}
	chain, err := parseVAAIDUint(parts[0], "chain", chainIDBitSize)
	if err != nil {
		return VAAID{}, err
	}
	emitter, err := parseEmitter(parts[1])
	if err != nil {
		return VAAID{}, err
	}
	seq, err := parseVAAIDUint(parts[2], "sequence", sequenceBitSize)
	if err != nil {
		return VAAID{}, err
	}
	return VAAID{
		Chain:    ChainID(chain), //nolint:gosec // chainIDBitSize is 16, so chain fits uint16
		Emitter:  emitter,
		Sequence: seq,
	}, nil
}

// parseVAAIDUint parses a decimal VAA ID part with no leading zeros.
func parseVAAIDUint(raw, part string, bitSize int) (uint64, error) {
	if raw == "" || (len(raw) > 1 && raw[0] == '0') {
		return 0, fmt.Errorf("%sVAA ID %s %q: invalid", errPrefix, part, raw)
	}
	n, err := strconv.ParseUint(raw, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("%sVAA ID %s %q: %w", errPrefix, part, raw, err)
	}
	return n, nil
}

// parseEmitter normalizes and validates a 32-byte hex emitter address.
func parseEmitter(raw string) (EmitterAddress, error) {
	s := strings.TrimPrefix(strings.ToLower(raw), hexPrefix)
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("%sVAA ID emitter %q: %w", errPrefix, raw, err)
	}
	if len(decoded) != emitterByteLength {
		return "", fmt.Errorf(
			"%sVAA ID emitter %q: want %d bytes, got %d",
			errPrefix, raw, emitterByteLength, len(decoded),
		)
	}
	return EmitterAddress(s), nil
}
