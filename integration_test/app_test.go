package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/parvez3019/go-swagger3/parser"
	"github.com/parvez3019/go-swagger3/writer"
	"github.com/stretchr/testify/assert"
)

// Characterisation test for the refactoring
func Test_ShouldGenerateExpectedSpec(t *testing.T) {
	if err := createSpecFile(); err != nil {
		panic(fmt.Sprintf("could not run app - Error %s", err.Error()))
	}
	actual := LoadJSONAsString("test_data/spec/actual.json")
	actual += "\n" // append new line for test
	assert.Equal(t, LoadJSONAsString("test_data/spec/expected.json"), actual)
}

func LoadJSONAsString(path string) string {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Unable to open at %s", path))
	}
	content, _ := ioutil.ReadAll(file)
	return string(content)
}

func createSpecFile() error {
	p, err := parser.NewParser(
		"test_data",
		"test_data/server/main.go",
		"",
		nil,
		false,
		false,
		false,
	).Init()

	if err != nil {
		return err
	}
	openApiObject, err := p.Parse()
	if err != nil {
		return err
	}

	fw := writer.NewFileWriter()
	return fw.Write(openApiObject, "test_data/spec/actual.json", false)
}
