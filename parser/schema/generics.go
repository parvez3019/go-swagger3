package schema

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"

	"github.com/iancoleman/orderedmap"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
)

// splitGenericTypeName parses Foo[Bar] or pkg.Foo[pkg.Bar, Baz] into base + type args.
func splitGenericTypeName(full string) (base string, args []string, ok bool) {
	full = stripSpaces(full)
	if full == "" || full[len(full)-1] != ']' {
		return "", nil, false
	}
	open := -1
	depth := 0
	for i, r := range full {
		switch r {
		case '[':
			if depth == 0 {
				open = i
			}
			depth++
		case ']':
			depth--
		}
	}
	if open <= 0 || depth != 0 {
		return "", nil, false
	}
	base = full[:open]
	inner := full[open+1 : len(full)-1]
	if inner == "" {
		return "", nil, false
	}
	args = splitTopLevel(inner, ',')
	if len(args) == 0 {
		return "", nil, false
	}
	return base, args, true
}

// splitCompositionTypeName parses Foo{data=Bar,meta=Baz}.
func splitCompositionTypeName(full string) (base string, overrides map[string]string, ok bool) {
	full = stripSpaces(full)
	open := strings.Index(full, "{")
	if open <= 0 || !strings.HasSuffix(full, "}") {
		return "", nil, false
	}
	// Generics use []; composition uses {}. Reject if this looks like nested generics only.
	if strings.Contains(full[:open], "[") {
		return "", nil, false
	}
	base = full[:open]
	inner := full[open+1 : len(full)-1]
	if inner == "" {
		return "", nil, false
	}
	overrides = map[string]string{}
	for _, part := range splitTopLevel(inner, ',') {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			return "", nil, false
		}
		overrides[kv[0]] = kv[1]
	}
	return base, overrides, true
}

func stripSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func splitTopLevel(s string, sep rune) []string {
	var parts []string
	depthBracket, depthBrace := 0, 0
	start := 0
	for i, r := range s {
		switch r {
		case '[':
			depthBracket++
		case ']':
			depthBracket--
		case '{':
			depthBrace++
		case '}':
			depthBrace--
		case sep:
			if depthBracket == 0 && depthBrace == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func genericSchemaName(base string, args []string) string {
	base = typeNameWithoutPkg(base)
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, base)
	for _, a := range args {
		parts = append(parts, typeNameWithoutPkg(a))
	}
	return strings.Join(parts, "_")
}

func compositionSchemaName(base string, overrides map[string]string) string {
	base = typeNameWithoutPkg(base)
	keys := make([]string, 0, len(overrides))
	for k := range overrides {
		keys = append(keys, k)
	}
	// Stable order by iterating sorted-ish via simple insertion for small maps.
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	parts := []string{base}
	for _, k := range keys {
		parts = append(parts, k+"_"+typeNameWithoutPkg(overrides[k]))
	}
	return strings.Join(parts, "_")
}

func typeNameWithoutPkg(name string) string {
	name = stripSpaces(name)
	// Strip generic/composition wrappers for ID segments.
	if i := strings.IndexAny(name, "[{"); i >= 0 {
		name = name[:i]
	}
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// schemaIDForTypeName returns the Components schema ID for a (possibly generic/composed) type name.
func (p *parser) schemaIDForTypeName(pkgName, typeName string) string {
	if base, args, ok := splitGenericTypeName(typeName); ok {
		return utils.GenSchemaObjectID(pkgName, genericSchemaName(base, args), p.SchemaWithoutPkg)
	}
	if base, overrides, ok := splitCompositionTypeName(typeName); ok {
		return utils.GenSchemaObjectID(pkgName, compositionSchemaName(base, overrides), p.SchemaWithoutPkg)
	}
	return utils.GenSchemaObjectID(pkgName, typeName, p.SchemaWithoutPkg)
}

func (p *parser) parseGenericSchemaObject(pkgPath, pkgName, typeName string, base string, args []string) (*SchemaObject, error) {
	idName := genericSchemaName(base, args)
	schemaID := utils.GenSchemaObjectID(pkgName, idName, p.SchemaWithoutPkg)
	if existing, ok := p.KnownIDSchema[schemaID]; ok {
		return existing, nil
	}

	basePkgPath, basePkgName, baseTypeName := p.resolveTypeNameParts(pkgPath, pkgName, base)
	typeSpec, exist := p.getTypeSpec(basePkgName, typeNameWithoutPkg(baseTypeName))
	if !exist {
		return &SchemaObject{}, nil
	}
	if typeSpec.TypeParams == nil || len(typeSpec.TypeParams.List) == 0 {
		return nil, fmt.Errorf("type %s is not generic", base)
	}

	formals := typeParamNames(typeSpec.TypeParams)
	if len(formals) != len(args) {
		return nil, fmt.Errorf("generic type %s expects %d type args, got %d", base, len(formals), len(args))
	}
	subs := map[string]string{}
	for i, formal := range formals {
		subs[formal] = args[i]
	}

	var schemaObject SchemaObject
	schemaObject.PkgName = basePkgName
	schemaObject.ID = schemaID
	p.KnownIDSchema[schemaID] = &schemaObject

	astStructType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return &schemaObject, nil
	}
	schemaObject.Type = "object"
	if desc := p.typeDescription(basePkgName, typeSpec.Name.String()); desc != "" {
		schemaObject.Description = desc
	}
	if astStructType.Fields != nil {
		p.parseSchemaPropertiesFromStructFields(basePkgPath, basePkgName, &schemaObject, astStructType.Fields.List, subs)
	}
	return &schemaObject, nil
}

func (p *parser) parseCompositionSchemaObject(pkgPath, pkgName, typeName string, base string, overrides map[string]string) (*SchemaObject, error) {
	idName := compositionSchemaName(base, overrides)
	schemaID := utils.GenSchemaObjectID(pkgName, idName, p.SchemaWithoutPkg)
	if existing, ok := p.KnownIDSchema[schemaID]; ok {
		return existing, nil
	}

	baseSchema, err := p.ParseSchemaObject(pkgPath, pkgName, base)
	if err != nil {
		return nil, err
	}
	if baseSchema == nil {
		return &SchemaObject{}, nil
	}

	cloned := cloneSchemaShallow(baseSchema)
	cloned.ID = schemaID
	cloned.Ref = ""
	if cloned.Properties == nil {
		return cloned, nil
	}

	for field, overrideType := range overrides {
		prop, exist := cloned.Properties.Get(field)
		if !exist {
			continue
		}
		fieldSchema, _ := prop.(*SchemaObject)
		if fieldSchema == nil {
			fieldSchema = &SchemaObject{}
		}
		newField := &SchemaObject{FieldName: fieldSchema.FieldName}
		if utils.IsBasicGoType(overrideType) || utils.IsGoTypeOASType(overrideType) {
			if utils.IsGoTypeOASType(overrideType) {
				newField.Type = utils.GoTypesOASTypes[overrideType]
			} else {
				newField.Type = overrideType
			}
		} else if strings.HasPrefix(overrideType, "[]") || strings.HasPrefix(overrideType, "map[]") {
			parsed, err := p.ParseSchemaObject(pkgPath, pkgName, overrideType)
			if err != nil {
				return nil, err
			}
			newField = parsed
		} else {
			refID, err := p.RegisterType(pkgPath, pkgName, overrideType)
			if err != nil {
				return nil, err
			}
			newField.Ref = utils.AddSchemaRefLinkPrefix(refID)
			newField.ID = refID
		}
		cloned.Properties.Set(field, newField)
	}

	p.KnownIDSchema[schemaID] = cloned
	return cloned, nil
}

func typeParamNames(params *ast.FieldList) []string {
	var names []string
	for _, field := range params.List {
		for _, ident := range field.Names {
			names = append(names, ident.Name)
		}
	}
	return names
}

func substituteTypeParams(typeName string, typeParams map[string]string) string {
	if typeParams == nil {
		return typeName
	}
	if sub, ok := typeParams[typeName]; ok {
		return sub
	}
	if strings.HasPrefix(typeName, "[]") {
		return "[]" + substituteTypeParams(typeName[2:], typeParams)
	}
	if strings.HasPrefix(typeName, "map[]") {
		return "map[]" + substituteTypeParams(typeName[5:], typeParams)
	}
	if base, args, ok := splitGenericTypeName(typeName); ok {
		newArgs := make([]string, len(args))
		for i, a := range args {
			newArgs[i] = substituteTypeParams(a, typeParams)
		}
		return base + "[" + strings.Join(newArgs, ",") + "]"
	}
	return typeName
}

func cloneSchemaShallow(src *SchemaObject) *SchemaObject {
	if src == nil {
		return nil
	}
	dst := *src
	if src.Properties != nil {
		dst.Properties = cloneOrderedSchemaProps(src.Properties)
	}
	if src.Required != nil {
		dst.Required = append([]string{}, src.Required...)
	}
	if src.DisabledFieldNames != nil {
		dst.DisabledFieldNames = map[string]struct{}{}
		for k, v := range src.DisabledFieldNames {
			dst.DisabledFieldNames[k] = v
		}
	}
	return &dst
}

func cloneOrderedSchemaProps(from *orderedmap.OrderedMap) *orderedmap.OrderedMap {
	if from == nil {
		return nil
	}
	to := orderedmap.New()
	for _, key := range from.Keys() {
		val, _ := from.Get(key)
		to.Set(key, val)
	}
	return to
}
