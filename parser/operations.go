package parser

import (
	"fmt"
	"github.com/iancoleman/orderedmap"
	. "github.com/parvez3019/goas/openApi3Schema"
	"go/ast"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func (p *parser) parseOperation(pkgPath, pkgName string, astComments []*ast.Comment) error {
	operation := &OperationObject{
		Responses: map[string]*ResponseObject{},
	}
	if !strings.HasPrefix(pkgPath, p.ModulePath) {
		// ignore this pkgName
		// p.debugf("parseOperation ignores %s", pkgPath)
		return nil
	} else if p.HandlerPath != "" && !strings.HasPrefix(pkgPath, p.HandlerPath) {
		return nil
	}
	var err error
	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if len(comment) == 0 {
			return nil
		}
		attribute := strings.Fields(comment)[0]
		switch strings.ToLower(attribute) {
		case "@title":
			operation.Summary = strings.TrimSpace(comment[len(attribute):])
		case "@description":
			operation.Description = strings.Join([]string{operation.Description, strings.TrimSpace(comment[len(attribute):])}, " ")
		case "@param":
			err = p.parseParamComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
		case "@header":
			err = p.parseHeaders(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
		case "@success", "@failure":
			err = p.parseResponseComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
		case "@resource", "@tag":
			resource := strings.TrimSpace(comment[len(attribute):])
			if resource == "" {
				resource = "others"
			}
			if !isInStringList(operation.Tags, resource) {
				operation.Tags = append(operation.Tags, resource)
			}
		case "@route", "@router":
			err = p.parseRouteComment(operation, comment)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseHeaders(pkgPath string, pkgName string, operation *OperationObject, comment string) error {
	schema, err := p.parseSchemaObject(pkgPath, pkgName, comment)
	if err != nil || schema.Properties == nil {
		return fmt.Errorf("parseHeaders can not parse Header schema %s", comment)
	}
	for _, key := range schema.Properties.Keys() {
		operation.Parameters = append(operation.Parameters, ParameterObject{
			Ref: addParametersRefLinkPrefix(key),
		})
	}
	return err
}

func (p *parser) parseParamComment(pkgPath, pkgName string, operation *OperationObject, comment string) error {
	// {name}  {in}  {goType}  {required}  {description}
	// user    body  User      true        "Info of a user."
	// f       file  ignored   true        "Upload a file."
	re := regexp.MustCompile(`([-\w]+)[\s]+([\w]+)[\s]+([\w./\[\]]+)[\s]+([\w]+)[\s]+"([^"]+)"`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) != 6 {
		return fmt.Errorf("parseParamComment can not parse param comment \"%s\"", comment)
	}

	parameterObject := ParameterObject{}
	appendName(&parameterObject, matches[1])
	appendIn(&parameterObject, matches[2])
	appendRequired(&parameterObject, matches[4])
	appendDescription(&parameterObject, matches[5])

	goType := getType(re, matches)

	// `file`, `form`
	appendRequestBody(operation, parameterObject, goType)

	// `path`, `query`, `header`, `cookie`
	if parameterObject.In != "body" {
		return p.appendQueryParam(pkgPath, pkgName, operation, parameterObject, goType)
	}

	if operation.RequestBody == nil {
		operation.RequestBody = &RequestBodyObject{
			Content:  map[string]*MediaTypeObject{},
			Required: parameterObject.Required,
		}
	}

	if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[]") || goType == "time.Time" {
		schema, err := p.parseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.debug("parseResponseComment cannot parse goType", goType)
		}
		operation.RequestBody.Content[ContentTypeJson] = &MediaTypeObject{
			Schema: *schema,
		}
	} else {
		typeName, err := p.registerType(pkgPath, pkgName, matches[3])
		if err != nil {
			return err
		}
		if isBasicGoType(typeName) {
			operation.RequestBody.Content[ContentTypeJson] = &MediaTypeObject{
				Schema: SchemaObject{
					Type: "string",
				},
			}
		} else {
			operation.RequestBody.Content[ContentTypeJson] = &MediaTypeObject{
				Schema: SchemaObject{
					Ref: addSchemaRefLinkPrefix(typeName),
				},
			}
		}
	}

	return nil
}

func (p *parser) appendQueryParam(pkgPath string, pkgName string, operation *OperationObject, parameterObject ParameterObject, goType string) error {
	if parameterObject.In == "path" {
		parameterObject.Required = true
	}
	if goType == "time.Time" {
		var err error
		parameterObject.Schema, err = p.parseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.debug("parseResponseComment cannot parse goType", goType)
		}
		operation.Parameters = append(operation.Parameters, parameterObject)
	} else if isGoTypeOASType(goType) {
		parameterObject.Schema = &SchemaObject{
			Type:        goTypesOASTypes[goType],
			Format:      goTypesOASFormats[goType],
			Description: parameterObject.Description,
		}
		operation.Parameters = append(operation.Parameters, parameterObject)
	} else if strings.Contains(goType, "model.") {
		typeName, err := p.registerType(pkgPath, pkgName, goType)
		if err != nil {
			p.debug("parse param model type failed", goType)
			return err
		}
		parameterObject.Schema = &SchemaObject{
			Ref:  addSchemaRefLinkPrefix(typeName),
			Type: typeName,
		}
		operation.Parameters = append(operation.Parameters, parameterObject)
	}
	return nil
}

func (p *parser) parseResponseComment(pkgPath, pkgName string, operation *OperationObject, comment string) error {
	// {status}  {jsonType}  {goType}     {description}
	// 201       object      models.User  "User Model"
	re := regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([\w\-\.\/\[\]]+)[^"]*(.*)?`)
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
	responseObject := &ResponseObject{
		Content: map[string]*MediaTypeObject{},
	}
	responseObject.Description = strings.Trim(matches[4], "\"")

	re = regexp.MustCompile(`\[\w*\]`)
	goType := re.ReplaceAllString(matches[3], "[]")
	if strings.HasPrefix(goType, "map[]") {
		schema, err := p.parseSchemaObject(pkgPath, pkgName, goType)
		if err != nil {
			p.debug("parseResponseComment cannot parse goType", goType)
		}
		responseObject.Content[ContentTypeJson] = &MediaTypeObject{
			Schema: *schema,
		}
	} else if strings.HasPrefix(goType, "[]") {
		goType = strings.Replace(goType, "[]", "", -1)
		typeName, err := p.registerType(pkgPath, pkgName, goType)
		if err != nil {
			return err
		}

		var s SchemaObject

		if isBasicGoType(typeName) {
			s = SchemaObject{
				Type: "string",
			}
		} else {
			s = SchemaObject{
				Ref: addSchemaRefLinkPrefix(typeName),
			}
		}

		responseObject.Content[ContentTypeJson] = &MediaTypeObject{
			Schema: SchemaObject{
				Type:  "array",
				Items: &s,
			},
		}
	} else {
		typeName, err := p.registerType(pkgPath, pkgName, matches[3])
		if err != nil {
			return err
		}
		if isBasicGoType(typeName) {
			responseObject.Content[ContentTypeText] = &MediaTypeObject{
				Schema: SchemaObject{
					Type: "string",
				},
			}
		} else {
			responseObject.Content[ContentTypeJson] = &MediaTypeObject{
				Schema: SchemaObject{
					Ref: addSchemaRefLinkPrefix(typeName),
				},
			}
		}
	}
	operation.Responses[status] = responseObject

	return nil
}

func (p *parser) parseRouteComment(operation *OperationObject, comment string) error {
	sourceString := strings.TrimSpace(comment[len("@Router"):])

	// /path [method]
	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	matches := re.FindStringSubmatch(sourceString)
	if len(matches) != 3 {
		return fmt.Errorf("Can not parse router comment \"%s\", skipped", comment)
	}

	_, ok := p.OpenAPI.Paths[matches[1]]
	if !ok {
		p.OpenAPI.Paths[matches[1]] = &PathItemObject{}
	}

	switch strings.ToUpper(matches[2]) {
	case http.MethodGet:
		p.OpenAPI.Paths[matches[1]].Get = operation
	case http.MethodPost:
		p.OpenAPI.Paths[matches[1]].Post = operation
	case http.MethodPatch:
		p.OpenAPI.Paths[matches[1]].Patch = operation
	case http.MethodPut:
		p.OpenAPI.Paths[matches[1]].Put = operation
	case http.MethodDelete:
		p.OpenAPI.Paths[matches[1]].Delete = operation
	case http.MethodOptions:
		p.OpenAPI.Paths[matches[1]].Options = operation
	case http.MethodHead:
		p.OpenAPI.Paths[matches[1]].Head = operation
	case http.MethodTrace:
		p.OpenAPI.Paths[matches[1]].Trace = operation
	}

	return nil
}

func appendRequestBody(operation *OperationObject, parameterObject ParameterObject, goType string) {
	if !(parameterObject.In == "file" || parameterObject.In == "form") {
		return
	}
	if operation.RequestBody == nil {
		operation.RequestBody = &RequestBodyObject{
			Content: map[string]*MediaTypeObject{
				ContentTypeForm: {Schema: SchemaObject{Type: "object", Properties: orderedmap.New()}},
			},
			Required: parameterObject.Required,
		}
	}
	if parameterObject.In == "file" {
		operation.RequestBody.Content[ContentTypeForm].Schema.Properties.Set(parameterObject.Name, &SchemaObject{
			Type:        "string",
			Format:      "binary",
			Description: parameterObject.Description,
		})
	}
	if isGoTypeOASType(goType) {
		operation.RequestBody.Content[ContentTypeForm].Schema.Properties.Set(parameterObject.Name, &SchemaObject{
			Type:        goTypesOASTypes[goType],
			Format:      goTypesOASFormats[goType],
			Description: parameterObject.Description,
		})
	}
}

func getType(re *regexp.Regexp, matches []string) string {
	re = regexp.MustCompile(`\[\w*\]`)
	goType := re.ReplaceAllString(matches[3], "[]")
	return goType
}

func appendRequired(paramObject *ParameterObject, isRequired string) {
	switch strings.ToLower(isRequired) {
	case "true", "required":
		paramObject.Required = true
	}
}

func appendDescription(parameterObject *ParameterObject, description string) {
	parameterObject.Description = description
}

func appendIn(parameterObject *ParameterObject, in string) {
	parameterObject.In = in
}

func appendName(parameterObject *ParameterObject, name string) {
	parameterObject.Name = name
}
