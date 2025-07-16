package plugin_test

import (
	"os"
	"testing"

	"github.com/cultureamp/migrations-runner-buildkite-plugin/plugin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailOnMissingRequiredEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	tests := []struct {
		name            string
		disabledEnvVars []string
		enabledEnvVars  map[string]string
		expectedErr     string
	}{
		{
			name: "all required parameters are unset",
			disabledEnvVars: []string{
				"BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_PARAMETER_NAME",
				"BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_COMMAND",
				"BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_TIMEOUT",
			},
			enabledEnvVars: map[string]string{},
			expectedErr:    "required key BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_PARAMETER_NAME missing value",
		},
		{
			name: "variable COMMAND set",
			disabledEnvVars: []string{
				"BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_PARAMETER_NAME",
			},
			enabledEnvVars: map[string]string{
				"BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_COMMAND": "bin/script",
			},
			expectedErr: "required key BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_PARAMETER_NAME missing value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range tc.disabledEnvVars {
				unsetEnv(t, key)
			}

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

func TestSucceedOnMissingOptionalEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	tests := []struct {
		name            string
		disabledEnvVars []string
		enabledEnvVars  map[string]string
	}{
		{
			name: "variable COMMAND unset",
			disabledEnvVars: []string{
				"BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_COMMAND",
			},
			enabledEnvVars: map[string]string{
				"BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_PARAMETER_NAME": "test-parameter",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range tc.disabledEnvVars {
				unsetEnv(t, key)
			}

			// set the environment variables
			for key, value := range tc.enabledEnvVars {
				t.Setenv(key, value)
			}

			// verify the fetcher throws the error specific to missing environment variable
			err := fetcher.Fetch(&config)
			assert.NoError(t, err)
		})
	}
}

func TestFetchConfigFromEnvironment(t *testing.T) {
	unsetEnv(t, "BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_PARAMETER_NAME")
	unsetEnv(t, "BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_COMMAND")
	unsetEnv(t, "BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_TIME_OUT")

	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	t.Setenv("BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_PARAMETER_NAME", "test-parameter")
	t.Setenv("BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_COMMAND", "hello-world")
	t.Setenv("BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_TIME_OUT", "600")

	err := fetcher.Fetch(&config)

	require.NoError(t, err, "fetch should not error")
	assert.Equal(t, "test-parameter", config.ParameterName, "fetched message should match environment")
	assert.Equal(t, "hello-world", config.Command, "fetched script should match environment")
	assert.Equal(t, 600, config.TimeOut, "fetched timeout should match environment")

	// test default value
	unsetEnv(t, "BUILDKITE_PLUGIN_MIGRATIONS_RUNNER_TIME_OUT")
	err = fetcher.Fetch(&config)
	require.NoError(t, err, "fetch should not error")
	assert.Equal(t, 2700, config.TimeOut, "fetched timeout should match environment")
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	// ensure state is restored correctly
	currValue, exists := os.LookupEnv(key)
	t.Cleanup(func() {
		if exists {
			os.Setenv(key, currValue)
		} else {
			os.Unsetenv(key)
		}
	})

	// clear the value
	os.Unsetenv(key)
}
