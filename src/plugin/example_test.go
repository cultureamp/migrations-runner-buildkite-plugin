package plugin_test

import (
	"context"
	"testing"

	"github.com/cultureamp/examplego/plugin"
	"github.com/stretchr/testify/assert"
)

func TestDoesAnnotate(t *testing.T) {
	agent := &AgentMock{}
	fetcher := plugin.EnvironmentConfigFetcher{}
	examplePlugin := plugin.ExamplePlugin{}
	ctx := context.Background()
	annotation := "test-message"

	t.Setenv("BUILDKITE_PLUGIN_EXAMPLE_GO_MESSAGE", annotation)
	agent.Mock.On("Annotate", ctx, annotation, "info", "message").Return(nil)

	err := examplePlugin.Run(ctx, fetcher, agent)

	assert.NoError(t, err, "should not error")
}
