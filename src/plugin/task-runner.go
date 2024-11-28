package plugin

import (
	"context"
	"log"
	"time"

	awsinternal "ecs-task-runner/aws"
	"ecs-task-runner/buildkite"

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

type Agent interface {
	Annotate(ctx context.Context, message string, style string, annotationContext string) error
}

func (trp TaskRunnerPlugin) Run(ctx context.Context, fetcher ConfigFetcher, agent Agent) error {
	var config Config
	err := fetcher.Fetch(&config)
	if err != nil {
		buildkite.LogFailuref("plugin configuration error: %s\n", err.Error())
		return err
	}

	buildkite.Log("Executing task-runner plugin\n")

	annotation := config.ParameterName
	err = agent.Annotate(ctx, annotation, "info", "message")
	if err != nil {
		buildkite.LogFailuref("buildkite annotation error: %s\n", err.Error())
		return err
	}
	annotation = config.Script
	err = agent.Annotate(ctx, annotation, "info", "message")
	if err != nil {
		buildkite.LogFailuref("buildkite annotation error: %s\n", err.Error())
		return err
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	ssmClient := ssm.NewFromConfig(cfg)
	log.Printf("Retrieving task configuration from: %s", config.ParameterName)
	configuration, err := awsinternal.RetrieveConfiguration(ctx, ssmClient, config.ParameterName)
	if err != nil {
		log.Fatalf("Failed to retrieve configuration: %v", err)
	}

	// append Script to configuration.Command slice
	configuration.Command = append(configuration.Command, config.Script)

	ecsClient := ecs.NewFromConfig(cfg)
	taskArn, err := awsinternal.SubmitTask(ctx, ecsClient, configuration)
	if err != nil {
		log.Fatalf("Failed to submit task: %v", err)
	}

	waiterClient := ecs.NewTasksStoppedWaiter(ecsClient, func(o *ecs.TasksStoppedWaiterOptions) {
		o.MinDelay = time.Second
		// TODO: This is currently a magic number. If we want this to be configurable, remove the nolint directive and fix it up
		o.MaxDelay = 10 * time.Second //nolint:mnd
	})
	result, err := awsinternal.WaitForCompletion(ctx, waiterClient, taskArn)
	if err != nil {
		// TODO: Do we wanna go from fatal, to print, and provide an opportunity to share logs from the task if there any?
		// That, or we include the log sharing logic within this condition
		log.Printf("error waiting for task completion: %v", err)
		log.Fatalf("failure information: %v", result.Failures[0])
	}
	// In a successful scenario for task completion, we would have a `tasks` slice with a single element
	task := result.Tasks[0]
	taskLogDetails, err := awsinternal.FindLogStreamFromTask(ctx, ecsClient, task)
	if err != nil {
		log.Fatalf("Failed to acquire log stream information for task: %v", err)
	}

	cloudwatchClient := cloudwatchlogs.NewFromConfig(cfg)
	logs, err := awsinternal.RetrieveLogs(ctx, cloudwatchClient, taskLogDetails)
	// Failing to retrieve the logs shouldn't be show-stopper if the task is able to complete successfully.
	// This can come from logs not being available yet, or the service lacking permissions to publish logs at the time
	// TODO: In the original implementation this is how it worked. Is there a possible way to "Wait" for logs?
	if err != nil {
		log.Printf("Failed to retrieve CloudWatch Logs for job, continuing... %v", err)
	}

	if len(logs) > 0 {
		log.Printf("CloudWatch Logs for job:")
		for _, l := range logs {
			if l.Timestamp != nil {
				// Applying ISO 8601 format, l.Timestamp is in milliseconds, not very useful in logging
				placeholder := time.UnixMilli(*l.Timestamp).Format(time.RFC3339)
				log.Printf("-> %s %s", placeholder, *l.Message)
			}
		}
	}

	// TODO: Assuming the task only has 1 container. What if there others? Like Datadog sideca
	if task.Containers[0].ExitCode != aws.Int32(0) {
		log.Fatalf("Task stopped with a non-zero exit code:: %d", task.Containers[0].ExitCode)
		// TODO: At about here, a structured return type of "success: true/false" and "error" is returned
	} else {
		log.Printf("Task completed successfully :)")
		// TODO: At about here, a structured return type of "success: true/false" and "error" is returned
	}

	buildkite.Log("done.")
	return nil
}
