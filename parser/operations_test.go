package parser

import (
	oas "github.com/mikunalpha/goas/openApi3Schema"
	"github.com/mikunalpha/goas/parser/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseHeaders(t *testing.T) {
	schemaParserMocks := &mocks.SchemaParser{}
	parser := parser{SchemaParser: schemaParserMocks}
	schemaParserMocks.
		On("ParseSchemaObject", "/test/path", "pkgName", "comment").
		Return(getSchemaObject(), nil)

	operationObject := &oas.OperationObject{}
	err := parser.parseHeaders("/test/path", "pkgName", operationObject, "comment")

	assert.Nil(t, err)
	expectedParameters := []oas.ParameterObject{
		{Ref: "#/components/parameters/ContentType"},
		{Ref: "#/components/parameters/Version"},
		{Ref: "#/components/parameters/Authorization"},
	}
	assert.Equal(t, expectedParameters, operationObject.Parameters)
}
