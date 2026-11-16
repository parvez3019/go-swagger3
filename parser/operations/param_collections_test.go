package operations

import (
	"errors"
	"testing"

	"github.com/parvez3019/go-swagger3/logger"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newParserWith returns a parser wired like the real one (enum registry and
// logger set) around the given schema parser mock.
func newParserWith(schema *MockSchemaParser) *parser {
	return &parser{
		Parser: schema,
		Utils: model.Utils{
			PkgAndSpecs: &model.PkgAndSpecs{RegisteredEnums: map[string]struct{}{}},
			Logger:      logger.SetDebugMode(false),
		},
	}
}

func newOp() *oas.OperationObject {
	return &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}
}

func TestNormalizeAnnotatedGoType(t *testing.T) {
	tests := map[string]string{
		"string":                   "string",
		"[]string":                 "[]string",
		"[5]int":                   "[]int",
		"[]model.Item":             "[]model.Item",
		"map[]model.Item":          "map[]model.Item",
		"map[string]model.Item":    "map[]model.Item",
		"map[string][]model.Item":  "map[][]model.Item",
		"map[string]map[int]bool":  "map[]map[]bool",
		"model.Page[model.Item]":   "model.Page[model.Item]",
		"[]model.Page[model.Item]": "[]model.Page[model.Item]",
		"map[string]Page[Item]":    "map[]Page[Item]",
		"JSONResult{data=Order}":   "JSONResult{data=Order}",
		"JSONResult{data=[]Order}": "JSONResult{data=[]Order}",
	}
	for in, want := range tests {
		assert.Equal(t, want, normalizeAnnotatedGoType(in), in)
	}
}

func TestArrayQueryParams(t *testing.T) {
	tests := []struct {
		name        string
		comment     string
		wantItems   oas.SchemaObject
		wantStyle   string
		wantExplode *bool
	}{
		{
			name:        "strings with multi",
			comment:     `@Param tags query []string false "Tags" collectionFormat(multi)`,
			wantItems:   oas.SchemaObject{Type: "string"},
			wantStyle:   "form",
			wantExplode: boolPtr(true),
		},
		{
			name:        "ints with pipes",
			comment:     `@Param ids query []int64 false "IDs" collectionFormat(pipes)`,
			wantItems:   oas.SchemaObject{Type: "integer", Format: "int64"},
			wantStyle:   "pipeDelimited",
			wantExplode: boolPtr(false),
		},
		{
			name:        "ssv",
			comment:     `@Param words query []string false "Words" collectionFormat(ssv)`,
			wantItems:   oas.SchemaObject{Type: "string"},
			wantStyle:   "spaceDelimited",
			wantExplode: boolPtr(false),
		},
		{
			name:        "csv",
			comment:     `@Param fields query []string false "Fields" collectionFormat(csv)`,
			wantItems:   oas.SchemaObject{Type: "string"},
			wantStyle:   "form",
			wantExplode: boolPtr(false),
		},
		{
			name:      "times without collection format",
			comment:   `@Param days query []time.Time false "Days"`,
			wantItems: oas.SchemaObject{Type: "string", Format: "date-time"},
		},
		{
			name:      "header array",
			comment:   `@Param X-Ids header []string false "IDs"`,
			wantItems: oas.SchemaObject{Type: "string"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &parser{}
			op := newOp()
			require.NoError(t, p.parseOperationFromComment("", "", tc.comment, op))
			require.Len(t, op.Parameters, 1, "array param must not be dropped")

			param := op.Parameters[0]
			require.NotNil(t, param.Schema)
			assert.Equal(t, "array", param.Schema.Type)
			require.NotNil(t, param.Schema.Items)
			assert.Equal(t, tc.wantItems.Type, param.Schema.Items.Type)
			assert.Equal(t, tc.wantItems.Format, param.Schema.Items.Format)
			assert.Equal(t, tc.wantStyle, param.Style)
			assert.Equal(t, tc.wantExplode, param.Explode)
		})
	}
}

func TestArrayQueryParamEnumsMoveToItems(t *testing.T) {
	p := &parser{}
	op := newOp()
	require.NoError(t, p.parseOperationFromComment("", "",
		`@Param status query []string false "Statuses" Enums(active,sold) collectionFormat(multi)`, op))

	require.Len(t, op.Parameters, 1)
	schema := op.Parameters[0].Schema
	assert.Nil(t, schema.Enum, "enum belongs on items for array params")
	assert.Equal(t, []interface{}{"active", "sold"}, schema.Items.Enum)
}

func TestArrayQueryParamOfModelAndEnumTypes(t *testing.T) {
	mockSchema := new(MockSchemaParser)
	mockSchema.On("RegisterType", "pkg", "pkg", "model.Filter").Return("Filter", nil)
	p := newParserWith(mockSchema)

	op := newOp()
	require.NoError(t, p.parseOperationFromComment("pkg", "pkg", `@Param filters query []model.Filter false "Filters"`, op))
	require.NoError(t, p.parseOperationFromComment("pkg", "pkg", `@Param countries query []model.CountriesEnum false "Countries"`, op))

	require.Len(t, op.Parameters, 2)
	assert.Equal(t, "#/components/schemas/Filter", op.Parameters[0].Schema.Items.Ref)
	assert.Equal(t, "#/components/schemas/CountriesEnum", op.Parameters[1].Schema.Items.Ref)
}

func TestArrayQueryParamUnresolvableTypeIsSkipped(t *testing.T) {
	mockSchema := new(MockSchemaParser)
	mockSchema.On("RegisterType", "pkg", "pkg", "uuid.UUID").Return("", errors.New("not found"))
	p := newParserWith(mockSchema)

	op := newOp()
	require.NoError(t, p.parseOperationFromComment("pkg", "pkg", `@Param ids query []uuid.UUID false "IDs"`, op),
		"unresolvable element types must not fail the whole generation")
	assert.Empty(t, op.Parameters)
}

func TestMapResponseWithExplicitKeyType(t *testing.T) {
	mockSchema := new(MockSchemaParser)
	mockSchema.On("ParseSchemaObject", "pkg", "pkg", "map[]model.Item").Return(&oas.SchemaObject{
		Type:                 "object",
		AdditionalProperties: &oas.SchemaObject{Ref: "#/components/schemas/Item"},
	}, nil)
	p := newParserWith(mockSchema)

	op := newOp()
	require.NoError(t, p.parseOperationFromComment("pkg", "pkg", `@Success 200 object map[string]model.Item "Items by ID"`, op))

	schema := op.Responses["200"].Content[oas.ContentTypeJson].Schema
	assert.Equal(t, "object", schema.Type)
	assert.Empty(t, schema.Ref, "map response must not collapse to the value type")
	mockSchema.AssertExpectations(t)
}

func boolPtr(b bool) *bool { return &b }

func TestMapRequestBodyWithExplicitKeyType(t *testing.T) {
	mockSchema := new(MockSchemaParser)
	mockSchema.On("ParseSchemaObject", "pkg", "pkg", "map[]model.Item").Return(&oas.SchemaObject{
		Type:                 "object",
		AdditionalProperties: &oas.SchemaObject{Ref: "#/components/schemas/Item"},
	}, nil)
	p := newParserWith(mockSchema)

	op := newOp()
	require.NoError(t, p.parseOperationFromComment("pkg", "pkg", `@Param items body map[string]model.Item true "Items by ID"`, op))

	schema := op.RequestBody.Content[oas.ContentTypeJson].Schema
	assert.Equal(t, "object", schema.Type)
	assert.Empty(t, schema.Ref)
	mockSchema.AssertExpectations(t)
}
