package codeartifact

import (
	_ "embed"
	"sync"

	"github.com/superplanehq/superplane/pkg/utils"
)

//go:embed example_data_on_package_version.json
var exampleDataOnPackageVersionBytes []byte

//go:embed example_output_describe_package_version.json
var exampleOutputDescribePackageVersionBytes []byte

var exampleDataOnPackageVersionOnce sync.Once
var exampleDataOnPackageVersion map[string]any

var exampleOutputDescribePackageVersionOnce sync.Once
var exampleOutputDescribePackageVersion map[string]any

func (t *OnPackageVersion) ExampleData() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleDataOnPackageVersionOnce, exampleDataOnPackageVersionBytes, &exampleDataOnPackageVersion)
}

func (c *DescribePackageVersion) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(
		&exampleOutputDescribePackageVersionOnce,
		exampleOutputDescribePackageVersionBytes,
		&exampleOutputDescribePackageVersion,
	)
}
