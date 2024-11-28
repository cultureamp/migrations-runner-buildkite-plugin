package plugin_test

import (
	"context"
	"ecs-task-runner/plugin"

	"github.com/stretchr/testify/mock"
)

type AgentMock struct {
	mock.Mock
}

type mockPlugin struct{}

func (m *AgentMock) Annotate(ctx context.Context, message string, style string, annotationContext string) error {
	args := m.Called(ctx, message, style, annotationContext)
	return args.Error(0)
}

// TODO: Run is overloaded with more than what was in the original template. We'd have to perform more-dependency injection if we use the original implementation
// In my head, it would make more sense to throw together a smaller mock Run function with just enough to test what we need
func (mp mockPlugin) RunMock(ctx context.Context, fetcher plugin.ConfigFetcher, agent plugin.Agent) error {
	var config plugin.Config
	err := fetcher.Fetch(&config)
	if err != nil {
		return err
	}

	err = agent.Annotate(ctx, config.ParameterName, "info", "message")
	if err != nil {
		return err
	}
	err = agent.Annotate(ctx, config.Script, "info", "message")
	if err != nil {
		return err
	}

	return nil
}
