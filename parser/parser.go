package parser

import (
	"encoding/json"
	"fmt"
	. "frontiercargroup.org/dealer/goas/openApi3Schema"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/iancoleman/orderedmap"
	"github.com/mikunalpha/go-module"
)

type parser struct {
	ModulePath string
	ModuleName string

	MainFilePath string

	HandlerPath string

	GoModFilePath string

	GoModCachePath string

	OpenAPI OpenAPIObject

	KnownPkgs     []pkg
	KnownNamePkg  map[string]*pkg
	KnownPathPkg  map[string]*pkg
	KnownIDSchema map[string]*SchemaObject

	TypeSpecs               map[string]map[string]*ast.TypeSpec
	PkgPathAstPkgCache      map[string]map[string]*ast.Package
	PkgNameImportedPkgAlias map[string]map[string][]string

	Debug  bool
	Strict bool
}

type pkg struct {
	Name string
	Path string
}

func NewParser(modulePath, mainFilePath, handlerPath string, debug, strict bool) (*parser, error) {
	p := &parser{
		KnownPkgs:               []pkg{},
		KnownNamePkg:            map[string]*pkg{},
		KnownPathPkg:            map[string]*pkg{},
		KnownIDSchema:           map[string]*SchemaObject{},
		TypeSpecs:               map[string]map[string]*ast.TypeSpec{},
		PkgPathAstPkgCache:      map[string]map[string]*ast.Package{},
		PkgNameImportedPkgAlias: map[string]map[string][]string{},
		Debug:                   debug,
		Strict:                  strict,
	}
	p.OpenAPI.OpenAPI = OpenAPIVersion
	p.OpenAPI.Paths = make(PathsObject)
	p.OpenAPI.Security = []map[string][]string{}
	p.OpenAPI.Components.Schemas = make(map[string]*SchemaObject)
	p.OpenAPI.Components.Parameters = make(map[string]*ParameterObject)
	p.OpenAPI.Components.SecuritySchemes = make(map[string]*SecuritySchemeObject)

	// check modulePath is exist
	modulePath, _ = filepath.Abs(modulePath)
	moduleInfo, err := os.Stat(modulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("cannot get information of %s: %s", modulePath, err)
	}
	if !moduleInfo.IsDir() {
		return nil, fmt.Errorf("modulePath should be a directory")
	}
	p.ModulePath = modulePath
	p.debugf("module path: %s", p.ModulePath)

	// check go.mod file is exist
	goModFilePath := filepath.Join(modulePath, "go.mod")
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
	p.debugf("go.mod file path: %s", p.GoModFilePath)

	// check mainFilePath is exist
	if mainFilePath == "" {
		fns, err := filepath.Glob(filepath.Join(modulePath, "*.go"))
		if err != nil {
			return nil, err
		}
		for _, fn := range fns {
			if isMainFile(fn) {
				mainFilePath = fn
				break
			}
		}
	} else {
		mainFileInfo, err := os.Stat(mainFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, err
			}
			return nil, fmt.Errorf("cannot get information of %s: %s", mainFilePath, err)
		}
		if mainFileInfo.IsDir() {
			return nil, fmt.Errorf("mainFilePath should not be a directory")
		}
	}
	p.MainFilePath = mainFilePath
	p.debugf("main file path: %s", p.MainFilePath)

	// get module name from go.mod file
	moduleName := getModuleNameFromGoMod(goModFilePath)
	if moduleName == "" {
		return nil, fmt.Errorf("cannot get module name from %s", goModFileInfo)
	}
	p.ModuleName = moduleName
	p.debugf("module name: %s", p.ModuleName)

	// check go module cache path is exist ($GOPATH/pkg/mod)
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		user, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("cannot get current user: %s", err)
		}
		goPath = filepath.Join(user.HomeDir, "go")
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
	p.debugf("go module cache path: %s", p.GoModCachePath)

	if handlerPath != "" {
		handlerPath, _ = filepath.Abs(handlerPath)
		_, err := os.Stat(handlerPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, err
			}
			return nil, fmt.Errorf("cannot get information of %s: %s", handlerPath, err)
		}
	}
	p.HandlerPath = handlerPath
	p.debugf("handler path: %s", p.HandlerPath)

	return p, nil
}

func (p *parser) CreateOASFile(path string) error {
	// parse basic info
	err := p.parseInfo()
	if err != nil {
		return err
	}

	// parse sub-package
	err = p.parseModule()
	if err != nil {
		return err
	}

	// parse go.mod info
	err = p.parseGoMod()
	if err != nil {
		return err
	}

	// parse APIs info
	err = p.parseAPIs()
	if err != nil {
		return err
	}

	fd, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Can not create the file %s: %v", path, err)
	}
	defer fd.Close()

	output, err := json.MarshalIndent(p.OpenAPI, "", "  ")
	if err != nil {
		return err
	}
	_, err = fd.WriteString(string(output))

	return err
}

func (p *parser) parseInfo() error {
	fileTree, err := goparser.ParseFile(token.NewFileSet(), p.MainFilePath, nil, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("can not parse general API information: %v", err)
	}

	// Security Scopes are defined at a different level in the hierarchy as where they need to end up in the OpenAPI structure,
	// so a temporary list is needed.
	oauthScopes := make(map[string]map[string]string, 0)

	if fileTree.Comments != nil {
		for i := range fileTree.Comments {
			for _, comment := range strings.Split(fileTree.Comments[i].Text(), "\n") {
				attribute := strings.ToLower(strings.Split(comment, " ")[0])
				if len(attribute) == 0 || attribute[0] != '@' {
					continue
				}
				value := strings.TrimSpace(comment[len(attribute):])
				if len(value) == 0 {
					continue
				}
				// p.debug(attribute, value)
				switch attribute {
				case "@version":
					p.OpenAPI.Info.Version = value
				case "@title":
					p.OpenAPI.Info.Title = value
				case "@description":
					p.OpenAPI.Info.Description = value
				case "@termsofserviceurl":
					p.OpenAPI.Info.TermsOfService = value
				case "@contactname":
					if p.OpenAPI.Info.Contact == nil {
						p.OpenAPI.Info.Contact = &ContactObject{}
					}
					p.OpenAPI.Info.Contact.Name = value
				case "@contactemail":
					if p.OpenAPI.Info.Contact == nil {
						p.OpenAPI.Info.Contact = &ContactObject{}
					}
					p.OpenAPI.Info.Contact.Email = value
				case "@contacturl":
					if p.OpenAPI.Info.Contact == nil {
						p.OpenAPI.Info.Contact = &ContactObject{}
					}
					p.OpenAPI.Info.Contact.URL = value
				case "@licensename":
					if p.OpenAPI.Info.License == nil {
						p.OpenAPI.Info.License = &LicenseObject{}
					}
					p.OpenAPI.Info.License.Name = value
				case "@licenseurl":
					if p.OpenAPI.Info.License == nil {
						p.OpenAPI.Info.License = &LicenseObject{}
					}
					p.OpenAPI.Info.License.URL = value
				case "@server":
					fields := strings.Split(value, " ")
					s := ServerObject{URL: fields[0], Description: value[len(fields[0]):]}
					p.OpenAPI.Servers = append(p.OpenAPI.Servers, s)
				case "@security":
					fields := strings.Split(value, " ")
					security := map[string][]string{
						fields[0]: fields[1:],
					}
					p.OpenAPI.Security = append(p.OpenAPI.Security, security)
				case "@securityscheme":
					fields := strings.Split(value, " ")

					var scheme *SecuritySchemeObject
					if strings.Contains(fields[1], "oauth2") {
						if oauthScheme, ok := p.OpenAPI.Components.SecuritySchemes[fields[0]]; ok {
							scheme = oauthScheme
						} else {
							scheme = &SecuritySchemeObject{
								Type:       "oauth2",
								OAuthFlows: &SecuritySchemeOauthObject{},
							}
						}
					}

					if scheme == nil {
						scheme = &SecuritySchemeObject{
							Type: fields[1],
						}
					}
					switch fields[1] {
					case "http":
						scheme.Scheme = fields[2]
						scheme.Description = strings.Join(fields[3:], " ")
					case "apiKey":
						scheme.In = fields[2]
						scheme.Name = fields[3]
						scheme.Description = strings.Join(fields[4:], "")
					case "openIdConnect":
						scheme.OpenIdConnectUrl = fields[2]
						scheme.Description = strings.Join(fields[3:], " ")
					case "oauth2AuthCode":
						scheme.OAuthFlows.AuthorizationCode = &SecuritySchemeOauthFlowObject{
							AuthorizationUrl: fields[2],
							TokenUrl:         fields[3],
							Scopes:           make(map[string]string, 0),
						}
					case "oauth2Implicit":
						scheme.OAuthFlows.Implicit = &SecuritySchemeOauthFlowObject{
							AuthorizationUrl: fields[2],
							Scopes:           make(map[string]string, 0),
						}
					case "oauth2ResourceOwnerCredentials":
						scheme.OAuthFlows.ResourceOwnerPassword = &SecuritySchemeOauthFlowObject{
							TokenUrl: fields[2],
							Scopes:   make(map[string]string, 0),
						}
					case "oauth2ClientCredentials":
						scheme.OAuthFlows.ClientCredentials = &SecuritySchemeOauthFlowObject{
							TokenUrl: fields[2],
							Scopes:   make(map[string]string, 0),
						}
					}
					p.OpenAPI.Components.SecuritySchemes[fields[0]] = scheme
				case "@securityscope":
					fields := strings.Split(value, " ")

					if _, ok := oauthScopes[fields[0]]; !ok {
						oauthScopes[fields[0]] = make(map[string]string, 0)
					}

					oauthScopes[fields[0]][fields[1]] = strings.Join(fields[2:], " ")
				}
			}
		}
	}

	// Apply security scopes to their security schemes
	for scheme, _ := range p.OpenAPI.Components.SecuritySchemes {
		if p.OpenAPI.Components.SecuritySchemes[scheme].Type == "oauth2" {
			if scopes, ok := oauthScopes[scheme]; ok {
				p.OpenAPI.Components.SecuritySchemes[scheme].OAuthFlows.ApplyScopes(scopes)
			}
		}
	}

	if len(p.OpenAPI.Servers) < 1 {
		p.OpenAPI.Servers = append(p.OpenAPI.Servers, ServerObject{URL: "/", Description: "Default Server URL"})
	}

	if p.OpenAPI.Info.Title == "" {
		return fmt.Errorf("info.title cannot not be empty")
	}
	if p.OpenAPI.Info.Version == "" {
		return fmt.Errorf("info.version cannot not be empty")
	}
	for i := range p.OpenAPI.Servers {
		if p.OpenAPI.Servers[i].URL == "" {
			return fmt.Errorf("servers[%d].url cannot not be empty", i)
		}
	}

	return nil
}

func (p *parser) parseModule() error {
	walker := func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			if strings.HasPrefix(strings.Trim(strings.TrimPrefix(path, p.ModulePath), "/"), ".git") {
				return nil
			}
			fns, err := filepath.Glob(filepath.Join(path, "*.go"))
			if len(fns) == 0 || err != nil {
				return nil
			}
			// p.debug(path)
			name := filepath.Join(p.ModuleName, strings.TrimPrefix(path, p.ModulePath))
			name = filepath.ToSlash(name)
			p.KnownPkgs = append(p.KnownPkgs, pkg{
				Name: name,
				Path: path,
			})
			p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
			p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		}
		return nil
	}
	filepath.Walk(p.ModulePath, walker)
	return nil
}

func (p *parser) parseGoMod() error {
	b, err := ioutil.ReadFile(p.GoModFilePath)
	if err != nil {
		return err
	}
	goMod, err := module.Parse(b)
	if err != nil {
		return err
	}
	for i := range goMod.Requires {
		pathRunes := []rune{}
		for _, v := range goMod.Requires[i].Path {
			if !unicode.IsUpper(v) {
				pathRunes = append(pathRunes, v)
				continue
			}
			pathRunes = append(pathRunes, '!')
			pathRunes = append(pathRunes, unicode.ToLower(v))
		}
		pkgName := goMod.Requires[i].Path
		pkgPath := filepath.Join(p.GoModCachePath, string(pathRunes)+"@"+goMod.Requires[i].Version)
		pkgName = filepath.ToSlash(pkgName)
		p.KnownPkgs = append(p.KnownPkgs, pkg{
			Name: pkgName,
			Path: pkgPath,
		})
		p.KnownNamePkg[pkgName] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		p.KnownPathPkg[pkgPath] = &p.KnownPkgs[len(p.KnownPkgs)-1]

		walker := func(path string, info os.FileInfo, err error) error {
			if info != nil && info.IsDir() {
				if strings.HasPrefix(strings.Trim(strings.TrimPrefix(path, p.ModulePath), "/"), ".git") {
					return nil
				}
				fns, err := filepath.Glob(filepath.Join(path, "*.go"))
				if len(fns) == 0 || err != nil {
					return nil
				}
				// p.debug(path)
				name := filepath.Join(pkgName, strings.TrimPrefix(path, pkgPath))
				name = filepath.ToSlash(name)
				p.KnownPkgs = append(p.KnownPkgs, pkg{
					Name: name,
					Path: path,
				})
				p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
				p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
			}
			return nil
		}
		filepath.Walk(pkgPath, walker)
	}
	if p.Debug {
		for i := range p.KnownPkgs {
			p.debug(p.KnownPkgs[i].Name, "->", p.KnownPkgs[i].Path)
		}
	}
	return nil
}

func (p *parser) getPkgAst(pkgPath string) (map[string]*ast.Package, error) {
	if cache, ok := p.PkgPathAstPkgCache[pkgPath]; ok {
		return cache, nil
	}
	ignoreFileFilter := func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}
	astPackages, err := goparser.ParseDir(token.NewFileSet(), pkgPath, ignoreFileFilter, goparser.ParseComments)
	if err != nil {
		return nil, err
	}
	p.PkgPathAstPkgCache[pkgPath] = astPackages
	return astPackages, nil
}

func (p *parser) parseAPIs() error {
	err := p.parseImportStatements()
	if err != nil {
		return err
	}

	err = p.parseTypeSpecs()
	if err != nil {
		return err
	}

	// err = p.parsePaths()
	// if err != nil {
	// 	return err
	// }

	return p.parsePaths()
}

func (p *parser) parseImportStatements() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		astPkgs, err := p.getPkgAst(pkgPath)
		if err != nil {
			if p.Strict {
				return fmt.Errorf("parseImportStatements: parse of %s package cause error: %s", pkgPath, err)
			}

			p.debugf("parseImportStatements: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		p.PkgNameImportedPkgAlias[pkgName] = map[string][]string{}
		for _, astPackage := range astPkgs {
			for _, astFile := range astPackage.Files {
				for _, astImport := range astFile.Imports {
					importedPkgName := strings.Trim(astImport.Path.Value, "\"")
					importedPkgAlias := ""

					// _, known := p.KnownNamePkg[importedPkgName]
					// if !known {
					// 	p.debug("unknown", importedPkgName)
					// }

					if astImport.Name != nil && astImport.Name.Name != "." && astImport.Name.Name != "_" {
						importedPkgAlias = astImport.Name.String()
						// p.debug(importedPkgAlias, importedPkgName)
					} else {
						s := strings.Split(importedPkgName, "/")
						importedPkgAlias = s[len(s)-1]
					}

					exist := false
					for _, v := range p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias] {
						if v == importedPkgName {
							exist = true
							break
						}
					}
					if !exist {
						p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias] = append(p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias], importedPkgName)
					}
				}
			}
		}
	}
	return nil
}

func (p *parser) parseTypeSpecs() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		_, ok := p.TypeSpecs[pkgName]
		if !ok {
			p.TypeSpecs[pkgName] = map[string]*ast.TypeSpec{}
		}
		astPkgs, err := p.getPkgAst(pkgPath)
		if err != nil {
			if p.Strict {
				return fmt.Errorf("parseTypeSpecs: parse of %s package cause error: %s", pkgPath, err)
			}

			p.debugf("parseTypeSpecs: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		for _, astPackage := range astPkgs {
			for _, astFile := range astPackage.Files {
				for _, astDeclaration := range astFile.Decls {
					if astGenDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && astGenDeclaration.Tok == token.TYPE {
						// find type declaration
						for _, astSpec := range astGenDeclaration.Specs {
							if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
								p.TypeSpecs[pkgName][typeSpec.Name.String()] = typeSpec
							}
						}
					} else if astFuncDeclaration, ok := astDeclaration.(*ast.FuncDecl); ok {
						// find type declaration in func, method
						if astFuncDeclaration.Doc != nil && astFuncDeclaration.Doc.List != nil && astFuncDeclaration.Body != nil {
							funcName := astFuncDeclaration.Name.String()
							for _, astStmt := range astFuncDeclaration.Body.List {
								if astDeclStmt, ok := astStmt.(*ast.DeclStmt); ok {
									if astGenDeclaration, ok := astDeclStmt.Decl.(*ast.GenDecl); ok {
										for _, astSpec := range astGenDeclaration.Specs {
											if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
												// type in func
												if astFuncDeclaration.Recv == nil {
													p.TypeSpecs[pkgName][strings.Join([]string{funcName, typeSpec.Name.String()}, "@")] = typeSpec
													continue
												}
												// type in method
												var recvTypeName string
												if astStarExpr, ok := astFuncDeclaration.Recv.List[0].Type.(*ast.StarExpr); ok {
													recvTypeName = fmt.Sprintf("%s", astStarExpr.X)
												} else if astIdent, ok := astFuncDeclaration.Recv.List[0].Type.(*ast.Ident); ok {
													recvTypeName = astIdent.String()
												}
												p.TypeSpecs[pkgName][strings.Join([]string{recvTypeName, funcName, typeSpec.Name.String()}, "@")] = typeSpec
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (p *parser) parsePaths() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name
		// p.debug(pkgName, "->", pkgPath)

		astPkgs, err := p.getPkgAst(pkgPath)
		if err != nil {
			if p.Strict {
				return fmt.Errorf("parsePaths: parse of %s package cause error: %s", pkgPath, err)
			}

			p.debugf("parsePaths: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		for _, astPackage := range astPkgs {
			for _, astFile := range astPackage.Files {
				for _, astDeclaration := range astFile.Decls {
					if astFuncDeclaration, ok := astDeclaration.(*ast.FuncDecl); ok {
						if astFuncDeclaration.Doc != nil && astFuncDeclaration.Doc.List != nil {
							err = p.parseOperation(pkgPath, pkgName, astFuncDeclaration.Doc.List)
							if err != nil {
								return err
							}
						}
					}

					if astFuncDeclaration, ok := astDeclaration.(*ast.GenDecl); ok {
						if astFuncDeclaration.Doc != nil && astFuncDeclaration.Doc.List != nil {
							err = p.parseParameters(pkgPath, pkgName, astFuncDeclaration.Doc.List)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (p *parser) parseOperation(pkgPath, pkgName string, astComments []*ast.Comment) error {
	operation := &OperationObject{
		Responses: map[string]*ResponseObject{},
	}
	if !strings.HasPrefix(pkgPath, p.ModulePath) {
		// ignore this pkgName
		// p.debugf("parseOperation ignores %s", pkgPath)
		return nil
	} else if p.HandlerPath != "" && !strings.HasPrefix(pkgPath, p.HandlerPath) {
		return nil
	}
	var err error
	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if len(comment) == 0 {
			return nil
		}
		attribute := strings.Fields(comment)[0]
		switch strings.ToLower(attribute) {
		case "@title":
			operation.Summary = strings.TrimSpace(comment[len(attribute):])
		case "@description":
			operation.Description = strings.Join([]string{operation.Description, strings.TrimSpace(comment[len(attribute):])}, " ")
		case "@param":
			err = p.parseParamComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
		//case "@parameters":
		//	err = p.parseHeaderComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
		case "@success", "@failure":
			err = p.parseResponseComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
		case "@resource", "@tag":
			resource := strings.TrimSpace(comment[len(attribute):])
			if resource == "" {
				resource = "others"
			}
			if !isInStringList(operation.Tags, resource) {
				operation.Tags = append(operation.Tags, resource)
			}
		case "@route", "@router":
			err = p.parseRouteComment(operation, comment)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseParameters(pkgPath, pkgName string, astComments []*ast.Comment) error {
	var err error
	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if len(comment) == 0 {
			return nil
		}
		attribute := strings.Fields(comment)[0]
		switch strings.ToLower(attribute) {
		case "@headerparameters":
			err = p.parseHeaderParameters(pkgPath, pkgName, strings.TrimSpace(comment[len(attribute):]))
		}
	}
	return err
}

func (p *parser) parseHeaderParameters(pkgPath string, pkgName string, comment string) error {
	schema, err := p.parseSchemaObject(pkgPath, pkgName, comment)
	if err != nil {
		return fmt.Errorf("parseHeaderComment can not parse Header comment schema %s", comment)
	}
	if schema.Properties == nil {
		return fmt.Errorf("parseHeaderComment can not parse Header comment schema %s", comment)
	}
	for _, key := range schema.Properties.Keys() {
		value, _ := schema.Properties.Get(key)
		currentSchemaObj, ok := value.(*SchemaObject)
		if !ok {
			return fmt.Errorf("parseHeaderComment can not parse Header Params %s", comment)
		}

		paramObj := &ParameterObject{
			Name:        key,
			In:          "header",
			Required:    isRequiredParam(schema.Required, key),
			Example:     currentSchemaObj.Example,
			Description: currentSchemaObj.Description,
			Schema:      currentSchemaObj,
		}
		p.OpenAPI.Components.Parameters[key] = paramObj
	}
	return nil
}

func isRequiredParam(requiredParams []string, key string) bool {
	for _, param := range requiredParams {
		if key == param {
			return true
		}
	}
	return false
}

func (p *parser) parseParamComment(pkgPath, pkgName string, operation *OperationObject, comment string) error {
	// {name}  {in}  {goType}  {required}  {description}
	// user    body  User      true        "Info of a user."
	// f       file  ignored   true        "Upload a file."
	re := regexp.MustCompile(`([-\w]+)[\s]+([\w]+)[\s]+([\w./\[\]]+)[\s]+([\w]+)[\s]+"([^"]+)"`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) != 6 {
		return fmt.Errorf("parseParamComment can not parse param comment \"%s\"", comment)
	}

	parameterObject := ParameterObject{}
	appendName(&parameterObject, matches[1])
	appendIn(&parameterObject, matches[2])
	appendRequired(&parameterObject, matches[4])
	appendDescription(&parameterObject, matches[5])

	goType := getType(re, matches)

	// `file`, `form`
	appendRequestBody(operation, parameterObject, goType)

	// `path`, `query`, `header`, `cookie`
	if parameterObject.In != "body" {
		return p.appendQueryParam(pkgPath, pkgName, operation, parameterObject, goType)
	}

	if operation.RequestBody == nil {
		operation.RequestBody = &RequestBodyObject{
			Content:  map[string]*MediaTypeObject{},
			Required: parameterObject.Required,
		}
	}

	if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[]") || goType == "time.Time" {
		schema, err := p.parseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.debug("parseResponseComment cannot parse goType", goType)
		}
		operation.RequestBody.Content[ContentTypeJson] = &MediaTypeObject{
			Schema: *schema,
		}
	} else {
		typeName, err := p.registerType(pkgPath, pkgName, matches[3])
		if err != nil {
			return err
		}
		if isBasicGoType(typeName) {
			operation.RequestBody.Content[ContentTypeJson] = &MediaTypeObject{
				Schema: SchemaObject{
					Type: "string",
				},
			}
		} else {
			operation.RequestBody.Content[ContentTypeJson] = &MediaTypeObject{
				Schema: SchemaObject{
					Ref: addSchemaRefLinkPrefix(typeName),
				},
			}
		}
	}

	return nil
}

func (p *parser) appendQueryParam(pkgPath string, pkgName string, operation *OperationObject, parameterObject ParameterObject, goType string) error {
	if parameterObject.In == "path" {
		parameterObject.Required = true
	}
	if goType == "time.Time" {
		var err error
		parameterObject.Schema, err = p.parseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.debug("parseResponseComment cannot parse goType", goType)
		}
		operation.Parameters = append(operation.Parameters, parameterObject)
	} else if isGoTypeOASType(goType) {
		parameterObject.Schema = &SchemaObject{
			Type:        goTypesOASTypes[goType],
			Format:      goTypesOASFormats[goType],
			Description: parameterObject.Description,
		}
		operation.Parameters = append(operation.Parameters, parameterObject)
	} else if strings.Contains(goType, "model.") {
		typeName, err := p.registerType(pkgPath, pkgName, goType)
		if err != nil {
			p.debug("parse param model type failed", goType)
			return err
		}
		parameterObject.Schema = &SchemaObject{
			Ref:  addSchemaRefLinkPrefix(typeName),
			Type: typeName,
		}
		operation.Parameters = append(operation.Parameters, parameterObject)
	}
	return nil
}

func appendRequestBody(operation *OperationObject, parameterObject ParameterObject, goType string) {
	if !(parameterObject.In == "file" || parameterObject.In == "form") {
		return
	}
	if operation.RequestBody == nil {
		operation.RequestBody = &RequestBodyObject{
			Content: map[string]*MediaTypeObject{
				ContentTypeForm: {Schema: SchemaObject{Type: "object", Properties: orderedmap.New()}},
			},
			Required: parameterObject.Required,
		}
	}
	if parameterObject.In == "file" {
		operation.RequestBody.Content[ContentTypeForm].Schema.Properties.Set(parameterObject.Name, &SchemaObject{
			Type:        "string",
			Format:      "binary",
			Description: parameterObject.Description,
		})
	}
	if isGoTypeOASType(goType) {
		operation.RequestBody.Content[ContentTypeForm].Schema.Properties.Set(parameterObject.Name, &SchemaObject{
			Type:        goTypesOASTypes[goType],
			Format:      goTypesOASFormats[goType],
			Description: parameterObject.Description,
		})
	}
}

func getType(re *regexp.Regexp, matches []string) string {
	re = regexp.MustCompile(`\[\w*\]`)
	goType := re.ReplaceAllString(matches[3], "[]")
	return goType
}

func appendRequired(paramObject *ParameterObject, isRequired string) {
	switch strings.ToLower(isRequired) {
	case "true", "required":
		paramObject.Required = true
	}
}

func appendDescription(parameterObject *ParameterObject, description string) {
	parameterObject.Description = description
}

func appendIn(parameterObject *ParameterObject, in string) {
	parameterObject.In = in
}

func appendName(parameterObject *ParameterObject, name string) {
	parameterObject.Name = name
}

func (p *parser) parseResponseComment(pkgPath, pkgName string, operation *OperationObject, comment string) error {
	// {status}  {jsonType}  {goType}     {description}
	// 201       object      models.User  "User Model"
	re := regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([\w\-\.\/\[\]]+)[^"]*(.*)?`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) != 5 {
		return fmt.Errorf("parseResponseComment can not parse response comment \"%s\"", comment)
	}

	status := matches[1]
	_, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("parseResponseComment: http status must be int, but got %s", status)
	}
	switch matches[2] {
	case "object", "array", "{object}", "{array}":
	default:
		return fmt.Errorf("parseResponseComment: invalid jsonType %s", matches[2])
	}
	responseObject := &ResponseObject{
		Content: map[string]*MediaTypeObject{},
	}
	responseObject.Description = strings.Trim(matches[4], "\"")

	re = regexp.MustCompile(`\[\w*\]`)
	goType := re.ReplaceAllString(matches[3], "[]")
	if strings.HasPrefix(goType, "map[]") {
		schema, err := p.parseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.debug("parseResponseComment cannot parse goType", goType)
		}
		responseObject.Content[ContentTypeJson] = &MediaTypeObject{
			Schema: *schema,
		}
	} else if strings.HasPrefix(goType, "[]") {
		goType = strings.Replace(goType, "[]", "", -1)
		typeName, err := p.registerType(pkgPath, pkgName, goType)
		if err != nil {
			return err
		}

		var s SchemaObject

		if isBasicGoType(typeName) {
			s = SchemaObject{
				Type: "string",
			}
		} else {
			s = SchemaObject{
				Ref: addSchemaRefLinkPrefix(typeName),
			}
		}

		responseObject.Content[ContentTypeJson] = &MediaTypeObject{
			Schema: SchemaObject{
				Type:  "array",
				Items: &s,
			},
		}
	} else {
		typeName, err := p.registerType(pkgPath, pkgName, matches[3])
		if err != nil {
			return err
		}
		if isBasicGoType(typeName) {
			responseObject.Content[ContentTypeText] = &MediaTypeObject{
				Schema: SchemaObject{
					Type: "string",
				},
			}
		} else {
			responseObject.Content[ContentTypeJson] = &MediaTypeObject{
				Schema: SchemaObject{
					Ref: addSchemaRefLinkPrefix(typeName),
				},
			}
		}
	}
	operation.Responses[status] = responseObject

	return nil
}

func (p *parser) parseRouteComment(operation *OperationObject, comment string) error {
	sourceString := strings.TrimSpace(comment[len("@Router"):])

	// /path [method]
	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	matches := re.FindStringSubmatch(sourceString)
	if len(matches) != 3 {
		return fmt.Errorf("Can not parse router comment \"%s\", skipped", comment)
	}

	_, ok := p.OpenAPI.Paths[matches[1]]
	if !ok {
		p.OpenAPI.Paths[matches[1]] = &PathItemObject{}
	}

	switch strings.ToUpper(matches[2]) {
	case http.MethodGet:
		p.OpenAPI.Paths[matches[1]].Get = operation
	case http.MethodPost:
		p.OpenAPI.Paths[matches[1]].Post = operation
	case http.MethodPatch:
		p.OpenAPI.Paths[matches[1]].Patch = operation
	case http.MethodPut:
		p.OpenAPI.Paths[matches[1]].Put = operation
	case http.MethodDelete:
		p.OpenAPI.Paths[matches[1]].Delete = operation
	case http.MethodOptions:
		p.OpenAPI.Paths[matches[1]].Options = operation
	case http.MethodHead:
		p.OpenAPI.Paths[matches[1]].Head = operation
	case http.MethodTrace:
		p.OpenAPI.Paths[matches[1]].Trace = operation
	}

	return nil
}

func (p *parser) registerType(pkgPath, pkgName, typeName string) (string, error) {
	var registerTypeName string

	if isBasicGoType(typeName) {
		registerTypeName = typeName
	} else if _, ok := p.KnownIDSchema[genSchemeaObjectID(pkgName, typeName)]; ok {
		return genSchemeaObjectID(pkgName, typeName), nil
	} else {
		schemaObject, err := p.parseSchemaObject(pkgPath, pkgName, typeName)
		if err != nil {
			return "", err
		}
		registerTypeName = schemaObject.ID
		_, ok := p.OpenAPI.Components.Schemas[replaceBackslash(registerTypeName)]
		if !ok {
			p.OpenAPI.Components.Schemas[replaceBackslash(registerTypeName)] = schemaObject
		}
	}
	return registerTypeName, nil
}

func (p *parser) parseSchemaObject(pkgPath, pkgName, typeName string) (*SchemaObject, error) {
	var typeSpec *ast.TypeSpec
	var exist bool
	var schemaObject SchemaObject
	var err error

	// handler basic and some specific typeName
	if strings.HasPrefix(typeName, "[]") {
		schemaObject.Type = "array"
		itemTypeName := typeName[2:]
		schema, ok := p.KnownIDSchema[genSchemeaObjectID(pkgName, itemTypeName)]
		if ok {
			schemaObject.Items = &SchemaObject{Ref: addSchemaRefLinkPrefix(schema.ID)}
			return &schemaObject, nil
		}
		schemaObject.Items, err = p.parseSchemaObject(pkgPath, pkgName, itemTypeName)
		if err != nil {
			return nil, err
		}
		return &schemaObject, nil
	} else if strings.HasPrefix(typeName, "map[]") {
		schemaObject.Type = "object"
		itemTypeName := typeName[5:]
		schema, ok := p.KnownIDSchema[genSchemeaObjectID(pkgName, itemTypeName)]
		if ok {
			schemaObject.Items = &SchemaObject{Ref: addSchemaRefLinkPrefix(schema.ID)}
			return &schemaObject, nil
		}
		schemaProperty, err := p.parseSchemaObject(pkgPath, pkgName, itemTypeName)
		if err != nil {
			return nil, err
		}
		schemaObject.Properties = orderedmap.New()
		schemaObject.Properties.Set("key", schemaProperty)
		return &schemaObject, nil
	} else if typeName == "time.Time" {
		schemaObject.Type = "string"
		schemaObject.Format = "date-time"
		return &schemaObject, nil
	} else if strings.HasPrefix(typeName, "interface{}") {
		return &schemaObject, nil
	} else if isGoTypeOASType(typeName) {
		schemaObject.Type = goTypesOASTypes[typeName]
		return &schemaObject, nil
	}

	// handler other type
	typeNameParts := strings.Split(typeName, ".")
	if len(typeNameParts) == 1 {
		typeSpec, exist = p.getTypeSpec(pkgPath, pkgName, typeName)
		if !exist {
			log.Fatalf("Can not find definition of %s ast.TypeSpec. Current package %s", typeName, pkgName)
		}
		schemaObject.PkgName = pkgName
		schemaObject.ID = genSchemeaObjectID(pkgName, typeName)
		p.KnownIDSchema[schemaObject.ID] = &schemaObject
	} else {
		guessPkgName := strings.Join(typeNameParts[:len(typeNameParts)-1], "/")
		guessPkgPath := ""
		for i := range p.KnownPkgs {
			if guessPkgName == p.KnownPkgs[i].Name {
				guessPkgPath = p.KnownPkgs[i].Path
				break
			}
		}
		guessTypeName := typeNameParts[len(typeNameParts)-1]
		typeSpec, exist = p.getTypeSpec(guessPkgName, guessPkgName, guessTypeName)
		if !exist {
			found := false
			for k := range p.PkgNameImportedPkgAlias[pkgName] {
				if k == guessPkgName && len(p.PkgNameImportedPkgAlias[pkgName][guessPkgName]) != 0 {
					found = true
					break
				}
			}
			if !found {
				p.debugf("unknown guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
				return &schemaObject, nil
			}
			guessPkgName = p.PkgNameImportedPkgAlias[pkgName][guessPkgName][0]
			guessPkgPath = ""
			for i := range p.KnownPkgs {
				if guessPkgName == p.KnownPkgs[i].Name {
					guessPkgPath = p.KnownPkgs[i].Path
					break
				}
			}
			// p.debugf("guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
			typeSpec, exist = p.getTypeSpec(guessPkgPath, guessPkgName, guessTypeName)
			if !exist {
				p.debugf("can not find definition of guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
				return &schemaObject, nil
			}
			schemaObject.PkgName = guessPkgName
			schemaObject.ID = genSchemeaObjectID(guessPkgName, guessTypeName)
			p.KnownIDSchema[schemaObject.ID] = &schemaObject
		}
		pkgPath, pkgName = guessPkgPath, guessPkgName
	}

	if astIdent, ok := typeSpec.Type.(*ast.Ident); ok {
		if astIdent != nil {
			schemaObject.Type = astIdent.Name
		}
	} else if astStructType, ok := typeSpec.Type.(*ast.StructType); ok {
		schemaObject.Type = "object"
		if astStructType.Fields != nil {
			p.parseSchemaPropertiesFromStructFields(pkgPath, pkgName, &schemaObject, astStructType.Fields.List)
		}
	} else if astArrayType, ok := typeSpec.Type.(*ast.ArrayType); ok {
		schemaObject.Type = "array"
		schemaObject.Items = &SchemaObject{}
		typeAsString := p.getTypeAsString(astArrayType.Elt)
		typeAsString = strings.TrimLeft(typeAsString, "*")
		if !isBasicGoType(typeAsString) {
			schemaItemsSchemeaObjectID, err := p.registerType(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug("parseSchemaObject parse array items err:", err)
			} else {
				schemaObject.Items.Ref = addSchemaRefLinkPrefix(schemaItemsSchemeaObjectID)
			}
		} else if isGoTypeOASType(typeAsString) {
			schemaObject.Items.Type = goTypesOASTypes[typeAsString]
		}
	} else if astMapType, ok := typeSpec.Type.(*ast.MapType); ok {
		schemaObject.Type = "object"
		schemaObject.Properties = orderedmap.New()
		propertySchema := &SchemaObject{}
		schemaObject.Properties.Set("key", propertySchema)
		typeAsString := p.getTypeAsString(astMapType.Value)
		typeAsString = strings.TrimLeft(typeAsString, "*")
		if !isBasicGoType(typeAsString) {
			schemaItemsSchemeaObjectID, err := p.registerType(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug("parseSchemaObject parse array items err:", err)
			} else {
				propertySchema.Ref = addSchemaRefLinkPrefix(schemaItemsSchemeaObjectID)
			}
		} else if isGoTypeOASType(typeAsString) {
			propertySchema.Type = goTypesOASTypes[typeAsString]
		}
	}

	return &schemaObject, nil
}

func (p *parser) getTypeSpec(pkgPath, pkgName, typeName string) (*ast.TypeSpec, bool) {
	pkgTypeSpecs, exist := p.TypeSpecs[pkgName]
	if !exist {
		return nil, false
	}
	astTypeSpec, exist := pkgTypeSpecs[typeName]
	if !exist {
		return nil, false
	}
	return astTypeSpec, true
}

func (p *parser) parseSchemaPropertiesFromStructFields(pkgPath, pkgName string, structSchema *SchemaObject, astFields []*ast.Field) {
	if astFields == nil {
		return
	}
	var err error
	structSchema.Properties = orderedmap.New()
	if structSchema.DisabledFieldNames == nil {
		structSchema.DisabledFieldNames = map[string]struct{}{}
	}
astFieldsLoop:
	for _, astField := range astFields {
		if len(astField.Names) == 0 {
			continue
		}
		fieldSchema := &SchemaObject{}
		typeAsString := p.getTypeAsString(astField.Type)
		typeAsString = strings.TrimLeft(typeAsString, "*")
		if strings.HasPrefix(typeAsString, "[]") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug(err)
				return
			}
		} else if strings.HasPrefix(typeAsString, "map[]") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug(err)
				return
			}
		} else if typeAsString == "time.Time" {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug(err)
				return
			}
		} else if strings.HasPrefix(typeAsString, "interface{}") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug(err)
				return
			}
		} else if !isBasicGoType(typeAsString) {
			fieldSchemaSchemeaObjectID, err := p.registerType(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug("parseSchemaPropertiesFromStructFields err:", err)
			} else {
				fieldSchema.ID = fieldSchemaSchemeaObjectID
				schema, ok := p.KnownIDSchema[fieldSchemaSchemeaObjectID]
				if ok {
					fieldSchema.Type = schema.Type
					if schema.Items != nil {
						fieldSchema.Items = schema.Items
					}
				}
				fieldSchema.Ref = addSchemaRefLinkPrefix(fieldSchemaSchemeaObjectID)
			}
		} else if isGoTypeOASType(typeAsString) {
			fieldSchema.Type = goTypesOASTypes[typeAsString]
		}

		name := astField.Names[0].Name
		fieldSchema.FieldName = name
		_, disabled := structSchema.DisabledFieldNames[name]
		if disabled {
			continue
		}

		if astField.Tag != nil {
			astFieldTag := reflect.StructTag(strings.Trim(astField.Tag.Value, "`"))
			tagText := ""

			if tag := astFieldTag.Get("go-swagger3"); tag != "" {
				tagText = tag
			}
			tagValues := strings.Split(tagText, ",")
			for _, v := range tagValues {
				if v == "-" {
					structSchema.DisabledFieldNames[name] = struct{}{}
					fieldSchema.Deprecated = true
					continue astFieldsLoop
				}
			}

			if tag := astFieldTag.Get("json"); tag != "" {
				tagText = tag
			}
			tagValues = strings.Split(tagText, ",")
			isRequired := false
			for _, v := range tagValues {
				if v == "-" {
					structSchema.DisabledFieldNames[name] = struct{}{}
					fieldSchema.Deprecated = true
					continue astFieldsLoop
				} else if v == "required" {
					isRequired = true
				} else if v != "" && v != "required" && v != "omitempty" {
					name = v
				}
			}

			if tag := astFieldTag.Get("example"); tag != "" {
				switch fieldSchema.Type {
				case "boolean":
					fieldSchema.Example, _ = strconv.ParseBool(tag)
				case "integer":
					fieldSchema.Example, _ = strconv.Atoi(tag)
				case "number":
					fieldSchema.Example, _ = strconv.ParseFloat(tag, 64)
				case "array":
					b, err := json.RawMessage(tag).MarshalJSON()
					if err != nil {
						fieldSchema.Example = "invalid example"
					} else {
						sliceOfInterface := []interface{}{}
						err := json.Unmarshal(b, &sliceOfInterface)
						if err != nil {
							fieldSchema.Example = "invalid example"
						} else {
							fieldSchema.Example = sliceOfInterface
						}
					}
				case "object":
					b, err := json.RawMessage(tag).MarshalJSON()
					if err != nil {
						fieldSchema.Example = "invalid example"
					} else {
						mapOfInterface := map[string]interface{}{}
						err := json.Unmarshal(b, &mapOfInterface)
						if err != nil {
							fieldSchema.Example = "invalid example"
						} else {
							fieldSchema.Example = mapOfInterface
						}
					}
				default:
					fieldSchema.Example = tag
				}

				if fieldSchema.Example != nil && len(fieldSchema.Ref) != 0 {
					fieldSchema.Ref = ""
				}
			}

			if _, ok := astFieldTag.Lookup("required"); ok || isRequired {
				structSchema.Required = append(structSchema.Required, name)
			}

			if desc := astFieldTag.Get("description"); desc != "" {
				fieldSchema.Description = desc
			}
		}

		structSchema.Properties.Set(name, fieldSchema)
	}
	for _, astField := range astFields {
		if len(astField.Names) > 0 {
			continue
		}
		fieldSchema := &SchemaObject{}
		typeAsString := p.getTypeAsString(astField.Type)
		typeAsString = strings.TrimLeft(typeAsString, "*")
		if strings.HasPrefix(typeAsString, "[]") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug(err)
				return
			}
		} else if strings.HasPrefix(typeAsString, "map[]") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug(err)
				return
			}
		} else if typeAsString == "time.Time" {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug(err)
				return
			}
		} else if strings.HasPrefix(typeAsString, "interface{}") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug(err)
				return
			}
		} else if !isBasicGoType(typeAsString) {
			fieldSchemaSchemeaObjectID, err := p.registerType(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.debug("parseSchemaPropertiesFromStructFields err:", err)
			} else {
				fieldSchema.ID = fieldSchemaSchemeaObjectID
				schema, ok := p.KnownIDSchema[fieldSchemaSchemeaObjectID]
				if ok {
					fieldSchema.Type = schema.Type
					if schema.Items != nil {
						fieldSchema.Items = schema.Items
					}
				}
				fieldSchema.Ref = addSchemaRefLinkPrefix(fieldSchemaSchemeaObjectID)
			}
		} else if isGoTypeOASType(typeAsString) {
			fieldSchema.Type = goTypesOASTypes[typeAsString]
		}
		// embedded type
		if len(astField.Names) == 0 {
			if fieldSchema.Properties != nil {
				for _, propertyName := range fieldSchema.Properties.Keys() {
					_, exist := structSchema.Properties.Get(propertyName)
					if exist {
						continue
					}
					propertySchema, _ := fieldSchema.Properties.Get(propertyName)
					structSchema.Properties.Set(propertyName, propertySchema)
				}
			} else if len(fieldSchema.Ref) != 0 && len(fieldSchema.ID) != 0 {
				refSchema, ok := p.KnownIDSchema[fieldSchema.ID]
				if ok {
					for _, propertyName := range refSchema.Properties.Keys() {
						refPropertySchema, _ := refSchema.Properties.Get(propertyName)
						_, disabled := structSchema.DisabledFieldNames[refPropertySchema.(*SchemaObject).FieldName]
						if disabled {
							continue
						}
						// p.debug(">", propertyName)
						_, exist := structSchema.Properties.Get(propertyName)
						if exist {
							continue
						}

						structSchema.Properties.Set(propertyName, refPropertySchema)
					}
				}
			}
			continue
		}
	}
}

func (p *parser) getTypeAsString(fieldType interface{}) string {
	astArrayType, ok := fieldType.(*ast.ArrayType)
	if ok {
		return fmt.Sprintf("[]%v", p.getTypeAsString(astArrayType.Elt))
	}

	astMapType, ok := fieldType.(*ast.MapType)
	if ok {
		return fmt.Sprintf("map[]%v", p.getTypeAsString(astMapType.Value))
	}

	_, ok = fieldType.(*ast.InterfaceType)
	if ok {
		return "interface{}"
	}

	astStarExpr, ok := fieldType.(*ast.StarExpr)
	if ok {
		// return fmt.Sprintf("*%v", p.getTypeAsString(astStarExpr.X))
		return fmt.Sprintf("%v", p.getTypeAsString(astStarExpr.X))
	}

	astSelectorExpr, ok := fieldType.(*ast.SelectorExpr)
	if ok {
		packageNameIdent, _ := astSelectorExpr.X.(*ast.Ident)
		if packageNameIdent != nil && packageNameIdent.Obj != nil && packageNameIdent.Obj.Decl != nil {
			a, ok := packageNameIdent.Obj.Decl.(DECL)
			if ok {
				fmt.Println(a)
			}
		}

		return packageNameIdent.Name + "." + astSelectorExpr.Sel.Name
	}

	return fmt.Sprint(fieldType)
}

type DECL struct {
	Type struct {
		Name string
	}
}

func (p *parser) debug(v ...interface{}) {
	if p.Debug {
		log.Println(v...)
	}
}

func (p *parser) debugf(format string, args ...interface{}) {
	if p.Debug {
		log.Printf(format, args...)
	}
}
