// Package wormholescan is a Go client for the Wormholescan API
// (https://api.wormholescan.io), the public explorer API for the Wormhole
// network.
//
// New returns a Client for mainnet; use WithBaseURL(TestnetURL) for testnet.
// Methods named Get* fetch one object, List* fetch one Page, and the bare
// plural (VAAs, Operations, ...) returns an iter.Seq2 that walks every page.
// Server errors are *APIError values that match ErrNotFound, ErrRateLimited
// and ErrBadRequest with [errors.Is]. GET requests are retried on 429 and 5xx.
//
// The guardian subpackage covers the /v1 guardian public API; the api
// subpackage is the generated one-method-per-operation client for everything
// else, reachable through Client.API.
package wormholescan
