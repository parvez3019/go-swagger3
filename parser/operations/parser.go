package operations

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"strings"

	"github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/schema"
)

type Parser interface {
	Parse(pkgPath, pkgName string, astComments []*ast.Comment) error
}

type parser struct {
	OpenAPI *openApi3Schema.OpenAPIObject

	model.Utils
	schema.Parser
	usedOperationIds map[string]struct{} // Track used operation IDs
}

func NewParser(utils model.Utils, api *openApi3Schema.OpenAPIObject, schemaParser schema.Parser) Parser {
	return &parser{
		Utils:            utils,
		OpenAPI:          api,
		Parser:           schemaParser,
		usedOperationIds: make(map[string]struct{}),
	}
}

func (p *parser) Parse(pkgPath, pkgName string, astComments []*ast.Comment) error {
	operation := &openApi3Schema.OperationObject{Responses: map[string]*openApi3Schema.ResponseObject{}}
	if !strings.HasPrefix(pkgPath, p.ModulePath) || (p.HandlerPath != "" && !strings.HasPrefix(pkgPath, p.HandlerPath)) {
		return nil
	}

	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if len(comment) == 0 {
			continue
		}
		if err := p.parseOperationFromComment(pkgPath, pkgName, comment, operation); err != nil {
			return err
		}
	}
	return nil
}

// validateOperationID checks if an operation ID is unique and registers it if it is.
// Returns an error if the operation ID is already used.
func (p *parser) validateOperationID(operationID string) error {
	if operationID == "" {
		return nil
	}
	if _, exists := p.usedOperationIds[operationID]; exists {
		return fmt.Errorf("operation ID '%s' is not unique", operationID)
	}
	p.usedOperationIds[operationID] = struct{}{}
	return nil
}

func (p *parser) parseOperationFromComment(pkgPath string, pkgName string, comment string, operation *openApi3Schema.OperationObject) error {
	attribute := strings.Fields(comment)[0]
	switch strings.ToLower(attribute) {
	case "@title":
		operation.Summary = strings.TrimSpace(comment[len(attribute):])
	case "@description":
		operation.Description = strings.Join([]string{operation.Description, strings.TrimSpace(comment[len(attribute):])}, " ")
	case "@param":
		return p.parseParamComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@header":
		return p.parseHeaders(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@success", "@failure":
		return p.parseResponseComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):]))
	case "@responseheader":
		return p.parseResponseHeaderComment(operation, strings.TrimSpace(comment[len(attribute):]))
	case "@resource", "@tag":
		p.parseResourceAndTag(comment, attribute, operation)
	case "@route", "@router":
		return p.parseRouteComment(operation, comment)
	case "@operationid":
		operationID := strings.TrimSpace(comment[len(attribute):])
		if err := p.validateOperationID(operationID); err != nil {
			return err
		}
		operation.OperationID = operationID
	case "@accept":
		operation.Accept = append(operation.Accept, parseMIMEList(strings.TrimSpace(comment[len(attribute):]))...)
	case "@produce":
		operation.Produce = append(operation.Produce, parseMIMEList(strings.TrimSpace(comment[len(attribute):]))...)
	case "@security":
		if sec := parseOperationSecurity(strings.TrimSpace(comment[len(attribute):])); sec != nil {
			operation.Security = append(operation.Security, sec)
		}
	case "@deprecated":
		operation.Deprecated = true
	case "@externaldocs":
		url, desc := parseExternalDocsValue(strings.TrimSpace(comment[len(attribute):]))
		docs := ensureOperationExternalDocs(operation)
		if url != "" {
			docs.URL = url
		}
		if desc != "" {
			docs.Description = desc
		}
	case "@externaldocs.description":
		ensureOperationExternalDocs(operation).Description = strings.TrimSpace(comment[len(attribute):])
	case "@externaldocs.url":
		ensureOperationExternalDocs(operation).URL = strings.TrimSpace(comment[len(attribute):])
	case "@extension":
		return parseOperationExtension(operation, strings.TrimSpace(comment[len(attribute):]))
	case "@discriminator":
		// Optional: store as x-discriminator hint when multiple responses share a status.
		ensureOperationExtensions(operation)["x-discriminator"] = strings.TrimSpace(comment[len(attribute):])
	default:
		if strings.HasPrefix(strings.ToLower(attribute), "@x-") {
			return parseOperationExtension(operation, comment)
		}
	}
	return nil
}

func ensureOperationExtensions(operation *openApi3Schema.OperationObject) map[string]interface{} {
	if operation.Extensions == nil {
		operation.Extensions = map[string]interface{}{}
	}
	return operation.Extensions
}

// parseOperationExtension handles:
//
//	@extension x-foo {"a":1}
//	@x-foo {"a":1}
//	@x-codeSample file
func parseOperationExtension(operation *openApi3Schema.OperationObject, comment string) error {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return nil
	}
	fields := strings.Fields(comment)
	if len(fields) == 0 {
		return nil
	}
	name := fields[0]
	rest := strings.TrimSpace(comment[len(name):])
	if !strings.HasPrefix(name, "@") && !strings.HasPrefix(name, "x-") {
		// @extension x-foo ...
	} else if strings.HasPrefix(name, "@") {
		name = name[1:] // strip leading @
	}
	if !strings.HasPrefix(name, "x-") {
		name = "x-" + name
	}
	ext := ensureOperationExtensions(operation)
	if rest == "" {
		ext[name] = true
		return nil
	}
	if (strings.HasPrefix(rest, "{") && strings.HasSuffix(rest, "}")) ||
		(strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]")) {
		var raw interface{}
		if err := json.Unmarshal([]byte(rest), &raw); err != nil {
			ext[name] = rest
			return nil
		}
		ext[name] = raw
		return nil
	}
	ext[name] = rest
	return nil
}
