package operations

import (
	"fmt"

	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
	"github.com/hanyue2020/go-swagger3/parser/utils"
)

func (p *parser) parseHeaders(pkgPath string, pkgName string, operation *oas.OperationObject, comment string) error {
	schema, err := p.ParseSchemaObject(pkgPath, pkgName, comment)
	if err != nil {
		return err
	}
	if schema.Properties == nil {
		return fmt.Errorf("NilSchemaProperties : parseHeaders can not parse Header schema %s", comment)
	}
	for _, key := range schema.Properties.Keys() {
		operation.Parameters = append(operation.Parameters, oas.ParameterObject{
			Ref: utils.AddParametersRefLinkPrefix(key),
		})
	}
	return err
}
