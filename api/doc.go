// Package api is the thin, generated client for the Wormholescan API: one
// method per operation, types that mirror the server's JSON, no retries, no
// pagination helpers.
//
// The code in api.gen.go is produced by oapi-codegen from the served Swagger
// document (spec/swagger.json) after the corrections in spec/patch.jq; each
// rule there names the upstream defect it works around. Regenerate with
// `moon run root:gen` after `moon run root:spec-fetch`.
//
// This package tracks the spec and changes when it does. It makes no
// compatibility promise of its own; the root wormholescan package is the
// stable, hand-written surface.
package api
