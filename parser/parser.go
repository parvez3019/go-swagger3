package parser

import (
	"go/ast"

	"github.com/hanyue2020/go-swagger3/logger"
	. "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/apis"
	"github.com/hanyue2020/go-swagger3/parser/gomod"
	"github.com/hanyue2020/go-swagger3/parser/info"
	"github.com/hanyue2020/go-swagger3/parser/model"
	"github.com/hanyue2020/go-swagger3/parser/module"
	"github.com/hanyue2020/go-swagger3/parser/schema"
	log "github.com/sirupsen/logrus"
)

type parser struct {
	OpenAPI *OpenAPIObject

	apiParser    apis.Parser
	infoParser   info.Parser
	goModParser  gomod.Parser
	moduleParser module.Parser
	schemaParser schema.Parser

	model.Utils
}

func NewParser(modulePath, mainFilePath, handlerPath string, debug, strict bool) *parser {
	return &parser{
		Utils: model.Utils{
			Path:        getPaths(modulePath, mainFilePath, handlerPath),
			Flags:       geFlags(debug, strict),
			PkgAndSpecs: initPkgAndSpecs(),
		},
		OpenAPI: initOpenApiObject(),
	}
}

func (p *parser) Init() (*parser, error) {
	p.Logger = logger.SetDebugMode(p.RunInDebugMode)

	if err := p.verifyAndSetPaths(); err != nil {
		return nil, err
	}

	p.schemaParser = schema.NewParser(p.Utils, p.OpenAPI)
	p.apiParser = apis.NewParser(p.Utils, p.OpenAPI, p.schemaParser)
	p.infoParser = info.NewParser(p.Utils, p.OpenAPI)
	p.goModParser = gomod.NewParser(p.Utils)
	p.moduleParser = module.NewParser(p.Utils)

	return p, nil
}

func (p *parser) Parse() (OpenAPIObject, error) {
	log.Info("Parsing Initialized")
	err := p.infoParser.Parse()
	if err != nil {
		return OpenAPIObject{}, err
	}

	err = p.moduleParser.Parse()
	if err != nil {
		return OpenAPIObject{}, err
	}

	err = p.goModParser.Parse()
	if err != nil {
		return OpenAPIObject{}, err
	}

	err = p.apiParser.Parse()
	if err != nil {
		return OpenAPIObject{}, err
	}

	log.Info("Parsing Completed ...")
	return *p.OpenAPI, nil
}

func initOpenApiObject() *OpenAPIObject {
	return &OpenAPIObject{
		Version:  OpenAPIVersion,
		Paths:    make(PathsObject),
		Security: make([]map[string][]string, 0),
		Components: ComponentsObject{
			Schemas:         make(map[string]*SchemaObject),
			Parameters:      make(map[string]*ParameterObject),
			SecuritySchemes: make(map[string]*SecuritySchemeObject),
		},
	}
}

func geFlags(debug bool, strict bool) model.Flags {
	return model.Flags{
		RunInDebugMode:  debug,
		RunInStrictMode: strict,
	}
}

func getPaths(modulePath string, mainFilePath string, handlerPath string) model.Path {
	return model.Path{
		ModulePath:   modulePath,
		MainFilePath: mainFilePath,
		HandlerPath:  handlerPath,
	}
}

func initPkgAndSpecs() *model.PkgAndSpecs {
	return &model.PkgAndSpecs{
		KnownPkgs:               make([]model.Pkg, 0),
		KnownNamePkg:            make(map[string]*model.Pkg, 0),
		KnownPathPkg:            make(map[string]*model.Pkg, 0),
		KnownIDSchema:           make(map[string]*SchemaObject, 0),
		TypeSpecs:               make(map[string]map[string]*ast.TypeSpec, 0),
		PkgPathAstPkgCache:      make(map[string]map[string]*ast.Package, 0),
		PkgNameImportedPkgAlias: make(map[string]map[string][]string, 0),
	}
}
