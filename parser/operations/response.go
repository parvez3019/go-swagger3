package operations

import (
	"fmt"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
	"regexp"
	"strconv"
	"strings"
)

func (p *parser) parseResponseComment(pkgPath, pkgName string, operation *oas.OperationObject, comment string) error {
	// {status}  {jsonType}  {goType}     {description}
	// 201       object      models.User  "User Model"
	re := regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([\w\-\.\/\[\]{}]+)[^"]*(.*)?`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) != 5 {
		return fmt.Errorf("parseResponseComment can not parse response comment \"%s\"", comment)
	}

	status := matches[1]
	_, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("parseResponseComment: http status must be int, but got %s", status)
	}
	switch matches[2] {
	case "object", "array", "{object}", "{array}":
	default:
		return fmt.Errorf("parseResponseComment: invalid jsonType %s", matches[2])
	}
	responseObject := &oas.ResponseObject{
		Content: map[string]*oas.MediaTypeObject{},
	}
	responseObject.Description = strings.Trim(matches[4], "\"")

	re = regexp.MustCompile(`\[\w*\]`)
	goType := re.ReplaceAllString(matches[3], "[]")
	if strings.HasPrefix(goType, "map[]") {
		schema, err := p.ParseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.Debug("parseResponseComment cannot parse goType", goType)
		}
		responseObject.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{
			Schema: *schema,
		}
	} else if strings.HasPrefix(goType, "[]") {
		goType = strings.Replace(goType, "[]", "", -1)
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
		typeName, err := p.RegisterType(pkgPath, pkgName, matches[3])
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
	operation.Responses[status] = responseObject
	return nil
}
