package integration_test

import (
	"fmt"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/parvez3019/go-swagger3/parser"
	"github.com/parvez3019/go-swagger3/writer"
	"github.com/stretchr/testify/assert"
)

// Characterisation test for roadmap features (separate from test_data goldens).
func Test_FeaturesSpec(t *testing.T) {
	if err := createFeaturesSpecFile(); err != nil {
		panic(fmt.Sprintf("could not run app - Error %s", err.Error()))
	}
	diff, _ := jsondiff.Compare(
		[]byte(LoadJSONAsString("features_data/spec/expected_features.json")),
		[]byte(LoadJSONAsString("features_data/spec/actual_features.json")),
		&jsondiff.Options{},
	)
	assert.Equal(t, jsondiff.FullMatch, diff)
}

func createFeaturesSpecFile() error {
	p, err := parser.NewParser(
		"features_data",
		"features_data/server/main.go",
		"",
		false,
		false,
		true,
		"",
		false,
	).Init()
	if err != nil {
		return err
	}
	openAPIObject, err := p.Parse()
	if err != nil {
		return err
	}
	return writer.NewFileWriter().Write(openAPIObject, "features_data/spec/actual_features.json", false, true)
}
