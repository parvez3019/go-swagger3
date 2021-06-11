package info

import (
	"fmt"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"strings"
)

type InfoParser interface {
	ParseInfo() error
}

type infoParser struct {
	model.Utils
	OpenAPI *OpenAPIObject
}

func NewInfoParser(utils model.Utils, api *OpenAPIObject) InfoParser {
	return &infoParser{
		Utils:   utils,
		OpenAPI: api,
	}
}

func (p *infoParser) ParseInfo() error {
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

func (p *infoParser) parseCommentGroups(commentGroup *ast.CommentGroup, oauthScopes map[string]map[string]string) {
	for _, comment := range strings.Split(commentGroup.Text(), "\n") {
		p.parseComment(comment, oauthScopes)
	}
}

func (p *infoParser) parseComment(comment string, oauthScopes map[string]map[string]string) {
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

func (p *infoParser) parseAttributeAndValue(comment string) (string, string, bool) {
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

func (p *infoParser) parseServerUrls(attribute string, value string) {
	if attribute == "@server" {
		p.parseServer(value)
	}
}

func (p *infoParser) parseOpenApiInfo(attribute string, value string) {
	switch attribute {
	case "@version":
		p.OpenAPI.Info.Version = value
	case "@title":
		p.OpenAPI.Info.Title = value
	case "@description":
		p.OpenAPI.Info.Description = value
	case "@termsofserviceurl":
		p.OpenAPI.Info.TermsOfService = value
	case "@contactname", "@contactemail", "contacturl":
		p.parseContact(attribute, value)
	case "@licensename":
		p.parseLicenseName(value)
	case "@licenseurl":
		p.parseLicenseUrl(value)
	}
}

func (p *infoParser) parseContact(attribute, value string) {
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

func (p *infoParser) parseLicenseName(value string) {
	if p.OpenAPI.Info.License == nil {
		p.OpenAPI.Info.License = &LicenseObject{}
	}
	p.OpenAPI.Info.License.Name = value
}

func (p *infoParser) parseLicenseUrl(value string) {
	if p.OpenAPI.Info.License == nil {
		p.OpenAPI.Info.License = &LicenseObject{}
	}
	p.OpenAPI.Info.License.URL = value
}

func (p *infoParser) parseServer(value string) {
	fields := strings.Split(value, " ")
	s := ServerObject{URL: fields[0], Description: value[len(fields[0]):]}
	p.OpenAPI.Servers = append(p.OpenAPI.Servers, s)
}

func (p *infoParser) parseSecurity(attribute, value string) {
	if attribute != "@security" {
		return
	}
	fields := strings.Split(value, " ")
	security := map[string][]string{
		fields[0]: fields[1:],
	}
	p.OpenAPI.Security = append(p.OpenAPI.Security, security)
}

func (p *infoParser) appendDefaultServer() {
	if len(p.OpenAPI.Servers) < 1 {
		p.OpenAPI.Servers = append(p.OpenAPI.Servers, ServerObject{URL: "/", Description: "Default Server URL"})
	}
}

func (p *infoParser) validateServers() error {
	for i := range p.OpenAPI.Servers {
		if p.OpenAPI.Servers[i].URL == "" {
			return fmt.Errorf("servers[%d].url cannot not be empty", i)
		}
	}
	return nil
}

func (p *infoParser) validateInfo() error {
	if p.OpenAPI.Info.Title == "" {
		return fmt.Errorf("info.title cannot not be empty")
	}
	if p.OpenAPI.Info.Version == "" {
		return fmt.Errorf("info.version cannot not be empty")
	}
	return nil
}

func (p *infoParser) applySecurityScopeToSecuritySchemes(oauthScopes map[string]map[string]string) {
	for scheme, _ := range p.OpenAPI.Components.SecuritySchemes {
		if p.OpenAPI.Components.SecuritySchemes[scheme].Type == "oauth2" {
			p.applySecurityScope(oauthScopes, scheme)
		}
	}
}

func (p *infoParser) applySecurityScope(oauthScopes map[string]map[string]string, scheme string) {
	if scopes, ok := oauthScopes[scheme]; ok {
		p.OpenAPI.Components.SecuritySchemes[scheme].OAuthFlows.ApplyScopes(scopes)
	}
}

func (p *infoParser) parseSecurityScope(attribute, value string, oauthScopes map[string]map[string]string) {
	if attribute != "@securityscope" {
		return
	}

	fields := strings.Split(value, " ")
	if _, ok := oauthScopes[fields[0]]; !ok {
		oauthScopes[fields[0]] = make(map[string]string, 0)
	}
	oauthScopes[fields[0]][fields[1]] = strings.Join(fields[2:], " ")
}

func (p *infoParser) parseSecurityScheme(attribute, value string) {
	if attribute != "@securityscheme" {
		return
	}
	fields := strings.Split(value, " ")

	var scheme *SecuritySchemeObject
	if strings.Contains(fields[1], "oauth2") {
		if oauthScheme, ok := p.OpenAPI.Components.SecuritySchemes[fields[0]]; ok {
			scheme = oauthScheme
		} else {
			scheme = &SecuritySchemeObject{
				Type:       "oauth2",
				OAuthFlows: &SecuritySchemeOauthObject{},
			}
		}
	}

	if scheme == nil {
		scheme = &SecuritySchemeObject{Type: fields[1]}
	}
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
	p.OpenAPI.Components.SecuritySchemes[fields[0]] = scheme
}
