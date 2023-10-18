package schema

import (
	"strings"

	. "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/utils"
	"github.com/iancoleman/orderedmap"
)

func (p *parser) parseBasicTypeSchemaObject(pkgPath string, pkgName string, typeName string) (*SchemaObject, error, bool) {
	var schemaObject SchemaObject
	var err error
	// handler basic and some specific typeName
	if strings.HasPrefix(typeName, "[]") {
		return p.parseArrayType(pkgPath, pkgName, typeName, schemaObject, err)
	} else if strings.HasPrefix(typeName, "map[]") {
		return p.parseMapType(pkgPath, pkgName, typeName, schemaObject)
	} else if typeName == "time.Time" {
		return p.parseTimeType(schemaObject)
	} else if strings.HasPrefix(typeName, "interface{}") {
		return p.parseInterfaceType()
	} else if utils.IsGoTypeOASType(typeName) {
		return p.parseBasicGoType(schemaObject, typeName)
	}
	return nil, nil, false
}

func (p *parser) parseBasicGoType(schemaObject SchemaObject, typeName string) (*SchemaObject, error, bool) {
	schemaObject.Type = utils.GoTypesOASTypes[typeName]
	return &schemaObject, nil, true
}

func (p *parser) parseInterfaceType() (*SchemaObject, error, bool) {
	return &SchemaObject{Type: "object"}, nil, true
}

func (p *parser) parseAnonymousStructType() (*SchemaObject, error, bool) {
	return &SchemaObject{Type: "object"}, nil, false
}

func (p *parser) parseTimeType(schemaObject SchemaObject) (*SchemaObject, error, bool) {
	schemaObject.Type = "string"
	schemaObject.Format = "date-time"
	return &schemaObject, nil, true
}

func (p *parser) parseArrayType(pkgPath string, pkgName string, typeName string, schemaObject SchemaObject, err error) (*SchemaObject, error, bool) {
	schemaObject.Type = "array"
	itemTypeName := typeName[2:]
	schema, ok := p.KnownIDSchema[utils.GenSchemaObjectID(pkgName, itemTypeName)]
	if ok {
		schemaObject.Items = &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(schema.ID)}
		return &schemaObject, nil, true
	}
	schemaObject.Items, err = p.ParseSchemaObject(pkgPath, pkgName, itemTypeName)
	if err != nil {
		return nil, err, true
	}
	return &schemaObject, nil, true
}

func (p *parser) parseMapType(pkgPath string, pkgName string, typeName string, schemaObject SchemaObject) (*SchemaObject, error, bool) {
	schemaObject.Type = "object"
	itemTypeName := typeName[5:]
	schema, ok := p.KnownIDSchema[utils.GenSchemaObjectID(pkgName, itemTypeName)]
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
}
