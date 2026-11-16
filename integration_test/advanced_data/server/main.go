package server

// @Title Advanced API
// @Version 2.1.0
// @Description Integration coverage for security schemes, schema edge cases and parameter styles
// @TermsOfServiceUrl https://example.com/terms
// @ContactName Advanced Support
// @ContactEmail advanced@example.com
// @ContactURL https://example.com/support
// @LicenseName Apache-2.0
// @LicenseURL https://www.apache.org/licenses/LICENSE-2.0
// @Server https://api.example.com/v2 Production
// @Server https://staging.example.com/v2 Staging
// @externalDocs https://example.com/handbook Developer handbook
// @tag.name items
// @tag.description Catalogue items
// @tag.docs.url https://example.com/docs/items
// @tag.docs.description Item guide
// @tag.name admin
// @tag.description Administrative operations
// @SecurityScheme BasicAuth http basic Admin credentials
// @SecurityScheme CookieAuth apiKey cookie session_id
// @SecurityScheme OIDC openIdConnect https://example.com/.well-known/openid-configuration Company SSO
// @SecurityScheme OAuth oauth2AuthCode https://example.com/oauth/authorize https://example.com/oauth/token
// @SecurityScheme OAuth oauth2Implicit https://example.com/oauth/authorize
// @SecurityScheme OAuth oauth2ResourceOwnerCredentials https://example.com/oauth/token
// @SecurityScheme OAuth oauth2ClientCredentials https://example.com/oauth/token
// @SecurityScope OAuth read:items Read catalogue items
// @SecurityScope OAuth write:items Modify catalogue items
// @Security OAuth read:items
func main() {}
