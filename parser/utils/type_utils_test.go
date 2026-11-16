package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestIsMainFile(t *testing.T) {
	tests := map[string]struct {
		src  string
		want bool
	}{
		"main package with main func": {"package main\n\nfunc main() {}\n", true},
		"main package without main":   {"package main\n\nfunc run() {}\n", false},
		"other package with main":     {"package server\n\nfunc main() {}\n", false},
		"empty file":                  {"", false},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsMainFile(writeFile(t, "main.go", tc.src)))
		})
	}
}

func TestGetModuleNameFromGoMod(t *testing.T) {
	path := writeFile(t, "go.mod", "// comment\nmodule   github.com/acme/api  \n\ngo 1.24.0\n")
	assert.Equal(t, "github.com/acme/api", GetModuleNameFromGoMod(path))

	assert.Equal(t, "", GetModuleNameFromGoMod(writeFile(t, "go.mod", "go 1.24.0\n")), "no module line")
	assert.Equal(t, "", GetModuleNameFromGoMod(filepath.Join(t.TempDir(), "missing.mod")), "missing file")
}

func TestTypePredicates(t *testing.T) {
	assert.True(t, IsInterfaceType("interface{}"))
	assert.True(t, IsInterfaceType("INTERFACE{}"))
	assert.False(t, IsInterfaceType("any"))

	assert.True(t, IsEnumType("CountriesEnum"))
	assert.False(t, IsEnumType("Country"))

	assert.True(t, IsBasicGoType("int64"))
	assert.True(t, IsBasicGoType("string"))
	assert.False(t, IsBasicGoType("model.User"))

	assert.True(t, IsGoTypeOASType("float32"))
	assert.False(t, IsGoTypeOASType("model.User"))
	assert.Equal(t, "number", GoTypesOASTypes["float64"])
	assert.Equal(t, "integer", GoTypesOASTypes["uint8"])

	assert.True(t, IsInStringList([]string{"a", "b"}, "b"))
	assert.False(t, IsInStringList([]string{"a", "b"}, "c"))
	assert.False(t, IsInStringList(nil, "a"))
}

func TestRefLinkPrefixes(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) string
		pfx  string
	}{
		{"schemas", AddSchemaRefLinkPrefix, "#/components/schemas/"},
		{"parameters", AddParametersRefLinkPrefix, "#/components/parameters/"},
		{"responses", AddResponsesRefLinkPrefix, "#/components/responses/"},
		{"requestBodies", AddRequestBodiesRefLinkPrefix, "#/components/requestBodies/"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.pfx+"User", tc.fn("User"), "adds prefix")
			assert.Equal(t, tc.pfx+"User", tc.fn(tc.pfx+"User"), "keeps an existing prefix")
		})
	}
}
