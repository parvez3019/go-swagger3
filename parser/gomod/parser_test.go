package gomod

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/parvez3019/go-swagger3/parser/model"
	log "github.com/sirupsen/logrus"
)

func init() { log.SetOutput(io.Discard) }

// writeSyntheticGoMod writes a go.mod with `direct` direct requires, `indirect`
// indirect requires (the shape a Go 1.24+ `tool` directive produces), plus a couple
// of fixed entries used by the assertions, and returns its path.
func writeSyntheticGoMod(t testing.TB, dir string, direct, indirect int) string {
	t.Helper()

	var b strings.Builder
	b.WriteString("module example.com/test\n\ngo 1.24\n\nrequire (\n")
	b.WriteString("\tgithub.com/parvez3019/go-swagger3 v1.2.3\n")
	b.WriteString("\texample.com/Example/Go-Example v1.0.0\n")
	for i := 0; i < direct; i++ {
		fmt.Fprintf(&b, "\texample.com/direct/dep%d v1.0.0\n", i)
	}
	b.WriteString(")\n\nrequire (\n")
	for i := 0; i < indirect; i++ {
		fmt.Fprintf(&b, "\texample.com/indirect/dep%d v1.0.0 // indirect\n", i)
	}
	b.WriteString(")\n\ntool (\n\texample.com/direct/dep/cmd/tool\n)\n")

	path := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	return path
}

func newTestParser(goModPath, cacheDir string) (Parser, *model.PkgAndSpecs) {
	specs := &model.PkgAndSpecs{}
	utils := model.Utils{
		Path:        model.Path{GoModFilePath: goModPath, GoModCachePath: cacheDir},
		PkgAndSpecs: specs,
	}
	return NewParser(utils), specs
}

// TestParseIndexesRequiresWithoutWalking is the regression guard for the on-demand
// rewrite: Parse must index every required module (direct and indirect) from go.mod
// without walking the module cache or registering any packages. If the eager
// filepath.Walk is ever reintroduced, KnownPkgs becomes non-empty and this fails.
func TestParseIndexesRequiresWithoutWalking(t *testing.T) {
	const direct, indirect = 5, 400
	dir := t.TempDir()
	goMod := writeSyntheticGoMod(t, dir, direct, indirect)
	cache := filepath.Join(dir, "cache")

	p, specs := newTestParser(goMod, cache)
	if err := p.Parse(); err != nil {
		t.Fatalf("Parse() with a tool directive present: %v", err)
	}

	// 2 fixed + direct + indirect requires; the tool directive entry is NOT a require.
	wantModules := 2 + direct + indirect
	if got := len(specs.DepModules); got != wantModules {
		t.Errorf("DepModules = %d, want %d", got, wantModules)
	}

	// No eager walk: nothing should be registered into the package index.
	if len(specs.KnownPkgs) != 0 {
		t.Errorf("KnownPkgs = %d, want 0 (Parse must not walk/register packages)", len(specs.KnownPkgs))
	}

	// Sorted longest import path first so prefix resolution picks the most specific module.
	for i := 1; i < len(specs.DepModules); i++ {
		if len(specs.DepModules[i-1].ImportPath) < len(specs.DepModules[i].ImportPath) {
			t.Fatalf("DepModules not sorted by import-path length descending at index %d", i)
		}
	}

	// Cache directory is computed (not walked), including uppercase escaping.
	assertCacheDir(t, specs, "github.com/parvez3019/go-swagger3", filepath.Join(cache, "github.com/parvez3019/go-swagger3@v1.2.3"))
	assertCacheDir(t, specs, "example.com/Example/Go-Example", filepath.Join(cache, "example.com/!example/!go-!example@v1.0.0"))
}

func assertCacheDir(t *testing.T, specs *model.PkgAndSpecs, importPath, want string) {
	t.Helper()
	for _, m := range specs.DepModules {
		if m.ImportPath == importPath {
			if m.CacheDir != want {
				t.Errorf("CacheDir(%s) = %q, want %q", importPath, m.CacheDir, want)
			}
			return
		}
	}
	t.Errorf("module %s not found in DepModules", importPath)
}

func TestModuleCacheDirEscapesUppercase(t *testing.T) {
	got := moduleCacheDir("/cache", "example.com/Example/GOExample", "v1.2.3")
	want := filepath.Join("/cache", "example.com/!example/!g!o!example@v1.2.3")
	if got != want {
		t.Errorf("moduleCacheDir = %q, want %q", got, want)
	}
}

func TestModuleCacheDirCases(t *testing.T) {
	cases := []struct {
		name, path, version, want string
	}{
		{"lowercase passthrough", "github.com/parvez3019/go-swagger3", "v1.2.3", "github.com/parvez3019/go-swagger3@v1.2.3"},
		{"v2 suffix", "github.com/foo/bar/v2", "v2.0.0", "github.com/foo/bar/v2@v2.0.0"},
		{"+incompatible version", "github.com/foo/bar", "v2.0.0+incompatible", "github.com/foo/bar@v2.0.0+incompatible"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := moduleCacheDir("/cache", c.path, c.version)
			want := filepath.Join("/cache", c.want)
			if got != want {
				t.Errorf("moduleCacheDir(%q, %q) = %q, want %q", c.path, c.version, got, want)
			}
		})
	}
}

// TestParseEmptyRequires: a go.mod with no requires yields an empty index, no error.
func TestParseEmptyRequires(t *testing.T) {
	dir := t.TempDir()
	goMod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	p, specs := newTestParser(goMod, filepath.Join(dir, "cache"))
	if err := p.Parse(); err != nil {
		t.Fatalf("Parse() = %v", err)
	}
	if len(specs.DepModules) != 0 {
		t.Errorf("DepModules = %d, want 0", len(specs.DepModules))
	}
}

// TestParseIgnoresReplaceDirective documents current behaviour: the index is built from
// `require` entries only. A `replace` directive is not applied, the module is indexed
// under its required path and version, and the replacement target is not added.
func TestParseIgnoresReplaceDirective(t *testing.T) {
	dir := t.TempDir()
	goMod := filepath.Join(dir, "go.mod")
	content := "module example.com/test\n\ngo 1.24\n\nrequire example.com/dep v1.0.0\n\nreplace example.com/dep => example.com/fork v2.0.0\n"
	if err := os.WriteFile(goMod, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(dir, "cache")
	p, specs := newTestParser(goMod, cache)
	if err := p.Parse(); err != nil {
		t.Fatalf("Parse() = %v", err)
	}
	if len(specs.DepModules) != 1 {
		t.Fatalf("DepModules = %d, want 1", len(specs.DepModules))
	}
	assertCacheDir(t, specs, "example.com/dep", filepath.Join(cache, "example.com/dep@v1.0.0"))
	for _, m := range specs.DepModules {
		if m.ImportPath == "example.com/fork" {
			t.Errorf("replace target should not be indexed")
		}
	}
}

// BenchmarkParse measures go.mod indexing over a module with ~2000 requires.
// It should stay low and scale with the require count, not with the size of the module cache.
func BenchmarkParse(b *testing.B) {
	dir := b.TempDir()
	goMod := writeSyntheticGoMod(b, dir, 50, 2000)
	cache := filepath.Join(dir, "cache")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := newTestParser(goMod, cache)
		if err := p.Parse(); err != nil {
			b.Fatal(err)
		}
	}
}
