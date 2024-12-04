package plugin_test

import (
	"os"
	"testing"

	"ecs-task-runner/plugin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailOnMissingRequiredEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	tests := []struct {
		name            string
		enabledEnvVars  map[string]string
		disabledEnvVars []string
		expectedErr     string
	}{
		{
			name:            "all required parameters are unset",
			enabledEnvVars:  map[string]string{},
			disabledEnvVars: []string{"BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME", "BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT"},
			expectedErr:     "required key BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME missing value",
		},
		{
			name: "variable PARAMETER_NAME set",
			enabledEnvVars: map[string]string{
				"BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME": "test-parameter",
			},
			disabledEnvVars: []string{"BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT"},
			expectedErr:     "required key BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT missing value",
		},
		{
			name: "variable SCRIPT set",
			enabledEnvVars: map[string]string{
				"BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT": "bin/script",
			},
			disabledEnvVars: []string{"BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME"},
			expectedErr:     "required key BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME missing value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// set the environment variables
			for key, value := range tc.enabledEnvVars {
				t.Setenv(key, value)
			}

			// unset the environment variables
			for _, key := range tc.disabledEnvVars {
				os.Unsetenv(key)
			}

			// verify the fetcher throws the error specific to missing environment variable
			err := fetcher.Fetch(&config)
			assert.EqualError(t, err, tc.expectedErr, "fetch should error on missing environment variable")
		})
	}
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
