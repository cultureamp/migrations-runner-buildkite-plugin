package plugin_test

import (
	"context"
	"testing"

	"ecs-task-runner/plugin"

	"github.com/stretchr/testify/require"
)

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
