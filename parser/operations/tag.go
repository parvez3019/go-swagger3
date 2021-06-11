package operations

import (
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	"github.com/parvez3019/go-swagger3/parser/utils"
	"strings"
)

func (p *parser) parseResourceAndTag(comment string, attribute string, operation *oas.OperationObject) {
	resource := strings.TrimSpace(comment[len(attribute):])
	if resource == "" {
		resource = "others"
	}
	if !utils.IsInStringList(operation.Tags, resource) {
		operation.Tags = append(operation.Tags, resource)
	}
}
