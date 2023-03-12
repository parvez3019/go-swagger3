package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	oas "github.com/parvez3019/go-swagger3/openApi3Schema"
	log "github.com/sirupsen/logrus"
)

type Writer interface {
	Write(openApiObject oas.OpenAPIObject, path string, isYaml bool) error
}

type fileWriter struct{}

func NewFileWriter() *fileWriter {
	return &fileWriter{}
}

func filterSchemaWithoutPkg(openApiObject oas.OpenAPIObject) {

	for key := range openApiObject.Components.Schemas {
		key_sep := strings.Split(key, ".")
		key_sep_last := key_sep[len(key_sep)-1]

		_, ok := openApiObject.Components.Schemas[key_sep_last]
		if len(key_sep) > 1 && ok {
			delete(openApiObject.Components.Schemas, key_sep_last)
		}
	}

}

func (w *fileWriter) Write(openApiObject oas.OpenAPIObject, path string, generateYAML bool, schemaWithoutPkg bool) error {
	if !schemaWithoutPkg {
		filterSchemaWithoutPkg(openApiObject)
	}
	log.Info("Writing to open api object file ...")
	fd, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Can not create the file %s: %v", path, err)
	}
	defer fd.Close()

	output, err := json.MarshalIndent(openApiObject, "", "  ")
	if err != nil {
		return err
	}
	if generateYAML {
		output, err = yaml.JSONToYAML(output)
		if err != nil {
			return err
		}
	}
	_, err = fd.WriteString(string(output))
	return err
}
