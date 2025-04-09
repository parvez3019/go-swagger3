package schema

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/orderedmap"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
	log "github.com/sirupsen/logrus"
)

func (p *parser) parseCustomTypeSchemaObject(pkgPath string, pkgName string, typeName string) (*SchemaObject, error) {
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
				for i := range p.KnownPkgs {
					if guessPkgName == p.KnownPkgs[i].Name {
						guessPkgPath = p.KnownPkgs[i].Path
						break
					}
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
			schemaObject.Type = astIdent.Name
		}
	} else if astStructType, ok := typeSpec.Type.(*ast.StructType); ok {
		schemaObject.Type = "object"
		if astStructType.Fields != nil {
			p.parseSchemaPropertiesFromStructFields(pkgPath, pkgName, &schemaObject, astStructType.Fields.List)
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
		schemaObject.Properties = orderedmap.New()
		propertySchema := &SchemaObject{}
		schemaObject.Properties.Set("key", propertySchema)
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

		if astField.Tag != nil {
			tag := reflect.StructTag(strings.Trim(astField.Tag.Value, "`"))
			if tag.Get("skip") == "true" {
				// If the field has a 'skip:"true"' tag, skip this iteration
				continue
			}
		}

		fieldSchema := &SchemaObject{}
		typeAsString := p.getTypeAsString(astField.Type)
		typeAsString = strings.TrimLeft(typeAsString, "*")
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

			if skip := astFieldTag.Get("skip"); skip == "true" {
				continue astFieldsLoop
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

			if p.RequiredByDefault {
				p.Debug(fmt.Sprintf("Setting field %s required (required-by-default)", name))
				isRequired = true
			}

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
			p.addFormat(astFieldTag, fieldSchema)
			p.addExample(astFieldTag, fieldSchema)
			p.addOverrideExample(astFieldTag, fieldSchema)
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
