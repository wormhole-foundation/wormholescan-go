package wormholescan

import (
	"fmt"
	"strconv"
	"strings"
)

const vaaIDNumParts = 3

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
func (id ChainID) String() string {
	if name, ok := chainIDNames()[id]; ok {
		return name
	}
	return "chain(" + strconv.FormatUint(uint64(id), 10) + ")"
}

// chainIDNames maps known chain ids to their names.
func chainIDNames() map[ChainID]string {
	return map[ChainID]string{
		ChainIDAlgorand:        "Algorand",
		ChainIDAptos:           "Aptos",
		ChainIDArbitrum:        "Arbitrum",
		ChainIDArbitrumSepolia: "ArbitrumSepolia",
		ChainIDArc:             "Arc",
		ChainIDAvalanche:       "Avalanche",
		ChainIDAztec:           "Aztec",
		ChainIDBOB:             "BOB",
		ChainIDBSC:             "BSC",
		ChainIDBase:            "Base",
		ChainIDBaseSepolia:     "BaseSepolia",
		ChainIDBerachain:       "Berachain",
		ChainIDBtc:             "Btc",
		ChainIDCelestia:        "Celestia",
		ChainIDCelo:            "Celo",
		ChainIDCodex:           "Codex",
		ChainIDConverge:        "Converge",
		ChainIDCosmoshub:       "Cosmoshub",
		ChainIDCreditCoin:      "CreditCoin",
		ChainIDDogecoin:        "Dogecoin",
		ChainIDDymension:       "Dymension",
		ChainIDEclipse:         "Eclipse",
		ChainIDEthereum:        "Ethereum",
		ChainIDEvmos:           "Evmos",
		ChainIDFileCoin:        "FileCoin",
		ChainIDFogo:            "Fogo",
		ChainIDGnosis:          "Gnosis",
		ChainIDHolesky:         "Holesky",
		ChainIDHyperCore:       "HyperCore",
		ChainIDHyperEVM:        "HyperEVM",
		ChainIDInjective:       "Injective",
		ChainIDInk:             "Ink",
		ChainIDKlaytn:          "Klaytn",
		ChainIDKujira:          "Kujira",
		ChainIDLinea:           "Linea",
		ChainIDMegaETH:         "MegaETH",
		ChainIDMezo:            "Mezo",
		ChainIDMoca:            "Moca",
		ChainIDMonad:           "Monad",
		ChainIDMonadTestnet:    "MonadTestnet",
		ChainIDMoonbeam:        "Moonbeam",
		ChainIDMovement:        "Movement",
		ChainIDNear:            "Near",
		ChainIDNeutron:         "Neutron",
		ChainIDNexus:           "Nexus",
		ChainIDNoble:           "Noble",
		ChainIDOptimism:        "Optimism",
		ChainIDOptimismSepolia: "OptimismSepolia",
		ChainIDOsmosis:         "Osmosis",
		ChainIDPlasma:          "Plasma",
		ChainIDPlume:           "Plume",
		ChainIDPolygon:         "Polygon",
		ChainIDPolygonSepolia:  "PolygonSepolia",
		ChainIDProvenance:      "Provenance",
		ChainIDPythNet:         "PythNet",
		ChainIDRootstock:       "Rootstock",
		ChainIDSeda:            "Seda",
		ChainIDSei:             "Sei",
		ChainIDSeiEVM:          "SeiEVM",
		ChainIDSepolia:         "Sepolia",
		ChainIDSolana:          "Solana",
		ChainIDSonic:           "Sonic",
		ChainIDStacks:          "Stacks",
		ChainIDStargaze:        "Stargaze",
		ChainIDStellar:         "Stellar",
		ChainIDSui:             "Sui",
		ChainIDTON:             "TON",
		ChainIDTempo:           "Tempo",
		ChainIDTerra2:          "Terra2",
		ChainIDTron:            "Tron",
		ChainIDUnichain:        "Unichain",
		ChainIDUnset:           "Unset",
		ChainIDWorldchain:      "Worldchain",
		ChainIDWormchain:       "Wormchain",
		ChainIDXRPL:            "XRPL",
		ChainIDXRPLEVM:         "XRPLEVM",
		ChainIDZeroGravity:     "ZeroGravity",
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
	chain, err := strconv.ParseUint(parts[0], 10, 16)
	if err != nil {
		return VAAID{}, fmt.Errorf("%sVAA ID chain %q: %w", errPrefix, parts[0], err)
	}
	if parts[1] == "" {
		return VAAID{}, fmt.Errorf("%sVAA ID emitter is empty", errPrefix)
	}
	seq, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return VAAID{}, fmt.Errorf("%sVAA ID sequence %q: %w", errPrefix, parts[2], err)
	}
	return VAAID{
		Chain:    ChainID(chain),
		Emitter:  EmitterAddress(parts[1]),
		Sequence: seq,
	}, nil
}
