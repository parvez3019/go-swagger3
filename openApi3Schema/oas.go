package openApi3Schema

import "github.com/iancoleman/orderedmap"

const (
	OpenAPIVersion = "3.0.0"

	ContentTypeText = "text/plain"
	ContentTypeJson = "application/json"
	ContentTypeForm = "multipart/form-data"
)

type OpenAPIObject struct {
	OpenAPI string         `json:"openapi"` // Required
	Info    InfoObject     `json:"info"`    // Required
	Servers []ServerObject `json:"servers,omitempty"`
	Paths   PathsObject    `json:"paths"` // Required

	Components ComponentsObject      `json:"components,omitempty"` // Required for Authorization header
	Security   []map[string][]string `json:"security,omitempty"`

	// Tags
	// ExternalDocs
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

	Tags        []string           `json:"tags,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	Description string             `json:"description,omitempty"`
	Parameters  []ParameterObject  `json:"parameters,omitempty"`
	RequestBody *RequestBodyObject `json:"requestBody,omitempty"`

	// Tags
	// ExternalDocs
	// OperationID
	// Callbacks
	// Deprecated
	// Security
	// Servers
}

type ParameterObject struct {
	Name        string        `json:"name,omitempty"` // Required
	In          string        `json:"in,omitempty"`   // Required. Possible values are "query", "header", "path" or "cookie"
	Description string        `json:"description,omitempty"`
	Required    bool          `json:"required,omitempty"`
	Example     interface{}   `json:"example,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`

	// Ref is used when ParameterOjbect is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`

	// Deprecated
	// AllowEmptyValue
	// Style
	// Explode
	// AllowReserved
	// Examples
	// Content
}

type RequestBodyObject struct {
	Content map[string]*MediaTypeObject `json:"content"` // Required

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
	ID                 string                 `json:"-"` // For goas
	PkgName            string                 `json:"-"` // For goas
	FieldName          string                 `json:"-"` // For goas
	DisabledFieldNames map[string]struct{}    `json:"-"` // For goas
	Type               string                 `json:"type,omitempty"`
	Format             string                 `json:"format,omitempty"`
	Required           []string               `json:"required,omitempty"`
	Properties         *orderedmap.OrderedMap `json:"properties,omitempty"`
	Description        string                 `json:"description,omitempty"`
	Items              *SchemaObject          `json:"items,omitempty"` // use ptr to prevent recursive error
	Example            interface{}            `json:"example,omitempty"`
	Deprecated         bool                   `json:"deprecated,omitempty"`
	Ref                string                 `json:"$ref,omitempty"` // Ref is used when SchemaObject is as a ReferenceObject
	Enum               interface{}            `json:"enum,omitempty"`

	// Title
	// MultipleOf
	// Maximum
	// ExclusiveMaximum
	// Minimum
	// ExclusiveMinimum
	// MaxLength
	// MinLength
	// Pattern
	// MaxItems
	// MinItems
	// UniqueItems
	// MaxProperties
	// MinProperties
	// Enum
	// AllOf
	// OneOf
	// AnyOf
	// Not
	// AdditionalProperties
	// Description
	// Default
	// Nullable
	// ReadOnly
	// WriteOnly
	// XML
	// ExternalDocs
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
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`

	// Ref is used when HeaderObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`
}

type ComponentsObject struct {
	Schemas         map[string]*SchemaObject         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecuritySchemeObject `json:"securitySchemes,omitempty"`
	Parameters      map[string]*ParameterObject      `json:"parameters,omitempty"`
	// Responses
	// Parameters

	// Examples
	// RequestBodies
	// Headers
	// Links
	// Callbacks
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
