package apis

import (
	"fmt"
	"go/ast"
	"strings"

	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
)

func (p *parser) parseParameters() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name
		// p.debug(pkgName, "->", pkgPath)

		astPkgs, err := p.schemaParser.GetPkgAst(pkgPath)
		if err != nil {
			if p.RunInStrictMode {
				return fmt.Errorf("parsePaths: parse of %s package cause error: %s", pkgPath, err)
			}

			p.Debugf("parsePaths: parse of %s package cause error: %s", pkgPath, err)
			continue
		}

		for _, astPackage := range astPkgs {
			if err := p.parseParametersFromPackage(astPackage, pkgPath, pkgName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *parser) parseParametersFromPackage(astPackage *ast.Package, pkgPath string, pkgName string) error {
	for _, astFile := range astPackage.Files {
		if err := p.parseParametersFromFile(astFile, pkgPath, pkgName); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseParametersFromFile(astFile *ast.File, pkgPath string, pkgName string) error {
	for _, astDeclaration := range astFile.Decls {
		if err := p.parseFuncDeclaration(astDeclaration, pkgPath, pkgName); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseFuncDeclaration(astDeclaration ast.Decl, pkgPath string, pkgName string) error {
	astFuncDeclaration, ok := astDeclaration.(*ast.GenDecl)
	if ok && astFuncDeclaration.Doc != nil && astFuncDeclaration.Doc.List != nil {
		if err := p.parseParameter(pkgPath, pkgName, astFuncDeclaration.Doc.List); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseParameter(pkgPath string, pkgName string, astComments []*ast.Comment) error {
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
		case "@enum":
			err = p.parseEnums(pkgPath, pkgName, strings.TrimSpace(comment[len(attribute):]))
		}
	}
	return err
}

func (p *parser) parseEnums(pkgPath string, pkgName string, comment string) error {
	schema, err := p.schemaParser.ParseSchemaObject(pkgPath, pkgName, comment)
	if err != nil {
		return fmt.Errorf("parseEnums can not parse enum schema %s", comment)
	}
	if schema.Properties == nil {
		return fmt.Errorf("parseHeaderComment can not parse Header comment schema %s", comment)
	}
	for _, key := range schema.Properties.Keys() {
		value, _ := schema.Properties.Get(key)
		currentSchemaObj, ok := value.(*oas.SchemaObject)
		if !ok {
			return fmt.Errorf("parseEnums can not parse enum params %s", comment)
		}

		p.OpenAPI.Components.Schemas[key] = currentSchemaObj
	}
	return nil
}

func (p *parser) parseHeaderParameters(pkgPath string, pkgName string, comment string) error {
	schema, err := p.schemaParser.ParseSchemaObject(pkgPath, pkgName, comment)
	if err != nil {
		return err
	}
	if schema.Properties == nil {
		return fmt.Errorf("NilSchemaProperties: parseHeaderComment can not parse Header comment schema, comment : %s", comment)
	}
	for _, key := range schema.Properties.Keys() {
		value, _ := schema.Properties.Get(key)
		currentSchemaObj, ok := value.(*oas.SchemaObject)
		if !ok {
			return fmt.Errorf("FailSchemaCasting: parseHeaderComment header param object to schema object casting failed, comment : %s", comment)
		}

		paramObj := &oas.ParameterObject{
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
