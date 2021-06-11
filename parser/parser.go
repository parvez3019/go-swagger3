package parser

import (
	"fmt"
	"github.com/parvez3019/go-swagger3/logger"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/apis"
	"github.com/parvez3019/go-swagger3/parser/gomod"
	"github.com/parvez3019/go-swagger3/parser/info"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/module"
	"github.com/parvez3019/go-swagger3/parser/schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
	log "github.com/sirupsen/logrus"
	"go/ast"
	"os"
	"os/user"
	"path/filepath"
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

func NewParser(modulePath, mainFilePath, handlerPath string, debug, strict, schemaWithoutPkg bool) *parser {
	return &parser{
		Utils: model.Utils{
			Path:        getPaths(modulePath, mainFilePath, handlerPath),
			Flags:       geFlags(debug, strict, schemaWithoutPkg),
			PkgAndSpecs: initPkgAndSpecs(),
		},
		OpenAPI: initOpenApiObject(),
	}
}

func (p *parser) Init() (*parser, error) {
	p.Logger = logger.SetDebugMode(p.RunInDebugMode)

	// check modulePath is exist
	var err error
	p.ModulePath, err = filepath.Abs(p.ModulePath)
	if err != nil {
		return nil, err
	}
	moduleInfo, err := os.Stat(p.ModulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("cannot get information of %s: %s", p.ModulePath, err)
	}
	if !moduleInfo.IsDir() {
		return nil, fmt.Errorf("modulePath should be a directory")
	}
	p.Debugf("module path: %s", p.ModulePath)

	// check go.mod file is exist
	goModFilePath := filepath.Join(p.ModulePath, "go.mod")
	goModFileInfo, err := os.Stat(goModFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("cannot get information of %s: %s", goModFilePath, err)
	}
	if goModFileInfo.IsDir() {
		return nil, fmt.Errorf("%s should be a file", goModFilePath)
	}
	p.GoModFilePath = goModFilePath
	p.Debugf("go.mod file path: %s", p.GoModFilePath)

	// check mainFilePath is exist
	if p.MainFilePath == "" {
		fns, err := filepath.Glob(filepath.Join(p.ModulePath, "*.go"))
		if err != nil {
			return nil, err
		}
		for _, fn := range fns {
			if utils.IsMainFile(fn) {
				p.MainFilePath = fn
				break
			}
		}
	} else {
		mainFileInfo, err := os.Stat(p.MainFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, err
			}
			return nil, fmt.Errorf("cannot get information of %s: %s", p.MainFilePath, err)
		}
		if mainFileInfo.IsDir() {
			return nil, fmt.Errorf("mainFilePath should not be a directory")
		}
	}
	p.Debugf("main file path: %s", p.MainFilePath)

	// get module name from go.mod file
	moduleName := utils.GetModuleNameFromGoMod(goModFilePath)
	if moduleName == "" {
		return nil, fmt.Errorf("cannot get module name from %s", goModFileInfo)
	}
	p.ModuleName = moduleName
	p.Debugf("module name: %s", p.ModuleName)

	// check go module cache path is exist ($GOPATH/pkg/mod)
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		current, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("cannot get current user: %s", err)
		}
		goPath = filepath.Join(current.HomeDir, "go")
	}
	goModCachePath := filepath.Join(goPath, "pkg", "mod")
	goModCacheInfo, err := os.Stat(goModCachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("cannot get information of %s: %s", goModCachePath, err)
	}
	if !goModCacheInfo.IsDir() {
		return nil, fmt.Errorf("%s should be a directory", goModCachePath)
	}
	p.GoModCachePath = goModCachePath
	p.Debugf("go module cache path: %s", p.GoModCachePath)

	if p.HandlerPath != "" {
		p.HandlerPath, err = filepath.Abs(p.HandlerPath)
		if err != nil {
			return nil, err
		}
		_, err := os.Stat(p.HandlerPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, err
			}
			return nil, fmt.Errorf("cannot get information of %s: %s", p.HandlerPath, err)
		}
	}
	p.Debugf("handler path: %s", p.HandlerPath)

	p.schemaParser = schema.NewParser(p.Utils, p.OpenAPI)
	p.apiParser = apis.NewParser(p.Utils, p.OpenAPI, p.schemaParser)
	p.infoParser = info.NewParser(p.Utils, p.OpenAPI)
	p.goModParser = gomod.NewParser(p.Utils)
	p.moduleParser = module.NewParser(p.Utils)

	return p, nil
}

func (p *parser) Parse() (OpenAPIObject, error) {
	log.Info("Parsing Initialized")
	// parse basic info
	log.Info("Parsing Info ...")
	err := p.infoParser.Parse()
	if err != nil {
		return OpenAPIObject{}, err
	}

	// parse sub-package
	log.Info("Parsing Modules ...")
	err = p.moduleParser.Parse()
	if err != nil {
		return OpenAPIObject{}, err
	}

	// parse go.mod info
	log.Info("Parsing GoMod Info ...")
	err = p.goModParser.Parse()
	if err != nil {
		return OpenAPIObject{}, err
	}

	// parse APIs info
	err = p.apiParser.Parse()
	log.Info("Parsing APIs ...")
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

func geFlags(debug bool, strict bool, schemaWithoutPkg bool) model.Flags {
	return model.Flags{
		RunInDebugMode:   debug,
		RunInStrictMode:  strict,
		SchemaWithoutPkg: schemaWithoutPkg,
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
