package parser

import (
	"fmt"
	. "github.com/parvez3019/goas/openApi3Schema"
	"go/ast"
	"go/token"
	"strings"
)

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
