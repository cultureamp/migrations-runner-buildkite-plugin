package plugin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	awsinternal "github.com/cultureamp/ecs-task-runner-buildkite-plugin/aws"
	"github.com/cultureamp/ecs-task-runner-buildkite-plugin/plugin"
	"github.com/stretchr/testify/require"
)

type MockBuildKiteAgent struct{}

func (m MockBuildKiteAgent) Annotate(ctx context.Context, message string, style string, annotationContext string) error {
	return nil
}

func TestRunPluginResponse(t *testing.T) {
	buildKiteAgent := MockBuildKiteAgent{}
	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_PARAMETER_NAME", "test-parameter")
	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_SCRIPT", "hello-world")
	t.Setenv("BUILDKITE_PLUGIN_ECS_TASK_RUNNER_TIME_OUT", "15")
	mockFetcher := plugin.EnvironmentConfigFetcher{}
	var config plugin.Config
	err := mockFetcher.Fetch(&config)
	require.NoError(t, err)

	mockContainers := map[string]types.Container{
		"success": {
			ExitCode: aws.Int32(0),
			Image:    aws.String("nginx"),
			Name:     aws.String("gateway"),
			Reason:   aws.String("Gracefully Terminated"),
		},
		"failed": {
			ExitCode: aws.Int32(1),
			Image:    aws.String("nginx"),
			Name:     aws.String("gateway"),
			Reason:   aws.String("Panicked"),
		},
		"running": {
			Image:  aws.String("nginx"),
			Name:   aws.String("gateway"),
			Reason: aws.String("Panicked"),
		},
	}

	mockResponses := map[string]plugin.WaitForCompletion{
		"success": func(ctx context.Context, waiter awsinternal.EcsWaiterAPI, taskArn string, timeOut int) (*ecs.DescribeTasksOutput, error) {
			return &ecs.DescribeTasksOutput{
				Tasks: []types.Task{{
					Containers: []types.Container{
						mockContainers["success"],
					},
					LastStatus: aws.String("STOPPED"),
				},
				},
			}, nil
		},
		"failed": func(ctx context.Context, waiter awsinternal.EcsWaiterAPI, taskArn string, timeOut int) (*ecs.DescribeTasksOutput, error) {
			return &ecs.DescribeTasksOutput{
				Tasks: []types.Task{{
					Containers: []types.Container{
						mockContainers["failed"],
					},
					LastStatus: aws.String("STOPPED"),
				},
				},
				Failures: []types.Failure{
					{
						Arn:    aws.String("test-task-arn"),
						Reason: aws.String("Panicked"),
						Detail: aws.String("Container gateway panicked with non-zero exit code 1"),
					},
				},
			}, nil
		},
		"running": func(ctx context.Context, waiter awsinternal.EcsWaiterAPI, taskArn string, timeOut int) (*ecs.DescribeTasksOutput, error) {
			return &ecs.DescribeTasksOutput{
				Tasks: []types.Task{{
					Containers: []types.Container{
						mockContainers["running"],
					},
					LastStatus: aws.String("RUNNING"),
				},
				},
			}, errors.New("exceeded max wait time for TasksStopped waiter")
		},
	}

	expectedString := map[string]string{
		"success": "",
		"failed":  "task did not complete successfully",
		"running": "task did not complete within the time limit",
	}
	// expectedError := map[string]error{
	// 	"success": nil,
	// 	"failed":  errors.New(expectedString["failed"]),
	// 	"running": errors.New(expectedString["running"]),
	// }

	for name, mockResponse := range mockResponses {
		t.Run(name, func(t *testing.T) {
			result, err := mockResponse(context.TODO(), nil, "test-task-arn", 15)
			plugin := plugin.TaskRunnerPlugin{}
			err = plugin.HandleResults(context.TODO(), result, err, buildKiteAgent, config)
			if err != nil {
				require.ErrorContains(t, err, expectedString[name])
				t.Logf("expected: %v, actual: %v", expectedString[name], err)
			}
		})
	}
}
