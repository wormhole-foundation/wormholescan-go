# wormholescan-go

> **Status: skeleton.** The module builds and exposes no client yet. Expect
> the API surface to change without notice before v1.

`wormholescan-go` is a Go client SDK for the
[Wormholescan API](https://api.wormholescan.io/swagger/index.html), the
public explorer API for the Wormhole network: VAAs, operations, observations,
governor state, guardian heartbeats, and network statistics.

```sh
go get github.com/wormhole-foundation/wormholescan-go
```

## Local Bootstrap

Prerequisites:

- [mise](https://mise.jdx.dev) provisions every pinned tool from `mise.toml` +
  `mise.lock`: Go, Moon, `golangci-lint`, mockery, GitHub CLI, and Python +
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
