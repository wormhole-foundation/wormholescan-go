# wormholescan-go

> **Status: early.** VAAs, observations, operations, governor state and the
> guardian public API are covered by the hand-written client; statistics and
> token endpoints are reachable through the generated `api` package. Expect
> the API surface to change without notice before v1.

`wormholescan-go` is a Go client SDK for the
[Wormholescan API](https://api.wormholescan.io/swagger/index.html), the
public explorer API for the Wormhole network: VAAs, operations, observations,
governor state, guardian heartbeats, and network statistics.

```sh
go get github.com/wormhole-foundation/wormholescan-go
```

## Usage

```go
c, err := wormholescan.New() // mainnet; wormholescan.WithBaseURL(wormholescan.TestnetURL) for testnet
if err != nil {
    return err
}

id, _ := wormholescan.ParseVAAID("1/19671a08a9cef6f3a04314ed478fc332a4966f41ad3e6fea76933dede9c6cdfe/755119")
vaa, err := c.GetVAA(ctx, id, wormholescan.VAAGetOptions{})
if errors.Is(err, wormholescan.ErrNotFound) {
    // no such VAA
}

// One page.
page, err := c.ListVAAsByChain(ctx, wormholescan.ChainIDEthereum,
    wormholescan.VAAListOptions{PageOptions: wormholescan.PageOptions{PageSize: 50}})

// Every page, lazily.
for vaa, err := range c.VAAsByChain(ctx, wormholescan.ChainIDEthereum, wormholescan.VAAListOptions{}) {
    if err != nil {
        return err
    }
    _ = vaa
}

// Guardian public API (/v1): works against Wormholescan or a guardian node.
g, err := guardian.New(wormholescan.WithBaseURL("https://guardian.example.org"))
heartbeats, err := g.Heartbeats(ctx)
```

Errors from the server are `*APIError` (status, server code, message,
request id) and match `ErrNotFound`, `ErrRateLimited`, `ErrBadRequest` with
`errors.Is`. GET requests are retried on 429 and 5xx with backoff, honouring
`Retry-After`; `WithRetry` and `WithoutRetry` tune that.

## Layout

Two layers:

- `wormholescan` (root) — hand-written. Typed domain values (`ChainID`,
  `VAAID`, `time.Time`, `[]byte`), pagination (`List*` returns one
  `Page[T]`; the bare plural returns an `iter.Seq2` over every page), error
  handling, retries. This is the supported surface and the one semver
  applies to.
- `wormholescan/guardian` — the `/v1` namespace: the guardiand public API
  that Wormholescan proxies. Same options; point it at a guardian node's
  own endpoint with `WithBaseURL`.
- `api` — thin, generated. One method per operation (`FindVaaById`,
  `GuardiansHearbeats`, ...), types that mirror the server's JSON, plus
  `*WithResponse` variants that decode the body. Tracks the spec; no
  compatibility promise of its own. `Client.API()` returns one that shares
  the root client's transport.

### Coverage

| Area | Root package | Not covered (use `api`) |
|---|---|---|
| VAAs | `GetVAA`, `GetDuplicatedVAAs`, `ListVAAs`/`VAAs`, `ListVAAsByChain`/`VAAsByChain`, `ListVAAsByEmitter`/`VAAsByEmitter` | `parse-vaa` (POST, body undeclared upstream), `get-vaa-counts` |
| Observations | `ListObservations`/`Observations`, `…ByChain`, `…ByEmitter`, `…ByVAA`, `GetObservation` (server currently answers 404 for every hash encoding), `ListDelegateObservations…`/`DelegateObservations…`, `GetDelegateObservation` | — |
| Operations | `GetOperation`, `ListOperations`/`Operations`, `SearchOperationsByTxHashes` (server currently answers 405), `GetRelay`, `GetGlobalTransaction` | `list-transactions`, `get-transaction-by-id`, `get-last-transactions` |
| Governor | `ListGovernorConfigs`/`GovernorConfigs`, `GetGovernorConfig`, `ListGovernorStatuses`/`GovernorStatuses`, `GetGovernorStatus`, `ListGovernorLimits`/`GovernorLimits`, `ListNotionalLimits`/`NotionalLimits`, `ListNotionalLimitsByChain`, `ListNotionalAvailable`/`NotionalAvailable`, `ListNotionalAvailableByChain`, `GetMaxNotionalAvailable`, `EnqueuedVAAs`, `EnqueuedVAAsByChain`, `GovernorVAAs` | — |
| Guardian `/v1` (`guardian` package) | `Heartbeats`, `CurrentGuardianSet`, `GetSignedVAA`, `GetSignedBatchVAA`, `AvailableNotionalByChain`, `EnqueuedVAAs`, `IsVAAEnqueued`, `GovernorTokenList` | — |
| Stats, supply, tokens, NTT, address, top-N | — | all |

## Generated Code

`api/api.gen.go` is produced by [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
from `api/spec/swagger.json`, a committed copy of the document the API
serves at `https://api.wormholescan.io/swagger.json`. The served document
has known defects (undeclared path parameters, `[]byte` fields typed as
integer arrays, wrong response shapes); `api/spec/patch.jq` corrects them
before generation, one commented rule per defect. `internal/specgen`
converts the patched Swagger 2.0 document to OpenAPI 3 for oapi-codegen.

```sh
moon run root:spec-fetch   # refresh api/spec/swagger.json from the server
moon run root:gen          # patch, convert, regenerate api/api.gen.go
moon run root:gen-check    # CI gate: committed output is current
```

`api/decode_test.go` decodes recorded responses in `api/testdata/` through
the generated types and fails if a key the server sent is lost or changed,
so a dropped patch rule shows up as a test failure. A weekly workflow
(`spec-drift.yml`) fetches the served spec and opens an issue when it
differs from the committed copy.

## Local Bootstrap

Prerequisites:

- [mise](https://mise.jdx.dev) provisions every pinned tool from `mise.toml` +
  `mise.lock`: Go, Moon, `golangci-lint`, mockery, jq, GitHub CLI, and Python +
  uv (used only by the repository settings script). Run `mise install` once.

Tool versions live in `mise.toml`; `mise.lock` records a per-platform download
URL and checksum for each. `mise install` runs with `locked = true`, so it fails
closed if a tool lacks a pre-resolved, checksummed entry for the current
platform. To bump a tool, edit its version in `mise.toml`, run
`mise lock --platform linux-x64,linux-arm64,macos-x64,macos-arm64`, and commit
`mise.toml` + `mise.lock`.

## Common Tasks

Moon is the task front door:

```sh
moon run root:format
moon run root:lint
moon run root:build
moon run root:test
moon run root:check     # everything CI runs
```

CI runs the same aggregate check:

```sh
moon ci --summary minimal
```

Tests in the library packages are unit tests: no Docker, no real network.
Anything that needs a container goes under `integration/` behind the
`integration` build tag once it exists.

## Releases

Release Please opens a release PR from Conventional Commit subjects and, on
merge, publishes a GitHub release and the `v*` tag that Go module proxies and
pkg.go.dev pick up. There are no binary artifacts.

## License

Apache-2.0. See [LICENSE](LICENSE).
