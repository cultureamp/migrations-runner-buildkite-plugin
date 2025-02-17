package plugin

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ParameterName string `required:"true"  split_words:"true"`
	Command       string `required:"false" split_words:"true"`
	TimeOut       int    `default:"2700"   split_words:"true"`
}

type EnvironmentConfigFetcher struct {
}

const pluginEnvironmentPrefix = "BUILDKITE_PLUGIN_ECS_TASK_RUNNER"

func (f EnvironmentConfigFetcher) Fetch(config *Config) error {
	return envconfig.Process(pluginEnvironmentPrefix, config)
}
