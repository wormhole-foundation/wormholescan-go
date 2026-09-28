// Command specgen converts a Swagger 2.0 document to OpenAPI 3.0 for
// oapi-codegen, which only reads OpenAPI 3.
//
// Usage: specgen <swagger2.json> <openapi3.json>
//
// The input is the served Wormholescan spec after api/spec/patch.jq has been
// applied; the output is an intermediate file that is not committed.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// errUsage is returned when the command line does not carry exactly two paths.
var errUsage = errors.New("usage: specgen <swagger2.json> <openapi3.json>")

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "specgen:", err)
		os.Exit(1)
	}
}

// run converts the Swagger 2.0 document at args[0] and writes OpenAPI 3.0 JSON
// to args[1].
func run(args []string) error {
	const argCount = 2
	if len(args) != argCount {
		return errUsage
	}
	in, out := args[0], args[1]

	raw, err := os.ReadFile(in) //nolint:gosec // paths come from the command line by design
	if err != nil {
		return fmt.Errorf("read %s: %w", in, err)
	}
	var v2 openapi2.T
	if err = json.Unmarshal(raw, &v2); err != nil {
		return fmt.Errorf("parse %s: %w", in, err)
	}
	v3, err := openapi2conv.ToV3(&v2)
	if err != nil {
		return fmt.Errorf("convert %s: %w", in, err)
	}
	keepEnumNames(&v2, v3)
	encoded, err := json.MarshalIndent(v3, "", "  ")
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	const fileMode = 0o644
	if err = os.WriteFile(out, encoded, fileMode); err != nil { //nolint:gosec // see ReadFile above
		return fmt.Errorf("write %s: %w", out, err)
	}
	return nil
}

// keepEnumNames copies the swag `x-enum-varnames` extension from every Swagger
// 2.0 definition to its OpenAPI 3.0 schema. openapi2conv drops schema
// extensions, and oapi-codegen uses this one to name enum constants
// (ChainIDEthereum instead of N2).
func keepEnumNames(v2 *openapi2.T, v3 *openapi3.T) {
	const ext = "x-enum-varnames"
	for name, def := range v2.Definitions {
		names, ok := def.Value.Extensions[ext]
		if !ok {
			continue
		}
		schema, ok := v3.Components.Schemas[name]
		if !ok || schema.Value == nil {
			continue
		}
		if schema.Value.Extensions == nil {
			schema.Value.Extensions = map[string]any{}
		}
		schema.Value.Extensions[ext] = names
	}
}
