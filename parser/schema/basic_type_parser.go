package schema

import (
	"github.com/iancoleman/orderedmap"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
	"strings"
)

func (p *parser) parseBasicTypeSchemaObject(pkgPath string, pkgName string, typeName string) (*SchemaObject, error, bool) {
	var schemaObject SchemaObject
	var err error

	// handler basic and some specific typeName
	if strings.HasPrefix(typeName, "[]") {
		schemaObject.Type = "array"
		itemTypeName := typeName[2:]
		schema, ok := p.KnownIDSchema[utils.GenSchemaObjectID(pkgName, itemTypeName, p.SchemaWithoutPkg)]
		if ok {
			schemaObject.Items = &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(schema.ID)}
			return &schemaObject, nil, true
		}
		schemaObject.Items, err = p.ParseSchemaObject(pkgPath, pkgName, itemTypeName)
		if err != nil {
			return nil, err, true
		}
		return &schemaObject, nil, true
	} else if strings.HasPrefix(typeName, "map[]") {
		schemaObject.Type = "object"
		itemTypeName := typeName[5:]
		schema, ok := p.KnownIDSchema[utils.GenSchemaObjectID(pkgName, itemTypeName, p.SchemaWithoutPkg)]
		if ok {
			schemaObject.Items = &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(schema.ID)}
			return &schemaObject, nil, true
		}
		schemaProperty, err := p.ParseSchemaObject(pkgPath, pkgName, itemTypeName)
		if err != nil {
			return nil, err, true
		}
		schemaObject.Properties = orderedmap.New()
		schemaObject.Properties.Set("key", schemaProperty)
		return &schemaObject, nil, true
	} else if typeName == "time.Time" {
		schemaObject.Type = "string"
		schemaObject.Format = "date-time"
		return &schemaObject, nil, true
	} else if strings.HasPrefix(typeName, "interface{}") {
		return &SchemaObject{Type: "object"}, nil, true
	} else if utils.IsGoTypeOASType(typeName) {
		schemaObject.Type = utils.GoTypesOASTypes[typeName]
		return &schemaObject, nil, true
	}
	return nil, nil, false
}
