# wormholescan-go

> **Status: early.** The generated `api` package covers every operation in
> the served spec; the hand-written client is not started. Expect the API
> surface to change without notice before v1.

`wormholescan-go` is a Go client SDK for the
[Wormholescan API](https://api.wormholescan.io/swagger/index.html), the
public explorer API for the Wormhole network: VAAs, operations, observations,
governor state, guardian heartbeats, and network statistics.

```sh
go get github.com/wormhole-foundation/wormholescan-go
```

## Layout

Two layers:

- `api` — thin, generated. One method per operation (`FindVaaById`,
  `GuardiansHearbeats`, ...), types that mirror the server's JSON, plus
  `*WithResponse` variants that decode the body. Tracks the spec; no
  compatibility promise of its own.
- `wormholescan` (root) — thick, hand-written. Typed domain values,
  pagination, error handling, retries. This is the supported surface and
  the one semver applies to. Not written yet.

Use `api` directly when the thick client does not cover an endpoint yet.

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
