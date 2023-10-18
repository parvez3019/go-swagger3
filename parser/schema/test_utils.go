package schema

import (
	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/schema/mocks"
	"github.com/iancoleman/orderedmap"
)

func GetSchemaObject() *oas.SchemaObject {
	properties := orderedmap.New()
	properties.Set("ContentType", ContentTypeHeaderSchema)
	properties.Set("Version", VersionHeaderSchema)
	properties.Set("Authorization", AuthorizationHeaderSchema)
	return &oas.SchemaObject{
		Properties: properties,
		Required:   []string{"ContentType", "Version"},
	}
}

func SetupUpSchemaParseMocks(schemaObject *oas.SchemaObject, err error) Parser {
	schemaParserMocks := &mocks.SchemaParser{}
	schemaParserMocks.
		On("ParseSchemaObject", "/test/path", "pkgName", "comment").
		Return(schemaObject, err)
	return schemaParserMocks
}

var ContentTypeHeaderSchema = &oas.SchemaObject{
	ID:          "ContentType",
	FieldName:   "ContentTypeFieldName",
	Type:        "string",
	Description: "Content Type Description",
	Example:     "json",
}

var VersionHeaderSchema = &oas.SchemaObject{
	ID:          "Version",
	FieldName:   "VersionFieldName",
	Type:        "int",
	Description: "Version Description",
	Example:     "101",
}

var AuthorizationHeaderSchema = &oas.SchemaObject{
	ID:          "Authorization",
	FieldName:   "AuthorizationFieldName",
	Type:        "string",
	Description: "Authorization Description",
	Example:     "Bearer 123",
}
