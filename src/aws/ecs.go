package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// internal interface for ecs
type EcsClientAPI interface {
	RunTask(ctx context.Context, params *ecs.RunTaskInput, optFns ...func(*ecs.Options)) (*ecs.RunTaskOutput, error)
	DescribeTasks(ctx context.Context, params *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error)
	DescribeTaskDefinition(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error)
}

type EcsWaiterAPI interface {
	WaitForOutput(ctx context.Context, params *ecs.DescribeTasksInput, maxWaitDur time.Duration, optFns ...func(*ecs.TasksStoppedWaiterOptions)) (*ecs.DescribeTasksOutput, error)
}

func SubmitTask(ctx context.Context, ecsAPI EcsClientAPI, input *TaskRunnerConfiguration) (string, error) {
	var containerOverrides = ContainerOverrideForConfig(input)

	response, err := ecsAPI.RunTask(ctx, &ecs.RunTaskInput{
		Cluster:    &input.Cluster,
		LaunchType: "FARGATE",
		Overrides: &types.TaskOverride{
			ContainerOverrides: containerOverrides,
		},
		TaskDefinition: &input.TaskDefinitionArn,
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				Subnets:        input.SubnetIds,
				SecurityGroups: input.SecurityGroupIds,
			},
		},
	})
	if err != nil {
		return "", err
	}

	if response.Tasks[0].TaskArn == nil {
		responseJSON, err := json.Marshal(response)
		if err != nil {
			return "", fmt.Errorf("error in unmarshalling response for failed RunTask: %w", err)
		}

		return "", fmt.Errorf("ecs:RunTask response contains no TaskArn: %v", string(responseJSON))
	}

	// this is working on the assumption that only one task is returned
	return *response.Tasks[0].TaskArn, nil
}

func WaitForCompletion(ctx context.Context, waiter EcsWaiterAPI, taskArn string, timeOut int) (*ecs.DescribeTasksOutput, error) {
	cluster := ClusterFromTaskArn(taskArn)

	maxWaitDuration := time.Duration(timeOut) * time.Second
	result, err := waiter.WaitForOutput(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []string{taskArn},
	}, maxWaitDuration)

	// the `DescribeTasksOutput` struct is returned even if there is an error. Counterintuitively, it happens to include failure information
	// which we may want to surface from the `Failures` struct field
	if err != nil {
		return result, err
	}

	// In a successful scenario, we should have a `tasks` slice with a single element
	return result, nil
}

func ContainerOverrideForConfig(input *TaskRunnerConfiguration) []types.ContainerOverride {
	if len(input.Command) == 0 {
		return []types.ContainerOverride{
			{
				Name: aws.String("migrations-runner"),
			},
		}
	}

	return []types.ContainerOverride{
		{
			Name:    aws.String("migrations-runner"),
			Command: input.Command,
		},
	}
}

func ClusterFromTaskArn(arn string) string {
	parts := strings.Split(arn, "/")
	return parts[len(parts)-2]
}

func TaskIDFromArn(taskArn string) string {
	parts := strings.Split(taskArn, "/")
	return parts[len(parts)-1]
}

// Acquires LogStream details for given ECS Task
func FindLogStreamFromTask(ctx context.Context, ecsClientAPI EcsClientAPI, task types.Task) (LogDetails, error) {
	response, err := ecsClientAPI.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: task.TaskDefinitionArn,
	})
	if err != nil {
		return LogDetails{}, err
	}

	if len(response.TaskDefinition.ContainerDefinitions) == 0 {
		return LogDetails{}, fmt.Errorf("ecs:DescribeTaskDefinition response is missing ContainerDefinitions data: %v", response)
	}

	container := response.TaskDefinition.ContainerDefinitions[0] // assume first container is the application container
	logGroupName := container.LogConfiguration.Options["awslogs-group"]
	//NOTE: Takes the format: prefix-name/container-name/ecs-task-id
	streamPrefix := container.LogConfiguration.Options["awslogs-stream-prefix"]

	// We need the logGroupName, streamPrefix, and a container name to be able to produce a FindLogStreamOutput in full
	if logGroupName == "" || streamPrefix == "" {
		return LogDetails{}, fmt.Errorf("cannot trace task output: container logging is not configured on task definition: %v", response.TaskDefinition.TaskDefinitionArn)
	}

	return LogDetails{
		logGroupName:  logGroupName,
		logStreamName: fmt.Sprintf("%s/%s/%s", streamPrefix, *container.Name, TaskIDFromArn(*task.TaskArn)),
	}, nil
}
