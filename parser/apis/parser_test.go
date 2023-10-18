package apis

import (
	"errors"
	"testing"

	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/schema"
	"github.com/iancoleman/orderedmap"
	"github.com/stretchr/testify/assert"
)

func Test_ParseHeaderParameters(t *testing.T) {
	tests := []struct {
		name               string
		schemaParser       schema.Parser
		wantErr            bool
		errMsg             string
		expectedParameters map[string]*oas.ParameterObject
	}{
		{
			name:               "Should return header parameters",
			schemaParser:       schema.SetupUpSchemaParseMocks(schema.GetSchemaObject(), nil),
			wantErr:            false,
			expectedParameters: getExpectedHeaderParameters(),
		},
		{
			name:         "Should return error when failed parsing schema object",
			schemaParser: schema.SetupUpSchemaParseMocks(nil, errors.New("someErr")),
			wantErr:      true,
			errMsg:       "someErr",
		},
		{
			name:         "Should return error when schema properties are nil",
			schemaParser: schema.SetupUpSchemaParseMocks(&oas.SchemaObject{}, nil),
			wantErr:      true,
			errMsg:       "NilSchemaProperties: parseHeaderComment can not parse Header comment schema, comment : comment",
		},
		{
			name: "Should return error when fails casting schema value to schema object",
			schemaParser: schema.SetupUpSchemaParseMocks(&oas.SchemaObject{
				Properties: getInvalidSchemaProperties(),
			}, nil),
			wantErr: true,
			errMsg:  "FailSchemaCasting: parseHeaderComment header param object to schema object casting failed, comment : comment",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			apiParser := parser{
				schemaParser: test.schemaParser,
				OpenAPI:      &oas.OpenAPIObject{Components: oas.ComponentsObject{Parameters: map[string]*oas.ParameterObject{}}},
			}
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

func getExpectedHeaderParameters() map[string]*oas.ParameterObject {
	params := map[string]*oas.ParameterObject{}
	params["ContentType"] = &oas.ParameterObject{
		Name:        "ContentType",
		In:          "header",
		Required:    true,
		Description: "Content Type Description",
		Example:     "json",
		Schema:      schema.ContentTypeHeaderSchema,
	}
	params["Version"] = &oas.ParameterObject{
		Name:        "Version",
		In:          "header",
		Required:    true,
		Description: "Version Description",
		Example:     "101",
		Schema:      schema.VersionHeaderSchema,
	}
	params["Authorization"] = &oas.ParameterObject{
		Name:        "Authorization",
		In:          "header",
		Required:    false,
		Description: "Authorization Description",
		Example:     "Bearer 123",
		Schema:      schema.AuthorizationHeaderSchema,
	}
	return params
}

func getInvalidSchemaProperties() *orderedmap.OrderedMap {
	properties := orderedmap.New()
	properties.Set("key", "value")
	return properties
}
