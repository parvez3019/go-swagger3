package apis

import (
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/model"
	"github.com/parvez3019/go-swagger3/parser/operations"
	"github.com/parvez3019/go-swagger3/parser/schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
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

	masker utils.Masker
}

func NewParser(utils model.Utils, api *oas.OpenAPIObject, schemaParser schema.Parser, masker *utils.Masker) Parser {
	return &parser{
		Utils:           utils,
		OpenAPI:         api,
		schemaParser:    schemaParser,
		operationParser: operations.NewParser(utils, api, schemaParser, masker),
		masker:          *masker,
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
