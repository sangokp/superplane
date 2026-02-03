package codeartifact

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/pkg/integrations/aws/common"
)

type DescribePackageVersion struct{}

type DescribePackageVersionConfiguration struct {
	Region           string `json:"region" mapstructure:"region"`
	DomainName       string `json:"domainName" mapstructure:"domainName"`
	DomainOwner      string `json:"domainOwner" mapstructure:"domainOwner"`
	RepositoryName   string `json:"repositoryName" mapstructure:"repositoryName"`
	PackageFormat    string `json:"packageFormat" mapstructure:"packageFormat"`
	PackageNamespace string `json:"packageNamespace" mapstructure:"packageNamespace"`
	PackageName      string `json:"packageName" mapstructure:"packageName"`
	PackageVersion   string `json:"packageVersion" mapstructure:"packageVersion"`
}

func (c *DescribePackageVersion) Name() string {
	return "aws.codeArtifact.describePackageVersion"
}

func (c *DescribePackageVersion) Label() string {
	return "CodeArtifact • Describe Package Version"
}

func (c *DescribePackageVersion) Description() string {
	return "Describe an AWS CodeArtifact package version"
}

func (c *DescribePackageVersion) Documentation() string {
	return `The Describe Package Version component retrieves metadata for a specific package version in AWS CodeArtifact.

## Use Cases

- **Release automation**: Resolve package metadata before promotion
- **Audit trails**: Capture version details for reporting
- **Dependency checks**: Validate status and origin of package versions

## Configuration

- **Region**: AWS region where the CodeArtifact domain lives
- **Domain Name**: CodeArtifact domain
- **Repository Name**: CodeArtifact repository
- **Package Format**: Package format (e.g., npm, maven, pypi)
- **Package Name**: Package name
- **Package Version**: Package version to describe

## Event Data

Outputs a single package version description object with metadata such as status, revision, and origin.`
}

func (c *DescribePackageVersion) Icon() string {
	return "aws"
}

func (c *DescribePackageVersion) Color() string {
	return "gray"
}

func (c *DescribePackageVersion) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *DescribePackageVersion) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:     "region",
			Label:    "Region",
			Type:     configuration.FieldTypeString,
			Required: true,
			Default:  "us-east-1",
		},
		{
			Name:     "domainName",
			Label:    "Domain Name",
			Type:     configuration.FieldTypeString,
			Required: true,
		},
		{
			Name:        "domainOwner",
			Label:       "Domain Owner",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Description: "Optional AWS account ID that owns the domain",
		},
		{
			Name:     "repositoryName",
			Label:    "Repository Name",
			Type:     configuration.FieldTypeString,
			Required: true,
		},
		{
			Name:        "packageFormat",
			Label:       "Package Format",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Placeholder: "npm",
		},
		{
			Name:        "packageNamespace",
			Label:       "Package Namespace",
			Type:        configuration.FieldTypeString,
			Required:    false,
			Description: "Required for maven, swift, and generic formats",
		},
		{
			Name:     "packageName",
			Label:    "Package Name",
			Type:     configuration.FieldTypeString,
			Required: true,
		},
		{
			Name:     "packageVersion",
			Label:    "Package Version",
			Type:     configuration.FieldTypeString,
			Required: true,
		},
	}
}

func (c *DescribePackageVersion) Setup(ctx core.SetupContext) error {
	var config DescribePackageVersionConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	config = normalizeDescribeConfig(config)

	if config.Region == "" {
		return fmt.Errorf("region is required")
	}
	if config.DomainName == "" {
		return fmt.Errorf("domain name is required")
	}
	if config.RepositoryName == "" {
		return fmt.Errorf("repository name is required")
	}
	if config.PackageFormat == "" {
		return fmt.Errorf("package format is required")
	}
	if config.PackageName == "" {
		return fmt.Errorf("package name is required")
	}
	if config.PackageVersion == "" {
		return fmt.Errorf("package version is required")
	}

	if requiresNamespace(config.PackageFormat) && config.PackageNamespace == "" {
		return fmt.Errorf("package namespace is required for format %s", config.PackageFormat)
	}

	return nil
}

func (c *DescribePackageVersion) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *DescribePackageVersion) Execute(ctx core.ExecutionContext) error {
	var config DescribePackageVersionConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	config = normalizeDescribeConfig(config)

	creds, err := common.CredentialsFromInstallation(ctx.Integration)
	if err != nil {
		return fmt.Errorf("failed to get AWS credentials: %w", err)
	}

	client := NewClient(ctx.HTTP, creds, config.Region)
	result, err := client.DescribePackageVersion(DescribePackageVersionInput{
		Domain:         config.DomainName,
		DomainOwner:    config.DomainOwner,
		Repository:     config.RepositoryName,
		Format:         config.PackageFormat,
		Namespace:      config.PackageNamespace,
		Package:        config.PackageName,
		PackageVersion: config.PackageVersion,
	})
	if err != nil {
		return fmt.Errorf("failed to describe package version: %w", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"aws.codeartifact.package.version",
		[]any{result},
	)
}

func (c *DescribePackageVersion) Actions() []core.Action {
	return []core.Action{}
}

func (c *DescribePackageVersion) HandleAction(ctx core.ActionContext) error {
	return nil
}

func (c *DescribePackageVersion) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return http.StatusOK, nil
}

func (c *DescribePackageVersion) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *DescribePackageVersion) Cleanup(ctx core.SetupContext) error {
	return nil
}

func normalizeDescribeConfig(config DescribePackageVersionConfiguration) DescribePackageVersionConfiguration {
	config.Region = strings.TrimSpace(config.Region)
	config.DomainName = strings.TrimSpace(config.DomainName)
	config.DomainOwner = strings.TrimSpace(config.DomainOwner)
	config.RepositoryName = strings.TrimSpace(config.RepositoryName)
	config.PackageFormat = strings.ToLower(strings.TrimSpace(config.PackageFormat))
	config.PackageNamespace = strings.TrimSpace(config.PackageNamespace)
	config.PackageName = strings.TrimSpace(config.PackageName)
	config.PackageVersion = strings.TrimSpace(config.PackageVersion)
	return config
}

func requiresNamespace(format string) bool {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "maven", "swift", "generic":
		return true
	default:
		return false
	}
}
