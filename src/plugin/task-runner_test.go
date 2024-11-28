package plugin_test

import (
	"context"
	"testing"

	"ecs-task-runner/plugin"

	"github.com/stretchr/testify/require"
)

type mockPlugin struct{}

func TestDoesAnnotate(t *testing.T) {
	agent := &AgentMock{}
	fetcher := plugin.EnvironmentConfigFetcher{}
	examplePlugin := mockPlugin{}
	ctx := context.Background()
	parameterName := "test-parameter"
	secretName := "hello-world"

	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME", parameterName)
	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT", secretName)
	agent.Mock.On("Annotate", ctx, parameterName, "info", "message").Return(nil)
	agent.Mock.On("Annotate", ctx, secretName, "info", "message").Return(nil)

	err := examplePlugin.RunMock(ctx, fetcher, agent)

	require.NoError(t, err, "should not error")
}

// TODO: Run is overloaded with more than what was in the original template. We'd have to perform more-dependency injection if we use the original implementation
// In my head, it would make more sense to throw together a smaller mock Run function with just enough to test what we need
func (mp mockPlugin) RunMock(ctx context.Context, fetcher plugin.ConfigFetcher, agent plugin.Agent) error {
	var config plugin.Config
	err := fetcher.Fetch(&config)
	if err != nil {
		return err
	}

	agent.Annotate(ctx, config.ParameterName, "info", "message")
	agent.Annotate(ctx, config.Script, "info", "message")

	return nil
}
