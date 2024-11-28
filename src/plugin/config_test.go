package plugin_test

import (
	"os"
	"testing"

	"github.com/cultureamp/examplego/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailOnMissingEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	t.Setenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE", "")
	os.Unsetenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE")

	err := fetcher.Fetch(&config)

	assert.Error(t, err, "fetch should error")
}

func TestFetchConfigFromEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	t.Setenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE", "test-message")

	err := fetcher.Fetch(&config)

	require.NoError(t, err, "fetch should not error")
	assert.Equal(t, "test-message", config.Message, "fetched message should match environment")
}
