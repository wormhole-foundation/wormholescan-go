// Package wormholescan is a Go client for the Wormholescan API
// (https://api.wormholescan.io), the public explorer API for the Wormhole
// network.
//
// [New] returns a client pointed at mainnet. Use [WithBaseURL] with
// [TestnetURL] for testnet. Endpoints this package does not wrap are
// available through [Client.API], the generated api subpackage.
package wormholescan
