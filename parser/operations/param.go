package operations

import (
	"fmt"
	"regexp"
	"strings"

	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/utils"
	"github.com/iancoleman/orderedmap"
)

func (p *parser) parseParamComment(pkgPath, pkgName string, operation *oas.OperationObject, comment string) error {
	// {name}  {in}  {goType}  {required}  {description}
	// user    body  User      true        "Info of a user."
	// f       file  ignored   true        "Upload a file."
	re := regexp.MustCompile(`([-.\w]+)[\s]+([\w]+)[\s]+([\w./\[\]]+)[\s]+([\w]+)[\s]+"([^"]+)"`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) != 6 {
		return fmt.Errorf("parseParamComment can not parse param comment \"%s\"", comment)
	}

	parameterObject := oas.ParameterObject{}
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

	return p.parseRequestBody(pkgPath, pkgName, operation, parameterObject, goType, matches)
}

func (p *parser) parseRequestBody(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string, matches []string) error {
	if operation.RequestBody == nil {
		operation.RequestBody = &oas.RequestBodyObject{
			Content:  map[string]*oas.MediaTypeObject{},
			Required: parameterObject.Required,
		}
	}
	if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[]") || goType == "time.Time" {
		return p.parseArrayMapOrTimeType(pkgPath, pkgName, operation, goType)
	}
	return p.parseGoBasicTypeOrStructType(pkgPath, pkgName, operation, matches)
}

func (p *parser) parseGoBasicTypeOrStructType(pkgPath string, pkgName string, operation *oas.OperationObject, matches []string) error {
	typeName, err := p.RegisterType(pkgPath, pkgName, matches[3])
	if err != nil {
		return err
	}
	if utils.IsBasicGoType(typeName) {
		operation.RequestBody.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{Schema: oas.SchemaObject{Type: "string"}}
		return nil
	}
	operation.RequestBody.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{Schema: oas.SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(typeName)}}
	return nil
}

func (p *parser) parseArrayMapOrTimeType(pkgPath string, pkgName string, operation *oas.OperationObject, goType string) error {
	parsedSchemaObject, err := p.ParseSchemaObject(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parseResponseComment cannot parse goType", goType)
		return err
	}
	if parsedSchemaObject != nil {
		operation.RequestBody.Content[oas.ContentTypeJson] = &oas.MediaTypeObject{Schema: *parsedSchemaObject}
	}
	return nil
}

func (p *parser) appendQueryParam(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) error {
	if parameterObject.In == "path" {
		parameterObject.Required = true
	}
	if goType == "time.Time" {
		return p.appendTimeParam(pkgPath, pkgName, operation, parameterObject, goType)
	}
	if utils.IsGoTypeOASType(goType) {
		p.appendGoTypeParams(parameterObject, goType, operation)
	}
	if utils.IsEnumType(goType) {
		p.appendEnumParamRef(goType, parameterObject, operation)
		return nil
	}
	if strings.Contains(goType, "model.") {
		return p.appendModelSchemaRef(pkgPath, pkgName, operation, parameterObject, goType)
	}
	return nil
}

func (p *parser) appendTimeParam(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) (err error) {
	parameterObject.Schema, err = p.ParseSchemaObject(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parseResponseComment cannot parse goType", goType)
	}
	operation.Parameters = append(operation.Parameters, parameterObject)
	return err
}

func (p *parser) appendGoTypeParams(parameterObject oas.ParameterObject, goType string, operation *oas.OperationObject) {
	parameterObject.Schema = &oas.SchemaObject{
		Type:        utils.GoTypesOASTypes[goType],
		Format:      utils.GoTypesOASFormats[goType],
		Description: parameterObject.Description,
	}
	operation.Parameters = append(operation.Parameters, parameterObject)
}

func (p *parser) appendModelSchemaRef(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) error {
	typeName, err := p.RegisterType(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parse param model type failed", goType)
		return err
	}
	parameterObject.Schema = &oas.SchemaObject{
		Ref:  utils.AddSchemaRefLinkPrefix(typeName),
		Type: typeName,
	}
	operation.Parameters = append(operation.Parameters, parameterObject)
	return nil
}

func (p *parser) appendEnumParamRef(goType string, parameterObject oas.ParameterObject, operation *oas.OperationObject) {
	if strings.Contains(goType, "model.") {
		goType = strings.Replace(goType, "model.", "", -1)
	}
	parameterObject.Schema = &oas.SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(goType)}
	operation.Parameters = append(operation.Parameters, parameterObject)
}

func appendRequestBody(operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) {
	if !(parameterObject.In == "file" || parameterObject.In == "form") {
		return
	}
	if operation.RequestBody == nil {
		operation.RequestBody = &oas.RequestBodyObject{
			Content: map[string]*oas.MediaTypeObject{
				oas.ContentTypeForm: {Schema: oas.SchemaObject{Type: "object", Properties: orderedmap.New()}},
			},
			Required: parameterObject.Required,
		}
	}
	if parameterObject.In == "file" {
		operation.RequestBody.Content[oas.ContentTypeForm].Schema.Properties.Set(parameterObject.Name, &oas.SchemaObject{
			Type:        "string",
			Format:      "binary",
			Description: parameterObject.Description,
		})
	}
	if utils.IsGoTypeOASType(goType) {
		operation.RequestBody.Content[oas.ContentTypeForm].Schema.Properties.Set(parameterObject.Name, &oas.SchemaObject{
			Type:        utils.GoTypesOASTypes[goType],
			Format:      utils.GoTypesOASFormats[goType],
			Description: parameterObject.Description,
		})
	}
}

func getType(re *regexp.Regexp, matches []string) string {
	re = regexp.MustCompile(`\[\w*\]`)
	goType := re.ReplaceAllString(matches[3], "[]")
	return goType
}

func appendRequired(paramObject *oas.ParameterObject, isRequired string) {
	switch strings.ToLower(isRequired) {
	case "true", "required":
		paramObject.Required = true
	}
}

func appendDescription(parameterObject *oas.ParameterObject, description string) {
	parameterObject.Description = description
}

func appendIn(parameterObject *oas.ParameterObject, in string) {
	parameterObject.In = in
}

func appendName(parameterObject *oas.ParameterObject, name string) {
	parameterObject.Name = name
}
