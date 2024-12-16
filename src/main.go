package main

import (
	"context"
	"os"

	"github.com/cultureamp/ecs-task-runner-buildkite-plugin/buildkite"
	"github.com/cultureamp/ecs-task-runner-buildkite-plugin/plugin"
)

func main() {
	ctx := context.Background()
	fetcher := plugin.EnvironmentConfigFetcher{}
	taskRunnerPlugin := plugin.TaskRunnerPlugin{}

	err := taskRunnerPlugin.Run(ctx, fetcher)

	if err != nil {
		buildkite.LogFailuref("plugin execution failed: %s\n", err.Error())
		os.Exit(1)
	}
}
