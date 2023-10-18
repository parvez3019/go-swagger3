package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/nsf/jsondiff"

	"github.com/hanyue2020/go-swagger3/parser"
	"github.com/hanyue2020/go-swagger3/writer"
	"github.com/stretchr/testify/assert"
)

// Characterisation test for the refactoring

func Test_GenerateExpectedSpecWithPkg(t *testing.T) {
	if err := createSpecFile(false); err != nil {
		panic(fmt.Sprintf("could not run app - Error %s", err.Error()))
	}
	diff, _ := jsondiff.Compare([]byte(LoadJSONAsString("test_data/spec/expected_with_pkg.json")),
		[]byte(LoadJSONAsString("test_data/spec/actual.json")), &jsondiff.Options{})

	// assert the diff is FullMatch
	assert.Equal(t, jsondiff.FullMatch, diff)

}

func LoadJSONAsString(path string) string {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Unable to open at %s", path))
	}
	content, _ := ioutil.ReadAll(file)
	return string(content)
}

func createSpecFile(generateYaml bool) error {
	p, err := parser.NewParser(
		"test_data",
		"test_data/server/main.go",
		"",
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
	return fw.Write(openApiObject, "test_data/spec/actual.json", generateYaml)
}
