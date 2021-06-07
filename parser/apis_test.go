package parser

import (
	"errors"
	"github.com/iancoleman/orderedmap"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseHeaderParameters1(t *testing.T) {
	tests := []struct {
		name               string
		schemaParser       SchemaParser
		wantErr            bool
		errMsg             string
		expectedParameters map[string]*oas.ParameterObject
	}{
		{
			name:               "Should return header parameters",
			schemaParser:       setupUpSchemaParseMocks(getSchemaObject(), nil),
			wantErr:            false,
			expectedParameters: getExpectedHeaderParameters(),
		},
		{
			name:         "Should return error when failed parsing schema object",
			schemaParser: setupUpSchemaParseMocks(nil, errors.New("someErr")),
			wantErr:      true,
			errMsg:       "someErr",
		},
		{
			name:         "Should return error when schema properties are nil",
			schemaParser: setupUpSchemaParseMocks(&oas.SchemaObject{}, nil),
			wantErr:      true,
			errMsg:       "NilSchemaProperties: parseHeaderComment can not parse Header comment schema, comment : comment",
		},
		{
			name: "Should return error when fails casting schema value to schema object",
			schemaParser: setupUpSchemaParseMocks(&oas.SchemaObject{
				Properties: getInvalidSchemaProperties(),
			}, nil),
			wantErr: true,
			errMsg:  "FailSchemaCasting: parseHeaderComment header param object to schema object casting failed, comment : comment",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			apiParser := apiParser{parser: &parser{
				SchemaParser: test.schemaParser,
				OpenAPI:      oas.OpenAPIObject{Components: oas.ComponentsObject{Parameters: map[string]*oas.ParameterObject{}}},
			}}
			err := apiParser.parseHeaderParameters("/test/path", "pkgName", "comment")
			if test.wantErr {
				assert.NotNil(t, err)
				assert.EqualError(t, err, test.errMsg)
			}
			if !test.wantErr {
				assertHeaderParameters(t, apiParser.OpenAPI.Components.Parameters, test.expectedParameters)
			}

		})
	}
}

func assertHeaderParameters(t *testing.T, actualParam map[string]*oas.ParameterObject, expectedParams map[string]*oas.ParameterObject) {
	assert.Len(t, actualParam, len(expectedParams))
	for key, parameterObject := range expectedParams {
		assert.Equal(t, parameterObject, actualParam[key])
	}
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

func getExpectedHeaderParameters() map[string]*oas.ParameterObject {
	params := map[string]*oas.ParameterObject{}
	params["ContentType"] = &oas.ParameterObject{
		Name:        "ContentType",
		In:          "header",
		Required:    true,
		Description: "Content Type Description",
		Example:     "json",
		Schema:      contentTypeHeaderSchema,
	}
	params["Version"] = &oas.ParameterObject{
		Name:        "Version",
		In:          "header",
		Required:    true,
		Description: "Version Description",
		Example:     "101",
		Schema:      versionHeaderSchema,
	}
	params["Authorization"] = &oas.ParameterObject{
		Name:        "Authorization",
		In:          "header",
		Required:    false,
		Description: "Authorization Description",
		Example:     "Bearer 123",
		Schema:      authorizationHeaderSchema,
	}
	return params
}

func getInvalidSchemaProperties() *orderedmap.OrderedMap {
	properties := orderedmap.New()
	properties.Set("key", "value")
	return properties
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
