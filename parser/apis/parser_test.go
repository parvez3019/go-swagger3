package apis

import (
	"errors"
	"go/ast"
	"testing"

	"github.com/iancoleman/orderedmap"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/schema"
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

func Test_ParseResponseComponent(t *testing.T) {
	schemaObj := &oas.SchemaObject{
		ID:          "ErrorResponse",
		Type:        "object",
		Description: "An error",
	}
	apiParser := parser{
		schemaParser: &componentSchemaMock{schema: schemaObj},
		OpenAPI: &oas.OpenAPIObject{Components: oas.ComponentsObject{
			Schemas:   map[string]*oas.SchemaObject{"ErrorResponse": schemaObj},
			Responses: map[string]*oas.ResponseObject{},
		}},
	}

	err := apiParser.parseResponseComponent("/test", "pkg", "ErrorResponse")
	assert.NoError(t, err)
	resp, ok := apiParser.OpenAPI.Components.Responses["ErrorResponse"]
	assert.True(t, ok)
	assert.Equal(t, "An error", resp.Description)
	assert.Equal(t, "#/components/schemas/ErrorResponse", resp.Content[oas.ContentTypeJson].Schema.Ref)
}

func Test_ParseRequestBodyComponent(t *testing.T) {
	schemaObj := &oas.SchemaObject{
		ID:   "CreateUser",
		Type: "object",
	}
	apiParser := parser{
		schemaParser: &componentSchemaMock{schema: schemaObj},
		OpenAPI: &oas.OpenAPIObject{Components: oas.ComponentsObject{
			Schemas:       map[string]*oas.SchemaObject{"CreateUser": schemaObj},
			RequestBodies: map[string]*oas.RequestBodyObject{},
		}},
	}
	err := apiParser.parseRequestBodyComponent("/test", "pkg", "CreateUser")
	assert.NoError(t, err)
	body, ok := apiParser.OpenAPI.Components.RequestBodies["CreateUser"]
	assert.True(t, ok)
	assert.True(t, body.Required)
	assert.Equal(t, "#/components/schemas/CreateUser", body.Content[oas.ContentTypeJson].Schema.Ref)
}

type componentSchemaMock struct {
	schema *oas.SchemaObject
}

func (m *componentSchemaMock) GetPkgAst(pkgPath string) (map[string]*ast.Package, error) {
	return nil, nil
}

func (m *componentSchemaMock) RegisterType(pkgPath, pkgName, typeName string) (string, error) {
	return m.schema.ID, nil
}

func (m *componentSchemaMock) ParseSchemaObject(pkgPath, pkgName, typeName string) (*oas.SchemaObject, error) {
	return m.schema, nil
}
