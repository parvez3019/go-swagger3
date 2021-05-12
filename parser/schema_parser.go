package parser

import (
	. "github.com/parvez3019/goas/openApi3Schema"
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

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
		typeSpec, exist = p.getTypeSpec(pkgName, typeName)
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
		typeSpec, exist = p.getTypeSpec(guessPkgName, guessTypeName)
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
			typeSpec, exist = p.getTypeSpec(guessPkgName, guessTypeName)
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

func (p *parser) getTypeSpec(pkgName, typeName string) (*ast.TypeSpec, bool) {
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

			if tag := astFieldTag.Get("goas"); tag != "" {
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
