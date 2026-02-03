package codeartifact

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	"github.com/superplanehq/superplane/test/support/contexts"
)

func Test__DescribePackageVersion__Setup(t *testing.T) {
	component := &DescribePackageVersion{}

	t.Run("invalid configuration -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: "invalid",
		})

		require.ErrorContains(t, err, "failed to decode configuration")
	})

	t.Run("missing region -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         " ",
				"domainName":     "example-domain",
				"repositoryName": "example-repo",
				"packageFormat":  "npm",
				"packageName":    "example-package",
				"packageVersion": "1.2.3",
			},
		})

		require.ErrorContains(t, err, "region is required")
	})

	t.Run("missing domain name -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"repositoryName": "example-repo",
				"packageFormat":  "npm",
				"packageName":    "example-package",
				"packageVersion": "1.2.3",
			},
		})

		require.ErrorContains(t, err, "domain name is required")
	})

	t.Run("missing repository name -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"domainName":     "example-domain",
				"packageFormat":  "npm",
				"packageName":    "example-package",
				"packageVersion": "1.2.3",
			},
		})

		require.ErrorContains(t, err, "repository name is required")
	})

	t.Run("missing package format -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"domainName":     "example-domain",
				"repositoryName": "example-repo",
				"packageName":    "example-package",
				"packageVersion": "1.2.3",
			},
		})

		require.ErrorContains(t, err, "package format is required")
	})

	t.Run("missing package name -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"domainName":     "example-domain",
				"repositoryName": "example-repo",
				"packageFormat":  "npm",
				"packageVersion": "1.2.3",
			},
		})

		require.ErrorContains(t, err, "package name is required")
	})

	t.Run("missing package version -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"domainName":     "example-domain",
				"repositoryName": "example-repo",
				"packageFormat":  "npm",
				"packageName":    "example-package",
			},
		})

		require.ErrorContains(t, err, "package version is required")
	})

	t.Run("missing namespace for maven -> error", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"domainName":     "example-domain",
				"repositoryName": "example-repo",
				"packageFormat":  "maven",
				"packageName":    "example-package",
				"packageVersion": "1.2.3",
			},
		})

		require.ErrorContains(t, err, "package namespace is required")
	})

	t.Run("valid configuration -> ok", func(t *testing.T) {
		err := component.Setup(core.SetupContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"domainName":     "example-domain",
				"repositoryName": "example-repo",
				"packageFormat":  "npm",
				"packageName":    "example-package",
				"packageVersion": "1.2.3",
			},
		})

		require.NoError(t, err)
	})
}

func Test__DescribePackageVersion__Execute(t *testing.T) {
	component := &DescribePackageVersion{}

	t.Run("invalid configuration -> error", func(t *testing.T) {
		err := component.Execute(core.ExecutionContext{
			Configuration:  "invalid",
			ExecutionState: &contexts.ExecutionStateContext{KVs: map[string]string{}},
		})

		require.ErrorContains(t, err, "failed to decode configuration")
	})

	t.Run("missing credentials -> error", func(t *testing.T) {
		err := component.Execute(core.ExecutionContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"domainName":     "example-domain",
				"repositoryName": "example-repo",
				"packageFormat":  "npm",
				"packageName":    "example-package",
				"packageVersion": "1.2.3",
			},
			Integration:    &contexts.IntegrationContext{Secrets: map[string]core.IntegrationSecret{}},
			ExecutionState: &contexts.ExecutionStateContext{KVs: map[string]string{}},
		})

		require.ErrorContains(t, err, "AWS session credentials are missing")
	})

	t.Run("valid request -> emits package version", func(t *testing.T) {
		httpContext := &contexts.HTTPContext{
			Responses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`
						{
							"packageVersion": {
								"packageName": "example-package",
								"version": "1.2.3",
								"status": "Published",
								"format": "npm"
							}
						}
					`)),
				},
			},
		}

		execState := &contexts.ExecutionStateContext{KVs: map[string]string{}}
		err := component.Execute(core.ExecutionContext{
			Configuration: map[string]any{
				"region":         "us-east-1",
				"domainName":     "example-domain",
				"repositoryName": "example-repo",
				"packageFormat":  "npm",
				"packageName":    "example-package",
				"packageVersion": "1.2.3",
			},
			HTTP:           httpContext,
			ExecutionState: execState,
			Integration: &contexts.IntegrationContext{
				Secrets: map[string]core.IntegrationSecret{
					"accessKeyId":     {Name: "accessKeyId", Value: []byte("key")},
					"secretAccessKey": {Name: "secretAccessKey", Value: []byte("secret")},
					"sessionToken":    {Name: "sessionToken", Value: []byte("token")},
				},
			},
		})

		require.NoError(t, err)
		require.Len(t, execState.Payloads, 1)
		payload := execState.Payloads[0].(map[string]any)["data"]
		packageVersion, ok := payload.(*PackageVersionDescription)
		require.True(t, ok)
		assert.Equal(t, "example-package", packageVersion.PackageName)
		assert.Equal(t, "1.2.3", packageVersion.Version)

		require.Len(t, httpContext.Requests, 1)
		assert.Equal(t, "https://codeartifact.us-east-1.amazonaws.com/v1/package/version?domain=example-domain&format=npm&package=example-package&repository=example-repo&version=1.2.3", httpContext.Requests[0].URL.String())
	})
}
