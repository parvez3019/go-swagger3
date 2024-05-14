package apis

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func (p *parser) parseTypeSpecs() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		_, ok := p.TypeSpecs[pkgName]
		if !ok {
			p.TypeSpecs[pkgName] = map[string]*ast.TypeSpec{}
		}
		astPkgs, err := p.schemaParser.GetPkgAst(pkgPath)
		if err != nil {
			if p.RunInStrictMode {
				return fmt.Errorf("parseTypeSpecs: parse of %s package cause error: %s", pkgPath, err)
			}

			p.Debugf("parseTypeSpecs: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		for _, astPackage := range astPkgs {
			p.parseTypeSpecsFromPackage(astPackage, pkgName)
		}
	}
	// After all type specifications have been parsed, resolve the type aliases
	for pkgName, aliases := range p.TypeAliases {
		for alias, original := range aliases {
			if originalTypeSpec, ok := p.TypeSpecs[pkgName][original]; ok {
				p.TypeSpecs[pkgName][alias] = originalTypeSpec
			}
		}
	}

	return nil
}

func (p *parser) parseTypeAlias(typeSpec *ast.TypeSpec, pkgName string) {
	if ident, ok := typeSpec.Type.(*ast.Ident); ok {
		if _, ok := p.TypeAliases[pkgName]; !ok {
			p.TypeAliases[pkgName] = map[string]string{}
		}
		p.TypeAliases[pkgName][typeSpec.Name.String()] = ident.Name
	}
}

func (p *parser) parseTypeSpecsFromPackage(astPackage *ast.Package, pkgName string) {
	for _, astFile := range astPackage.Files {
		p.parseTypeSpecsFromFile(astFile, pkgName)
	}
}

func (p *parser) parseTypeSpecsFromFile(astFile *ast.File, pkgName string) {
	for _, astDeclaration := range astFile.Decls {
		p.parseTypeSpecFromDeclaration(astDeclaration, pkgName)
	}
}

func (p *parser) parseTypeSpecFromDeclaration(astDeclaration ast.Decl, pkgName string) {
	if astGenDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && astGenDeclaration.Tok == token.TYPE {
		p.parseTypeSpecFromGenDeclaration(astGenDeclaration, pkgName)
	} else if astFuncDeclaration, ok := astDeclaration.(*ast.FuncDecl); ok {
		p.parseTypeSpecInFuncDeclaration(astFuncDeclaration, pkgName)
	}
}

func (p *parser) parseTypeSpecFromGenDeclaration(astGenDeclaration *ast.GenDecl, pkgName string) {
	for _, astSpec := range astGenDeclaration.Specs {
		if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
			p.TypeSpecs[pkgName][typeSpec.Name.String()] = typeSpec
			p.parseTypeAlias(typeSpec, pkgName)
		}
	}
}

// parseTypeSpecInFuncDeclaration find type declaration in func, method
func (p *parser) parseTypeSpecInFuncDeclaration(astFuncDeclaration *ast.FuncDecl, pkgName string) {
	if astFuncDeclaration.Doc != nil && astFuncDeclaration.Doc.List != nil && astFuncDeclaration.Body != nil {
		funcName := astFuncDeclaration.Name.String()
		for _, astStmt := range astFuncDeclaration.Body.List {
			p.parseTypeSpecFromFunctionBlockStmt(astFuncDeclaration, pkgName, astStmt, funcName)
		}
	}
}

func (p *parser) parseTypeSpecFromFunctionBlockStmt(astFuncDeclaration *ast.FuncDecl, pkgName string, astStmt ast.Stmt, funcName string) {
	if astDeclStmt, ok := astStmt.(*ast.DeclStmt); ok {
		if astGenDeclaration, ok := astDeclStmt.Decl.(*ast.GenDecl); ok {
			p.parseTypeSpecFromFunctionGenDeclaration(astFuncDeclaration, pkgName, astGenDeclaration, funcName)
		}
	}
}

func (p *parser) parseTypeSpecFromFunctionGenDeclaration(astFuncDeclaration *ast.FuncDecl, pkgName string, astGenDeclaration *ast.GenDecl, funcName string) {
	for _, astSpec := range astGenDeclaration.Specs {
		typeSpec, ok := astSpec.(*ast.TypeSpec)
		if !ok {
			continue
		}
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
