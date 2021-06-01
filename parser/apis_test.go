package parser

import (
	"github.com/iancoleman/orderedmap"
	oas "github.com/mikunalpha/goas/openApi3Schema"
	"github.com/mikunalpha/goas/parser/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseHeaderParameters(t *testing.T) {
	schemaParserMocks := &mocks.SchemaParser{}
	parser := parser{
		SchemaParser: schemaParserMocks,
		OpenAPI:      oas.OpenAPIObject{Components: oas.ComponentsObject{Parameters: map[string]*oas.ParameterObject{}}},
	}
	schemaParserMocks.
		On("ParseSchemaObject", "/test/path", "pkgName", "comment").
		Return(getSchemaObject(), nil)

	err := parser.parseHeaderParameters("/test/path", "pkgName", "comment")

	assert.Nil(t, err)
	assertHeaderParameters(t, parser.OpenAPI.Components.Parameters)
}

func assertHeaderParameters(t *testing.T, parameters map[string]*oas.ParameterObject) {
	assert.Len(t, parameters, 3)
	assert.Equal(t, &oas.ParameterObject{
		Name:        "ContentType",
		In:          "header",
		Required:    true,
		Description: "Content Type Description",
		Example:     "json",
		Schema:      contentTypeHeaderSchema,
	}, parameters["ContentType"])

	assert.Equal(t, &oas.ParameterObject{
		Name:        "Version",
		In:          "header",
		Required:    true,
		Description: "Version Description",
		Example:     "101",
		Schema:      versionHeaderSchema,
	}, parameters["Version"])

	assert.Equal(t, &oas.ParameterObject{
		Name:        "Authorization",
		In:          "header",
		Required:    false,
		Description: "Authorization Description",
		Example:     "Bearer 123",
		Schema:      authorizationHeaderSchema,
	}, parameters["Authorization"])

}

func getSchemaObject() *oas.SchemaObject {
	properties := orderedmap.New()
	properties.Set("ContentType", contentTypeHeaderSchema)
	properties.Set("Version", versionHeaderSchema)
	properties.Set("Authorization", authorizationHeaderSchema)

	return &oas.SchemaObject{
		Properties: properties,
		Required:   []string{"ContentType", "Version"},
	}
}

var contentTypeHeaderSchema = &oas.SchemaObject{
	ID:          "ContentType",
	FieldName:   "ContentTypeFieldName",
	Type:        "string",
	Description: "Content Type Description",
	Example:     "json",
}

var versionHeaderSchema = &oas.SchemaObject{
	ID:          "Version",
	FieldName:   "VersionFieldName",
	Type:        "int",
	Description: "Version Description",
	Example:     "101",
}

var authorizationHeaderSchema = &oas.SchemaObject{
	ID:          "Authorization",
	FieldName:   "AuthorizationFieldName",
	Type:        "string",
	Description: "Authorization Description",
	Example:     "Bearer 123",
}
