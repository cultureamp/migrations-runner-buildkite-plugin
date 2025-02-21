package plugin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	awsinternal "github.com/cultureamp/ecs-task-runner-buildkite-plugin/aws"
	"github.com/cultureamp/ecs-task-runner-buildkite-plugin/buildkite"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type TaskRunnerPlugin struct {
}

type WaitForCompletion func(ctx context.Context, waiter awsinternal.EcsWaiterAPI, taskArn string, timeOut int) (*ecs.DescribeTasksOutput, error)
type ConfigFetcher interface {
	Fetch(config *Config) error
}

func (trp TaskRunnerPlugin) Run(ctx context.Context, fetcher ConfigFetcher, waiter WaitForCompletion) error {
	var config Config

	err := fetcher.Fetch(&config)
	if err != nil {
		return fmt.Errorf("plugin configuration error: %w", err)
	}
	buildKiteAgent := buildkite.Agent{}

	buildkite.Log("Executing task-runner plugin\n")

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	ssmClient := ssm.NewFromConfig(cfg)
	buildkite.Logf("Retrieving task configuration from: %s \n", config.ParameterName)
	configuration, err := awsinternal.RetrieveConfiguration(ctx, ssmClient, config.ParameterName)
	if err != nil {
		return fmt.Errorf("failed to retrieve configuration: %w", err)
	}

	// The `Command` configuration is optional. If it's not provided, we don't want to update the configuration struct
	// This check is here because otherwise it inserts a command with the value of an empty string and causes a panic
	// TODO: Can we decompose this?
	if config.Command != "" {
		configuration.Command = strings.Split(config.Command, " ")
	}

	ecsClient := ecs.NewFromConfig(cfg)
	taskArn, err := awsinternal.SubmitTask(ctx, ecsClient, configuration)
	if err != nil {
		return fmt.Errorf("failed to submit task: %w", err)
	}

	// FIXME: Confirm how this AWS Library code returns an error, and how it interacts with our `waiter` interface
	waiterClient := ecs.NewTasksStoppedWaiter(ecsClient, func(o *ecs.TasksStoppedWaiterOptions) {
		o.MinDelay = time.Second
		// TODO: This is currently a magic number. If we want this to be configurable, remove the nolint directive and fix it up
		o.MaxDelay = 10 * time.Second //nolint:mnd
	})

	//FIXME: sussing out why email-service ain't returning logs
	buildkite.Logf("Waiting for task to complete: %s\n", taskArn)
	result, err := waiter(ctx, waiterClient, taskArn, config.TimeOut)
	//FIXME: sussing out why email-service ain't returning logs
	buildkite.Logf("result after waiter: %+v\n", result)
	err = trp.HandleResults(ctx, result, err, buildKiteAgent, config)
	if err != nil {
		return fmt.Errorf("failed to handle task results: %w", err)
	}

	// In a successful scenario for task completion, we would have a `tasks` slice with a single element
	task := result.Tasks[0]
	taskLogDetails, err := awsinternal.FindLogStreamFromTask(ctx, ecsClient, task)
	if err != nil {
		return fmt.Errorf("failed to acquire log stream information for task: %w", err)
	}

	cloudwatchClient := cloudwatchlogs.NewFromConfig(cfg)
	logs, err := awsinternal.RetrieveLogs(ctx, cloudwatchClient, taskLogDetails)
	// Failing to retrieve the logs shouldn't be show-stopper if the task is able to complete successfully.
	// This can come from logs not being available yet, or the service lacking permissions to publish logs at the time
	// TODO: In the original implementation this is how it worked. Is there a possible way to "Wait" for logs?
	if err != nil {
		buildkite.LogFailuref("failed to retrieve CloudWatch Logs for job, continuing... %v", err)
	}

	if len(logs) > 0 {
		buildkite.Logf("CloudWatch Logs for job: \n")
		for _, l := range logs {
			if l.Timestamp != nil {
				// Applying ISO 8601 format, l.Timestamp is in milliseconds, not very useful in logging
				placeholder := time.UnixMilli(*l.Timestamp).Format(time.RFC3339)
				buildkite.Logf("-> %s %s\n", placeholder, *l.Message)
			}
		}
	}

	// TODO: Assuming the task only has 1 container. What if there others? Like Datadog sidecar
	if *task.Containers[0].ExitCode != 0 {
		buildkite.LogFailuref("Task stopped with a non-zero exit code:: %d\n", *task.Containers[0].ExitCode)
		// TODO: At about here, a structured return type of "success: true/false" and "error" is returned
	} else {
		buildkite.Log("Task completed successfully :) \n")
		// TODO: At about here, a structured return type of "success: true/false" and "error" is returned
	}

	buildkite.Log("done. \n")
	return nil
}

func (trp TaskRunnerPlugin) HandleResults(ctx context.Context, output *ecs.DescribeTasksOutput, err error, bkAgent buildkite.AgentAPI, config Config) error {
	if err != nil {
		// This comparison is hacky, but is the only way that I could get the wrapped errors surfaced
		// from the AWS library to be properly handled. It would be better if this was done using errors.As
		if strings.Contains(err.Error(), "exceeded max wait time for TasksStopped waiter") {
			err := bkAgent.Annotate(ctx, fmt.Sprintf("Task did not complete successfully within timeout (%d seconds)", config.TimeOut), "error", "ecs-task-runner")
			if err != nil {
				return fmt.Errorf("failed to annotate buildkite with task timeout failure: %w", err)
			}
			return errors.New("task did not complete within the time limit")
		}
		bkerr := bkAgent.Annotate(ctx, fmt.Sprintf("failed to wait for task completion: %v\n", err), "error", "ecs-task-runner")
		if bkerr != nil {
			return fmt.Errorf("failed to annotate buildkite with task wait failure: %w, annotation error: %w", err, bkerr)
		}
	} else if len(output.Failures) > 0 {
		// There is still a scenario where the task could return failures but this isn't handled by the waiter
		// This is due to the waiter only returning errors in scenarios where there are issues querying the task
		// or scheduling the task. For a list of the Failures that can be returned in this case, see:
		// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/api_failures_messages.html
		// specifically, under the `DescribeTasks` API.
		err := bkAgent.Annotate(ctx, fmt.Sprintf("Task did not complete successfully: %v", output.Failures[0]), "error", "ecs-task-runner")
		if err != nil {
			return fmt.Errorf("failed to annotate buildkite with task failure: %w", err)
		}
		return fmt.Errorf("task did not complete successfully: %v", output.Failures[0])
	}
	return nil
}
