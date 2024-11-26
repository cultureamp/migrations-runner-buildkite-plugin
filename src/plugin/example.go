package plugin

import (
	"context"

	"github.com/cultureamp/examplego/buildkite"
)

type ExamplePlugin struct {
}

type ConfigFetcher interface {
	Fetch(config *Config) error
}

type Agent interface {
	Annotate(ctx context.Context, message string, style string, annotationContext string) error
}

func (ep ExamplePlugin) Run(ctx context.Context, fetcher ConfigFetcher, agent Agent) error {
	var config Config
	err := fetcher.Fetch(&config)
	if err != nil {
		buildkite.LogFailuref("plugin configuration error: %s\n", err.Error())
		return err
	}
	annotation := config.Message

	buildkite.Logf("Annotating with message: %s\n", annotation)

	err = agent.Annotate(ctx, annotation, "info", "message")
	if err != nil {
		buildkite.LogFailuref("buildkite annotation error: %s\n", err.Error())
		return err
	}

	buildkite.Log("done.")
	return nil
}
