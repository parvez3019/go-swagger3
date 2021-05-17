package writer

import (
	"encoding/json"
	"fmt"
	. "github.com/parvez3019/go-swagger3/openApi3Schema"
	log "github.com/sirupsen/logrus"
	"os"
)

type Writer interface {
	Write(OpenAPIObject, string) error
}

type fileWriter struct{}

func NewFileWriter() *fileWriter {
	return &fileWriter{}
}

func (w *fileWriter) Write(openApiObject OpenAPIObject, path string) error {
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
	_, err = fd.WriteString(string(output))
	return err
}
