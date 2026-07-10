package model

import (
	"go/ast"

	"github.com/parvez3019/go-swagger3/logger"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
)

type Utils struct {
	Path
	Flags
	*PkgAndSpecs

	*logger.Logger
}

type Path struct {
	ModulePath     string
	ModuleName     string
	MainFilePath   string
	HandlerPath    string
	GoModFilePath  string
	GoModCachePath string
	ExcludePaths   []string
}

type PkgAndSpecs struct {
	KnownPkgs     []Pkg
	KnownNamePkg  map[string]*Pkg
	KnownPathPkg  map[string]*Pkg
	KnownIDSchema map[string]*oas.SchemaObject

	TypeSpecs               map[string]map[string]*ast.TypeSpec
	TypeDescriptions        map[string]map[string]string // pkgName -> typeName -> @Description
	PkgPathAstPkgCache      map[string]map[string]*ast.Package
	PkgNameImportedPkgAlias map[string]map[string][]string

	// RegisteredEnums tracks type names registered via @Enum annotations.
	// Used so enum params do not rely solely on the type name containing "Enum".
	RegisteredEnums map[string]struct{}

	// DepModules indexes every module in go.mod by import path to its module-cache
	// directory, without walking it. The schema parser uses this to locate and index
	// dependency packages on demand (only the ones an annotation actually references),
	// instead of walking every required module up front. Sorted by import path
	// length descending so the most specific module wins a prefix match.
	DepModules []DepModule
}

type DepModule struct {
	ImportPath string
	CacheDir   string
}

type Flags struct {
	RunInDebugMode   bool
	RunInStrictMode  bool
	SchemaWithoutPkg bool
	Quiet            bool
}

type Pkg struct {
	Name string
	Path string
}
