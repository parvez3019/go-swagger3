package openApi3Schema

import (
	"encoding/json"

	"github.com/iancoleman/orderedmap"
)

const (
	OpenAPIVersion = "3.0.0"

	ContentTypeText = "text/plain"
	ContentTypeJson = "application/json"
	ContentTypeForm = "multipart/form-data"
)

type OpenAPIObject struct {
	Version      string                       `json:"openapi"` // Required
	Info         InfoObject                   `json:"info"`    // Required
	Servers      []ServerObject               `json:"servers,omitempty"`
	Paths        PathsObject                  `json:"paths"` // Required
	Components   ComponentsObject             `json:"components,omitempty"` // Required for Authorization header
	Security     []map[string][]string        `json:"security,omitempty"`
	Tags         []TagObject                  `json:"tags,omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
}

type ServerObject struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`

	// Variables
}

type InfoObject struct {
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty"`
	Contact        *ContactObject `json:"contact,omitempty"`
	License        *LicenseObject `json:"license,omitempty"`
	Version        string         `json:"version"`
}

type ContactObject struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type LicenseObject struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type PathsObject map[string]*PathItemObject

type PathItemObject struct {
	Ref         string           `json:"$ref,omitempty"`
	Summary     string           `json:"summary,omitempty"`
	Description string           `json:"description,omitempty"`
	Get         *OperationObject `json:"get,omitempty"`
	Post        *OperationObject `json:"post,omitempty"`
	Patch       *OperationObject `json:"patch,omitempty"`
	Put         *OperationObject `json:"put,omitempty"`
	Delete      *OperationObject `json:"delete,omitempty"`
	Options     *OperationObject `json:"options,omitempty"`
	Head        *OperationObject `json:"head,omitempty"`
	Trace       *OperationObject `json:"trace,omitempty"`

	// Servers
	// Parameters
}

type OperationObject struct {
	Responses ResponsesObject `json:"responses"` // Required

	Tags         []string                     `json:"tags,omitempty"`
	Summary      string                       `json:"summary,omitempty"`
	Description  string                       `json:"description,omitempty"`
	OperationID  string                       `json:"operationId,omitempty"`
	Parameters   []ParameterObject            `json:"parameters,omitempty"`
	RequestBody  *RequestBodyObject           `json:"requestBody,omitempty"`
	Deprecated   bool                         `json:"deprecated,omitempty"`
	Security     []map[string][]string        `json:"security,omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`

	// Extensions holds OpenAPI x-* vendor extensions; flattened by MarshalJSON.
	Extensions map[string]interface{} `json:"-"`

	// Accept / Produce are internal (not serialized); used when building content maps.
	Accept  []string `json:"-"`
	Produce []string `json:"-"`

	// Callbacks
	// Servers
}

// MarshalJSON flattens Extensions as top-level x-* keys.
func (o OperationObject) MarshalJSON() ([]byte, error) {
	type Alias OperationObject
	b, err := json.Marshal(Alias(o))
	if err != nil {
		return nil, err
	}
	if len(o.Extensions) == 0 {
		return b, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for k, v := range o.Extensions {
		m[k] = v
	}
	return json.Marshal(m)
}

type ParameterObject struct {
	Name        string        `json:"name,omitempty"` // Required
	In          string        `json:"in,omitempty"`   // Required. Possible values are "query", "header", "path" or "cookie"
	Description string        `json:"description,omitempty"`
	Required    bool          `json:"required,omitempty"`
	Example     interface{}   `json:"example,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`
	Style       string        `json:"style,omitempty"`
	Explode     *bool         `json:"explode,omitempty"`

	// Ref is used when ParameterOjbect is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`

	// Deprecated
	// AllowEmptyValue
	// AllowReserved
	// Examples
	// Content
}

type RequestBodyObject struct {
	Content map[string]*MediaTypeObject `json:"content,omitempty"` // Required when not a $ref

	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`

	// Ref is used when RequestBodyObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`
}

type MediaTypeObject struct {
	Schema SchemaObject `json:"schema,omitempty"`
	// Example string       `json:"example,omitempty"`

	// Examples
	// Encoding
}

type SchemaObject struct {
	ID                         string                 `json:"-"` // For go-swagger3
	PkgName                    string                 `json:"-"` // For go-swagger3
	FieldName                  string                 `json:"-"` // For go-swagger3
	DisabledFieldNames         map[string]struct{}    `json:"-"` // For go-swagger3
	Type                       string                 `json:"type,omitempty"`
	Format                     string                 `json:"format,omitempty"`
	Required                   []string               `json:"required,omitempty"`
	Properties                 *orderedmap.OrderedMap `json:"properties,omitempty"`
	Description                string                 `json:"description,omitempty"`
	Items                      *SchemaObject          `json:"items,omitempty"` // use ptr to prevent recursive error
	Example                    interface{}            `json:"example,omitempty"`
	Deprecated                 bool                   `json:"deprecated,omitempty"`
	Ref                        string                 `json:"$ref,omitempty"` // Ref is used when SchemaObject is as a ReferenceObject
	Enum                       interface{}            `json:"enum,omitempty"`
	Title                      string                 `json:"title,omitempty"`
	Default                    interface{}            `json:"default,omitempty"`
	MultipleOf                 float64                `json:"multipleOf,omitempty"`
	Maximum                    float64                `json:"maximum,omitempty"`
	ExclusiveMaximum           bool                   `json:"exclusiveMaximum,omitempty"`
	Minimum                    float64                `json:"minimum,omitempty"`
	ExclusiveMinimum           bool                   `json:"exclusiveMinimum,omitempty"`
	MaxLength                  uint                   `json:"maxLength,omitempty"`
	MinLength                  uint                   `json:"minLength,omitempty"`
	Pattern                    string                 `json:"pattern,omitempty"`
	MaxItems                   uint                   `json:"maxItems,omitempty"`
	MinItems                   uint                   `json:"minItems,omitempty"`
	UniqueItems                bool                   `json:"uniqueItems,omitempty"`
	MaxProperties              uint                   `json:"maxProperties,omitempty"`
	MinProperties              uint                   `json:"minProperties,omitempty"`
	AdditionalProperties interface{} `json:"additionalProperties,omitempty"`
	Nullable             bool        `json:"nullable,omitempty"`
	ReadOnly                   bool                   `json:"readOnly,omitempty"`
	WriteOnly                  bool                   `json:"writeOnly,omitempty"`
	AllOf                      []*SchemaObject        `json:"allOf,omitempty"`
	OneOf                      []*SchemaObject        `json:"oneOf,omitempty"`
	AnyOf                      []*SchemaObject        `json:"anyOf,omitempty"`
	Discriminator              *DiscriminatorObject   `json:"discriminator,omitempty"`

	// Extensions holds OpenAPI x-* vendor extensions; flattened by MarshalJSON.
	Extensions map[string]interface{} `json:"-"`

	// Not
	// XML
	// ExternalDocs
}

// MarshalJSON flattens Extensions as top-level x-* keys.
func (s SchemaObject) MarshalJSON() ([]byte, error) {
	type Alias SchemaObject
	b, err := json.Marshal(Alias(s))
	if err != nil {
		return nil, err
	}
	if len(s.Extensions) == 0 {
		return b, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for k, v := range s.Extensions {
		m[k] = v
	}
	return json.Marshal(m)
}

type DiscriminatorObject struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

type ResponsesObject map[string]*ResponseObject // [status]ResponseObject

type ResponseObject struct {
	Description string `json:"description"` // Required

	Headers map[string]*HeaderObject    `json:"headers,omitempty"`
	Content map[string]*MediaTypeObject `json:"content,omitempty"`

	// Ref is for ReferenceObject
	Ref string `json:"$ref,omitempty"`

	// Links
}

type HeaderObject struct {
	Description string        `json:"description,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`

	// Ref is used when HeaderObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`
}

type ComponentsObject struct {
	Schemas         map[string]*SchemaObject         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecuritySchemeObject `json:"securitySchemes,omitempty"`
	Parameters      map[string]*ParameterObject      `json:"parameters,omitempty"`
	Responses       map[string]*ResponseObject       `json:"responses,omitempty"`
	RequestBodies   map[string]*RequestBodyObject    `json:"requestBodies,omitempty"`
	// Examples
	// Headers
	// Links
	// Callbacks
}

type TagObject struct {
	Name         string                       `json:"name"`
	Description  string                       `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
}

type ExternalDocumentationObject struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"` // Required
}

type SecuritySchemeObject struct {
	// Generic fields
	Type        string `json:"type"` // Required
	Description string `json:"description,omitempty"`

	// http
	Scheme string `json:"scheme,omitempty"`

	// apiKey
	In   string `json:"in,omitempty"`
	Name string `json:"name,omitempty"`

	// OpenID
	OpenIdConnectUrl string `json:"openIdConnectUrl,omitempty"`

	// OAuth2
	OAuthFlows *SecuritySchemeOauthObject `json:"flows,omitempty"`

	// BearerFormat
}

type SecuritySchemeOauthObject struct {
	Implicit              *SecuritySchemeOauthFlowObject `json:"implicit,omitempty"`
	AuthorizationCode     *SecuritySchemeOauthFlowObject `json:"authorizationCode,omitempty"`
	ResourceOwnerPassword *SecuritySchemeOauthFlowObject `json:"password,omitempty"`
	ClientCredentials     *SecuritySchemeOauthFlowObject `json:"clientCredentials,omitempty"`
}

func (s *SecuritySchemeOauthObject) ApplyScopes(scopes map[string]string) {
	if s.Implicit != nil {
		s.Implicit.Scopes = scopes
	}

	if s.AuthorizationCode != nil {
		s.AuthorizationCode.Scopes = scopes
	}

	if s.ResourceOwnerPassword != nil {
		s.ResourceOwnerPassword.Scopes = scopes
	}

	if s.ClientCredentials != nil {
		s.ClientCredentials.Scopes = scopes
	}
}

type SecuritySchemeOauthFlowObject struct {
	AuthorizationUrl string            `json:"authorizationUrl,omitempty"`
	TokenUrl         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}
