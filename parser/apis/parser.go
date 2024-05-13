package apis

import (
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/operations"
	"github.com/parvez3019/go-swagger3/parser/schema"
	log "github.com/sirupsen/logrus"
)

type Parser interface {
	Parse() error
}

type parser struct {
	OpenAPI *oas.OpenAPIObject

	model.Utils
	schemaParser    schema.Parser
	operationParser operations.Parser
	TypeAliases     map[string]map[string]string // pkgName -> alias -> original
}

func NewParser(utils model.Utils, api *oas.OpenAPIObject, schemaParser schema.Parser) Parser {
	return &parser{
		Utils:           utils,
		OpenAPI:         api,
		schemaParser:    schemaParser,
		operationParser: operations.NewParser(utils, api, schemaParser),
		TypeAliases:     make(map[string]map[string]string),
	}
}

// Parse parse APIs info
func (p *parser) Parse() error {
	log.Info("Parsing APIs ...")
	err := p.parseImportStatements()
	if err != nil {
		return err
	}

	err = p.parseTypeSpecs()
	if err != nil {
		return err
	}

	err = p.parseParameters()
	if err != nil {
		return err
	}

	return p.parsePaths()
}
