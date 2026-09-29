package wormholescan

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVAAIDRoundTrip(t *testing.T) {
	t.Parallel()

	const raw = "1/19671a08a9cef6f3a04314ed478fc332a4966f41ad3e6fea76933dede9c6cdfe/755119"
	id, err := ParseVAAID(raw)
	require.NoError(t, err)
	assert.Equal(t, ChainIDSolana, id.Chain)
	assert.Equal(t, EmitterAddress("19671a08a9cef6f3a04314ed478fc332a4966f41ad3e6fea76933dede9c6cdfe"), id.Emitter)
	assert.Equal(t, uint64(755119), id.Sequence)
	assert.Equal(t, raw, id.String())
}

func TestParseVAAIDRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		part string
	}{
		{name: "empty", raw: "", part: "chain/emitter/sequence"},
		{name: "two parts", raw: "1/abc", part: "chain/emitter/sequence"},
		{name: "four parts", raw: "1/abc/2/extra", part: "chain/emitter/sequence"},
		{name: "non-numeric chain", raw: "x/abc/1", part: "chain"},
		{name: "chain overflow", raw: "65536/abc/1", part: "chain"},
		{name: "empty emitter", raw: "1//1", part: "emitter"},
		{name: "non-numeric sequence", raw: "1/abc/x", part: "sequence"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseVAAID(tt.raw)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.part)
		})
	}
}

func TestChainIDString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ethereum", ChainID(2).String())
	const unknown ChainID = 999
	assert.Equal(t, "chain(999)", unknown.String())
}

func TestChainIDConstantsMatchAPI(t *testing.T) {
	t.Parallel()

	apiConsts := constValues(t, "api/api.gen.go", "VaaChainID")
	ours := constValues(t, "types.go", "ChainID")
	require.NotEmpty(t, apiConsts)
	for name, value := range apiConsts {
		got, ok := ours[name]
		require.True(t, ok, "missing root constant %s", name)
		assert.Equal(t, value, got, "constant %s", name)
	}
}

func TestDeref(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, deref[int](nil))
	n := 7
	assert.Equal(t, 7, deref(&n))

	assert.True(t, derefTime(nil).IsZero())
	loc := time.FixedZone("X", 3600)
	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, loc)
	got := derefTime(&ts)
	assert.True(t, got.Equal(ts.UTC()))
	assert.Equal(t, time.UTC, got.Location())
}

func constValues(t *testing.T, path, typeName string) map[string]int64 {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	require.NoError(t, err)

	out := make(map[string]int64)
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			ident, ok := vs.Type.(*ast.Ident)
			if !ok || ident.Name != typeName {
				continue
			}
			for i, name := range vs.Names {
				require.Less(t, i, len(vs.Values))
				lit, ok := vs.Values[i].(*ast.BasicLit)
				require.True(t, ok, "constant %s is not a basic literal", name.Name)
				n, err := strconv.ParseInt(lit.Value, 10, 64)
				require.NoError(t, err)
				out[name.Name] = n
			}
		}
	}
	return out
}
