package plugin_test

import (
	"os"
	"testing"

	"ecs-task-runner/plugin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailOnMissingEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	t.Setenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE", "")
	os.Unsetenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE")

	err := fetcher.Fetch(&config)

	require.Error(t, err, "fetch should error")
}

func TestFetchConfigFromEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME", "test-parameter")
	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT", "hello-world")

	err := fetcher.Fetch(&config)

	require.NoError(t, err, "fetch should not error")
	assert.Equal(t, "test-parameter", config.ParameterName, "fetched message should match environment")
	assert.Equal(t, "hello-world", config.Script, "fetched message should match environment")
}
