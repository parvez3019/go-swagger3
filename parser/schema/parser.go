package schema

import (
	"go/ast"
	goParser "go/parser"
	"go/token"
	"os"
	"strings"

	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/utils"
)

type Parser interface {
	GetPkgAst(pkgPath string) (map[string]*ast.Package, error)
	RegisterType(pkgPath, pkgName, typeName string) (string, error)
	ParseSchemaObject(pkgPath, pkgName, typeName string) (*SchemaObject, error)
}

type parser struct {
	model.Utils
	OpenAPI *OpenAPIObject

	masker *utils.Masker
}

func NewParser(utils model.Utils, openAPIObject *OpenAPIObject, masker *utils.Masker) Parser {
	return &parser{
		Utils:   utils,
		OpenAPI: openAPIObject,
		masker:  masker,
	}
}

func (p *parser) GetPkgAst(pkgPath string) (map[string]*ast.Package, error) {
	if cache, ok := p.PkgPathAstPkgCache[pkgPath]; ok {
		return cache, nil
	}
	ignoreFileFilter := func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}
	astPackages, err := goParser.ParseDir(token.NewFileSet(), pkgPath, ignoreFileFilter, goParser.ParseComments)
	if err != nil {
		return nil, err
	}
	p.PkgPathAstPkgCache[pkgPath] = astPackages
	return astPackages, nil
}

func (p *parser) RegisterType(pkgPath, pkgName, typeName string) (string, error) {
	var registerTypeName string

	if utils.IsBasicGoType(typeName) || utils.IsInterfaceType(typeName) {
		registerTypeName = typeName
	} else if schemaObject, ok := p.KnownIDSchema[utils.GenSchemaObjectID(pkgName, typeName, p.SchemaWithoutPkg)]; ok {
		_, ok := p.OpenAPI.Components.Schemas[p.masker.ReplaceBackslash(typeName)]
		if !ok {
			p.OpenAPI.Components.Schemas[p.masker.ReplaceBackslash(typeName)] = schemaObject
		}
		return utils.GenSchemaObjectID(pkgName, typeName, p.SchemaWithoutPkg), nil
	} else {
		schemaObject, err := p.ParseSchemaObject(pkgPath, pkgName, typeName)
		if err != nil {
			return "", err
		}
		registerTypeName = schemaObject.ID
		_, ok := p.OpenAPI.Components.Schemas[p.masker.ReplaceBackslash(registerTypeName)]
		if !ok {
			p.OpenAPI.Components.Schemas[p.masker.ReplaceBackslash(registerTypeName)] = schemaObject
		}
	}
	return registerTypeName, nil
}

func (p *parser) ParseSchemaObject(pkgPath, pkgName, typeName string) (*SchemaObject, error) {
	schemaObject, err, isBasicType := p.parseBasicTypeSchemaObject(pkgPath, pkgName, typeName)
	if isBasicType {
		return schemaObject, err
	}

	return p.parseCustomTypeSchemaObject(pkgPath, pkgName, typeName)
}
