package plugin

import (
	"context"
	"fmt"
	"time"

	awsinternal "github.com/cultureamp/ecs-task-runner-buildkite-plugin/aws"
	"github.com/cultureamp/ecs-task-runner-buildkite-plugin/buildkite"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type TaskRunnerPlugin struct {
}

type ConfigFetcher interface {
	Fetch(config *Config) error
}

func (trp TaskRunnerPlugin) Run(ctx context.Context, fetcher ConfigFetcher) error {
	var config Config
	err := fetcher.Fetch(&config)
	if err != nil {
		return fmt.Errorf("plugin configuration error: %w", err)
	}

	buildkite.Log("Executing task-runner plugin\n")

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	ssmClient := ssm.NewFromConfig(cfg)
	buildkite.Logf("Retrieving task configuration from: %s", config.ParameterName)
	configuration, err := awsinternal.RetrieveConfiguration(ctx, ssmClient, config.ParameterName)
	if err != nil {
		return fmt.Errorf("failed to retrieve configuration: %w", err)
	}

	// append Script to configuration.Command. The Script value specifies what script needs to be
	// executed by the task.
	configuration.Command = append(configuration.Command, config.Script)

	ecsClient := ecs.NewFromConfig(cfg)
	taskArn, err := awsinternal.SubmitTask(ctx, ecsClient, configuration)
	if err != nil {
		return fmt.Errorf("failed to submit task: %w", err)
	}

	waiterClient := ecs.NewTasksStoppedWaiter(ecsClient, func(o *ecs.TasksStoppedWaiterOptions) {
		o.MinDelay = time.Second
		// TODO: This is currently a magic number. If we want this to be configurable, remove the nolint directive and fix it up
		o.MaxDelay = 10 * time.Second //nolint:mnd
	})
	result, err := awsinternal.WaitForCompletion(ctx, waiterClient, taskArn)
	if err != nil {
		return fmt.Errorf("failed to wait for task completion: %w\nFailure information: %v", err, result.Failures[0])
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
		buildkite.Logf("CloudWatch Logs for job:")
		for _, l := range logs {
			if l.Timestamp != nil {
				// Applying ISO 8601 format, l.Timestamp is in milliseconds, not very useful in logging
				placeholder := time.UnixMilli(*l.Timestamp).Format(time.RFC3339)
				buildkite.Logf("-> %s %s\n", placeholder, *l.Message)
			}
		}
	}

	// TODO: Assuming the task only has 1 container. What if there others? Like Datadog sideca
	if task.Containers[0].ExitCode != aws.Int32(0) {
		buildkite.LogFailuref("Task stopped with a non-zero exit code:: %d", task.Containers[0].ExitCode)
		// TODO: At about here, a structured return type of "success: true/false" and "error" is returned
	} else {
		buildkite.Log("Task completed successfully :)")
		// TODO: At about here, a structured return type of "success: true/false" and "error" is returned
	}

	buildkite.Log("done.")
	return nil
}
