package info

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"strings"

	. "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/model"
	log "github.com/sirupsen/logrus"
)

type Parser interface {
	Parse() error
}

type parser struct {
	model.Utils
	OpenAPI *OpenAPIObject
}

func NewParser(utils model.Utils, api *OpenAPIObject) Parser {
	return &parser{
		Utils:   utils,
		OpenAPI: api,
	}
}

// Parse parse basic info
func (p *parser) Parse() error {
	log.Info("Parsing Info ...")
	fileTree, err := goparser.ParseFile(token.NewFileSet(), p.MainFilePath, nil, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("can not parse general API information: %v", err)
	}
	// Security Scopes are defined at a different level in the hierarchy as where they need to end up in the OpenAPI structure,
	// so a temporary list is needed.
	oauthScopes := make(map[string]map[string]string, 0)

	// Parse comments
	for _, commentGroup := range fileTree.Comments {
		p.parseCommentGroups(commentGroup, oauthScopes)
	}

	// Apply security scopes to their security schemes
	p.applySecurityScopeToSecuritySchemes(oauthScopes)
	p.appendDefaultServer()
	err = p.validateInfo()
	if err != nil {
		return err
	}
	return p.validateServers()
}

func (p *parser) parseCommentGroups(commentGroup *ast.CommentGroup, oauthScopes map[string]map[string]string) {
	for _, comment := range strings.Split(commentGroup.Text(), "\n") {
		p.parseComment(comment, oauthScopes)
	}
}

func (p *parser) parseComment(comment string, oauthScopes map[string]map[string]string) {
	attribute, value, notPresent := p.parseAttributeAndValue(comment)
	if notPresent {
		return
	}
	p.Debug(attribute, value)
	p.parseOpenApiInfo(attribute, value)
	p.parseServerUrls(attribute, value)
	p.parseSecurity(attribute, value)
	p.parseSecurityScheme(attribute, value)
	p.parseSecurityScope(attribute, value, oauthScopes)
}

func (p *parser) parseAttributeAndValue(comment string) (string, string, bool) {
	attribute := strings.ToLower(strings.Split(comment, " ")[0])
	if len(attribute) == 0 || attribute[0] != '@' {
		return "", "", true
	}
	value := strings.TrimSpace(comment[len(attribute):])
	if len(value) == 0 {
		return "", "", true
	}
	return attribute, value, false
}

func (p *parser) parseServerUrls(attribute string, value string) {
	if attribute == "@server" {
		p.parseServer(value)
	}
}

func (p *parser) parseOpenApiInfo(attribute string, value string) {
	switch attribute {
	case "@info.version", "@version":
		p.OpenAPI.Info.Version = value
	case "@info.title", "@title":
		p.OpenAPI.Info.Title = value
	case "@info.description", "@info.desc", "@desc", "@description":
		p.OpenAPI.Info.Description = value
	case "@info.termsofserviceurl":
		p.OpenAPI.Info.TermsOfService = value
	case "@contactname", "@contactemail", "contacturl":
		p.parseContact(attribute, value)
	case "@licensename":
		p.parseLicenseName(value)
	case "@licenseurl":
		p.parseLicenseUrl(value)
	}
}

func (p *parser) parseContact(attribute, value string) {
	if p.OpenAPI.Info.Contact == nil {
		p.OpenAPI.Info.Contact = &ContactObject{}
	}
	switch attribute {
	case "@contactname":
		p.OpenAPI.Info.Contact.Name = value
	case "@contactemail":
		p.OpenAPI.Info.Contact.Email = value
	case "@contacturl":
		p.OpenAPI.Info.Contact.URL = value
	}
}

func (p *parser) parseLicenseName(value string) {
	if p.OpenAPI.Info.License == nil {
		p.OpenAPI.Info.License = &LicenseObject{}
	}
	p.OpenAPI.Info.License.Name = value
}

func (p *parser) parseLicenseUrl(value string) {
	if p.OpenAPI.Info.License == nil {
		p.OpenAPI.Info.License = &LicenseObject{}
	}
	p.OpenAPI.Info.License.URL = value
}

func (p *parser) parseServer(value string) {
	fields := strings.Split(value, " ")
	s := ServerObject{URL: fields[0], Description: value[len(fields[0]):]}
	p.OpenAPI.Servers = append(p.OpenAPI.Servers, s)
}

func (p *parser) parseSecurity(attribute, value string) {
	if attribute != "@security" {
		return
	}
	fields := strings.Split(value, " ")
	security := map[string][]string{
		fields[0]: fields[1:],
	}
	p.OpenAPI.Security = append(p.OpenAPI.Security, security)
}

func (p *parser) appendDefaultServer() {
	if len(p.OpenAPI.Servers) < 1 {
		p.OpenAPI.Servers = append(p.OpenAPI.Servers, ServerObject{URL: "/", Description: "Default Server URL"})
	}
}

func (p *parser) validateServers() error {
	for i := range p.OpenAPI.Servers {
		if p.OpenAPI.Servers[i].URL == "" {
			return fmt.Errorf("servers[%d].url cannot not be empty", i)
		}
	}
	return nil
}

func (p *parser) validateInfo() error {
	if p.OpenAPI.Info.Title == "" {
		return fmt.Errorf("info.title cannot not be empty")
	}
	if p.OpenAPI.Info.Version == "" {
		return fmt.Errorf("info.version cannot not be empty")
	}
	return nil
}

func (p *parser) applySecurityScopeToSecuritySchemes(oauthScopes map[string]map[string]string) {
	for scheme := range p.OpenAPI.Components.SecuritySchemes {
		if p.OpenAPI.Components.SecuritySchemes[scheme].Type == "oauth2" {
			p.applySecurityScope(oauthScopes, scheme)
		}
	}
}

func (p *parser) applySecurityScope(oauthScopes map[string]map[string]string, scheme string) {
	if scopes, ok := oauthScopes[scheme]; ok {
		p.OpenAPI.Components.SecuritySchemes[scheme].OAuthFlows.ApplyScopes(scopes)
	}
}

func (p *parser) parseSecurityScope(attribute, value string, oauthScopes map[string]map[string]string) {
	if attribute != "@securityscope" {
		return
	}

	fields := strings.Split(value, " ")
	if _, ok := oauthScopes[fields[0]]; !ok {
		oauthScopes[fields[0]] = make(map[string]string, 0)
	}
	oauthScopes[fields[0]][fields[1]] = strings.Join(fields[2:], " ")
}

func (p *parser) parseSecurityScheme(attribute, value string) {
	if attribute != "@securityscheme" {
		return
	}
	fields := strings.Split(value, " ")

	var scheme *SecuritySchemeObject
	if strings.Contains(fields[1], "oauth2") {
		scheme = p.parseOauth2Scheme(fields)
	}
	if scheme == nil {
		scheme = &SecuritySchemeObject{Type: fields[1]}
	}

	p.OpenAPI.Components.SecuritySchemes[fields[0]] = p.parseSecuritySchemeFromFields(fields, scheme)
}

func (p *parser) parseSecuritySchemeFromFields(fields []string, scheme *SecuritySchemeObject) *SecuritySchemeObject {
	switch fields[1] {
	case "http":
		scheme.Scheme = fields[2]
		scheme.Description = strings.Join(fields[3:], " ")
	case "apiKey":
		scheme.In = fields[2]
		scheme.Name = fields[3]
		scheme.Description = strings.Join(fields[4:], "")
	case "openIdConnect":
		scheme.OpenIdConnectUrl = fields[2]
		scheme.Description = strings.Join(fields[3:], " ")
	case "oauth2AuthCode":
		scheme.OAuthFlows.AuthorizationCode = &SecuritySchemeOauthFlowObject{
			AuthorizationUrl: fields[2],
			TokenUrl:         fields[3],
			Scopes:           make(map[string]string, 0),
		}
	case "oauth2Implicit":
		scheme.OAuthFlows.Implicit = &SecuritySchemeOauthFlowObject{
			AuthorizationUrl: fields[2],
			Scopes:           make(map[string]string, 0),
		}
	case "oauth2ResourceOwnerCredentials":
		scheme.OAuthFlows.ResourceOwnerPassword = &SecuritySchemeOauthFlowObject{
			TokenUrl: fields[2],
			Scopes:   make(map[string]string, 0),
		}
	case "oauth2ClientCredentials":
		scheme.OAuthFlows.ClientCredentials = &SecuritySchemeOauthFlowObject{
			TokenUrl: fields[2],
			Scopes:   make(map[string]string, 0),
		}
	}
	return scheme
}

func (p *parser) parseOauth2Scheme(fields []string) *SecuritySchemeObject {
	if oauthScheme, ok := p.OpenAPI.Components.SecuritySchemes[fields[0]]; ok {
		return oauthScheme
	}
	return &SecuritySchemeObject{
		Type:       "oauth2",
		OAuthFlows: &SecuritySchemeOauthObject{},
	}
}
