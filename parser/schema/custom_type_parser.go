package schema

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/orderedmap"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/utils"
	log "github.com/sirupsen/logrus"
)

func (p *parser) parseCustomTypeSchemaObject(pkgPath string, pkgName string, typeName string) (*SchemaObject, error) {
	if base, overrides, ok := splitCompositionTypeName(typeName); ok {
		return p.parseCompositionSchemaObject(pkgPath, pkgName, typeName, base, overrides)
	}
	if base, args, ok := splitGenericTypeName(typeName); ok {
		return p.parseGenericSchemaObject(pkgPath, pkgName, typeName, base, args)
	}

	var typeSpec *ast.TypeSpec
	var exist bool
	var schemaObject SchemaObject

	// handler other type
	typeNameParts := strings.Split(typeName, ".")
	if len(typeNameParts) == 1 {
		typeSpec, exist = p.getTypeSpec(pkgName, typeName)
		if !exist {
			log.Fatalf("Can not find definition of %s ast.TypeSpec. Current package %s", typeName, pkgName)
		}
		schemaObject.PkgName = pkgName
		schemaObject.ID = utils.GenSchemaObjectID(pkgName, typeName, p.SchemaWithoutPkg)
		p.KnownIDSchema[schemaObject.ID] = &schemaObject
	} else {
		guessPkgName := strings.Join(typeNameParts[:len(typeNameParts)-1], "/")
		guessPkgPath := ""
		if dir, ok := p.resolvePkgDir(guessPkgName); ok {
			guessPkgPath = dir
		}
		guessTypeName := typeNameParts[len(typeNameParts)-1]
		typeSpec, exist = p.getTypeSpec(guessPkgName, guessTypeName)
		if !exist {
			found := false
			aliases := p.PkgNameImportedPkgAlias[pkgName][guessPkgName]
			for k := range p.PkgNameImportedPkgAlias[pkgName] {
				if k == guessPkgName && len(aliases) != 0 {
					found = true
					break
				}
			}
			if !found {
				p.Debugf("unknown guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
				return &schemaObject, nil
			}
			for index, currentAliasName := range aliases {
				guessPkgName = currentAliasName
				guessPkgPath = ""
				if dir, ok := p.resolvePkgDir(guessPkgName); ok {
					guessPkgPath = dir
				}
				// p.debugf("guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
				typeSpec, exist = p.getTypeSpec(guessPkgName, guessTypeName)
				if exist {
					break
				}
				if !exist && index == len(aliases)-1 {
					p.Debugf("can not find definition of guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
					return &schemaObject, nil
				}
			}

			schemaObject.PkgName = guessPkgName
			schemaObject.ID = utils.GenSchemaObjectID(guessPkgName, guessTypeName, p.SchemaWithoutPkg)
			p.KnownIDSchema[schemaObject.ID] = &schemaObject
		}
		pkgPath, pkgName = guessPkgPath, guessPkgName
	}

	if astIdent, ok := typeSpec.Type.(*ast.Ident); ok {
		if astIdent != nil {
			if utils.IsGoTypeOASType(astIdent.Name) {
				schemaObject.Type = utils.GoTypesOASTypes[astIdent.Name]
			} else {
				schemaObject.Type = astIdent.Name
			}

			if schemaObject.ID != "" {
				componentsKey := schemaObject.ID
				if _, exists := p.OpenAPI.Components.Schemas[utils.ReplaceBackslash(componentsKey)]; !exists {
					p.OpenAPI.Components.Schemas[utils.ReplaceBackslash(componentsKey)] = &schemaObject
				}
			}
		}
	} else if astStructType, ok := typeSpec.Type.(*ast.StructType); ok {
		schemaObject.Type = "object"
		if schemaObject.Description == "" {
			if desc := p.typeDescription(pkgName, typeSpec.Name.String()); desc != "" {
				schemaObject.Description = desc
			}
		}
		if astStructType.Fields != nil {
			p.parseSchemaPropertiesFromStructFields(pkgPath, pkgName, &schemaObject, astStructType.Fields.List, nil)
		}
		typeNameParts := strings.Split(typeName, ".")
		if len(typeNameParts) > 1 {
			typeName = typeNameParts[len(typeNameParts)-1]
		}
		if !utils.IsBasicGoType(typeName) {
			_, err := p.RegisterType(pkgPath, pkgName, typeName)
			if err != nil {
				p.Debugf("ParseSchemaObject parse array items err: %s", err.Error())
			}
		}
	} else if astArrayType, ok := typeSpec.Type.(*ast.ArrayType); ok {
		schemaObject.Type = "array"
		schemaObject.Items = &SchemaObject{}
		typeAsString := p.getTypeAsString(astArrayType.Elt)
		typeAsString = strings.TrimLeft(typeAsString, "*")
		if !utils.IsBasicGoType(typeAsString) {
			schemaItemsSchemeaObjectID, err := p.RegisterType(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.Debugf("ParseSchemaObject parse array items err: %s", err.Error())
			} else {
				schemaObject.Items.Ref = utils.AddSchemaRefLinkPrefix(schemaItemsSchemeaObjectID)
			}
		} else if utils.IsGoTypeOASType(typeAsString) {
			schemaObject.Items.Type = utils.GoTypesOASTypes[typeAsString]
		}
	} else if astMapType, ok := typeSpec.Type.(*ast.MapType); ok {
		schemaObject.Type = "object"
		propertySchema := &SchemaObject{}
		typeAsString := p.getTypeAsString(astMapType.Value)
		typeAsString = strings.TrimLeft(typeAsString, "*")
		if !utils.IsBasicGoType(typeAsString) {
			schemaItemsSchemeaObjectID, err := p.RegisterType(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.Debugf("ParseSchemaObject parse array items err: %s", err.Error())
			} else {
				propertySchema.Ref = utils.AddSchemaRefLinkPrefix(schemaItemsSchemeaObjectID)
			}
		} else if utils.IsGoTypeOASType(typeAsString) {
			propertySchema.Type = utils.GoTypesOASTypes[typeAsString]
		}
		schemaObject.AdditionalProperties = propertySchema
	}
	return &schemaObject, nil
}

// resolveTypeNameParts resolves pkg.Type or Type into package path/name and local type name.
func (p *parser) resolveTypeNameParts(pkgPath, pkgName, typeName string) (string, string, string) {
	typeNameParts := strings.Split(typeName, ".")
	if len(typeNameParts) == 1 {
		return pkgPath, pkgName, typeName
	}
	guessPkgName := strings.Join(typeNameParts[:len(typeNameParts)-1], "/")
	guessTypeName := typeNameParts[len(typeNameParts)-1]
	guessPkgPath := ""
	if dir, ok := p.resolvePkgDir(guessPkgName); ok {
		guessPkgPath = dir
	}
	if _, exist := p.getTypeSpec(guessPkgName, guessTypeName); exist {
		return guessPkgPath, guessPkgName, guessTypeName
	}
	aliases := p.PkgNameImportedPkgAlias[pkgName][guessPkgName]
	for _, currentAliasName := range aliases {
		guessPkgPath = ""
		if dir, ok := p.resolvePkgDir(currentAliasName); ok {
			guessPkgPath = dir
		}
		if _, exist := p.getTypeSpec(currentAliasName, guessTypeName); exist {
			return guessPkgPath, currentAliasName, guessTypeName
		}
	}
	return guessPkgPath, guessPkgName, guessTypeName
}

func (p *parser) getTypeSpec(pkgName, typeName string) (*ast.TypeSpec, bool) {
	if _, indexed := p.TypeSpecs[pkgName]; !indexed {
		// Best-effort: a dependency package's type specs are built on first reference.
		// No-op for the project's own packages (already indexed) and for unresolvable
		// names (standard library, short aliases).
		p.resolvePkgDir(pkgName)
	}
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

// resolvePkgDir returns the directory of a package by its import path, indexing the
// package (building its TypeSpecs and import aliases) on first use. Project packages
// are already in KnownNamePkg; dependency packages are located in the module cache
// via the DepModules index built by the gomod parser and indexed lazily here, so only
// the dependency packages an annotation actually references are ever parsed.
func (p *parser) resolvePkgDir(importPath string) (string, bool) {
	if kp, ok := p.KnownNamePkg[importPath]; ok {
		p.ensurePkgIndexed(importPath, kp.Path)
		return kp.Path, true
	}
	dir, ok := p.locateDepPkgDir(importPath)
	if !ok {
		return "", false
	}
	pk := &model.Pkg{Name: importPath, Path: dir}
	p.KnownNamePkg[importPath] = pk
	p.KnownPathPkg[dir] = pk
	if p.RunInDebugMode {
		p.Debugf("on-demand %s -> %s", importPath, dir)
	}
	p.ensurePkgIndexed(importPath, dir)
	return dir, true
}

// locateDepPkgDir maps an import path to its module-cache directory using the longest
// matching module from the DepModules index, and confirms the directory exists.
func (p *parser) locateDepPkgDir(importPath string) (string, bool) {
	for _, m := range p.DepModules {
		if importPath != m.ImportPath && !strings.HasPrefix(importPath, m.ImportPath+"/") {
			continue
		}
		rel := strings.TrimPrefix(importPath, m.ImportPath)
		dir := filepath.Join(m.CacheDir, filepath.FromSlash(rel))
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, true
		}
		return "", false
	}
	return "", false
}

// ensurePkgIndexed builds TypeSpecs and import aliases for a single package from its
// AST, mirroring what the eager apis passes do for project packages. Idempotent.
func (p *parser) ensurePkgIndexed(pkgName, dir string) {
	if _, done := p.TypeSpecs[pkgName]; done {
		return
	}
	p.TypeSpecs[pkgName] = map[string]*ast.TypeSpec{}

	astPkgs, err := p.GetPkgAst(dir)
	if err != nil {
		return
	}
	if _, ok := p.PkgNameImportedPkgAlias[pkgName]; !ok {
		p.PkgNameImportedPkgAlias[pkgName] = map[string][]string{}
	}

	for _, astPackage := range astPkgs {
		for _, astFile := range astPackage.Files {
			for _, astDeclaration := range astFile.Decls {
				astGenDeclaration, ok := astDeclaration.(*ast.GenDecl)
				if !ok || astGenDeclaration.Tok != token.TYPE {
					continue
				}
				for _, astSpec := range astGenDeclaration.Specs {
					if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
						p.TypeSpecs[pkgName][typeSpec.Name.String()] = typeSpec
						genDesc := extractTypeDescription(astGenDeclaration.Doc)
						desc := extractTypeDescription(typeSpec.Doc)
						if desc == "" {
							desc = genDesc
						}
						if desc != "" {
							if _, ok := p.TypeDescriptions[pkgName]; !ok {
								p.TypeDescriptions[pkgName] = map[string]string{}
							}
							p.TypeDescriptions[pkgName][typeSpec.Name.String()] = desc
						}
					}
				}
			}
			for _, astImport := range astFile.Imports {
				p.indexImportAlias(pkgName, astImport)
			}
		}
	}
}

func (p *parser) indexImportAlias(pkgName string, astImport *ast.ImportSpec) {
	importedPkgName := strings.Trim(astImport.Path.Value, "\"")
	importedPkgAlias := ""
	if astImport.Name != nil && astImport.Name.Name != "." && astImport.Name.Name != "_" {
		importedPkgAlias = astImport.Name.String()
	} else {
		s := strings.Split(importedPkgName, "/")
		importedPkgAlias = s[len(s)-1]
	}
	for _, v := range p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias] {
		if v == importedPkgName {
			return
		}
	}
	p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias] = append(p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias], importedPkgName)
}

func (p *parser) parseSchemaPropertiesFromStructFields(pkgPath, pkgName string, structSchema *SchemaObject, astFields []*ast.Field, typeParams map[string]string) {
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

		if astField.Tag != nil {
			tag := reflect.StructTag(strings.Trim(astField.Tag.Value, "`"))
			if tag.Get("skip") == "true" {
				// If the field has a 'skip:"true"' tag, skip this iteration
				continue
			}
		}

		fieldSchema := &SchemaObject{}
		isPointer := false
		if _, ok := astField.Type.(*ast.StarExpr); ok {
			isPointer = true
		}
		typeAsString := p.getTypeAsString(astField.Type)
		if strings.HasPrefix(typeAsString, "*") {
			isPointer = true
			typeAsString = strings.TrimLeft(typeAsString, "*")
		}
		typeAsString = substituteTypeParams(typeAsString, typeParams)

		// Apply swaggertype / type overrides before resolving the Go type so custom
		// types (sql.NullInt64, []byte, …) can be documented without a TypeSpec.
		skipGoTypeResolve := false
		if astField.Tag != nil {
			astFieldTag := reflect.StructTag(strings.Trim(astField.Tag.Value, "`"))
			if astFieldTag.Get("swaggertype") != "" || strings.Contains(astFieldTag.Get("go-swagger3"), "type=") || astFieldTag.Get("type") != "" {
				p.addType(astFieldTag, fieldSchema)
				p.addSwaggerType(astFieldTag, fieldSchema)
				if fieldSchema.Type != "" {
					skipGoTypeResolve = true
				}
			}
		}

		if !skipGoTypeResolve {
			if strings.HasPrefix(typeAsString, "[]") {
				fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString)
				if err != nil {
					p.Debug(err)
					return
				}
			} else if strings.HasPrefix(typeAsString, "map[]") {
				fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString)
				if err != nil {
					p.Debug(err)
					return
				}
			} else if typeAsString == "time.Time" {
				fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString)
				if err != nil {
					p.Debug(err)
					return
				}
			} else if strings.HasPrefix(typeAsString, "interface{}") {
				fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString)
				if err != nil {
					p.Debug(err)
					return
				}
			} else if _, _, isGeneric := splitGenericTypeName(typeAsString); isGeneric {
				fieldSchemaSchemeaObjectID, err := p.RegisterType(pkgPath, pkgName, typeAsString)
				if err != nil {
					p.Debug("parseSchemaPropertiesFromStructFields err:", err)
				} else {
					fieldSchema.ID = fieldSchemaSchemeaObjectID
					schema, ok := p.KnownIDSchema[fieldSchemaSchemeaObjectID]
					if ok {
						fieldSchema.Type = schema.Type
						if schema.Items != nil {
							fieldSchema.Items = schema.Items
						}
					}
					fieldSchema.Ref = utils.AddSchemaRefLinkPrefix(fieldSchemaSchemeaObjectID)
				}
			} else if !utils.IsBasicGoType(typeAsString) {
				fieldSchemaSchemeaObjectID, err := p.RegisterType(pkgPath, pkgName, typeAsString)
				if err != nil {
					p.Debug("parseSchemaPropertiesFromStructFields err:", err)
				} else {
					fieldSchema.ID = fieldSchemaSchemeaObjectID
					schema, ok := p.KnownIDSchema[fieldSchemaSchemeaObjectID]
					if ok {
						fieldSchema.Type = schema.Type
						if schema.Items != nil {
							fieldSchema.Items = schema.Items
						}
					}
					fieldSchema.Ref = utils.AddSchemaRefLinkPrefix(fieldSchemaSchemeaObjectID)
				}
			} else if utils.IsGoTypeOASType(typeAsString) {
				fieldSchema.Type = utils.GoTypesOASTypes[typeAsString]
			}
		}

		name := astField.Names[0].Name
		fieldSchema.FieldName = name
		_, disabled := structSchema.DisabledFieldNames[name]
		if disabled {
			continue
		}

		nullableTag := ""
		if astField.Tag != nil {
			astFieldTag := reflect.StructTag(strings.Trim(astField.Tag.Value, "`"))
			tagText := ""

			if tag := astFieldTag.Get("go-swagger3"); tag != "" {
				for _, v := range strings.Split(tag, ",") {
					v = strings.TrimSpace(v)
					if v == "-" {
						structSchema.DisabledFieldNames[name] = struct{}{}
						fieldSchema.Deprecated = true
						continue astFieldsLoop
					}
				}
			}

			if skip := astFieldTag.Get("skip"); skip == "true" {
				continue astFieldsLoop
			}

			if tag := astFieldTag.Get("json"); tag != "" {
				tagText = tag
			}
			tagValues := strings.Split(tagText, ",")
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
			p.addType(astFieldTag, fieldSchema)
			p.addSwaggerType(astFieldTag, fieldSchema)
			p.addFormat(astFieldTag, fieldSchema)
			p.addExample(astFieldTag, fieldSchema)
			p.addOverrideExample(astFieldTag, fieldSchema)
			p.addDefault(astFieldTag, fieldSchema)
			p.addRequiredField(astFieldTag, isRequired, structSchema, name)
			p.addDescription(astFieldTag, fieldSchema)
			p.addReference(astFieldTag, fieldSchema)
			p.addEnum(astFieldTag, fieldSchema)
			p.addTitle(astFieldTag, fieldSchema)
			p.addMaxLimit(astFieldTag, fieldSchema)
			p.addIsExclusiveMaximum(astFieldTag, fieldSchema)
			p.addMinimumLimit(astFieldTag, fieldSchema)
			p.addIsExclusiveMinimum(astFieldTag, fieldSchema)
			p.addMaxLength(astFieldTag, fieldSchema)
			p.addMinLength(astFieldTag, fieldSchema)
			p.addPattern(astFieldTag, fieldSchema)
			p.addMaxItems(astFieldTag, fieldSchema)
			p.addMinItems(astFieldTag, fieldSchema)
			p.addUniqueItems(astFieldTag, fieldSchema)
			p.addMaxProperties(astFieldTag, fieldSchema)
			p.addMinProperties(astFieldTag, fieldSchema)
			p.addAdditionalProperties(astFieldTag, fieldSchema)
			p.addNullable(astFieldTag, fieldSchema)
			p.addReadOnly(astFieldTag, fieldSchema)
			p.addWriteOnly(astFieldTag, fieldSchema)
			p.addDeprecated(astFieldTag, fieldSchema)
			p.addComposition(pkgPath, pkgName, astFieldTag, fieldSchema)
			p.addSchemaExtensions(astFieldTag, fieldSchema)
			nullableTag = astFieldTag.Get("nullable")
		}
		if isPointer && nullableTag != "false" {
			fieldSchema.Nullable = true
		}
		p.addFieldCommentDescription(astField, fieldSchema)
		structSchema.Properties.Set(name, fieldSchema)
	}
	for _, astField := range astFields {
		if len(astField.Names) > 0 {
			continue
		}
		fieldSchema := &SchemaObject{}
		typeAsString := p.getTypeAsString(astField.Type)
		typeAsString = strings.TrimLeft(typeAsString, "*")
		typeAsString = substituteTypeParams(typeAsString, typeParams)
		if strings.HasPrefix(typeAsString, "[]") {
			fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.Debug(err)
				return
			}
		} else if strings.HasPrefix(typeAsString, "map[]") {
			fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.Debug(err)
				return
			}
		} else if typeAsString == "time.Time" {
			fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.Debug(err)
				return
			}
		} else if strings.HasPrefix(typeAsString, "interface{}") {
			fieldSchema, err = p.ParseSchemaObject(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.Debug(err)
				return
			}
		} else if !utils.IsBasicGoType(typeAsString) {
			fieldSchemaSchemeaObjectID, err := p.RegisterType(pkgPath, pkgName, typeAsString)
			if err != nil {
				p.Debug("parseSchemaPropertiesFromStructFields err:", err)
			} else {
				fieldSchema.ID = fieldSchemaSchemeaObjectID
				schema, ok := p.KnownIDSchema[fieldSchemaSchemeaObjectID]
				if ok {
					fieldSchema.Type = schema.Type
					if schema.Items != nil {
						fieldSchema.Items = schema.Items
					}
				}
				fieldSchema.Ref = utils.AddSchemaRefLinkPrefix(fieldSchemaSchemeaObjectID)
			}
		} else if utils.IsGoTypeOASType(typeAsString) {
			fieldSchema.Type = utils.GoTypesOASTypes[typeAsString]
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
					structSchema.Required = append(structSchema.Required, refSchema.Required...)
				}
			}
			continue
		}
	}
}

func (p *parser) addType(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if tag := astFieldTag.Get("type"); tag != "" {
		fieldSchema.Type = tag
		fieldSchema.Ref = ""
		fieldSchema.Items = nil
	}
}

// addSwaggerType applies swaggertype:"integer" / "primitive,integer" / "array,number"
// and go-swagger3:"type=string" overrides (swag-compatible).
func (p *parser) addSwaggerType(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	tag := astFieldTag.Get("swaggertype")
	if tag == "" {
		if gs := astFieldTag.Get("go-swagger3"); gs != "" {
			for _, part := range strings.Split(gs, ",") {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "type=") {
					tag = strings.TrimPrefix(part, "type=")
					break
				}
			}
		}
	}
	if tag == "" {
		return
	}
	var types []string
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" || part == "primitive" {
			continue
		}
		types = append(types, part)
	}
	if len(types) == 0 {
		return
	}
	fieldSchema.Ref = ""
	if types[0] == "array" {
		fieldSchema.Type = "array"
		if len(types) > 1 {
			fieldSchema.Items = &SchemaObject{Type: types[1]}
		} else {
			fieldSchema.Items = nil
		}
	} else {
		fieldSchema.Type = types[0]
		fieldSchema.Items = nil
	}
}

func (p *parser) addDefault(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	tag := astFieldTag.Get("default")
	if tag == "" {
		return
	}
	switch fieldSchema.Type {
	case "boolean":
		fieldSchema.Default, _ = strconv.ParseBool(tag)
	case "integer":
		fieldSchema.Default, _ = strconv.Atoi(tag)
	case "number":
		fieldSchema.Default, _ = strconv.ParseFloat(tag, 64)
	default:
		fieldSchema.Default = tag
	}
}

func (p *parser) addFormat(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if tag := astFieldTag.Get("format"); tag != "" {
		fieldSchema.Format = tag
	}
}

func (p *parser) addExample(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
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
}

func (p *parser) addWriteOnly(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if writeOnly := astFieldTag.Get("writeOnly"); writeOnly == "true" {
		fieldSchema.WriteOnly = true
	}
}

func (p *parser) addReadOnly(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if readOnly := astFieldTag.Get("readOnly"); readOnly == "true" {
		fieldSchema.ReadOnly = true
	}
}

func (p *parser) addNullable(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if nullable := astFieldTag.Get("nullable"); nullable == "true" {
		fieldSchema.Nullable = true
	}
}

func (p *parser) addAdditionalProperties(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if additionalProperties := astFieldTag.Get("additionalProperties"); additionalProperties == "true" {
		fieldSchema.AdditionalProperties = true
	}
}

func (p *parser) addMinProperties(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if minProperties := astFieldTag.Get("minProperties"); minProperties != "" {
		fieldSchema.MinProperties = parseUint(minProperties)
	}
}

func (p *parser) addMaxProperties(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if maxProperties := astFieldTag.Get("maxProperties"); maxProperties != "" {
		fieldSchema.MaxProperties = parseUint(maxProperties)
	}
}

func (p *parser) addUniqueItems(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if uniqueItems := astFieldTag.Get("uniqueItems"); uniqueItems == "true" {
		fieldSchema.UniqueItems = true
	}
}

func (p *parser) addMinItems(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if minItems := astFieldTag.Get("minItems"); minItems != "" {
		fieldSchema.MinItems = parseUint(minItems)
	}
}

func (p *parser) addMaxItems(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if maxItems := astFieldTag.Get("maxItems"); maxItems != "" {
		fieldSchema.MaxItems = parseUint(maxItems)
	}
}

func (p *parser) addPattern(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if pattern := astFieldTag.Get("pattern"); pattern != "" {
		fieldSchema.Pattern = pattern
	}
}

func (p *parser) addMinLength(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if minLength := astFieldTag.Get("minLength"); minLength != "" {
		fieldSchema.MinLength = parseUint(minLength)
	}
}

func (p *parser) addMaxLength(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if maxLength := astFieldTag.Get("maxLength"); maxLength != "" {
		fieldSchema.MaxLength = parseUint(maxLength)
	}
}

func (p *parser) addIsExclusiveMinimum(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if exclusiveMinimum := astFieldTag.Get("exclusiveMinimum"); exclusiveMinimum == "true" {
		fieldSchema.ExclusiveMinimum = true
	}
}

func (p *parser) addMinimumLimit(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if minimum := astFieldTag.Get("minimum"); minimum != "" {
		fieldSchema.Minimum = parseFloat64(minimum)
	}
}

func (p *parser) addIsExclusiveMaximum(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if exclusiveMaximum := astFieldTag.Get("exclusiveMaximum"); exclusiveMaximum == "true" {
		fieldSchema.ExclusiveMaximum = true
	}
}

func (p *parser) addMaxLimit(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if maximum := astFieldTag.Get("maximum"); maximum != "" {
		fieldSchema.Maximum = parseFloat64(maximum)
	}
}

func (p *parser) addTitle(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if title := astFieldTag.Get("title"); title != "" {
		fieldSchema.Title = title
	}
}

func (p *parser) addEnum(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if enumValues := astFieldTag.Get("enum"); enumValues != "" {
		fieldSchema.Enum = parseEnumValues(enumValues)
	}
}

func (p *parser) addReference(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if ref := astFieldTag.Get("$ref"); ref != "" {
		fieldSchema.Ref = utils.AddSchemaRefLinkPrefix(ref)
		fieldSchema.Type = "" // remove default type in case of reference link
	}
}

func (p *parser) addDescription(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if desc := astFieldTag.Get("description"); desc != "" {
		fieldSchema.Description = desc
	}
}

func (p *parser) addFieldCommentDescription(astField *ast.Field, fieldSchema *SchemaObject) {
	if fieldSchema.Description != "" {
		return
	}
	text := ""
	if astField.Doc != nil {
		text = astField.Doc.Text()
	} else if astField.Comment != nil {
		text = astField.Comment.Text()
	}
	if text == "" {
		return
	}
	desc := strings.TrimSpace(text)
	// Prefer explicit @description annotation in the comment.
	if extracted := extractTypeDescriptionFromText(desc); extracted != "" {
		fieldSchema.Description = extracted
		return
	}
	// Plain godoc comments (no leading @annotation) become the description.
	lines := strings.Split(desc, "\n")
	var plain []string
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimLeft(line, "/"))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "@") {
			continue
		}
		plain = append(plain, line)
	}
	if len(plain) > 0 {
		fieldSchema.Description = strings.Join(plain, " ")
	}
}

func (p *parser) addDeprecated(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if deprecated := astFieldTag.Get("deprecated"); deprecated == "true" {
		fieldSchema.Deprecated = true
	}
}

// addComposition applies oneOf / anyOf / allOf / discriminator struct tags.
// Example: `oneOf:"Cat,Dog" discriminator:"petType"`
func (p *parser) addComposition(pkgPath, pkgName string, astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if refs := p.resolveCompositionRefs(pkgPath, pkgName, astFieldTag.Get("oneOf")); len(refs) > 0 {
		fieldSchema.OneOf = refs
		clearCompositionConflicts(fieldSchema)
	}
	if refs := p.resolveCompositionRefs(pkgPath, pkgName, astFieldTag.Get("anyOf")); len(refs) > 0 {
		fieldSchema.AnyOf = refs
		clearCompositionConflicts(fieldSchema)
	}
	if refs := p.resolveCompositionRefs(pkgPath, pkgName, astFieldTag.Get("allOf")); len(refs) > 0 {
		fieldSchema.AllOf = refs
		clearCompositionConflicts(fieldSchema)
	}
	if prop := strings.TrimSpace(astFieldTag.Get("discriminator")); prop != "" {
		fieldSchema.Discriminator = &DiscriminatorObject{PropertyName: prop}
	}
}

func (p *parser) resolveCompositionRefs(pkgPath, pkgName, tag string) []*SchemaObject {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil
	}
	var refs []*SchemaObject
	for _, part := range strings.Split(tag, ",") {
		typeName := strings.TrimSpace(part)
		if typeName == "" {
			continue
		}
		id, err := p.RegisterType(pkgPath, pkgName, typeName)
		if err != nil {
			p.Debug("addComposition RegisterType err:", err)
			continue
		}
		refs = append(refs, &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(id)})
	}
	return refs
}

func clearCompositionConflicts(fieldSchema *SchemaObject) {
	fieldSchema.Type = ""
	fieldSchema.Ref = ""
	fieldSchema.Items = nil
	fieldSchema.Properties = nil
}

// addSchemaExtensions parses extensions:"x-foo=bar,x-baz=qux" into SchemaObject.Extensions.
func (p *parser) addSchemaExtensions(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	tag := strings.TrimSpace(astFieldTag.Get("extensions"))
	if tag == "" {
		return
	}
	if fieldSchema.Extensions == nil {
		fieldSchema.Extensions = map[string]interface{}{}
	}
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		if !strings.HasPrefix(key, "x-") {
			key = "x-" + key
		}
		fieldSchema.Extensions[key] = strings.TrimSpace(val)
	}
}

func (p *parser) typeDescription(pkgName, typeName string) string {
	if descs, ok := p.TypeDescriptions[pkgName]; ok {
		return descs[typeName]
	}
	return ""
}

func extractTypeDescription(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	return extractTypeDescriptionFromText(group.Text())
}

func extractTypeDescriptionFromText(text string) string {
	var parts []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(line, "/"))
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if strings.EqualFold(fields[0], "@description") {
			parts = append(parts, strings.TrimSpace(line[len(fields[0]):]))
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func (p *parser) addRequiredField(astFieldTag reflect.StructTag, isRequired bool, structSchema *SchemaObject, name string) {
	if _, ok := astFieldTag.Lookup("required"); ok || isRequired {
		structSchema.Required = append(structSchema.Required, name)
	}
}

func (p *parser) addOverrideExample(astFieldTag reflect.StructTag, fieldSchema *SchemaObject) {
	if tag := astFieldTag.Get("override-example"); tag != "" {
		fieldSchema.Example = tag

		if fieldSchema.Example != nil && len(fieldSchema.Ref) != 0 {
			fieldSchema.Ref = ""
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
		return fmt.Sprintf("*%v", p.getTypeAsString(astStarExpr.X))
	}

	astIndexExpr, ok := fieldType.(*ast.IndexExpr)
	if ok {
		return fmt.Sprintf("%s[%s]", p.getTypeAsString(astIndexExpr.X), p.getTypeAsString(astIndexExpr.Index))
	}

	astIndexListExpr, ok := fieldType.(*ast.IndexListExpr)
	if ok {
		args := make([]string, 0, len(astIndexListExpr.Indices))
		for _, idx := range astIndexListExpr.Indices {
			args = append(args, p.getTypeAsString(idx))
		}
		return fmt.Sprintf("%s[%s]", p.getTypeAsString(astIndexListExpr.X), strings.Join(args, ","))
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

func parseEnumValues(enumString string) interface{} {
	var result []interface{}
	for _, currentEnumValue := range strings.Split(enumString, EnumValueSeparator) {
		result = append(result, currentEnumValue)
	}
	return result
}

func parseUint(uintString string) uint {
	value, err := strconv.ParseUint(uintString, 10, 64)
	if err != nil {
		value = 0
	}

	return uint(value)
}

func parseFloat64(float64String string) float64 {
	value, err := strconv.ParseFloat(float64String, 64)
	if err != nil {
		value = 0
	}

	return value
}

type DECL struct {
	Type struct {
		Name string
	}
}

const EnumValueSeparator = ","
