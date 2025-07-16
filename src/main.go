package main

import (
	"context"
	"os"

	awsinternal "github.com/cultureamp/migrations-runner-buildkite-plugin/aws"
	"github.com/cultureamp/migrations-runner-buildkite-plugin/buildkite"
	"github.com/cultureamp/migrations-runner-buildkite-plugin/plugin"
)

func main() {
	ctx := context.Background()
	fetcher := plugin.EnvironmentConfigFetcher{}
	taskRunnerPlugin := plugin.TaskRunnerPlugin{}

	err := taskRunnerPlugin.Run(ctx, fetcher, awsinternal.WaitForCompletion)
	if err != nil {
		buildkite.LogFailuref("plugin execution failed: %s\n", err.Error())
		os.Exit(1)
	}
}
