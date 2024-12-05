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
		name           string
		enabledEnvVars map[string]string
		expectedErr    string
	}{
		{
			name:           "all required parameters are unset",
			enabledEnvVars: map[string]string{},
			expectedErr:    "required key BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME missing value",
		},
		{
			name: "variable PARAMETER_NAME set",
			enabledEnvVars: map[string]string{
				"BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME": "test-parameter",
			},
			expectedErr: "required key BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT missing value",
		},
		{
			name: "variable SCRIPT set",
			enabledEnvVars: map[string]string{
				"BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT": "bin/script",
			},
			expectedErr: "required key BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME missing value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// clear the environment variables *before* each test case is run
			unsetEnvironmentVariables()
			// clear the environment variables *after* each test case is run
			defer unsetEnvironmentVariables()

			// set the environment variables
			for key, value := range tc.enabledEnvVars {
				t.Setenv(key, value)
			}

			// verify the fetcher throws the error specific to missing environment variable
			err := fetcher.Fetch(&config)
			assert.EqualError(t, err, tc.expectedErr, "fetch should error on missing environment variable")
		})
	}
}

func TestFetchConfigFromEnvironment(t *testing.T) {
	unsetEnvironmentVariables()
	defer unsetEnvironmentVariables()

	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME", "test-parameter")
	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT", "hello-world")

	err := fetcher.Fetch(&config)

	require.NoError(t, err, "fetch should not error")
	assert.Equal(t, "test-parameter", config.ParameterName, "fetched message should match environment")
	assert.Equal(t, "hello-world", config.Script, "fetched message should match environment")
}

// Unsets environment variables through an all-in-one function. Extend this with additional environment variables as
// needed.
func unsetEnvironmentVariables() {
	os.Unsetenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME")
	os.Unsetenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT")
}
