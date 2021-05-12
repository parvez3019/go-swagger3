package parser

import (
	"fmt"
	. "github.com/parvez3019/goas/openApi3Schema"
	goparser "go/parser"
	"go/token"
	"strings"
)

func (p *parser) parseInfo() error {
	fileTree, err := goparser.ParseFile(token.NewFileSet(), p.MainFilePath, nil, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("can not parse general API information: %v", err)
	}

	// Security Scopes are defined at a different level in the hierarchy as where they need to end up in the OpenAPI structure,
	// so a temporary list is needed.
	oauthScopes := make(map[string]map[string]string, 0)

	if fileTree.Comments != nil {
		for i := range fileTree.Comments {
			for _, comment := range strings.Split(fileTree.Comments[i].Text(), "\n") {
				attribute := strings.ToLower(strings.Split(comment, " ")[0])
				if len(attribute) == 0 || attribute[0] != '@' {
					continue
				}
				value := strings.TrimSpace(comment[len(attribute):])
				if len(value) == 0 {
					continue
				}
				// p.debug(attribute, value)
				switch attribute {
				case "@version":
					p.OpenAPI.Info.Version = value
				case "@title":
					p.OpenAPI.Info.Title = value
				case "@description":
					p.OpenAPI.Info.Description = value
				case "@termsofserviceurl":
					p.OpenAPI.Info.TermsOfService = value
				case "@contactname":
					if p.OpenAPI.Info.Contact == nil {
						p.OpenAPI.Info.Contact = &ContactObject{}
					}
					p.OpenAPI.Info.Contact.Name = value
				case "@contactemail":
					if p.OpenAPI.Info.Contact == nil {
						p.OpenAPI.Info.Contact = &ContactObject{}
					}
					p.OpenAPI.Info.Contact.Email = value
				case "@contacturl":
					if p.OpenAPI.Info.Contact == nil {
						p.OpenAPI.Info.Contact = &ContactObject{}
					}
					p.OpenAPI.Info.Contact.URL = value
				case "@licensename":
					if p.OpenAPI.Info.License == nil {
						p.OpenAPI.Info.License = &LicenseObject{}
					}
					p.OpenAPI.Info.License.Name = value
				case "@licenseurl":
					if p.OpenAPI.Info.License == nil {
						p.OpenAPI.Info.License = &LicenseObject{}
					}
					p.OpenAPI.Info.License.URL = value
				case "@server":
					fields := strings.Split(value, " ")
					s := ServerObject{URL: fields[0], Description: value[len(fields[0]):]}
					p.OpenAPI.Servers = append(p.OpenAPI.Servers, s)
				case "@security":
					fields := strings.Split(value, " ")
					security := map[string][]string{
						fields[0]: fields[1:],
					}
					p.OpenAPI.Security = append(p.OpenAPI.Security, security)
				case "@securityscheme":
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
						scheme = &SecuritySchemeObject{
							Type: fields[1],
						}
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
				case "@securityscope":
					fields := strings.Split(value, " ")

					if _, ok := oauthScopes[fields[0]]; !ok {
						oauthScopes[fields[0]] = make(map[string]string, 0)
					}

					oauthScopes[fields[0]][fields[1]] = strings.Join(fields[2:], " ")
				}
			}
		}
	}

	// Apply security scopes to their security schemes
	for scheme, _ := range p.OpenAPI.Components.SecuritySchemes {
		if p.OpenAPI.Components.SecuritySchemes[scheme].Type == "oauth2" {
			if scopes, ok := oauthScopes[scheme]; ok {
				p.OpenAPI.Components.SecuritySchemes[scheme].OAuthFlows.ApplyScopes(scopes)
			}
		}
	}

	if len(p.OpenAPI.Servers) < 1 {
		p.OpenAPI.Servers = append(p.OpenAPI.Servers, ServerObject{URL: "/", Description: "Default Server URL"})
	}

	if p.OpenAPI.Info.Title == "" {
		return fmt.Errorf("info.title cannot not be empty")
	}
	if p.OpenAPI.Info.Version == "" {
		return fmt.Errorf("info.version cannot not be empty")
	}
	for i := range p.OpenAPI.Servers {
		if p.OpenAPI.Servers[i].URL == "" {
			return fmt.Errorf("servers[%d].url cannot not be empty", i)
		}
	}

	return nil
}
