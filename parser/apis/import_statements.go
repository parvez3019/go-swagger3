package apis

import (
	"fmt"
	"go/ast"
	"strings"
)

func (p *parser) parseImportStatements() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		astPkgs, err := p.schemaParser.GetPkgAst(pkgPath)
		if err != nil {
			if p.RunInStrictMode {
				return fmt.Errorf("parseImportStatements: parse of %s package cause error: %s", pkgPath, err)
			}
			p.Debugf("parseImportStatements: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		p.PkgNameImportedPkgAlias[pkgName] = map[string][]string{}
		for _, astPackage := range astPkgs {
			p.parseImportStatementsFromPackage(astPackage, pkgName)
		}
	}
	return nil
}

func (p *parser) parseImportStatementsFromPackage(astPackage *ast.Package, pkgName string) {
	for _, astFile := range astPackage.Files {
		p.parseImportStatementsFromFile(astFile, pkgName)
	}
}

func (p *parser) parseImportStatementsFromFile(astFile *ast.File, pkgName string) {
	for _, astImport := range astFile.Imports {
		p.parseImportStatementFromImportSpec(astImport, pkgName)
	}
}

func (p *parser) parseImportStatementFromImportSpec(astImport *ast.ImportSpec, pkgName string) {
	importedPkgName := strings.Trim(astImport.Path.Value, "\"")
	importedPkgAlias := ""

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
