package writer

import (
	"encoding/json"
	"fmt"
	"os"

	oas "github.com/hanyue2020/go-swagger3/openApi3Schema"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type Writer interface {
	Write(openApiObject oas.OpenAPIObject, path string, isYaml bool) error
}

type fileWriter struct{}

func NewFileWriter() *fileWriter {
	return &fileWriter{}
}

func (w *fileWriter) Write(openApiObject oas.OpenAPIObject, path string, generateYAML bool) error {
	var (
		fd     *os.File
		err    error
		output []byte
	)
	log.Info("Writing to open api object file ...")
	fd, err = os.Create(path)
	if err != nil {
		return fmt.Errorf("Can not create the file %s: %v", path, err)
	}
	defer fd.Close()

	if generateYAML {
		output, err = yaml.Marshal(openApiObject)
		if err != nil {
			return err
		}
	} else {
		output, err = json.MarshalIndent(openApiObject, "", "  ")
		if err != nil {
			return err
		}
	}
	_, err = fd.WriteString(string(output))
	return err
}
