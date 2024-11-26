package plugin_test

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type AgentMock struct {
	mock.Mock
}

func (m *AgentMock) Annotate(ctx context.Context, message string, style string, annotationContext string) error {
	args := m.Called(ctx, message, style, annotationContext)
	return args.Error(0)
}
