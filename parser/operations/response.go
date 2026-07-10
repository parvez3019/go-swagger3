package operations

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
)

func (p *parser) parseResponseComment(pkgPath, pkgName string, operation *oas.OperationObject, comment string) error {
	// {status}  {jsonType}  {goType}     {description}
	// 201       object      models.User  "User Model"
	// for cases of empty return payload
	// {status} {description}
	// 204 "User Model"
	// for cases of simple types
	// 200 {string} string "..."
	re := regexp.MustCompile(`(?P<status>[\d]+)[\s]*(?P<jsonType>[\w\{\}]+)?[\s]+(?P<goType>[\w\-\.\/\[\]]+)?[^"]*(?P<description>.*)?`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) <= 2 {
		return fmt.Errorf("parseResponseComment can not parse response comment \"%s\"", comment)
	}

	status := matches[1]
	statusInt, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("parseResponseComment: http status must be int, but got %s", status)
	}
	if !utils.IsValidHTTPStatusCode(statusInt) {
		return fmt.Errorf("parseResponseComment: Invalid http status code %s", status)
	}

	responseObject := &oas.ResponseObject{
		Content: map[string]*oas.MediaTypeObject{},
	}
	responseObject.Description = strings.Trim(matches[4], "\"")

	switch matches[2] {

	case "object", "array", "{object}", "{array}":
		err = p.complexResponseObject(pkgPath, pkgName, matches[3], responseObject)
	case "{string}", "{integer}", "{boolean}", "string", "integer", "boolean":
		err = p.simpleResponseObject(matches[2], responseObject)
	case "":

	default:
		return fmt.Errorf("parseResponseComment: invalid jsonType %s", matches[2])
	}

	if err != nil {
		return err
	}

	operation.Responses[status] = responseObject
	return nil
}

func (p *parser) parseResponseHeaderComment(operation *oas.OperationObject, comment string) error {
	// status code e.g. 200
	// header name e.g. Set-Cookie
	// type e.g. string
	// description e.g. JWT cookie
	// example (optional) e.g. accessToken=eyJ...
	re := regexp.MustCompile(`(\d+)\s+([\w-]+)\s+(\w+)\s+"([^"]*)"(?:\s+"([^"]*)")?`)
	matches := re.FindStringSubmatch(comment)

	// minimum required status + headerName + type + description
	if len(matches) < 5 {
		return fmt.Errorf("parseResponseHeaderComment can not parse comment \"%s\"", comment)
	}

	status := matches[1]
	statusInt, err := strconv.Atoi(status)
	if err != nil {
		return fmt.Errorf("parseResponseHeaderComment: http status must be int, but got %s", status)
	}
	if !utils.IsValidHTTPStatusCode(statusInt) {
		return fmt.Errorf("parseResponseHeaderComment: invalid http status code %s", status)
	}

	headerName := matches[2]
	headerType := matches[3]
	description := matches[4]
	example := matches[5] // empty string if no match

	// reusing ResponseObject so we don’t nuke the parsed content
	responseObject, exists := operation.Responses[status]
	if !exists {
		responseObject = &oas.ResponseObject{
			Content: map[string]*oas.MediaTypeObject{},
		}
		operation.Responses[status] = responseObject
	}

	// only init the map when we actually need i
	if responseObject.Headers == nil {
		responseObject.Headers = map[string]*oas.HeaderObject{}
	}

	schema := &oas.SchemaObject{
		Type: headerType,
	}
	if example != "" {
		schema.Example = example
	}

	responseObject.Headers[headerName] = &oas.HeaderObject{
		Description: description,
		Schema:      schema,
	}

	return nil
}

// function to parse cases of jsonType in case "object", "array", "{object}", "{array}":
func (p *parser) complexResponseObject(pkgPath, pkgName, typ string, responseObject *oas.ResponseObject) error {

	re := regexp.MustCompile(`\[\w*\]`)
	goType := re.ReplaceAllString(typ, "[]")
	if strings.HasPrefix(goType, "map[]") {
		schema, err := p.ParseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.Debug("parseResponseComment cannot parse goType", goType)
		}
		responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
			Schema: *schema,
		}
	} else if strings.HasPrefix(goType, "[]") {
		goType = strings.ReplaceAll(goType, "[]", "")
		typeName, err := p.RegisterType(pkgPath, pkgName, goType)
		if err != nil {
			return err
		}

		var s oas.SchemaObject

		if utils.IsBasicGoType(typeName) {
			s = oas.SchemaObject{
				Type: "string",
			}
		} else {
			s = oas.SchemaObject{
				Ref: utils.AddSchemaRefLinkPrefix(typeName),
			}
		}

		responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
			Schema: oas.SchemaObject{
				Type:  "array",
				Items: &s,
			},
		}
	} else {
		typeName, err := p.RegisterType(pkgPath, pkgName, typ)
		if err != nil {
			return err
		}
		if utils.IsBasicGoType(typeName) {
			responseObject.Content[oas.ContentTypeText] = &oas.MediaTypeObject{
				Schema: oas.SchemaObject{
					Type: "string",
				},
			}
		} else if utils.IsInterfaceType(typeName) {
			responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
				Schema: oas.SchemaObject{
					Type: "object",
				},
			}
		} else {
			responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
				Schema: oas.SchemaObject{
					Ref: utils.AddSchemaRefLinkPrefix(typeName),
				},
			}
		}
	}
	return nil
}

func (p *parser) simpleResponseObject(jsonType string, responseObject *oas.ResponseObject) error {
	formattedType := jsonType
	if strings.HasPrefix(jsonType, "{") && strings.HasSuffix(jsonType, "}") {
		formattedType = jsonType[1 : len(jsonType)-1]
	}

	responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{Schema: oas.SchemaObject{Type: formattedType}}
	return nil
}
