package schema

import (
	"go/ast"
	"os"
	"path/filepath"
	"testing"

	"github.com/parvez3019/go-swagger3/logger"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
)

// newSchemaTestParser builds a schema parser with initialised, empty indexes and the
// given DepModules (which the gomod parser would normally populate, sorted longest
// import-path first).
func newSchemaTestParser(deps []model.DepModule) *parser {
	return &parser{
		Utils: model.Utils{
			PkgAndSpecs: &model.PkgAndSpecs{
				KnownPkgs:               []model.Pkg{},
				KnownNamePkg:            map[string]*model.Pkg{},
				KnownPathPkg:            map[string]*model.Pkg{},
				KnownIDSchema:           map[string]*SchemaObject{},
				TypeSpecs:               map[string]map[string]*ast.TypeSpec{},
				PkgPathAstPkgCache:      map[string]map[string]*ast.Package{},
				PkgNameImportedPkgAlias: map[string]map[string][]string{},
				DepModules:              deps,
			},
			Logger: logger.SetDebugMode(false),
		},
		OpenAPI: &OpenAPIObject{},
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

// writePkg creates a package directory with a single source file.
func writePkg(t *testing.T, dir, file, content string) {
	t.Helper()
	mustMkdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLocateDepPkgDir(t *testing.T) {
	root := t.TempDir()
	// Module-cache fixtures; contents don't matter locateDepPkgDir only stats dirs.
	barV1 := filepath.Join(root, "example.com/foo/bar@v1.0.0")
	barV2 := filepath.Join(root, "example.com/foo/bar/v2@v2.0.0")
	mustMkdir(t, barV1)
	mustMkdir(t, filepath.Join(barV2, "sub"))

	// Sorted longest import path first, as gomod.Parse produces.
	deps := []model.DepModule{
		{ImportPath: "example.com/foo/bar/v2", CacheDir: barV2},
		{ImportPath: "example.com/foo/bar", CacheDir: barV1},
	}
	p := newSchemaTestParser(deps)

	cases := []struct {
		name, importPath, wantDir string
		wantOK                    bool
	}{
		{"exact v1", "example.com/foo/bar", barV1, true},
		{"exact v2 (longest prefix wins)", "example.com/foo/bar/v2", barV2, true},
		{"subpath of v2", "example.com/foo/bar/v2/sub", filepath.Join(barV2, "sub"), true},
		{"matched module but missing dir", "example.com/foo/bar/baz", "", false},
		{"stdlib / unknown path", "net/http", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, ok := p.locateDepPkgDir(c.importPath)
			if ok != c.wantOK || dir != c.wantDir {
				t.Errorf("locateDepPkgDir(%q) = (%q, %v), want (%q, %v)", c.importPath, dir, ok, c.wantDir, c.wantOK)
			}
		})
	}
}

const depPkgSrc = `package dep

import "time"

type Widget struct {
	Name string
	When time.Time
}

type Gadget struct {
	ID int
}
`

func TestResolvePkgDirIndexesDependencyOnDemand(t *testing.T) {
	root := t.TempDir()
	depDir := filepath.Join(root, "example.com/dep@v1.0.0")
	writePkg(t, depDir, "dep.go", depPkgSrc)

	p := newSchemaTestParser([]model.DepModule{{ImportPath: "example.com/dep", CacheDir: depDir}})

	dir, ok := p.resolvePkgDir("example.com/dep")
	if !ok || dir != depDir {
		t.Fatalf("resolvePkgDir = (%q, %v), want (%q, true)", dir, ok, depDir)
	}

	// Registered into the lookup maps (but not into the eager KnownPkgs slice).
	if kp := p.KnownNamePkg["example.com/dep"]; kp == nil || kp.Path != depDir {
		t.Errorf("KnownNamePkg not set correctly: %+v", p.KnownNamePkg["example.com/dep"])
	}
	if p.KnownPathPkg[depDir] == nil {
		t.Errorf("KnownPathPkg[%q] not set", depDir)
	}
	if len(p.KnownPkgs) != 0 {
		t.Errorf("KnownPkgs = %d, want 0 (on-demand resolution must not grow the eager slice)", len(p.KnownPkgs))
	}

	// Type specs built from the package AST.
	for _, typ := range []string{"Widget", "Gadget"} {
		if _, ok := p.TypeSpecs["example.com/dep"][typ]; !ok {
			t.Errorf("TypeSpecs missing %q", typ)
		}
	}

	// Import aliases built from the package's imports.
	if got := p.PkgNameImportedPkgAlias["example.com/dep"]["time"]; len(got) != 1 || got[0] != "time" {
		t.Errorf("import alias for time = %v, want [time]", got)
	}
}

func TestResolvePkgDirIsIdempotent(t *testing.T) {
	root := t.TempDir()
	depDir := filepath.Join(root, "example.com/dep@v1.0.0")
	writePkg(t, depDir, "dep.go", depPkgSrc)
	p := newSchemaTestParser([]model.DepModule{{ImportPath: "example.com/dep", CacheDir: depDir}})

	if _, ok := p.resolvePkgDir("example.com/dep"); !ok {
		t.Fatal("first resolvePkgDir failed")
	}
	if _, cached := p.PkgPathAstPkgCache[depDir]; !cached {
		t.Errorf("package AST not cached after first resolution")
	}
	if _, ok := p.resolvePkgDir("example.com/dep"); !ok {
		t.Fatal("second resolvePkgDir failed")
	}
	// A second resolution must not duplicate the import-alias entry.
	if got := p.PkgNameImportedPkgAlias["example.com/dep"]["time"]; len(got) != 1 {
		t.Errorf("import alias duplicated on re-resolution: %v", got)
	}
}

func TestResolvePkgDirPrefersAlreadyKnownPackage(t *testing.T) {
	root := t.TempDir()
	projDir := filepath.Join(root, "project/handlers")
	writePkg(t, projDir, "h.go", "package handlers\n\ntype Req struct{ ID int }\n")

	// A different module in the index - if resolvePkgDir fell through to the cache
	// lookup it would not find example.com/proj, so returning projDir proves the
	// KnownNamePkg short-circuit is taken.
	p := newSchemaTestParser([]model.DepModule{{ImportPath: "example.com/other", CacheDir: root}})
	p.KnownNamePkg["example.com/proj/handlers"] = &model.Pkg{Name: "example.com/proj/handlers", Path: projDir}

	dir, ok := p.resolvePkgDir("example.com/proj/handlers")
	if !ok || dir != projDir {
		t.Fatalf("resolvePkgDir = (%q, %v), want (%q, true)", dir, ok, projDir)
	}
}

func TestResolvePkgDirUnresolvable(t *testing.T) {
	p := newSchemaTestParser([]model.DepModule{{ImportPath: "example.com/dep", CacheDir: t.TempDir()}})

	dir, ok := p.resolvePkgDir("net/http") // not in the index
	if ok || dir != "" {
		t.Fatalf("resolvePkgDir(net/http) = (%q, %v), want (\"\", false)", dir, ok)
	}
	if len(p.KnownNamePkg) != 0 || len(p.TypeSpecs) != 0 {
		t.Errorf("nothing should be registered for an unresolvable path")
	}
}

func TestGetTypeSpecLazilyIndexes(t *testing.T) {
	root := t.TempDir()
	depDir := filepath.Join(root, "example.com/dep@v1.0.0")
	writePkg(t, depDir, "dep.go", depPkgSrc)
	p := newSchemaTestParser([]model.DepModule{{ImportPath: "example.com/dep", CacheDir: depDir}})

	// TypeSpecs starts empty; getTypeSpec must resolve + index the package on first use.
	spec, ok := p.getTypeSpec("example.com/dep", "Widget")
	if !ok || spec == nil {
		t.Fatalf("getTypeSpec(Widget) = (%v, %v), want a spec", spec, ok)
	}
	if _, ok := p.getTypeSpec("example.com/dep", "DoesNotExist"); ok {
		t.Errorf("getTypeSpec for an absent type should return false")
	}
	if _, ok := p.getTypeSpec("net/http", "Client"); ok {
		t.Errorf("getTypeSpec for an unresolvable package should return false")
	}
}
