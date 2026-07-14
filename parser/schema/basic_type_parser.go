package schema

import (
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
	"strings"
)

func (p *parser) parseBasicTypeSchemaObject(pkgPath string, pkgName string, typeName string) (*SchemaObject, error, bool) {
	var schemaObject SchemaObject
	// handler basic and some specific typeName
	if strings.HasPrefix(typeName, "[]") {
		return p.parseArrayType(pkgPath, pkgName, typeName, schemaObject)
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

func (p *parser) parseTimeType(schemaObject SchemaObject) (*SchemaObject, error, bool) {
	schemaObject.Type = "string"
	schemaObject.Format = "date-time"
	return &schemaObject, nil, true
}

func (p *parser) parseArrayType(pkgPath string, pkgName string, typeName string, schemaObject SchemaObject) (*SchemaObject, error, bool) {
	schemaObject.Type = "array"
	itemTypeName := typeName[2:]

	if !utils.IsBasicGoType(itemTypeName) {
		itemID, err := p.RegisterType(pkgPath, pkgName, itemTypeName)
		if err != nil {
			return nil, err, true
		}
		if itemID != "" {
			schemaObject.Items = &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(itemID)}
			return &schemaObject, nil, true
		}
	}

	var err error
	schemaObject.Items, err = p.ParseSchemaObject(pkgPath, pkgName, itemTypeName)
	if err != nil {
		return nil, err, true
	}
	return &schemaObject, nil, true
}

func (p *parser) parseMapType(pkgPath string, pkgName string, typeName string, schemaObject SchemaObject) (*SchemaObject, error, bool) {
	schemaObject.Type = "object"
	itemTypeName := typeName[5:]

	if !utils.IsBasicGoType(itemTypeName) {
		itemID, err := p.RegisterType(pkgPath, pkgName, itemTypeName)
		if err != nil {
			return nil, err, true
		}
		if itemID != "" {
			schemaObject.AdditionalProperties = &SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(itemID)}
			return &schemaObject, nil, true
		}
	}

	schemaProperty, err := p.ParseSchemaObject(pkgPath, pkgName, itemTypeName)
	if err != nil {
		return nil, err, true
	}
	schemaObject.AdditionalProperties = schemaProperty
	return &schemaObject, nil, true
}
