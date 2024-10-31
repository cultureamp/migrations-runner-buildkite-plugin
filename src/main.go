package main

import (
	"context"
	"os"

	awsinternal "ecs-task-runner/aws"
	"ecs-task-runner/buildkite"
	"ecs-task-runner/config"
	"ecs-task-runner/plugin"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {
	// This came from the naive translation. This probably needs a new home, and some refactoring. Mayhaps the "Run" function
	env := config.New()
	rp := config.RunParameters{
		ParameterName: env.ParameterName,
		Script:        env.Script,
	}
	// END of Naive translation here

	// This is from the template
	ctx := context.Background()
	agent := &buildkite.Agent{}
	fetcher := plugin.EnvironmentConfigFetcher{}
	examplePlugin := plugin.ExamplePlugin{}

	err := examplePlugin.Run(ctx, fetcher, agent)

	if err != nil {
		buildkite.LogFailuref("plugin execution failed: %s\n", err.Error())
		os.Exit(1)
	}
	// Template Ends

	// Reamainder of the naive translation
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	ssmClient := ssm.NewFromConfig(cfg)
	log.Printf("Retrieving task configuration from: %s", rp.ParameterName)
	configuration, err := awsinternal.RetrieveConfiguration(ctx, ssmClient, rp.ParameterName)
	if err != nil {
		log.Fatalf("Failed to retrieve configuration: %v", err)
	}

	// append Script to configuration.Command slice
	configuration.Command = append(configuration.Command, rp.Script)

	ecsClient := ecs.NewFromConfig(cfg)
	taskArn, err := awsinternal.SubmitTask(ctx, ecsClient, configuration)
	if err != nil {
		log.Fatalf("Failed to submit task: %v", err)
	}

	//TODO: later down the road, we need to check if a task completed successfully. i.e. it returns a SUCCESS state.
	// In the murmur implementation, this was done by checking the result of the waiter. Our waiter doesn't include
	// this information. So we may need to look at another way

	waiterClient := ecs.NewTasksStoppedWaiter(ecsClient, func(o *ecs.TasksStoppedWaiterOptions) {
		o.MinDelay = time.Second
		o.MaxDelay = 10 * time.Second
	})
	result, err := awsinternal.WaitForCompletion(ctx, waiterClient, taskArn)
	if err != nil {
		// TODO: Do we wanna wanna go from fatal, to print, and provide an opportunity to share logs from the task if there any?
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

}
