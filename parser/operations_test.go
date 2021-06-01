package parser

import (
	"errors"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseHeader(t *testing.T) {
	tests := []struct {
		name               string
		schemaParser       SchemaParser
		wantErr            bool
		errMsg             string
		expectedParameters []oas.ParameterObject
	}{
		{
			name:         "Should add parameters with ref",
			schemaParser: setupUpSchemaParseMocks(getSchemaObject(), nil),
			wantErr:      false,
			expectedParameters: []oas.ParameterObject{
				{Ref: "#/components/parameters/ContentType"},
				{Ref: "#/components/parameters/Version"},
				{Ref: "#/components/parameters/Authorization"},
			},
		},
		{
			name:         "Should return error if fails parsing the schema",
			schemaParser: setupUpSchemaParseMocks(getSchemaObject(), errors.New("someErr")),
			wantErr:      true,
			errMsg:       "someErr",
		},
		{
			name:         "Should return error schema properties are nil",
			schemaParser: setupUpSchemaParseMocks(&oas.SchemaObject{}, nil),
			wantErr:      true,
			errMsg:       "NilSchemaProperties : parseHeaders can not parse Header schema comment",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := parser{SchemaParser: test.schemaParser}
			operationObject := &oas.OperationObject{}
			err := parser.parseHeaders("/test/path", "pkgName", operationObject, "comment")
			if test.wantErr {
				assert.NotNil(t, err)
				assert.EqualError(t, err, test.errMsg)
			}
			assert.Equal(t, test.expectedParameters, operationObject.Parameters)
		})
	}
}

func setupUpSchemaParseMocks(schemaObject *oas.SchemaObject, err error) SchemaParser {
	schemaParserMocks := &mocks.SchemaParser{}
	schemaParserMocks.
		On("ParseSchemaObject", "/test/path", "pkgName", "comment").
		Return(schemaObject, err)
	return schemaParserMocks
}
