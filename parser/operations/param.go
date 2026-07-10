package operations

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/iancoleman/orderedmap"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
)

var paramCommentRegex = regexp.MustCompile(`([-.\w]+)[\s]+([\w]+)[\s]+([\w./\[\]\{\}=,$:]+)[\s]+([\w]+)[\s]+"([^"]+)"(?:[\s]+"([^"]*)")?`)
var paramAttrRegex = regexp.MustCompile(`(?i)(enums|enum|default|minimum|maximum|minlength|maxlength|format|collectionformat|example|explode|style)\s*\(([^)]*)\)`)

func (p *parser) parseParamComment(pkgPath, pkgName string, operation *oas.OperationObject, comment string) error {
	// {name}  {in}  {goType}  {required}  {description}  {example (optional)}  {attributes...}
	loc := paramCommentRegex.FindStringSubmatchIndex(comment)
	matches := paramCommentRegex.FindStringSubmatch(comment)
	if len(matches) < 6 || loc == nil {
		return fmt.Errorf("parseParamComment can not parse param comment \"%s\"", comment)
	}

	parameterObject := oas.ParameterObject{}
	appendName(&parameterObject, matches[1])
	appendIn(&parameterObject, matches[2])
	appendRequired(&parameterObject, matches[4])
	appendDescription(&parameterObject, matches[5])
	example := ""
	if len(matches) > 6 {
		example = matches[6]
	}
	appendExample(&parameterObject, example)

	remainder := strings.TrimSpace(comment[loc[1]:])
	applyParamAttributes(&parameterObject, remainder)

	goType := getType(matches)

	switch parameterObject.In {
	// file, form
	case "form", "file":
		appendRequestBody(operation, parameterObject, goType)
		return nil
	// body
	case "body":
		return p.parseRequestBody(pkgPath, pkgName, operation, parameterObject, goType, matches)

	// path, query, header, cookie
	default:
		return p.appendQueryParam(pkgPath, pkgName, operation, parameterObject, goType)
	}
}

func applyParamAttributes(parameterObject *oas.ParameterObject, remainder string) {
	if remainder == "" {
		return
	}
	if parameterObject.Schema == nil {
		parameterObject.Schema = &oas.SchemaObject{}
	}
	for _, match := range paramAttrRegex.FindAllStringSubmatch(remainder, -1) {
		key := strings.ToLower(match[1])
		value := strings.TrimSpace(match[2])
		switch key {
		case "enums", "enum":
			parameterObject.Schema.Enum = splitCSV(value)
		case "default":
			parameterObject.Schema.Default = coerceScalar(value)
		case "minimum":
			parameterObject.Schema.Minimum = parseFloat(value)
		case "maximum":
			parameterObject.Schema.Maximum = parseFloat(value)
		case "minlength":
			parameterObject.Schema.MinLength = parseUintAttr(value)
		case "maxlength":
			parameterObject.Schema.MaxLength = parseUintAttr(value)
		case "format":
			parameterObject.Schema.Format = value
		case "collectionformat":
			applyCollectionFormat(parameterObject, value)
		case "example":
			parameterObject.Example = coerceScalar(value)
			parameterObject.Schema.Example = coerceScalar(value)
		case "explode":
			b := strings.EqualFold(value, "true")
			parameterObject.Explode = &b
		case "style":
			parameterObject.Style = value
		}
	}
}

func applyCollectionFormat(parameterObject *oas.ParameterObject, format string) {
	switch strings.ToLower(format) {
	case "multi":
		parameterObject.Style = "form"
		b := true
		parameterObject.Explode = &b
	case "ssv":
		parameterObject.Style = "spaceDelimited"
		b := false
		parameterObject.Explode = &b
	case "pipes":
		parameterObject.Style = "pipeDelimited"
		b := false
		parameterObject.Explode = &b
	case "csv", "tsv":
		parameterObject.Style = "form"
		b := false
		parameterObject.Explode = &b
	}
}

func splitCSV(value string) []interface{} {
	parts := strings.Split(value, ",")
	out := make([]interface{}, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func coerceScalar(value string) interface{} {
	if value == "" {
		return ""
	}
	if b, err := strconv.ParseBool(value); err == nil && (value == "true" || value == "false") {
		return b
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}
	return value
}

func parseFloat(value string) float64 {
	f, _ := strconv.ParseFloat(value, 64)
	return f
}

func parseUintAttr(value string) uint {
	u, _ := strconv.ParseUint(value, 10, 64)
	return uint(u)
}

func (p *parser) parseRequestBody(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string, matches []string) error {
	if strings.HasPrefix(goType, "$ref:") {
		refName := strings.TrimPrefix(goType, "$ref:")
		operation.RequestBody = &oas.RequestBodyObject{
			Ref:         utils.AddRequestBodiesRefLinkPrefix(refName),
			Required:    parameterObject.Required,
			Description: parameterObject.Description,
		}
		return nil
	}
	contentType := contentTypeForRequest(operation)
	if operation.RequestBody == nil {
		operation.RequestBody = &oas.RequestBodyObject{
			Content:  map[string]*oas.MediaTypeObject{},
			Required: parameterObject.Required,
		}
	}
	if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[]") || goType == "time.Time" {
		return p.parseArrayMapOrTimeType(pkgPath, pkgName, operation, goType, contentType)
	}
	return p.parseGoBasicTypeOrStructType(pkgPath, pkgName, operation, matches, contentType)
}

func (p *parser) parseGoBasicTypeOrStructType(pkgPath string, pkgName string, operation *oas.OperationObject, matches []string, contentType string) error {
	typeName, err := p.RegisterType(pkgPath, pkgName, matches[3])
	if err != nil {
		return err
	}
	if utils.IsBasicGoType(typeName) {
		operation.RequestBody.Content[contentType] = &oas.MediaTypeObject{Schema: oas.SchemaObject{Type: "string"}}
		return nil
	}
	operation.RequestBody.Content[contentType] = &oas.MediaTypeObject{Schema: oas.SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(typeName)}}
	return nil
}

func (p *parser) parseArrayMapOrTimeType(pkgPath string, pkgName string, operation *oas.OperationObject, goType string, contentType string) error {
	parsedSchemaObject, err := p.ParseSchemaObject(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parseResponseComment cannot parse goType", goType)
		return err
	}
	if parsedSchemaObject != nil {
		operation.RequestBody.Content[contentType] = &oas.MediaTypeObject{Schema: *parsedSchemaObject}
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
		return nil
	}
	if p.isEnumType(goType) {
		p.appendEnumParamRef(goType, parameterObject, operation)
		return nil
	}
	if strings.Contains(goType, "model.") {
		return p.appendModelSchemaRef(pkgPath, pkgName, operation, parameterObject, goType)
	}
	return nil
}

func (p *parser) appendTimeParam(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) (err error) {
	schema, err := p.ParseSchemaObject(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parseResponseComment cannot parse goType", goType)
	}
	parameterObject.Schema = mergeParamSchema(parameterObject.Schema, schema)
	operation.Parameters = append(operation.Parameters, parameterObject)
	return err
}

func (p *parser) appendGoTypeParams(parameterObject oas.ParameterObject, goType string, operation *oas.OperationObject) {
	base := &oas.SchemaObject{
		Type:        utils.GoTypesOASTypes[goType],
		Format:      utils.GoTypesOASFormats[goType],
		Description: parameterObject.Description,
	}
	parameterObject.Schema = mergeParamSchema(parameterObject.Schema, base)
	operation.Parameters = append(operation.Parameters, parameterObject)
}

func (p *parser) appendModelSchemaRef(pkgPath string, pkgName string, operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) error {
	typeName, err := p.RegisterType(pkgPath, pkgName, goType)
	if err != nil {
		p.Debug("parse param model type failed", goType)
		return err
	}
	base := &oas.SchemaObject{
		Ref:  utils.AddSchemaRefLinkPrefix(typeName),
		Type: typeName,
	}
	parameterObject.Schema = mergeParamSchema(parameterObject.Schema, base)
	operation.Parameters = append(operation.Parameters, parameterObject)
	return nil
}

func (p *parser) appendEnumParamRef(goType string, parameterObject oas.ParameterObject, operation *oas.OperationObject) {
	goType = strings.ReplaceAll(goType, "model.", "")
	base := &oas.SchemaObject{Ref: utils.AddSchemaRefLinkPrefix(goType)}
	parameterObject.Schema = mergeParamSchema(parameterObject.Schema, base)
	operation.Parameters = append(operation.Parameters, parameterObject)
}

func (p *parser) isEnumType(goType string) bool {
	name := strings.ReplaceAll(goType, "model.", "")
	if p.RegisteredEnums != nil {
		if _, ok := p.RegisteredEnums[name]; ok {
			return true
		}
		if _, ok := p.RegisteredEnums[goType]; ok {
			return true
		}
	}
	return utils.IsEnumType(goType)
}

// mergeParamSchema keeps attributes already parsed from the comment (enum, default, …)
// and fills in type/ref/format from the Go type when missing.
func mergeParamSchema(existing, base *oas.SchemaObject) *oas.SchemaObject {
	if existing == nil {
		return base
	}
	if base == nil {
		return existing
	}
	if existing.Type == "" {
		existing.Type = base.Type
	}
	if existing.Format == "" {
		existing.Format = base.Format
	}
	if existing.Description == "" {
		existing.Description = base.Description
	}
	if existing.Ref == "" {
		existing.Ref = base.Ref
	}
	if existing.Items == nil {
		existing.Items = base.Items
	}
	return existing
}

func appendRequestBody(operation *oas.OperationObject, parameterObject oas.ParameterObject, goType string) {
	if parameterObject.In != "file" && parameterObject.In != "form" {
		return
	}
	contentType := oas.ContentTypeForm
	if len(operation.Accept) > 0 {
		contentType = operation.Accept[0]
	}
	if operation.RequestBody == nil {
		operation.RequestBody = &oas.RequestBodyObject{
			Content: map[string]*oas.MediaTypeObject{
				contentType: {Schema: oas.SchemaObject{Type: "object", Properties: orderedmap.New()}},
			},
			Required: parameterObject.Required,
		}
	}
	media := operation.RequestBody.Content[contentType]
	if media == nil {
		media = &oas.MediaTypeObject{Schema: oas.SchemaObject{Type: "object", Properties: orderedmap.New()}}
		operation.RequestBody.Content[contentType] = media
	}
	if media.Schema.Properties == nil {
		media.Schema.Properties = orderedmap.New()
	}
	if parameterObject.In == "file" {
		media.Schema.Properties.Set(parameterObject.Name, &oas.SchemaObject{
			Type:        "string",
			Format:      "binary",
			Description: parameterObject.Description,
		})
	}
	if utils.IsGoTypeOASType(goType) {
		media.Schema.Properties.Set(parameterObject.Name, &oas.SchemaObject{
			Type:        utils.GoTypesOASTypes[goType],
			Format:      utils.GoTypesOASFormats[goType],
			Description: parameterObject.Description,
		})
	}
}

func getType(matches []string) string {
	return normalizeAnnotatedGoType(matches[3])
}

// normalizeAnnotatedGoType converts legacy leading [Type] array notation to []Type
// while preserving Generic[T] and Composition{field=Type} forms.
func normalizeAnnotatedGoType(goType string) string {
	if idx := strings.Index(goType, "["); idx > 0 {
		return goType
	}
	if strings.Contains(goType, "{") {
		return goType
	}
	re := regexp.MustCompile(`\[\w*\]`)
	return re.ReplaceAllString(goType, "[]")
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

func appendExample(parameterObject *oas.ParameterObject, example string) {
	if example == "" {
		parameterObject.Example = nil
	} else {
		parameterObject.Example = example
	}
}
