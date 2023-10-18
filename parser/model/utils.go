package model

import (
	"go/ast"

	"github.com/hanyue2020/go-swagger3/logger"
	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
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
}

type PkgAndSpecs struct {
	KnownPkgs     []Pkg
	KnownNamePkg  map[string]*Pkg
	KnownPathPkg  map[string]*Pkg
	KnownIDSchema map[string]*oas.SchemaObject

	TypeSpecs               map[string]map[string]*ast.TypeSpec
	PkgPathAstPkgCache      map[string]map[string]*ast.Package
	PkgNameImportedPkgAlias map[string]map[string][]string
}

type Flags struct {
	RunInDebugMode  bool
	RunInStrictMode bool
}

type Pkg struct {
	Name string
	Path string
}
