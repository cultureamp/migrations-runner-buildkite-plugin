package aws

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockECSClient struct {
	mockRunTask                func(ctx context.Context, params *ecs.RunTaskInput, optFns ...func(*ecs.Options)) (*ecs.RunTaskOutput, error)
	mockDescribeTasks          func(ctx context.Context, params *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error)
	mockDescribeTaskDefinition func(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error)
}

type mockECSWaiter struct {
	mockWaitForOutput func(ctx context.Context, params *ecs.DescribeTasksInput, maxWaitDur time.Duration, optFns ...func(*ecs.TasksStoppedWaiterOptions)) (*ecs.DescribeTasksOutput, error)
}

func (m mockECSClient) RunTask(ctx context.Context, params *ecs.RunTaskInput, optFns ...func(*ecs.Options)) (*ecs.RunTaskOutput, error) {
	return m.mockRunTask(ctx, params, optFns...)
}

func (m mockECSClient) DescribeTasks(ctx context.Context, params *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error) {
	return m.mockDescribeTasks(ctx, params, optFns...)
}

func (m mockECSClient) DescribeTaskDefinition(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
	return m.mockDescribeTaskDefinition(ctx, params, optFns...)
}

func (m mockECSWaiter) WaitForOutput(ctx context.Context, params *ecs.DescribeTasksInput, maxWaitDur time.Duration, optFns ...func(*ecs.TasksStoppedWaiterOptions)) (*ecs.DescribeTasksOutput, error) {
	return m.mockWaitForOutput(ctx, params, maxWaitDur, optFns...)
}

func TestSubmitTask(t *testing.T) {
	taskConfig := TaskRunnerConfiguration{
		Cluster:           "test-cluster",
		Command:           []string{"echo", "hello"},
		SecurityGroupIds:  []string{"sg-123456"},
		SubnetIds:         []string{"subnet-123456"},
		TaskDefinitionArn: "arn:aws:ecs:us-west-2:123456789012:task-definition/test-task-1",
	}

	result1 := "arn:aws:ecs:us-west-2:123456789012:task/test-cluster/07cc583696bd44e0be450bff7314ddaf"
	result2 := "arn:aws:ecs:us-west-2:123456789012:task/test-cluster/fda4dd137ffb054eb0e44696b385cc70"
	mockedClient1 := mockECSClient{
		mockRunTask: func(ctx context.Context, params *ecs.RunTaskInput, optFns ...func(*ecs.Options)) (*ecs.RunTaskOutput, error) {
			return &ecs.RunTaskOutput{
				Tasks: []types.Task{
					{
						TaskArn: aws.String(result1),
					},
				},
			}, nil
		},
	}
	mockedClient2 := mockECSClient{
		mockRunTask: func(ctx context.Context, params *ecs.RunTaskInput, optFns ...func(*ecs.Options)) (*ecs.RunTaskOutput, error) {
			return &ecs.RunTaskOutput{
				Tasks: []types.Task{
					{
						TaskArn: aws.String(result2),
					},
				},
			}, nil
		},
	}

	tests := []struct {
		name     string
		input    *TaskRunnerConfiguration
		client   EcsClientAPI
		expected string
	}{
		{
			name:     "given taskConfig, it should return the ARN of the ECS task",
			input:    &taskConfig,
			client:   mockedClient1,
			expected: result1,
		},
		{
			name:     "given taskConfig, it should return the ARN of the ECS task",
			input:    &taskConfig,
			client:   mockedClient2,
			expected: result2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SubmitTask(context.TODO(), tc.client, tc.input)

			t.Logf("result: %v", result)
			t.Logf("expected: %v", tc.expected)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestClusterFromTaskArn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "given a task belonging to the 'test-cluster', it should return 'test-cluster' as the cluster name",
			input:    "arn:aws:ecs:us-west-2:123456789012:task/test-cluster/07cc583696bd44e0be450bff7314ddaf",
			expected: "test-cluster",
		},
		{
			name:     "given a task belonging to the 'jabroni-cluster', it should return 'jabroni-cluster' as the cluster name",
			input:    "arn:aws:ecs:us-west-2:123456789012:task/jabroni-cluster/07cc583696bd44e0be450bff7314ddaf",
			expected: "jabroni-cluster",
		},
		{
			name:     "given a task name that is shorter than usual, it should still return its cluster name",
			input:    "arn:aws:ecs:us-west-2:123456789012:task/jabroni-cluster/aaa",
			expected: "jabroni-cluster",
		},
		{
			name:     "given a task ARN is in a different region, it should still return it's cluster name",
			input:    "arn:aws:ecs:ap-southeast-2:123456789012:task/jabroni-cluster/07cc583696bd44e0be450bff7314ddaf",
			expected: "jabroni-cluster",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ClusterFromTaskArn(tc.input)

			t.Logf("result: %v", result)
			t.Logf("expected: %v", tc.expected)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTaskIdFromArn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "given a task belonging to the 'test-cluster', it should return '07cc583696bd44e0be450bff7314ddaf' as the task ID",
			input:    "arn:aws:ecs:us-west-2:123456789012:task/test-cluster/07cc583696bd44e0be450bff7314ddaf",
			expected: "07cc583696bd44e0be450bff7314ddaf",
		},
		{
			name:     "given an ARN that does not include the region, it should still return '07cc583696bd44e0be450bff7314ddaf' as the task ID",
			input:    "arn:aws:ecs::123456789012:task/test-cluster/07cc583696bd44e0be450bff7314ddaf",
			expected: "07cc583696bd44e0be450bff7314ddaf",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TaskIDFromArn(tc.input)

			t.Logf("result: %v", result)
			t.Logf("expected: %v", tc.expected)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFindLogStreamFromTask(t *testing.T) {
	task := types.Task{
		TaskArn:           aws.String("arn:aws:ecs:us-west-2:123456789012:task/test-cluster/07cc583696bd44e0be450bff7314ddaf"),
		TaskDefinitionArn: aws.String("arn:aws:ecs:us-west-2:123456789012:task-definition/test-task-1"),
	}

	taskDefinitionOutput := &ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &types.TaskDefinition{
			ContainerDefinitions: []types.ContainerDefinition{
				{
					Name: aws.String("test-container"),
					LogConfiguration: &types.LogConfiguration{
						Options: map[string]string{
							"awslogs-group":         "test-group",
							"awslogs-stream-prefix": "test-stream",
						},
					},
				},
			},
		},
	}

	positiveTests := []struct {
		name     string
		input    types.Task
		client   EcsClientAPI
		expected LogDetails
	}{
		{
			name:  "given a task ARN, it constructs a complete FindLogStreamOutput struct",
			input: task,
			client: mockECSClient{
				mockDescribeTaskDefinition: func(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
					return taskDefinitionOutput, nil
				},
			},
			expected: LogDetails{
				logGroupName:  "test-group",
				logStreamName: "test-stream/test-container/07cc583696bd44e0be450bff7314ddaf",
			},
		},
	}

	for _, tc := range positiveTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FindLogStreamFromTask(context.TODO(), tc.client, tc.input)

			t.Logf("result: %v", result)
			t.Logf("expected: %v", tc.expected)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}

	negativeTests := []struct {
		name     string
		input    types.Task
		client   EcsClientAPI
		expected LogDetails
	}{
		{
			name:  "when ContainerDefinitions is empty, it should return an error indicating that the ContainerDefinitions data is missing",
			input: task,
			client: mockECSClient{
				mockDescribeTaskDefinition: func(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
					return &ecs.DescribeTaskDefinitionOutput{
						TaskDefinition: &types.TaskDefinition{
							ContainerDefinitions: []types.ContainerDefinition{},
						},
					}, nil
				},
			},
			expected: LogDetails{},
		},
		{
			name:  "when logGroupName is empty, it should return an error indicating the logging configuration is incomplete",
			input: task,
			client: mockECSClient{
				mockDescribeTaskDefinition: func(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
					return &ecs.DescribeTaskDefinitionOutput{
						TaskDefinition: &types.TaskDefinition{
							ContainerDefinitions: []types.ContainerDefinition{
								{
									Name: aws.String("test-container"),
									LogConfiguration: &types.LogConfiguration{
										Options: map[string]string{
											"awslogs-group":         "",
											"awslogs-stream-prefix": "test-stream",
										},
									},
								},
							},
						},
					}, nil
				},
			},
			expected: LogDetails{},
		},
		{
			name:  "when streamPrefix is empty, it should return an error indicating the logging configuration is incomplete",
			input: task,
			client: mockECSClient{
				mockDescribeTaskDefinition: func(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
					return &ecs.DescribeTaskDefinitionOutput{
						TaskDefinition: &types.TaskDefinition{
							ContainerDefinitions: []types.ContainerDefinition{
								{
									Name: aws.String("test-container"),
									LogConfiguration: &types.LogConfiguration{
										Options: map[string]string{
											"awslogs-group":         "test-group",
											"awslogs-stream-prefix": "",
										},
									},
								},
							},
						},
					}, nil
				},
			},
			expected: LogDetails{},
		},
	}

	for _, tc := range negativeTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FindLogStreamFromTask(context.TODO(), tc.client, tc.input)

			t.Logf("result: %v", result)
			t.Logf("expected: %v", tc.expected)
			t.Logf("error: %v", err)

			require.Error(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// we probably don't need to test this function in its current state because it is effectively just a "pause" function
// to allow thing to finish in the background. The return value is used only for when a task fails, and we push
// this to a log.
func TestWaitForCompletion(t *testing.T) {
	mockedWaiter := mockECSWaiter{
		mockWaitForOutput: func(context.Context, *ecs.DescribeTasksInput, time.Duration, ...func(*ecs.TasksStoppedWaiterOptions)) (*ecs.DescribeTasksOutput, error) {
			return &ecs.DescribeTasksOutput{
				Failures: []types.Failure{
					{
						Arn:    aws.String("arn:aws:ecs:us-west-2:123456789012:task/test-cluster/07cc583696bd44e0be450bff7314ddaf"),
						Detail: aws.String("task stopped"),
						Reason: aws.String("computer is full of beanz"),
					},
				}}, errors.New("task stopped: computer is full of beanz")
		},
	}

	type expectedReturn struct {
		*ecs.DescribeTasksOutput
		error
	}

	//TODO: testing failure cases and how it looks in the logger
	tests := []struct {
		name     string
		input    string
		waiter   ecsWaiterAPI
		expected expectedReturn
	}{
		{
			name:   "given a task ARN, it should return the task details",
			input:  "arn:aws:ecs:us-west-2:123456789012:task/test-cluster/07cc583696bd44e0be450bff7314ddaf",
			waiter: mockedWaiter,
			expected: expectedReturn{&ecs.DescribeTasksOutput{
				Failures: []types.Failure{
					{
						Arn:    aws.String("arn:aws:ecs:us-west-2:123456789012:task/test-cluster/07cc583696bd44e0be450bff7314ddaf"),
						Detail: aws.String("task stopped"),
						Reason: aws.String("computer is full of beanz"),
					},
				}}, errors.New("task stopped: computer is full of beanz"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := WaitForCompletion(context.TODO(), tc.waiter, tc.input)
			t.Logf("result: '%v'", err)
			t.Logf("expected: detail: %v, reason: %v", *tc.expected.Failures[0].Detail, *tc.expected.Failures[0].Reason)

			// The function is most-useful when the underlying task fails. i.e. no news is good news in a real-world scenario
			// So, we will test the failure cases
			require.Error(t, err)
			assert.Equal(t, tc.expected.Failures[0], result.Failures[0])
		})
	}
}
