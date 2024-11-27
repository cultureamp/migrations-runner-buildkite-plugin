package plugin_test

import (
	"os"
	"testing"

	"github.com/cultureamp/examplego/plugin"
	"github.com/stretchr/testify/assert"
)

func TestFailOnMissingEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	t.Setenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE", "")
	os.Unsetenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE")

	err := fetcher.Fetch(&config)

	assert.NotNil(t, err, "fetch should error")
}

func TestFetchConfigFromEnvironment(t *testing.T) {
	var config plugin.Config
	fetcher := plugin.EnvironmentConfigFetcher{}

	t.Setenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE", "test-message")

	err := fetcher.Fetch(&config)

	assert.Nil(t, err, "fetch should not error")
	assert.Equal(t, config.Message, "test-message", "fetched message should match environment")
}
