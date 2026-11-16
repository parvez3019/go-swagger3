package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/parvez3019/go-swagger3/app"
)

const (
	advancedModule   = "advanced_data"
	advancedMainFile = "advanced_data/server/main.go"
)

type specSummary struct {
	Info struct {
		Title string `json:"title"`
	} `json:"info"`
	Paths map[string]json.RawMessage `json:"paths"`
}

func loadSpec(t *testing.T, path string) specSummary {
	t.Helper()
	var s specSummary
	require.NoError(t, json.Unmarshal([]byte(LoadJSONAsString(path)), &s))
	return s
}

// runCLI drives the real CLI entry point (flag parsing, parser, writer).
func runCLI(t *testing.T, args ...string) error {
	t.Helper()
	return app.NewApp().Run(append([]string{"go-swagger3"}, args...))
}

// Characterisation test for security schemes, schema edge cases (named slice/map
// types, multi-parameter generics, generic fields, typed defaults/examples,
// time.Time/interface{}/map fields, comment descriptions) and request/response
// body shapes, generated through the CLI.
func Test_AdvancedSpec(t *testing.T) {
	err := runCLI(t,
		"--module-path", advancedModule,
		"--main-file-path", advancedMainFile,
		"--output", "advanced_data/spec/actual_advanced.json",
		"--schema-without-pkg",
		"--quiet",
	)
	require.NoError(t, err)

	diff, report := jsondiff.Compare(
		[]byte(LoadJSONAsString("advanced_data/spec/expected_advanced.json")),
		[]byte(LoadJSONAsString("advanced_data/spec/actual_advanced.json")),
		&jsondiff.Options{},
	)
	assert.Equal(t, jsondiff.FullMatch, diff, report)
}

func Test_AdvancedSpec_YAMLOutput(t *testing.T) {
	out := filepath.Join(t.TempDir(), "oas.yaml")
	err := runCLI(t,
		"--module-path", advancedModule,
		"--main-file-path", advancedMainFile,
		"--output", out,
		"--schema-without-pkg",
		"--generate-yaml",
		"--quiet",
	)
	require.NoError(t, err)

	content, err := os.ReadFile(out)
	require.NoError(t, err)
	yaml := string(content)
	assert.Contains(t, yaml, "openapi: 3.0.0")
	assert.Contains(t, yaml, "/items/{id}/image:")
	assert.Contains(t, yaml, "clientCredentials:")
	assert.False(t, strings.HasPrefix(strings.TrimSpace(yaml), "{"), "expected YAML, got JSON")
}

func Test_AdvancedSpec_HandlerPathLimitsOperations(t *testing.T) {
	out := filepath.Join(t.TempDir(), "oas.json")
	err := runCLI(t,
		"--module-path", advancedModule,
		"--main-file-path", advancedMainFile,
		"--handler-path", "advanced_data/server",
		"--output", out,
		"--schema-without-pkg",
		"--quiet",
	)
	require.NoError(t, err)

	spec := loadSpec(t, out)
	assert.Equal(t, "Advanced API", spec.Info.Title)
	assert.Empty(t, spec.Paths, "handlers outside --handler-path must be skipped")
}

func Test_AdvancedSpec_ExcludeSkipsDirectories(t *testing.T) {
	out := filepath.Join(t.TempDir(), "oas.json")
	err := runCLI(t,
		"--module-path", advancedModule,
		"--main-file-path", advancedMainFile,
		"--exclude", "handler",
		"--output", out,
		"--schema-without-pkg",
		"--quiet",
	)
	require.NoError(t, err)
	assert.Empty(t, loadSpec(t, out).Paths, "excluded handler directory must be skipped")
}

func Test_CLI_ErrorsOnInvalidPaths(t *testing.T) {
	tests := map[string][]string{
		"missing module path": {"--module-path", "does-not-exist"},
		"module without go.mod": {
			"--module-path", "advanced_data/handler",
		},
		"missing main file": {
			"--module-path", advancedModule,
			"--main-file-path", "advanced_data/server/nope.go",
		},
		"missing handler path": {
			"--module-path", advancedModule,
			"--main-file-path", advancedMainFile,
			"--handler-path", "advanced_data/nope",
		},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "oas.json")
			err := runCLI(t, append(args, "--output", out, "--quiet")...)
			assert.Error(t, err)
			_, statErr := os.Stat(out)
			assert.True(t, os.IsNotExist(statErr), "no spec should be written on error")
		})
	}
}

func Test_CLI_FmtCommandAlignsAnnotations(t *testing.T) {
	dir := t.TempDir()
	src := "package handler\n\n// @Title Get item\n// @Param   id   path int true \"Item ID\"\n// @Success 200 object  model.Item \"Item\"\nfunc GetItem() {}\n"
	path := filepath.Join(dir, "handler.go")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	require.NoError(t, runCLI(t, "fmt", "-d", dir))

	formatted, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotEqual(t, src, string(formatted), "fmt should rewrite unaligned annotations")
	assert.Contains(t, string(formatted), "func GetItem() {}")
}
