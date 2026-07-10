package operations

import (
	"testing"

	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAcceptProduce(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}

	require.NoError(t, p.parseOperationFromComment("", "", "@Accept json xml", op))
	require.NoError(t, p.parseOperationFromComment("", "", "@Produce plain", op))

	assert.Equal(t, []string{"application/json", "text/xml"}, op.Accept)
	assert.Equal(t, []string{"text/plain"}, op.Produce)
}

func TestParseAcceptDoesNotAffectResponseContentType(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}

	require.NoError(t, p.parseOperationFromComment("", "", "@Accept multipart/form-data", op))
	require.NoError(t, p.parseOperationFromComment("", "", `@Success 201 {string} string "created"`, op))

	assert.Equal(t, []string{"multipart/form-data"}, op.Accept)
	media, ok := op.Responses["201"].Content[oas.ContentTypeJson]
	require.True(t, ok, "Accept must not change response content type")
	assert.Equal(t, "string", media.Schema.Type)
	_, hasForm := op.Responses["201"].Content[oas.ContentTypeForm]
	assert.False(t, hasForm)
}

func TestParseProduceAffectsResponseContentType(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}

	require.NoError(t, p.parseOperationFromComment("", "", "@Produce xml", op))
	require.NoError(t, p.parseOperationFromComment("", "", `@Success 200 {string} string "ok"`, op))

	_, ok := op.Responses["200"].Content["text/xml"]
	assert.True(t, ok)
	_, hasJSON := op.Responses["200"].Content[oas.ContentTypeJson]
	assert.False(t, hasJSON)
}

func TestParseAcceptAffectsRequestBodyContentType(t *testing.T) {
	mockSchema := new(MockSchemaParser)
	mockSchema.On("RegisterType", "pkg", "pkg", "User").Return("User", nil)

	p := &parser{Parser: mockSchema}
	op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}

	require.NoError(t, p.parseOperationFromComment("pkg", "pkg", "@Accept xml", op))
	require.NoError(t, p.parseOperationFromComment("pkg", "pkg", `@Param user body User true "user"`, op))

	_, ok := op.RequestBody.Content["text/xml"]
	assert.True(t, ok)
	_, hasJSON := op.RequestBody.Content[oas.ContentTypeJson]
	assert.False(t, hasJSON)
	mockSchema.AssertExpectations(t)
}

func TestParseDeprecatedAndSecurity(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}

	require.NoError(t, p.parseOperationFromComment("", "", "@deprecated", op))
	require.NoError(t, p.parseOperationFromComment("", "", "@Security ApiKeyAuth read write", op))

	assert.True(t, op.Deprecated)
	require.Len(t, op.Security, 1)
	assert.Equal(t, []string{"read", "write"}, op.Security[0]["ApiKeyAuth"])
}

func TestParseExternalDocs(t *testing.T) {
	t.Run("split annotations", func(t *testing.T) {
		p := &parser{}
		op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}
		require.NoError(t, p.parseOperationFromComment("", "", "@externalDocs.description More info", op))
		require.NoError(t, p.parseOperationFromComment("", "", "@externalDocs.url https://example.com/docs", op))
		require.NotNil(t, op.ExternalDocs)
		assert.Equal(t, "More info", op.ExternalDocs.Description)
		assert.Equal(t, "https://example.com/docs", op.ExternalDocs.URL)
	})

	t.Run("single annotation", func(t *testing.T) {
		p := &parser{}
		op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}
		require.NoError(t, p.parseOperationFromComment("", "", "@externalDocs https://example.com/docs More info", op))
		require.NotNil(t, op.ExternalDocs)
		assert.Equal(t, "https://example.com/docs", op.ExternalDocs.URL)
		assert.Equal(t, "More info", op.ExternalDocs.Description)
	})
}

func TestParseParamAttributes(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{}

	err := p.parseParamComment("pkg", "pkg", op,
		`status query string false "filter status" Enums(A, B, C) default(A) minimum(1) maximum(10) minlength(1) maxlength(5) Format(email) collectionFormat(multi) example(foo) explode(true)`)
	require.NoError(t, err)
	require.Len(t, op.Parameters, 1)

	param := op.Parameters[0]
	require.NotNil(t, param.Schema)
	assert.Equal(t, []interface{}{"A", "B", "C"}, param.Schema.Enum)
	assert.Equal(t, "A", param.Schema.Default)
	assert.Equal(t, float64(1), param.Schema.Minimum)
	assert.Equal(t, float64(10), param.Schema.Maximum)
	assert.Equal(t, uint(1), param.Schema.MinLength)
	assert.Equal(t, uint(5), param.Schema.MaxLength)
	assert.Equal(t, "email", param.Schema.Format)
	assert.Equal(t, "form", param.Style)
	require.NotNil(t, param.Explode)
	assert.True(t, *param.Explode)
	assert.Equal(t, "foo", param.Example)
}

func TestParseParamStyleAttribute(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{}

	err := p.parseParamComment("pkg", "pkg", op,
		`tags query string false "tags" style(form) explode(false)`)
	require.NoError(t, err)
	require.Len(t, op.Parameters, 1)
	assert.Equal(t, "form", op.Parameters[0].Style)
	require.NotNil(t, op.Parameters[0].Explode)
	assert.False(t, *op.Parameters[0].Explode)

	op2 := &oas.OperationObject{}
	err = p.parseParamComment("pkg", "pkg", op2,
		`id path string true "id" style(simple)`)
	require.NoError(t, err)
	assert.Equal(t, "simple", op2.Parameters[0].Style)
}

func TestParseOperationExtensions(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}

	require.NoError(t, p.parseOperationFromComment("", "", `@x-codeSample curl`, op))
	require.NoError(t, p.parseOperationFromComment("", "", `@extension x-internal {"team":"api"}`, op))
	require.NoError(t, p.parseOperationFromComment("", "", `@x-rate-limit {"limit":100}`, op))

	assert.Equal(t, "curl", op.Extensions["x-codeSample"])
	assert.Equal(t, map[string]interface{}{"team": "api"}, op.Extensions["x-internal"])
	assert.Equal(t, map[string]interface{}{"limit": float64(100)}, op.Extensions["x-rate-limit"])
}

func TestParseResponseRef(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}

	require.NoError(t, p.parseOperationFromComment("", "", `@Success 400 $ref:ErrorResponse "bad request"`, op))
	resp := op.Responses["400"]
	require.NotNil(t, resp)
	assert.Equal(t, "#/components/responses/ErrorResponse", resp.Ref)
	assert.Equal(t, "bad request", resp.Description)
	assert.Nil(t, resp.Content)
}

func TestParseRequestBodyRef(t *testing.T) {
	p := &parser{}
	op := &oas.OperationObject{Responses: map[string]*oas.ResponseObject{}}

	require.NoError(t, p.parseOperationFromComment("pkg", "pkg", `@Param body body $ref:CreateUser true "payload"`, op))
	require.NotNil(t, op.RequestBody)
	assert.Equal(t, "#/components/requestBodies/CreateUser", op.RequestBody.Ref)
	assert.True(t, op.RequestBody.Required)
	assert.Nil(t, op.RequestBody.Content)
}

func TestResolveMIMETypeAliases(t *testing.T) {
	assert.Equal(t, "application/json", resolveMIMEType("json"))
	assert.Equal(t, "multipart/form-data", resolveMIMEType("mpfd"))
	assert.Equal(t, "application/x-www-form-urlencoded", resolveMIMEType("x-www-form-urlencoded"))
	assert.Equal(t, "application/custom", resolveMIMEType("application/custom"))
}

func TestContentTypeHelpers(t *testing.T) {
	op := &oas.OperationObject{}
	assert.Equal(t, oas.ContentTypeJson, contentTypeForRequest(op))
	assert.Equal(t, oas.ContentTypeJson, contentTypeForResponse(op))

	op.Accept = []string{"text/xml"}
	op.Produce = []string{"text/plain"}
	assert.Equal(t, "text/xml", contentTypeForRequest(op))
	assert.Equal(t, "text/plain", contentTypeForResponse(op))
}
